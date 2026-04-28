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
	"fmt"
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

type NodeState string

const (
	STATE_ACTIVE      NodeState = "ACTIVE"
	STATE_INACTIVE    NodeState = "INACTIVE"
	STATE_SPLIT_BRAIN NodeState = "SPLIT_BRAIN"
)

type ConnState string

const (
	StateFailed    ConnState = "failed"
	StateConnected ConnState = "connected"
	StateUnknown   ConnState = "unknown"
)

type (
	SkipperFunc              func() bool
	ConnStateFunc            func(node string) ConnState
	ReadyFunc                func(nodeName string) bool
	SplitBrainDetectorConfig struct {
		nodeState     map[string]NodeState
		mutex         sync.RWMutex
		stopCh        chan struct{}
		config        *Config
		rc            *RaftNode
		kc            client.Client
		kv            *Kvstore
		skipperFunc   SkipperFunc
		connStateFunc ConnStateFunc
		readyFunc     ReadyFunc
	}
)

type Config struct {
	Namespace  string
	PetsetName string
	PodLists   []string
	Replicas   int
	PodName    string
	ID         map[string]uint64
	sel        map[string]string
}

func NewSplitBrainDetectorConfig(cfg *Config, rc *RaftNode, kc client.Client, kv *Kvstore, csf ConnStateFunc, stopCh chan struct{}) (*SplitBrainDetectorConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.ID == nil {
		return nil, fmt.Errorf("config ID map cannot be nil")
	}
	if cfg.sel == nil {
		return nil, fmt.Errorf("config selector map cannot be nil")
	}
	return &SplitBrainDetectorConfig{
		nodeState:     make(map[string]NodeState),
		stopCh:        stopCh,
		config:        cfg,
		rc:            rc,
		kc:            kc,
		kv:            kv,
		connStateFunc: csf,
	}, nil
}

func (sbd *SplitBrainDetectorConfig) SetSkipperFunc(skipper SkipperFunc) {
	sbd.skipperFunc = skipper
}

func (sbd *SplitBrainDetectorConfig) SetReadyFunc(readyFunc ReadyFunc) {
	sbd.readyFunc = readyFunc
}

func (sbd *SplitBrainDetectorConfig) SetNodeState(node string, fs NodeState) {
	sbd.mutex.Lock()
	defer sbd.mutex.Unlock()
	sbd.nodeState[node] = fs
}

func (sbd *SplitBrainDetectorConfig) GetNodeState(node string) NodeState {
	sbd.mutex.RLock()
	defer sbd.mutex.RUnlock()
	return sbd.nodeState[node]
}

func (sbd *SplitBrainDetectorConfig) RemoveNodeState(node string) {
	sbd.mutex.Lock()
	defer sbd.mutex.Unlock()
	delete(sbd.nodeState, node)
}

func (sbd *SplitBrainDetectorConfig) InitNodeState() {
	sbd.mutex.Lock()
	defer sbd.mutex.Unlock()
	for i := 0; i < sbd.config.Replicas; i++ {
		podName := fmt.Sprintf("%s-%d", sbd.config.PetsetName, i)
		sbd.nodeState[podName] = STATE_ACTIVE
	}
}

func (sbd *SplitBrainDetectorConfig) DetectSplitBrain(stopCh chan struct{}) {
	go sbd.DetectSplitBrainWithRaft(stopCh)
}

func (sbd *SplitBrainDetectorConfig) DetectSplitBrainWithRaft(stopCh chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Starting DetectSplitBrainWithRaft goroutine - detects split brain by checking raft quorum and primary labels")
	ticker := time.NewTicker(time.Second * 2)
	defer func() {
		ticker.Stop()
		klog.Infoln("[SplitBrainDetector] Stopping Split Brain Detector using Raft")
	}()
	repCount := 0
	for {
		select {
		case <-stopCh:
			klog.Infoln("[SplitBrainDetector] DetectSplitBrainWithRaft stopped (stop signal)")
			return
		case <-sbd.stopCh:
			klog.Infoln("[SplitBrainDetector] DetectSplitBrainWithRaft stopped (shutdown signal)")
			return
		case <-ticker.C:
			if sbd.skipperFunc != nil && sbd.skipperFunc() {
				klog.V(6).Infoln("[SplitBrainDetector] Leader heartbeat check skipped, skipper function returned true")
				continue
			}

			t := time.Now()
			if sbd.rc.Node.Status().Lead != sbd.rc.Node.Status().ID {
				continue
			}
			active := 1
			r := len(sbd.config.PodLists)
			if r%2 == 0 {
				r++
			}
			quorum := (r + 1) / 2
			for i := 0; i < len(sbd.config.PodLists); i++ {
				podName := sbd.config.PodLists[i]
				if podName == sbd.config.PodName {
					continue
				}
				nodeState := sbd.GetNodeState(podName)
				if nodeState == STATE_ACTIVE {
					active++
				}
			}
			if active < quorum {
				repCount++
			} else {
				repCount = 0
				sbd.SetNodeState(sbd.config.PodName, STATE_ACTIVE)
			}

			if repCount > 5 {
				pods := core.PodList{}
				sel := labels.SelectorFromSet(sbd.config.sel)
				err := sbd.kc.List(context.TODO(), &pods, &client.ListOptions{
					LabelSelector: sel,
					Namespace:     sbd.config.Namespace,
				})
				if err == nil {
					// TODO: check split brain commenting out this code
					primaryCounter := 0
					for _, pod := range pods.Items {
						if pod.Labels[kubedb.LabelRole] == kubedb.PostgresPodPrimary {
							primaryCounter++
						}
					}

					if primaryCounter <= 1 {
						continue
					}
					klog.Warningf("[SplitBrainDetector] Split brain detected! %d primaries found, took: %v", primaryCounter, time.Since(t))
				}
				sbd.SetNodeState(sbd.config.PodName, STATE_SPLIT_BRAIN)
			}
			klog.V(6).Infof("[SplitBrainDetector] Split brain check completed, active=%d, quorum=%d, repCount=%d, took: %v", active, quorum, repCount, time.Since(t))
		}
	}
}

func (sbd *SplitBrainDetectorConfig) StartLeaderNode(shutCh chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Starting StartLeaderNode goroutine - leader sends heartbeat proposals when quorum is lost")
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		ticker.Stop()
		klog.Infoln("[SplitBrainDetector] Stopped sending custom entries from leader node.")
	}()
	tryCount := 0
	failedOnce := false
	for {
		select {
		case <-sbd.stopCh:
			klog.Infoln("[SplitBrainDetector] StartLeaderNode stopped (shutdown signal)")
			return
		case <-shutCh:
			klog.Infoln("[SplitBrainDetector] StartLeaderNode stopped (stop signal)")
			return
		case <-ticker.C:
			if sbd.skipperFunc != nil && sbd.skipperFunc() {
				klog.V(6).Infoln("[SplitBrainDetector] Leader heartbeat check skipped, an ops request is in progress")
				continue
			}
			t := time.Now()
			r := len(sbd.config.PodLists)
			if r%2 == 0 {
				r++
			}
			quorum := (r + 1) / 2
			active := 0
			for i := 0; i < len(sbd.config.PodLists); i++ {
				podName := sbd.config.PodLists[i]
				s := sbd.connStateFunc(podName)
				if s == StateConnected {
					active++
				}
			}
			if active >= quorum {
				tryCount = 0
				failedOnce = false
				klog.V(6).Infof("[SplitBrainDetector] Quorum satisfied: active=%d, quorum=%d, took: %v", active, quorum, time.Since(t))
				continue
			}
			tryCount++
			klog.V(5).Infof("[SplitBrainDetector] Quorum lost: active=%d, quorum=%d, tryCount=%d", active, quorum, tryCount)
			if !failedOnce || tryCount > 2 {
				tryCount = 0
				failedOnce = true

				if sbd.rc.Node.Status().Lead == sbd.rc.Node.Status().ID {
					proposeStart := time.Now()
					klog.V(3).Infoln("[SplitBrainDetector] trying to propose random value to keep raft alive")
					sbd.kv.Propose(sbd.config.PodName, strconv.Itoa(rand.Int()))
					klog.V(3).Infoln("[SplitBrainDetector] proposed random value to keep raft alive, took:", time.Since(proposeStart))
				}
			}
			klog.V(6).Infof("[SplitBrainDetector] Leader heartbeat check completed, took: %v", time.Since(t))
		}
	}
}

func (sbd *SplitBrainDetectorConfig) StartFollowerNodes(shutCh chan struct{}) {
	klog.Infoln("[SplitBrainDetector] Starting StartFollowerNodes goroutine - monitors raft progress for each peer to detect inactive nodes")
	ticker := time.NewTicker(2 * time.Second)
	defer func() {
		ticker.Stop()
		klog.Infoln("[SplitBrainDetector] Stopped monitoring split brain from follower nodes.")
	}()
	peerMap := make(map[string]chan struct{})

	for {
		select {
		case <-sbd.stopCh:
			klog.Infoln("[SplitBrainDetector] StartFollowerNodes stopped (shutdown signal), closing all peer channels")
			for pr, ch := range peerMap {
				close(ch)
				delete(peerMap, pr)
			}
			return
		case <-shutCh:
			klog.Infoln("[SplitBrainDetector] StartFollowerNodes stopped (stop signal), closing all peer channels")
			for pr, ch := range peerMap {
				close(ch)
				delete(peerMap, pr)
			}
			return
		case <-ticker.C:
			m := make(map[string]struct{})
			for i := 0; i < len(sbd.config.PodLists); i++ {
				peerPodName := sbd.config.PodLists[i]
				if peerPodName == sbd.config.PodName {
					continue
				}
				m[peerPodName] = struct{}{}
			}
			for pr, ch := range peerMap {
				if _, exists := m[pr]; !exists {
					klog.Infof("[SplitBrainDetector] Removing peer monitor for: %s", pr)
					close(ch)
					delete(peerMap, pr)
				}
			}
			for pr := range m {
				if _, exists := peerMap[pr]; !exists {
					klog.Infof("[SplitBrainDetector] Starting peer monitor goroutine for: %s", pr)
					peerMap[pr] = make(chan struct{})
					go func(peerPod string, stopCh chan struct{}, shutCh chan struct{}) {
						klog.Infof("[SplitBrainDetector] Peer monitor goroutine started for %s - monitors raft match progress", peerPod)
						btc := time.NewTicker(1 * time.Second)
						defer btc.Stop()

						prevMatch := uint64(0)
						nmc := 0
						for {
							select {
							case <-stopCh:
								klog.Infof("[SplitBrainDetector] Peer monitor for %s stopped (peer removed)", peerPod)
								return
							case <-shutCh:
								klog.Infof("[SplitBrainDetector] Peer monitor for %s stopped (shutdown signal)", peerPod)
								return
							case <-btc.C:

								if sbd.skipperFunc != nil && sbd.skipperFunc() {
									klog.V(6).Infoln("[SplitBrainDetector] Leader heartbeat check skipped, an ops request is in progress")
									continue
								}

								t := time.Now()
								lid := sbd.rc.Node.Status().Lead
								if lid == 0 {
									continue
								}
								id := sbd.config.ID[peerPod]
								if id == 0 {
									klog.Warningf("[SplitBrainDetector] No ID found for peer %s, skipping match check", peerPod)
									continue
								}
								pm := sbd.rc.Node.Status().Progress
								leaderMatch, e1 := pm[lid]
								peerMatch, e2 := pm[id]
								if e1 && e2 {
									if leaderMatch.Match == peerMatch.Match || peerMatch.Match > prevMatch {
										prevMatch = peerMatch.Match
										sbd.SetNodeState(peerPod, STATE_ACTIVE)
										nmc = 0
										klog.V(6).Infof("[SplitBrainDetector] Peer %s is ACTIVE, match=%d, took: %v", peerPod, peerMatch.Match, time.Since(t))
										continue
									}
									nmc++
									klog.V(6).Infof("[SplitBrainDetector] Peer %s match not progressing, nmc=%d, leaderMatch=%d, peerMatch=%d", peerPod, nmc, leaderMatch.Match, peerMatch.Match)
								}
								if nmc <= 10 {
									continue
								}
								pd := &core.Pod{}
								err := sbd.kc.Get(context.TODO(), types.NamespacedName{
									Name:      peerPod,
									Namespace: sbd.config.Namespace,
								}, pd)
								if err == nil && pd != nil && (pd.DeletionTimestamp != nil || pd.Status.Phase != core.PodRunning) {
									nmc = 5
									klog.V(6).Infof("[SplitBrainDetector] Peer %s is being deleted or not running, resetting counter", peerPod)
									continue
								}
								klog.V(6).Infof("[SplitBrainDetector] Setting peer %s to INACTIVE, nmc=%d, took: %v", peerPod, nmc, time.Since(t))
								sbd.SetNodeState(peerPod, STATE_INACTIVE)
							}
						}
					}(pr, peerMap[pr], sbd.stopCh)
				}
			}
		}
	}
}

func (sbd *SplitBrainDetectorConfig) RunSplitBrainManager() {
	klog.Infoln("[SplitBrainDetector] Starting RunSplitBrainManager - orchestrates all split brain detection goroutines")
	u := time.After(time.Second * 40)
	t := time.NewTicker(1 * time.Second)

	klog.Infoln("[SplitBrainDetector] Waiting for LSN data from all replicas or 40 seconds timeout...")
	for {
		stepDown := false
		select {
		case <-sbd.stopCh:
			klog.Infoln("[SplitBrainDetector] RunSplitBrainManager stopped during initialization")
			return
		case <-u:
			klog.Infoln("[SplitBrainDetector] 40 second timeout reached, starting split brain manager")
			stepDown = true
		case <-t.C:
			got := 0
			for i := 0; i < len(sbd.config.PodLists); i++ {
				podName := sbd.config.PodLists[i]
				if sbd.readyFunc == nil || sbd.readyFunc(podName) {
					got++
				}
			}
			if got >= sbd.config.Replicas {
				klog.Infof("[SplitBrainDetector] All %d replicas have LSN data, starting split brain manager", got)
				stepDown = true
			}
		}
		if stepDown {
			break
		}
	}
	t.Stop()

	ticker := time.NewTicker(2 * time.Second)
	defer func() {
		ticker.Stop()
		klog.Infoln("[SplitBrainDetector] Stopping Split Brain Manager")
	}()
	chanMap := make(map[string]chan struct{})
	s := []string{"start"}
	initialize := false

	for {
		select {
		case <-sbd.stopCh:
			klog.Infoln("[SplitBrainDetector] RunSplitBrainManager stopped")
			return
		case <-ticker.C:
			checkStart := time.Now()
			if sbd.rc.Node.Status().Lead != sbd.rc.Node.Status().ID {
				if !initialize {
					initialize = true
					sbd.InitNodeState()
					klog.Infoln("[SplitBrainDetector] Not a leader, initialized node states")
				}
				for k, ch := range chanMap {
					klog.V(5).Infof("[SplitBrainDetector] Stopping split brain monitors (not a leader)")
					close(ch)
					delete(chanMap, k)
				}
				klog.V(6).Infof("[SplitBrainDetector] Manager check completed (not leader), took: %v", time.Since(checkStart))
				continue
			}
			initialize = false
			for _, v := range s {
				if _, ok := chanMap[v]; !ok {
					klog.Infoln("[SplitBrainDetector] I am the leader, starting split brain monitor goroutines")
					sbd.InitNodeState()
					chanMap[v] = make(chan struct{})
					go sbd.StartLeaderNode(chanMap[v])
					go sbd.DetectSplitBrain(chanMap[v])
					go sbd.StartFollowerNodes(chanMap[v])
					klog.Infoln("[SplitBrainDetector] All split brain monitor goroutines started, took:", time.Since(checkStart))
				}
			}
		}
	}
}
