// // Copyright 2015 The etcd Authors
// //
// // Licensed under the Apache License, Version 2.0 (the "License");
// // you may not use this file except in compliance with the License.
// // You may obtain a copy of the License at
// //
// //     http://www.apache.org/licenses/LICENSE-2.0
// //
// // Unless required by applicable law or agreed to in writing, software
// // distributed under the License is distributed on an "AS IS" BASIS,
// // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// // See the License for the specific language governing permissions and
// // limitations under the License.
package raft

//
//import (
//	"context"
//	"os"
//	"strconv"
//	"time"
//
//	"kubedb.dev/apimachinery/apis/kubedb"
//
//	"go.etcd.io/etcd/client/pkg/v3/fileutil"
//	"go.etcd.io/etcd/client/pkg/v3/types"
//	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
//	etcdsnap "go.etcd.io/etcd/server/v3/etcdserver/api/snap"
//	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
//	"go.etcd.io/etcd/server/v3/storage/wal"
//	"go.etcd.io/raft/v3"
//	"go.uber.org/zap"
//	"k8s.io/klog/v2"
//)
//
//var defaultSnapshotCount uint64 = 10000
//
//// startRaft initializes and starts the Raft consensus algorithm with specified configuration.
//// It sets up snapshots, WAL replay, node configuration, and transport layers.
//func (rc *RaftNode) startRaft(period time.Duration, electionTick uint64, heartbeatTick uint64) {
//	if !fileutil.Exist(rc.snapdir) {
//		if err := os.MkdirAll(rc.snapdir, 0o750); err != nil {
//			klog.Fatalln("cannot create dir for snapshot", err.Error())
//		}
//	}
//	rc.snapshotter = etcdsnap.New(zap.NewExample(), rc.snapdir)
//
//	oldwal := wal.Exist(rc.waldir)
//	rc.wal = rc.replayWAL()
//
//	// signal replay has finished
//	rc.snapshotterReady <- rc.snapshotter
//
//	rpeers := make([]raft.Peer, len(rc.peers))
//	for i := range rpeers {
//		rpeers[i] = raft.Peer{ID: uint64(i + 1)}
//	}
//	c := &raft.Config{
//		ID:                        uint64(rc.id),
//		ElectionTick:              int(electionTick),
//		HeartbeatTick:             int(heartbeatTick),
//		Storage:                   rc.raftStorage,
//		MaxSizePerMsg:             1024 * 1024,
//		MaxInflightMsgs:           256,
//		MaxUncommittedEntriesSize: 1 << 30,
//		PreVote:                   true,
//	}
//	// this two are the keys that we need to update incase of data corruption in raft replica nodes
//	// this two are the only value that we are going to save in raft.
//	// one is the current timeline value
//	// another one is the postgres coordinator status
//	keyList := []string{
//		StickyLeader,
//		leaderTimeline,
//		kubedb.PostgresPgCoordinatorStatus,
//	}
//
//	if oldwal || rc.join {
//		rc.Node = raft.RestartNode(c, keyList)
//	} else {
//		rc.Node = raft.StartNode(c, rpeers, keyList)
//	}
//
//	rc.Transport = &rafthttp.Transport{
//		Logger:      rc.logger,
//		ID:          types.ID(rc.id),
//		ClusterID:   0x1000,
//		Raft:        rc,
//		ServerStats: stats.NewServerStats("", ""),
//		LeaderStats: stats.NewLeaderStats(zap.NewExample(), strconv.Itoa(rc.id)),
//		ErrorC:      make(chan error),
//	}
//
//	err := rc.Transport.Start()
//	if err != nil {
//		klog.Errorln(err)
//	}
//	for i := range rc.peers {
//		if i+1 != rc.id {
//			rc.Transport.AddPeer(types.ID(i+1), []string{rc.peers[i]})
//		}
//	}
//	go rc.serveRaft()
//	go rc.serveChannels()
//}
