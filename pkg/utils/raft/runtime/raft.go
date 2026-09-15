// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal"
	"go.etcd.io/etcd/server/v3/wal/walpb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/klog/v2"
)

// RaftNode A key-value stream backed by raft
type RaftNode struct {
	transferLeadershipC <-chan TransferLeadershipConfig
	proposeC            <-chan string            // proposed messages (k,v)
	confChangeC         <-chan raftpb.ConfChange // proposed cluster config changes
	commitC             chan<- *Commit           // entries committed to log (k,v)
	errorC              chan<- error             // errors from raft session

	id          int      // client ID for raft session
	peers       []string // raft peer URLs
	join        bool     // Node is joining an existing cluster
	waldir      string   // path to WAL directory
	snapdir     string   // path to snapshot directory
	getSnapshot func() ([]byte, error)

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	Node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL

	snapshotter      *snap.Snapshotter
	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready

	snapCount uint64
	Transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shut down
	httpdonec chan struct{} // signals http server shutdown complete
	logger    *zap.Logger

	// keyList contains keys that will be stored in raft
	keyList []string
	// transferLeadershipTimeout is the timeout for transfer leadership
	transferLeadershipTimeout time.Duration
}

var defaultSnapshotCount uint64 = 10000

// RaftNodeConfig holds configuration for creating a new RaftNode
type RaftNodeConfig struct {
	ID                        int
	Period                    time.Duration
	ElectionTick              uint64
	HeartbeatTick             uint64
	Peers                     []string
	Join                      bool
	WalDir                    string
	SnapDir                   string
	GetSnapshot               func() ([]byte, error)
	KeyList                   []string
	TransferLeadershipTimeout time.Duration
}

// NewRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shut down, close proposeC and read errorC.
func NewRaftNode(cfg RaftNodeConfig, proposeC <-chan string,
	confChangeC <-chan raftpb.ConfChange, transferLeaderC <-chan TransferLeadershipConfig,
) (<-chan *Commit, <-chan error, <-chan *snap.Snapshotter, *RaftNode) {
	commitC := make(chan *Commit)
	errorC := make(chan error)

	encoderCfg := zap.NewProductionConfig()
	encoderCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	encoderCfg.Encoding = "console"
	encoderCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := encoderCfg.Build()
	if err != nil {
		panic(fmt.Errorf("failed to create logger, %v", err))
	}

	rc := &RaftNode{
		transferLeadershipC:       transferLeaderC,
		proposeC:                  proposeC,
		confChangeC:               confChangeC,
		commitC:                   commitC,
		errorC:                    errorC,
		id:                        cfg.ID,
		peers:                     cfg.Peers,
		join:                      cfg.Join,
		waldir:                    cfg.WalDir,
		snapdir:                   cfg.SnapDir,
		getSnapshot:               cfg.GetSnapshot,
		snapCount:                 defaultSnapshotCount,
		stopc:                     make(chan struct{}),
		httpstopc:                 make(chan struct{}),
		httpdonec:                 make(chan struct{}),
		logger:                    logger,
		snapshotterReady:          make(chan *snap.Snapshotter, 1),
		keyList:                   cfg.KeyList,
		transferLeadershipTimeout: cfg.TransferLeadershipTimeout,
	}
	go rc.startRaft(cfg.Period, cfg.ElectionTick, cfg.HeartbeatTick)
	return commitC, errorC, rc.snapshotterReady, rc
}

func (rc *RaftNode) saveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index:     snap.Metadata.Index,
		Term:      snap.Metadata.Term,
		ConfState: &snap.Metadata.ConfState,
	}
	// save the snapshot file before writing the snapshot to the wal.
	// This makes it possible for the snapshot file to become orphaned, but prevents
	// a WAL snapshot entry from having no corresponding snapshot file.
	if err := rc.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	if err := rc.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	return rc.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (rc *RaftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}
	firstIdx := ents[0].Index
	if firstIdx > rc.appliedIndex+1 {
		klog.Fatalf("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rc.appliedIndex)
	}
	if rc.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rc.appliedIndex-firstIdx+1:]
	}
	return nents
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rc *RaftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
	if len(ents) == 0 {
		return nil, true
	}
	data := make([]string, 0, len(ents))
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}
			s := string(ents[i].Data)
			data = append(data, s)

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			err := cc.Unmarshal(ents[i].Data)
			if err != nil {
				klog.Errorln(err)
			}
			rc.confState = *rc.Node.ApplyConfChange(cc)
			switch cc.Type {

			case raftpb.ConfChangeAddNode:
				if cc.NodeID == uint64(rc.id) {
					continue
				}
				if len(cc.Context) == 0 {
					// tempUrl is not given means, this node was removed,
					// now we re applying the wal files so just adding
					// garbage url, this node is removed in upcoming
					// records
					tempUrl := "http://ds.demo.svc:2380"
					cc.Context = []byte(tempUrl)
				}
				rc.Transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rc.id) {
					continue
				}
				rc.Transport.RemovePeer(types.ID(cc.NodeID))
			}
		}
	}
	var applyDoneC chan struct{}
	if len(data) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case rc.commitC <- &Commit{Data: data, ApplyDoneC: applyDoneC}:
		case <-rc.stopc:
			return nil, false
		}
	}
	// after commit, update appliedIndex
	rc.appliedIndex = ents[len(ents)-1].Index

	return applyDoneC, true
}

func (rc *RaftNode) loadSnapshot() *raftpb.Snapshot {
	if wal.Exist(rc.waldir) {
		walSnaps, err := wal.ValidSnapshotEntries(rc.logger, rc.waldir)
		if err != nil {
			klog.Fatalf("error listing snapshots (%v)", err)
		}
		snapshot, err := rc.snapshotter.LoadNewestAvailable(walSnaps)
		if err != nil && err != snap.ErrNoSnapshot {
			klog.Fatalf("error loading snapshot (%v)", err)
		}
		return snapshot
	}
	return &raftpb.Snapshot{}
}

// openWAL returns a WAL ready for reading.
func (rc *RaftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
	if !wal.Exist(rc.waldir) {
		if err := os.MkdirAll(rc.waldir, 0o750); err != nil {
			klog.Fatalf("cannot create dir for wal (%v)", err)
		}
		w, err := wal.Create(zap.NewExample(), rc.waldir, nil)
		if err != nil {
			klog.Fatalf("create wal error (%v)", err)
		}
		if err := w.Close(); err != nil {
			klog.Errorf("failed to close WAL: %v", err)
		}
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	klog.Infoln(fmt.Sprintf("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index))
	w, err := wal.Open(zap.NewExample(), rc.waldir, walsnap)
	if err != nil {
		klog.Fatalf("error loading wal (%v)", err)
	}

	return w
}

// replayWAL replays WAL entries into the raft instance.
func (rc *RaftNode) replayWAL() *wal.WAL {
	klog.Infoln(fmt.Sprintf("replaying WAL of member %d", rc.id))
	snapshot := rc.loadSnapshot()
	w := rc.openWAL(snapshot)
	_, st, ents, err := w.ReadAll()
	if err != nil {
		klog.Fatalf("failed to read WAL (%v)", err)
	}
	rc.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		err := rc.raftStorage.ApplySnapshot(*snapshot)
		if err != nil {
			klog.Errorln(err.Error())
		}
	}
	err = rc.raftStorage.SetHardState(st)
	if err != nil {
		klog.Errorln(err.Error())
	}

	// append to storage so raft starts at the right place in log
	err = rc.raftStorage.Append(ents)
	if err != nil {
		klog.Errorln(err.Error())
	}

	return w
}

func (rc *RaftNode) writeError(err error) {
	rc.stopHTTP()
	close(rc.commitC)
	rc.errorC <- err
	close(rc.errorC)
	rc.Node.Stop()
}

func (rc *RaftNode) startRaft(period time.Duration, electionTick uint64, heartbeatTick uint64) {
	if !fileutil.Exist(rc.snapdir) {
		if err := os.MkdirAll(rc.snapdir, 0o750); err != nil {
			klog.Fatalf("cannot create dir for snapshot (%v)", err)
		}
	}
	rc.snapshotter = snap.New(zap.NewExample(), rc.snapdir)

	oldwal := wal.Exist(rc.waldir)
	rc.wal = rc.replayWAL()

	// signal replay has finished
	rc.snapshotterReady <- rc.snapshotter

	rpeers := make([]raft.Peer, len(rc.peers))
	for i := range rpeers {
		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
	}
	c := &raft.Config{
		ID:                        uint64(rc.id),
		ElectionTick:              int(electionTick),
		HeartbeatTick:             int(heartbeatTick),
		Storage:                   rc.raftStorage,
		MaxSizePerMsg:             1024 * 1024,
		MaxInflightMsgs:           256,
		MaxUncommittedEntriesSize: 1 << 30,
	}

	if oldwal || rc.join {
		rc.Node = raft.RestartNode(c, rc.keyList)
	} else {
		rc.Node = raft.StartNode(c, rpeers, rc.keyList)
	}

	rc.Transport = &rafthttp.Transport{
		Logger:      rc.logger,
		ID:          types.ID(rc.id),
		ClusterID:   0x1000,
		Raft:        rc,
		ServerStats: stats.NewServerStats("", ""),
		LeaderStats: stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(rc.id)),
		ErrorC:      make(chan error),
	}

	err := rc.Transport.Start()
	if err != nil {
		klog.Errorln(err)
	}
	for i := range rc.peers {
		if i+1 != rc.id {
			rc.Transport.AddPeer(types.ID(i+1), []string{rc.peers[i]})
		}
	}
	go rc.serveRaft()
	go rc.serveChannels(period)
}

// stop closes http, closes all channels, and stops raft.
func (rc *RaftNode) stop() {
	rc.stopHTTP()
	close(rc.commitC)
	close(rc.errorC)
	rc.Node.Stop()
}

func (rc *RaftNode) stopHTTP() {
	rc.Transport.Stop()
	close(rc.httpstopc)
	<-rc.httpdonec
}

func (rc *RaftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	klog.Infoln("publishing snapshot at index ", rc.snapshotIndex)
	defer klog.Infoln("finished publishing snapshot at index ", rc.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rc.appliedIndex {
		klog.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rc.appliedIndex)
	}
	rc.commitC <- nil // trigger kvstore to load snapshot

	rc.confState = snapshotToSave.Metadata.ConfState
	rc.snapshotIndex = snapshotToSave.Metadata.Index
	rc.appliedIndex = snapshotToSave.Metadata.Index
}

var snapshotCatchUpEntriesN uint64 = 10000

func (rc *RaftNode) maybeTriggerSnapshot(applyDoneC <-chan struct{}) {
	if rc.appliedIndex-rc.snapshotIndex <= rc.snapCount {
		return
	}

	// wait until all committed entries are applied (or server is closed)
	if applyDoneC != nil {
		select {
		case <-applyDoneC:
		case <-rc.stopc:
			return
		}
	}

	klog.Infoln(fmt.Sprintf("start snapshot [applied index: %d | last snapshot index: %d]", rc.appliedIndex, rc.snapshotIndex))
	data, err := rc.getSnapshot()
	if err != nil {
		klog.Fatalln(err)
	}
	snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
	if err != nil {
		panic(err)
	}
	if err := rc.saveSnap(snap); err != nil {
		panic(err)
	}

	compactIndex := uint64(1)
	if rc.appliedIndex > snapshotCatchUpEntriesN {
		compactIndex = rc.appliedIndex - snapshotCatchUpEntriesN
	}
	if err := rc.raftStorage.Compact(compactIndex); err != nil {
		panic(err)
	}

	klog.Infoln("compacted log at index ", compactIndex)
	rc.snapshotIndex = rc.appliedIndex
}

func (rc *RaftNode) serveChannels(period time.Duration) {
	snap, err := rc.raftStorage.Snapshot()
	if err != nil {
		panic(err)
	}
	rc.confState = snap.Metadata.ConfState
	rc.snapshotIndex = snap.Metadata.Index
	rc.appliedIndex = snap.Metadata.Index

	defer func() {
		if err := rc.wal.Close(); err != nil {
			klog.Errorf("failed to close WAL: %v", err)
		}
	}()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	// send proposals over raft
	go func() {
		confChangeCount := uint64(0)
		timeout := rc.transferLeadershipTimeout

		for rc.proposeC != nil && rc.confChangeC != nil && rc.transferLeadershipC != nil {
			select {
			case prop, ok := <-rc.proposeC:
				if !ok {
					rc.proposeC = nil
				} else {
					// blocks until accepted by raft state machine
					err = rc.Node.Propose(context.TODO(), []byte(prop))
					if err != nil {
						klog.Errorln(err)
					}
				}
			case transferLeader, ok := <-rc.transferLeadershipC:
				if !ok {
					rc.transferLeadershipC = nil
				} else {
					// blocks until accepted by raft state machine
					err = func() error {
						ctx, cancel := context.WithTimeout(context.TODO(), timeout/4)
						defer cancel()
						err = rc.MoveLeadership(ctx, rc.Node.Status().ID, uint64(*transferLeader.Transferee))
						if err != nil {
							return err
						}
						return nil
					}()
					if err != nil {
						klog.Errorln(err)
					}
				}
			case cc, ok := <-rc.confChangeC:
				if !ok {
					rc.confChangeC = nil
				} else {
					confChangeCount++
					cc.ID = confChangeCount
					err = rc.Node.ProposeConfChange(context.TODO(), cc)
					if err != nil {
						klog.Errorln(err)
					}
				}
			}
		}
		// client closed channel; shutdown raft if not already
		close(rc.stopc)
	}()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.Node.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-rc.Node.Ready():
			err = rc.wal.Save(rd.HardState, rd.Entries)
			if err != nil {
				klog.Errorln(err)
			}

			if !raft.IsEmptySnap(rd.Snapshot) {
				err = rc.saveSnap(rd.Snapshot)
				if err != nil {
					klog.Errorln(err)
				}

				err = rc.raftStorage.ApplySnapshot(rd.Snapshot)
				if err != nil {
					klog.Errorln(err)
				}
				rc.publishSnapshot(rd.Snapshot)

			}

			err = rc.raftStorage.Append(rd.Entries)
			if err != nil {
				klog.Errorln(err)
			}

			rc.Transport.Send(rc.processMessages(rd.Messages))
			applyDoneC, _ := rc.publishEntries(rc.entriesToApply(rd.CommittedEntries))
			rc.maybeTriggerSnapshot(applyDoneC)
			rc.Node.Advance()

		case err := <-rc.Transport.ErrorC:
			rc.writeError(err)
			return

		case <-rc.stopc:
			rc.stop()
			return
		}
	}
}

// When there is a `raftpb.EntryConfChange` after creating the snapshot,
// then the confState included in the snapshot is out of date. so We need
// to update the confState before sending a snapshot to a follower.
func (rc *RaftNode) processMessages(ms []raftpb.Message) []raftpb.Message {
	for i := range ms {
		if ms[i].Type == raftpb.MsgSnap {
			ms[i].Snapshot.Metadata.ConfState = rc.confState
		}
	}
	return ms
}

func (rc *RaftNode) serveRaft() {
	url, err := url.Parse(rc.peers[rc.id-1])
	if err != nil {
		klog.Fatalf("Failed parsing URL (%v)", err)
	}

	ln, err := newStoppableListener(url.Host, rc.httpstopc)
	if err != nil {
		klog.Fatalf("Failed to listen rafthttp (%v)", err)
	}

	err = (&http.Server{Handler: rc.Transport.Handler()}).Serve(ln)
	select {
	case <-rc.httpstopc:
	default:
		klog.Fatalf("Failed to serve rafthttp (%v)", err)
	}
	close(rc.httpdonec)
}

func (rc *RaftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rc.Node.Step(ctx, m)
}

func (rc *RaftNode) IsIDRemoved(id uint64) bool  { return false }
func (rc *RaftNode) ReportUnreachable(id uint64) { rc.Node.ReportUnreachable(id) }
func (rc *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	rc.Node.ReportSnapshot(id, status)
}

func (rc *RaftNode) MoveLeadership(ctx context.Context, myid uint64, transferee uint64) error {
	interval := 500 * time.Millisecond
	if rc.Node.Status().Lead != myid {
		return fmt.Errorf("local node is not the leader. ")
	}

	rc.Node.TransferLeadership(ctx, myid, transferee)

	for rc.Node.Status().Lead != transferee {
		select {
		case <-ctx.Done(): // time out
			return fmt.Errorf(" ErrTimeoutLeaderTransfer")
		case <-time.After(interval):
			klog.V(4).Infoln("waiting for the target to be leader...")
		}
	}

	return nil
}

// ConfState returns the current raft configuration state
func (rc *RaftNode) ConfState() raftpb.ConfState {
	return rc.confState
}
