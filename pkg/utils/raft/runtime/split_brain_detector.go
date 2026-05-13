/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package raft

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// -----------------------------------------------------------------------------
// Constants & tuning knobs
// -----------------------------------------------------------------------------

const (
	// Timing intervals.
	managerTickInterval  = 2 * time.Second
	leaderTickInterval   = 3 * time.Second
	followerTickInterval = 2 * time.Second
	peerTickInterval     = 1 * time.Second

	// Thresholds.
	splitBrainRepCountThreshold = 5  // consecutive quorum-loss ticks before declaring split-brain
	peerInactiveThreshold       = 10 // consecutive no-progress ticks before marking a peer INACTIVE
	peerDeletingThreshold       = 5  // counter reset when a peer pod is being deleted
	leaderHeartbeatTryThreshold = 2  // leader heartbeat retry count before proposing
)

// -----------------------------------------------------------------------------
// Node / connection state types
// -----------------------------------------------------------------------------

// NodeState represents the perceived health of a cluster node.
type NodeState string

const (
	StateActive     NodeState = "ACTIVE"
	StateInactive   NodeState = "INACTIVE"
	StateSplitBrain NodeState = "SPLIT_BRAIN"
)

// ConnState represents the result of a connectivity probe to a peer.
type ConnState string

const (
	ConnStateFailed    ConnState = "failed"
	ConnStateConnected ConnState = "connected"
	ConnStateUnknown   ConnState = "unknown"
)

// -----------------------------------------------------------------------------
// Callback function types
// -----------------------------------------------------------------------------

// SkipperFunc returns true when split-brain detection should be paused
// (e.g. during an ongoing ops-request). A nil SkipperFunc never skips.
type SkipperFunc func() bool

// ConnStateChecker probes connectivity to a named pod.
// It must return ConnStateConnected, ConnStateFailed, or ConnStateUnknown.
type ConnStateChecker func(podName string) ConnState

// ReadyChecker returns true once a pod has bootstrapped and joined the cluster.
// The detector waits for all pods to be ready before starting, to avoid
// false-positives at startup. A nil ReadyChecker treats all pods as ready.
type ReadyChecker func(podName string) bool

// -----------------------------------------------------------------------------
// Configuration
// -----------------------------------------------------------------------------

// ClusterConfig describes the static topology of the database cluster.
// All fields must be set before passing to NewDetector; none may be mutated
// after construction.
type ClusterConfig struct {
	// Namespace is the Kubernetes namespace that contains all cluster pods.
	Namespace string

	// PodNames lists every pod in the cluster (including the local pod).
	PodNames []string

	// LocalPod is the name of the pod running this detector instance.
	LocalPod string

	// RaftIDs maps each pod name to its Raft node ID.
	// Used to track replication progress per peer.
	RaftIDs map[string]uint64

	// PodSelector is the label selector used to list cluster pods.
	PodSelector map[string]string

	StartUpTimeOut *time.Duration
}

func (c *ClusterConfig) validate() {
	if c.RaftIDs == nil || c.PodSelector == nil {
		panic("ClusterConfig: RaftIDs and PodSelector must not be nil")
	}
}

// quorum returns the minimum number of active nodes required to maintain
// a healthy cluster, adjusted so that even-sized clusters use odd quorum math.
func (c *ClusterConfig) quorum() int {
	n := len(c.PodNames)
	if n%2 == 0 {
		n++
	}
	return (n + 1) / 2
}

// peers returns all pod names except the local pod.
func (c *ClusterConfig) peers() []string {
	out := make([]string, 0, len(c.PodNames)-1)
	for _, name := range c.PodNames {
		if name != c.LocalPod {
			out = append(out, name)
		}
	}
	return out
}

// -----------------------------------------------------------------------------
// Node-state registry
// -----------------------------------------------------------------------------

// nodeStateRegistry is a thread-safe store of per-pod NodeState values.
type nodeStateRegistry struct {
	mu    sync.RWMutex
	state map[string]NodeState
}

func newNodeStateRegistry() *nodeStateRegistry {
	return &nodeStateRegistry{state: make(map[string]NodeState)}
}

func (r *nodeStateRegistry) set(pod string, s NodeState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state[pod] = s
}

func (r *nodeStateRegistry) get(pod string) NodeState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state[pod]
}

func (r *nodeStateRegistry) remove(pod string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.state, pod)
}

// initAll marks every pod in the cluster as StateActive.
func (r *nodeStateRegistry) initAll(pods []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range pods {
		r.state[p] = StateActive
	}
}

// -----------------------------------------------------------------------------
// Detector
// -----------------------------------------------------------------------------

// Detector orchestrates split-brain detection for a single Raft cluster member.
// It runs three goroutine loops (leader heartbeat, follower monitoring, and
// split-brain declaration) that are started and stopped in lock-step with
// Raft leadership changes.
type Detector struct {
	cfg      *ClusterConfig
	rc       *RaftNode
	kc       client.Client
	kv       *Kvstore
	registry *nodeStateRegistry

	connCheck  ConnStateChecker
	skipCheck  SkipperFunc  // may be nil
	readyCheck ReadyChecker // may be nil

	// stopCh is closed by the caller to request a full shutdown.
	stopCh <-chan struct{}
}

// DetectorOptions carries optional callbacks for the Detector.
type DetectorOptions struct {
	SkipCheck  SkipperFunc
	ReadyCheck ReadyChecker
}

// NewDetector constructs a Detector. stopCh should be closed by the caller
// when the entire component is shutting down.
func NewDetector(
	cfg *ClusterConfig,
	rc *RaftNode,
	kc client.Client,
	kv *Kvstore,
	connCheck ConnStateChecker,
	stopCh <-chan struct{},
	opts DetectorOptions,
) *Detector {
	cfg.validate()
	return &Detector{
		cfg:        cfg,
		rc:         rc,
		kc:         kc,
		kv:         kv,
		registry:   newNodeStateRegistry(),
		connCheck:  connCheck,
		skipCheck:  opts.SkipCheck,
		readyCheck: opts.ReadyCheck,
		stopCh:     stopCh,
	}
}

// -----------------------------------------------------------------------------
// Public entry point
// -----------------------------------------------------------------------------

// Run blocks and orchestrates all detection goroutines until stopCh is closed.
// It should be called in its own goroutine.
func (d *Detector) Run() {
	klog.Infoln("[SplitBrainDetector] Waiting for all replicas to become ready...")
	if !d.waitForAllReady() {
		return // stopCh fired during wait
	}
	klog.Infoln("[SplitBrainDetector] All replicas ready, entering management loop")
	d.managerLoop()
}

// waitForAllReady blocks until every pod reports ready (via ReadyChecker) or
// startupTimeout elapses. Returns false if stopCh fires first.
func (d *Detector) waitForAllReady() bool {
	startupTimeout := 40 * time.Second
	if d.cfg.StartUpTimeOut != nil {
		startupTimeout = *d.cfg.StartUpTimeOut
	}
	timeout := time.After(startupTimeout)
	ticker := time.NewTicker(peerTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return false
		case <-timeout:
			klog.Infoln("[SplitBrainDetector] Startup timeout reached, proceeding anyway")
			return true
		case <-ticker.C:
			if d.allPodsReady() {
				klog.Infof("[SplitBrainDetector] All %d replicas ready", len(d.cfg.PodNames))
				return true
			}
		}
	}
}

func (d *Detector) allPodsReady() bool {
	if d.readyCheck == nil {
		return true
	}
	for _, pod := range d.cfg.PodNames {
		if !d.readyCheck(pod) {
			return false
		}
	}
	return true
}

// -----------------------------------------------------------------------------
// Manager loop
// -----------------------------------------------------------------------------

// managerLoop watches for leadership changes and starts/stops the three
// detection sub-loops accordingly.
func (d *Detector) managerLoop() {
	ticker := time.NewTicker(managerTickInterval)
	defer ticker.Stop()

	var (
		cancel      func() // nil when not leading
		leading     bool
		initialized bool
	)

	klog.Infoln("[SplitBrainDetector] Manager loop started")
	defer klog.Infoln("[SplitBrainDetector] Manager loop stopped")

	for {
		select {
		case <-d.stopCh:
			if cancel != nil {
				cancel()
			}
			return

		case <-ticker.C:
			isLeader := d.isLeader()

			switch {
			case isLeader && !leading:
				// Became leader: start sub-loops.
				leading = true
				initialized = false
				d.registry.initAll(d.cfg.PodNames)
				var ctx context.Context
				ctx, cancel = newCancelContext(d.stopCh)
				klog.Infoln("[SplitBrainDetector] Became leader, starting detection sub-loops")
				go d.leaderHeartbeatLoop(ctx.Done())
				go d.splitBrainLoop(ctx.Done())
				go d.followerMonitorLoop(ctx.Done())

			case !isLeader && leading:
				// Lost leadership: stop sub-loops.
				leading = false
				cancel()
				cancel = nil
				klog.Infoln("[SplitBrainDetector] Lost leadership, stopping detection sub-loops")

			case !isLeader && !initialized:
				// Follower initialisation (once per non-leader tenure).
				initialized = true
				d.registry.initAll(d.cfg.PodNames)
				klog.Infoln("[SplitBrainDetector] Follower: initialized node states")
			}
		}
	}
}

// newCancelContext returns a context that is cancelled when either doneCh or
// the returned cancel function fires.
func newCancelContext(doneCh <-chan struct{}) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-doneCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// -----------------------------------------------------------------------------
// Leader heartbeat loop
// -----------------------------------------------------------------------------

// leaderHeartbeatLoop proposes a random Raft entry when quorum is lost, to
// keep the Raft log advancing and prevent the leader from stepping down
// prematurely.
func (d *Detector) leaderHeartbeatLoop(stopCh <-chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Leader heartbeat loop started")
	defer klog.Infoln("[SplitBrainDetector] Leader heartbeat loop stopped")

	ticker := time.NewTicker(leaderTickInterval)
	defer ticker.Stop()

	tryCount := 0
	failedOnce := false

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if d.shouldSkip("Leader heartbeat") {
				continue
			}

			active := d.countConnectedPeers() + 1 // +1 for self
			if active >= d.cfg.quorum() {
				tryCount, failedOnce = 0, false
				klog.V(6).Infof("[SplitBrainDetector] Heartbeat: quorum OK (active=%d)", active)
				continue
			}

			tryCount++
			klog.V(5).Infof("[SplitBrainDetector] Heartbeat: quorum lost (active=%d, try=%d)", active, tryCount)

			if !failedOnce || tryCount > leaderHeartbeatTryThreshold {
				tryCount = 0
				failedOnce = true
				if d.isLeader() {
					klog.V(3).Infoln("[SplitBrainDetector] Proposing random value to keep Raft alive")
					d.kv.Propose(d.cfg.LocalPod, strconv.Itoa(rand.Int()))
				}
			}
		}
	}
}

// countConnectedPeers returns the number of peers reachable via connCheck.
func (d *Detector) countConnectedPeers() int {
	count := 0
	for _, peer := range d.cfg.peers() {
		if d.connCheck(peer) == ConnStateConnected {
			count++
		}
	}
	return count
}

// -----------------------------------------------------------------------------
// Split-brain detection loop
// -----------------------------------------------------------------------------

// splitBrainLoop runs on the leader and declares a split-brain when active
// node count falls below quorum for several consecutive ticks.
func (d *Detector) splitBrainLoop(stopCh <-chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Split-brain detection loop started")
	defer klog.Infoln("[SplitBrainDetector] Split-brain detection loop stopped")

	ticker := time.NewTicker(managerTickInterval)
	defer ticker.Stop()

	repCount := 0

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if !d.isLeader() || d.shouldSkip("Split-brain check") {
				continue
			}

			active := d.countActiveNodes()
			if active >= d.cfg.quorum() {
				repCount = 0
				d.registry.set(d.cfg.LocalPod, StateActive)
				klog.V(6).Infof("[SplitBrainDetector] Split-brain check OK (active=%d)", active)
				continue
			}

			repCount++
			klog.V(6).Infof("[SplitBrainDetector] Quorum loss count=%d (active=%d)", repCount, active)

			if repCount > splitBrainRepCountThreshold {
				d.handlePotentialSplitBrain()
			}
		}
	}
}

// countActiveNodes counts how many pods (including self) are in StateActive.
func (d *Detector) countActiveNodes() int {
	count := 1 // self is always active on the leader
	for _, peer := range d.cfg.peers() {
		if d.registry.get(peer) == StateActive {
			count++
		}
	}
	return count
}

// handlePotentialSplitBrain verifies a split-brain by checking how many pods
// carry the primary label; if more than one, the local node is marked as split-brain.
func (d *Detector) handlePotentialSplitBrain() {
	primaryCount, err := d.countPrimaryPods()
	if err != nil {
		klog.Warningf("[SplitBrainDetector] Could not list pods: %v", err)
		return
	}
	if primaryCount <= 1 {
		return
	}
	klog.Warningf("[SplitBrainDetector] Split brain! %d primaries detected", primaryCount)
	d.registry.set(d.cfg.LocalPod, StateSplitBrain)
}

// countPrimaryPods lists pods matching the cluster selector and returns how
// many carry the primary role label.
func (d *Detector) countPrimaryPods() (int, error) {
	pods := &core.PodList{}
	sel := labels.SelectorFromSet(d.cfg.PodSelector)
	if err := d.kc.List(context.TODO(), pods, &client.ListOptions{
		LabelSelector: sel,
		Namespace:     d.cfg.Namespace,
	}); err != nil {
		return 0, err
	}

	count := 0
	for _, pod := range pods.Items {
		if pod.Labels[kubedb.LabelRole] == kubedb.PostgresPodPrimary {
			count++
		}
	}
	return count, nil
}

// -----------------------------------------------------------------------------
// Follower / peer monitoring loop
// -----------------------------------------------------------------------------

// followerMonitorLoop manages per-peer goroutines that track Raft match
// progress, marking lagging peers as INACTIVE.
func (d *Detector) followerMonitorLoop(stopCh <-chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Follower monitor loop started")
	defer klog.Infoln("[SplitBrainDetector] Follower monitor loop stopped")

	ticker := time.NewTicker(followerTickInterval)
	defer ticker.Stop()

	// peerCancel maps peer pod names to their per-goroutine cancel functions.
	peerCancel := make(map[string]func())

	cleanup := func() {
		for pod, cancel := range peerCancel {
			cancel()
			delete(peerCancel, pod)
		}
	}
	defer cleanup()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			desired := d.peerSet()

			// Stop goroutines for pods no longer in the cluster.
			for pod, cancel := range peerCancel {
				if _, ok := desired[pod]; !ok {
					klog.Infof("[SplitBrainDetector] Removing peer monitor: %s", pod)
					cancel()
					delete(peerCancel, pod)
				}
			}

			// Start goroutines for newly appearing pods.
			for pod := range desired {
				if _, ok := peerCancel[pod]; !ok {
					klog.Infof("[SplitBrainDetector] Starting peer monitor: %s", pod)
					ctx, cancel := newCancelContext(stopCh)
					peerCancel[pod] = cancel
					go d.monitorPeer(pod, ctx.Done())
				}
			}
		}
	}
}

// peerSet returns the current set of peer pod names (excluding local).
func (d *Detector) peerSet() map[string]struct{} {
	m := make(map[string]struct{}, len(d.cfg.PodNames)-1)
	for _, pod := range d.cfg.peers() {
		m[pod] = struct{}{}
	}
	return m
}

// monitorPeer tracks Raft match progress for a single peer and flips its
// state between StateActive and StateInactive.
func (d *Detector) monitorPeer(pod string, stopCh <-chan struct{}) {
	klog.Infof("[SplitBrainDetector] Peer monitor started: %s", pod)
	defer klog.Infof("[SplitBrainDetector] Peer monitor stopped: %s", pod)

	ticker := time.NewTicker(peerTickInterval)
	defer ticker.Stop()

	prevMatch := uint64(0)
	noProgressCount := 0

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if d.shouldSkip("Peer monitor " + pod) {
				continue
			}

			if d.peerIsProgressing(pod, &prevMatch) {
				noProgressCount = 0
				d.registry.set(pod, StateActive)
				continue
			}

			noProgressCount++
			klog.V(6).Infof("[SplitBrainDetector] Peer %s not progressing (count=%d)", pod, noProgressCount)

			if noProgressCount <= peerInactiveThreshold {
				continue
			}

			// Before marking inactive, check if the pod is being deleted.
			if d.podIsDeletingOrNotRunning(pod) {
				noProgressCount = peerDeletingThreshold
				klog.V(6).Infof("[SplitBrainDetector] Peer %s deleting/not-running, deferring inactive", pod)
				continue
			}

			klog.V(4).Infof("[SplitBrainDetector] Marking peer %s INACTIVE", pod)
			d.registry.set(pod, StateInactive)
		}
	}
}

// peerIsProgressing checks whether a peer's Raft match index has advanced
// since the last check, or has caught up with the leader.
// prevMatch is updated in place.
func (d *Detector) peerIsProgressing(pod string, prevMatch *uint64) bool {
	leaderID := d.rc.Node.Status().Lead
	if leaderID == 0 {
		return true // No leader yet; don't penalise.
	}

	peerID := d.cfg.RaftIDs[pod]
	if peerID == 0 {
		klog.Warningf("[SplitBrainDetector] No Raft ID for peer %s", pod)
		return true // Unknown peer; be conservative.
	}

	progress := d.rc.Node.Status().Progress
	leaderProg, leaderOK := progress[leaderID]
	peerProg, peerOK := progress[peerID]
	if !leaderOK || !peerOK {
		return true // Progress not yet available.
	}

	if leaderProg.Match == peerProg.Match || peerProg.Match > *prevMatch {
		*prevMatch = peerProg.Match
		return true
	}
	return false
}

// podIsDeletingOrNotRunning returns true when the pod is being terminated or
// is not in the Running phase (so we don't prematurely declare it inactive).
func (d *Detector) podIsDeletingOrNotRunning(pod string) bool {
	p := &core.Pod{}
	err := d.kc.Get(context.TODO(), types.NamespacedName{
		Name:      pod,
		Namespace: d.cfg.Namespace,
	}, p)
	if err != nil {
		return false // Assume running if we can't fetch.
	}
	return p.DeletionTimestamp != nil || p.Status.Phase != core.PodRunning
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// isLeader returns true when this node is the current Raft leader.
func (d *Detector) isLeader() bool {
	status := d.rc.Node.Status()
	return status.Lead == status.ID
}

// shouldSkip returns true when the SkipperFunc signals that detection should
// pause. Logs at V(6) if skipping.
func (d *Detector) shouldSkip(context string) bool {
	if d.skipCheck != nil && d.skipCheck() {
		klog.V(6).Infof("[SplitBrainDetector] %s skipped (ops request in progress)", context)
		return true
	}
	return false
}
