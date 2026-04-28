package raft

//
//import (
//	"fmt"
//	"net"
//	"time"
//
//	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
//	"go.etcd.io/raft/v3/raftpb"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//)
//
//// newStoppableListener creates a TCP listener that can be gracefully stopped via a channel.
//// It wraps a standard TCP listener with stop signal handling for clean shutdown.
//func newStoppableListener(addr string, stopc <-chan struct{}) (*stoppableListener, error) {
//	ln, err := net.Listen("tcp", addr)
//	if err != nil {
//		return nil, err
//	}
//	return &stoppableListener{ln.(*net.TCPListener), stopc}, nil
//}
//
//// NewRaftNode creates and initializes a new Raft node instance for distributed consensus.
//// It sets up the necessary channels, storage, and transport layers for Raft communication.
//// Returns commit channel, error channel, snapshotter ready channel, and the RaftNode instance.
//func NewRaftNode(config *Config, getSnapshot func() ([]byte, error), proposeC <-chan string,
//	confChangeC <-chan raftpb.ConfChange, transferLeaderC <-chan TransferLeadershipConfig,
//) (<-chan *commit, <-chan error, <-chan *snap.Snapshotter, *RaftNode) {
//	commitC := make(chan *commit)
//	errorC := make(chan error)
//
//	encoderCfg := zap.NewProductionConfig()
//	encoderCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
//	encoderCfg.Encoding = "console"
//	encoderCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
//	logger, err := encoderCfg.Build()
//	if err != nil {
//		panic(fmt.Errorf("failed to create logger, %v", err))
//	}
//
//	rc := &RaftNode{
//		transferLeadershipC: transferLeaderC,
//		proposeC:            proposeC,
//		confChangeC:         confChangeC,
//		commitC:             commitC,
//		errorC:              errorC,
//		id:                  id,
//		peers:               peers,
//		join:                join,
//		waldir:              "/var/pv/raftwal",
//		snapdir:             "/var/pv/raftsnapshot",
//		getSnapshot:         getSnapshot,
//		snapCount:           defaultSnapshotCount,
//		stopc:               make(chan struct{}),
//		httpstopc:           make(chan struct{}),
//		httpdonec:           make(chan struct{}),
//		logger:              logger,
//		snapshotterReady:    make(chan *snap.Snapshotter, 1),
//	}
//	go rc.startRaft(period, electionTick, heartbeatTick)
//	return commitC, errorC, rc.snapshotterReady, rc
//}
//
//func (rc *RaftNode) WithKeyList(keyList []string) *RaftNode {
//	rc.keyList = keyList
//	return rc
//}
