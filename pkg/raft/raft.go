package raft

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"net/http"
//	"net/url"
//	"os"
//	"time"
//
//	"go.etcd.io/etcd/client/pkg/v3/types"
//	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
//	"go.etcd.io/etcd/server/v3/storage/wal"
//	"go.etcd.io/etcd/server/v3/storage/wal/walpb"
//	"go.etcd.io/raft/v3"
//	"go.etcd.io/raft/v3/raftpb"
//	"go.uber.org/zap"
//	"k8s.io/klog/v2"
//)
//
//// saveSnap saves a Raft snapshot to disk and updates the WAL with snapshot metadata.
//// It ensures the snapshot file is saved before writing to WAL to prevent orphaned entries.
//func (rc *RaftNode) saveSnap(snap raftpb.Snapshot) error {
//	walSnap := walpb.Snapshot{
//		Index:     snap.Metadata.Index,
//		Term:      snap.Metadata.Term,
//		ConfState: &snap.Metadata.ConfState,
//	}
//	// save the snapshot file before writing the snapshot to the wal.
//	// This makes it possible for the snapshot file to become orphaned, but prevents
//	// a WAL snapshot entry from having no corresponding snapshot file.
//	if err := rc.snapshotter.SaveSnap(snap); err != nil {
//		return err
//	}
//	if err := rc.wal.SaveSnapshot(walSnap); err != nil {
//		return err
//	}
//	return rc.wal.ReleaseLockTo(snap.Metadata.Index)
//}
//
//// entriesToApply filters committed entries to find those that need to be applied.
//// It ensures entries are applied in order and prevents duplicate applications.
//func (rc *RaftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
//	if len(ents) == 0 {
//		return ents
//	}
//	firstIdx := ents[0].Index
//	if firstIdx > rc.appliedIndex+1 {
//		klog.Fatalln("first index of committed entry[", firstIdx, "] should <= progress.appliedIndex[", rc.snapshotIndex, "]+1")
//	}
//	if rc.appliedIndex-firstIdx+1 < uint64(len(ents)) {
//		nents = ents[rc.appliedIndex-firstIdx+1:]
//	}
//	return nents
//}
//
//// publishEntries writes committed log entries to commit channel and returns
//// whether all entries could be published.
//func (rc *RaftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
//	if len(ents) == 0 {
//		return nil, true
//	}
//	data := make([]string, 0, len(ents))
//	for i := range ents {
//		switch ents[i].Type {
//		case raftpb.EntryNormal:
//			if len(ents[i].Data) == 0 {
//				// ignore empty messages
//				break
//			}
//			s := string(ents[i].Data)
//			data = append(data, s)
//
//		case raftpb.EntryConfChange:
//			var cc raftpb.ConfChange
//			err := cc.Unmarshal(ents[i].Data)
//			if err != nil {
//				klog.Errorln(err)
//			}
//			rc.confState = *rc.Node.ApplyConfChange(cc)
//			switch cc.Type {
//
//			case raftpb.ConfChangeAddNode:
//				if cc.NodeID == uint64(rc.id) {
//					continue
//				}
//				if len(cc.Context) == 0 {
//					// tempUrl is not given means, this node was removed,
//					// now we re applying the wal files so just adding
//					// garbage url, this node is removed in upcoming
//					// records
//					tempUrl := "http://ds.demo.svc:2380"
//					cc.Context = []byte(tempUrl)
//				}
//				rc.Transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
//			case raftpb.ConfChangeRemoveNode:
//				if cc.NodeID == uint64(rc.id) {
//					continue
//				}
//				rc.Transport.RemovePeer(types.ID(cc.NodeID))
//			}
//		}
//	}
//	var applyDoneC chan struct{}
//	if len(data) > 0 {
//		applyDoneC = make(chan struct{}, 1)
//		select {
//		case rc.commitC <- &commit{data, applyDoneC}:
//		case <-rc.stopc:
//			return nil, false
//		}
//	}
//	// after commit, update appliedIndex
//	rc.appliedIndex = ents[len(ents)-1].Index
//
//	return applyDoneC, true
//}
//
//// loadSnapshot loads the most recent snapshot from disk for Raft state restoration.
//// It validates available snapshots and returns the newest one for cluster recovery.
//func (rc *RaftNode) loadSnapshot() *raftpb.Snapshot {
//	if wal.Exist(rc.waldir) {
//		walSnaps, err := wal.ValidSnapshotEntries(rc.logger, rc.waldir)
//		if err != nil {
//			klog.Fatalln("error listing snapshots", err.Error())
//		}
//		snapshot, err := rc.snapshotter.LoadNewestAvailable(walSnaps)
//		if err != nil && !errors.Is(err, snap.ErrNoSnapshot) {
//			klog.Fatalln("error loading snapshot", err.Error())
//		}
//		return snapshot
//	}
//	return &raftpb.Snapshot{}
//}
//
//// openWAL returns a WAL ready for reading, creating directories and WAL files if needed.
//// It handles both new cluster initialization and existing cluster recovery scenarios.
//func (rc *RaftNode) openWAL(snapshot *raftpb.Snapshot) *wal.WAL {
//	if !wal.Exist(rc.waldir) {
//		if err := os.MkdirAll(rc.waldir, 0o750); err != nil {
//			klog.Fatalln("cannot create dir for wal ", err.Error())
//		}
//		w, err := wal.Create(zap.NewExample(), rc.waldir, nil)
//		if err != nil {
//			klog.Fatalln("create wal error", err.Error())
//		}
//		w.Close() // nolint:errcheck
//	}
//
//	walsnap := walpb.Snapshot{}
//	if snapshot != nil {
//		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
//	}
//	klog.Infoln(fmt.Sprintf("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index))
//	w, err := wal.Open(zap.NewExample(), rc.waldir, walsnap)
//	if err != nil {
//		klog.Fatalln("error loading wal ", err.Error())
//	}
//
//	return w
//}
//
//// replayWAL replays WAL entries into the raft instance for state recovery.
//// It loads snapshots, reads all WAL entries, and restores the Raft state machine.
//func (rc *RaftNode) replayWAL() *wal.WAL {
//	klog.Infoln(fmt.Sprintf("replaying WAL of member %d", rc.id))
//	snapshot := rc.loadSnapshot()
//	w := rc.openWAL(snapshot)
//	_, st, ents, err := w.ReadAll()
//	if err != nil {
//		klog.Fatalln("failed to read WAL ", err.Error())
//	}
//	rc.raftStorage = raft.NewMemoryStorage()
//	if snapshot != nil {
//		err := rc.raftStorage.ApplySnapshot(*snapshot)
//		if err != nil {
//			klog.Errorln(err.Error())
//		}
//	}
//	err = rc.raftStorage.SetHardState(st)
//	if err != nil {
//		klog.Errorln(err.Error())
//	}
//
//	// append to storage so raft starts at the right place in log
//	err = rc.raftStorage.Append(ents)
//	if err != nil {
//		klog.Errorln(err.Error())
//	}
//
//	return w
//}
//
//// writeError handles fatal errors by stopping HTTP transport and closing channels.
//// It ensures clean shutdown when unrecoverable errors occur in the Raft node.
//func (rc *RaftNode) writeError(err error) {
//	rc.stopHTTP()
//	close(rc.commitC)
//	rc.errorC <- err
//	close(rc.errorC)
//	rc.Node.Stop()
//}
//
//// stop closes http, closes all channels, and stops raft.
//func (rc *RaftNode) stop() {
//	rc.stopHTTP()
//	close(rc.commitC)
//	close(rc.errorC)
//	rc.Node.Stop()
//}
//
//// stopHTTP stops the HTTP transport and waits for shutdown completion.
//// It ensures clean termination of all HTTP-related goroutines and connections.
//func (rc *RaftNode) stopHTTP() {
//	rc.Transport.Stop()
//	close(rc.httpstopc)
//	<-rc.httpdonec
//}
//
//// publishSnapshot publishes a snapshot to the commit channel for application by the state machine.
//// It updates the node's applied index and configuration state based on the snapshot metadata.
//func (rc *RaftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
//	if raft.IsEmptySnap(snapshotToSave) {
//		return
//	}
//
//	klog.Infoln("publishing snapshot at index ", rc.snapshotIndex)
//	defer klog.Infoln("finished publishing snapshot at index ", rc.snapshotIndex)
//
//	if snapshotToSave.Metadata.Index <= rc.appliedIndex {
//		klog.Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rc.appliedIndex)
//	}
//	rc.commitC <- nil // trigger kvstore to load snapshot
//
//	rc.confState = snapshotToSave.Metadata.ConfState
//	rc.snapshotIndex = snapshotToSave.Metadata.Index
//	rc.appliedIndex = snapshotToSave.Metadata.Index
//}
//
//var snapshotCatchUpEntriesN uint64 = 10000
//
//// maybeTriggerSnapshot checks if a snapshot should be created based on the number of applied entries.
//// It creates, saves, and compacts the log when the snapshot threshold is reached.
//func (rc *RaftNode) maybeTriggerSnapshot(applyDoneC <-chan struct{}) {
//	if rc.appliedIndex-rc.snapshotIndex <= rc.snapCount {
//		return
//	}
//
//	// wait until all committed entries are applied (or server is closed)
//	if applyDoneC != nil {
//		select {
//		case <-applyDoneC:
//		case <-rc.stopc:
//			return
//		}
//	}
//
//	klog.Infoln(fmt.Sprintf("start snapshot [applied index: %d | last snapshot index: %d]", rc.appliedIndex, rc.snapshotIndex))
//	data, err := rc.getSnapshot()
//	if err != nil {
//		klog.Fatalln(err)
//	}
//	snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
//	if err != nil {
//		panic(err)
//	}
//	if err := rc.saveSnap(snap); err != nil {
//		panic(err)
//	}
//
//	compactIndex := uint64(1)
//	if rc.appliedIndex > snapshotCatchUpEntriesN {
//		compactIndex = rc.appliedIndex - snapshotCatchUpEntriesN
//	}
//	if err := rc.raftStorage.Compact(compactIndex); err != nil {
//		if !errors.Is(err, raft.ErrCompacted) {
//			panic(err)
//		}
//	} else {
//		klog.Infof("compacted log at index %d", compactIndex)
//	}
//	klog.Infoln("compacted log at index ", compactIndex)
//	rc.snapshotIndex = rc.appliedIndex
//}
//
//// serveRaft starts the HTTP server for Raft inter-node communication.
//// It handles incoming Raft protocol messages and manages the transport layer.
//func (rc *RaftNode) serveRaft() {
//	url, err := url.Parse(rc.peers[rc.id-1])
//	if err != nil {
//		klog.Fatalf("Failed parsing URL (%v)", err)
//	}
//
//	ln, err := newStoppableListener(url.Host, rc.httpstopc)
//	if err != nil {
//		klog.Fatalf("Failed to listen rafthttp (%v)", err)
//	}
//
//	err = (&http.Server{Handler: rc.Transport.Handler()}).Serve(ln)
//	select {
//	case <-rc.httpstopc:
//	default:
//		klog.Fatalf("Failed to serve rafthttp (%v)", err)
//	}
//	close(rc.httpdonec)
//}
//
//// Process handles incoming Raft messages from other nodes in the cluster.
//// It forwards messages to the Raft state machine for processing.
//func (rc *RaftNode) Process(ctx context.Context, m raftpb.Message) error {
//	return rc.Node.Step(ctx, m)
//}
//
//// IsIDRemoved always returns false as nodes are not permanently removed in this implementation.
//// This is part of the Raft transport interface requirements.
//func (rc *RaftNode) IsIDRemoved(_ uint64) bool   { return false }
//func (rc *RaftNode) ReportUnreachable(id uint64) { rc.Node.ReportUnreachable(id) }
//func (rc *RaftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
//	rc.Node.ReportSnapshot(id, status)
//}
//
//// MoveLeadership transfers Raft leadership from the current node to the specified target node.
//// It includes timeout handling and validation that the transfer completed successfully.
//func (rc *RaftNode) MoveLeadership(ctx context.Context, myid uint64, transferee uint64) error {
//	interval := time.Duration(500) * time.Millisecond
//	if rc.Node.Status().Lead != myid {
//		return fmt.Errorf("local node is not the leader. ")
//	}
//
//	rc.Node.TransferLeadership(ctx, myid, transferee)
//
//	for rc.Node.Status().Lead != transferee {
//		select {
//		case <-ctx.Done(): // time out
//			return fmt.Errorf(" ErrTimeoutLeaderTransfer")
//		case <-time.After(interval):
//			klog.Infoln("waiting for the target to be leader...")
//		}
//	}
//
//	return nil
//}
//
//// processMessages updates message metadata, particularly snapshot configuration state.
//// It ensures snapshots contain the most current cluster configuration information.
//func (rc *RaftNode) processMessages(ms []raftpb.Message) []raftpb.Message {
//	for i := 0; i < len(ms); i++ {
//		if ms[i].Type == raftpb.MsgSnap {
//			ms[i].Snapshot.Metadata.ConfState = rc.confState
//		}
//	}
//	return ms
//}
//
//// serveChannels handles the main event loop for Raft message processing and state transitions.
//// It manages proposals, configuration changes, leadership transfers, and WAL persistence.
//func (rc *RaftNode) serveChannels() {
//	snap, err := rc.raftStorage.Snapshot()
//	if err != nil {
//		panic(err)
//	}
//	rc.confState = snap.Metadata.ConfState
//	rc.snapshotIndex = snap.Metadata.Index
//	rc.appliedIndex = snap.Metadata.Index
//
//	defer rc.wal.Close() // nolint:errcheck
//
//	ticker := time.NewTicker(rc.period)
//	defer ticker.Stop()
//
//	// send proposals over raft
//	go func() {
//		confChangeCount := uint64(0)
//
//		for rc.proposeC != nil && rc.confChangeC != nil && rc.transferLeadershipC != nil {
//			select {
//			case prop, ok := <-rc.proposeC:
//				if !ok {
//					rc.proposeC = nil
//				} else {
//					// blocks until accepted by raft state machine
//					err = rc.Node.Propose(context.TODO(), []byte(prop))
//					if err != nil {
//						klog.Errorln(err)
//					}
//				}
//			case transferLeader, ok := <-rc.transferLeadershipC:
//				if !ok {
//					rc.transferLeadershipC = nil
//				} else {
//					// blocks until accepted by raft state machine
//					err = func() error {
//						// TODO: make sure transfer leadership timeout is set
//						ctx, cancel := context.WithTimeout(context.TODO(), rc.transferLeaderShipTimeout/4)
//						defer cancel()
//						err = rc.MoveLeadership(ctx, rc.Node.Status().ID, uint64(*transferLeader.Transferee))
//						if err != nil {
//							return err
//						}
//						return nil
//					}()
//					if err != nil {
//						klog.Errorln(err)
//					}
//				}
//			case cc, ok := <-rc.confChangeC:
//				if !ok {
//					rc.confChangeC = nil
//				} else {
//					confChangeCount++
//					cc.ID = confChangeCount
//					err = rc.Node.ProposeConfChange(context.TODO(), cc)
//					if err != nil {
//						klog.Errorln(err)
//					}
//				}
//			}
//		}
//		// client closed channel; shutdown raft if not already
//		close(rc.stopc)
//	}()
//
//	// event loop on raft state machine updates
//	for {
//		select {
//		case <-ticker.C:
//			rc.Node.Tick()
//
//		// store raft entries to wal, then publish over commit channel
//		case rd := <-rc.Node.Ready():
//			// Must save the snapshot file and WAL snapshot entry before saving any other entries
//			// or hardstate to ensure that recovery after a snapshot restore is possible.
//			if !raft.IsEmptySnap(rd.Snapshot) {
//				_ = rc.saveSnap(rd.Snapshot)
//			}
//			err = rc.wal.Save(rd.HardState, rd.Entries)
//			if err != nil {
//				klog.Errorln(err)
//			}
//
//			if !raft.IsEmptySnap(rd.Snapshot) {
//				err = rc.saveSnap(rd.Snapshot)
//				if err != nil {
//					klog.Errorln(err)
//				}
//
//				err = rc.raftStorage.ApplySnapshot(rd.Snapshot)
//				if err != nil {
//					klog.Errorln(err)
//				}
//				rc.publishSnapshot(rd.Snapshot)
//
//			}
//
//			err = rc.raftStorage.Append(rd.Entries)
//			if err != nil {
//				klog.Errorln(err)
//			}
//
//			// Debug logging for messages including heartbeats
//			//if len(rd.Messages) > 0 {
//			//	for _, msg := range rd.Messages {
//			//		if msg.Type == raftpb.MsgHeartbeat {
//			//			klog.V(4).Infof("Sending heartbeat from node %d to node %d", rc.id, msg.To)
//			//		} else if msg.Type == raftpb.MsgHeartbeatResp {
//			//			klog.V(4).Infof("Sending heartbeat response from node %d to node %d", rc.id, msg.To)
//			//		} else {
//			//			klog.V(5).Infof("Sending message type %s from node %d to node %d", msg.Type, rc.id, msg.To)
//			//		}
//			//	}
//			//}
//
//			rc.Transport.Send(rc.processMessages(rd.Messages))
//			applyDoneC, _ := rc.publishEntries(rc.entriesToApply(rd.CommittedEntries))
//			// eikhane rc.stop() stop method ta baad diye dichi
//			// TODO: rc.stop() ta use korte hoibe
//			// plan hoilo lastIndex == rc.totalCommited && confchange remove node is itself
//			rc.maybeTriggerSnapshot(applyDoneC)
//			rc.Node.Advance()
//
//		case err := <-rc.Transport.ErrorC:
//			rc.writeError(err)
//			return
//
//		case <-rc.stopc:
//			rc.stop()
//			return
//		}
//	}
//}
