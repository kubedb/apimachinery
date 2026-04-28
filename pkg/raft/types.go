package raft

//
//import (
//	"net"
//	"time"
//
//	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
//	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
//	"go.etcd.io/etcd/server/v3/storage/wal"
//	"go.etcd.io/raft/v3"
//	"go.etcd.io/raft/v3/raftpb"
//	"go.uber.org/zap"
//)
//
//type commit struct {
//	data       []string
//	applyDoneC chan<- struct{}
//}
//
//type TransferLeadershipConfig struct {
//	Transferee *int `json:"transferee" protobuf:"varint,1,opt,name=transferee"`
//}
//type KeyValue struct {
//	Key   *string `json:"key" protobuf:"bytes,1,opt,name=key"`
//	Value *string `json:"value" protobuf:"bytes,2,opt,name=value"`
//}
//
//type NodeInfo struct {
//	NodeId *int    `json:"id" protobuf:"varint,1,opt,name=id"`
//	Url    *string `json:"url" protobuf:"bytes,2,opt,name=url"`
//}
//
//// stoppableListener sets TCP keep-alive timeouts on accepted
//// connections and waits on stopc message
//type stoppableListener struct {
//	*net.TCPListener
//	stopc <-chan struct{}
//}
//
//// RaftNode A key-value stream backed by raft
//type RaftNode struct {
//	*Config
//
//	id          int      // client ID for raft session
//	peers       []string // raft peer URLs
//	join        bool     // Node is joining an existing cluster
//	waldir      string   // path to WAL directory
//	snapdir     string   // path to snapshot directory
//	getSnapshot func() ([]byte, error)
//
//	confState     raftpb.ConfState
//	snapshotIndex uint64
//	appliedIndex  uint64
//
//	// raft backing for the commit/error channel
//	Node        raft.Node
//	raftStorage *raft.MemoryStorage
//	wal         *wal.WAL
//
//	snapshotter      *snap.Snapshotter
//	snapshotterReady chan *snap.Snapshotter // signals when snapshotter is ready
//
//	snapCount uint64
//	Transport *rafthttp.Transport
//	stopc     chan struct{} // signals proposal channel closed
//	httpstopc chan struct{} // signals http server to shutdown
//	httpdonec chan struct{} // signals http server shutdown complete
//	logger    *zap.Logger
//	keyList   []string // TODO:
//}
//
//type Config struct {
//	period                    time.Duration
//	transferLeaderShipTimeout time.Duration
//	id                        int
//	electionTick              uint64
//	heartbeatTick             uint64
//	peers                     []string
//	join                      bool
//
//	transferLeadershipC <-chan TransferLeadershipConfig
//	proposeC            <-chan string            // proposed messages (k,v)
//	confChangeC         <-chan raftpb.ConfChange // proposed cluster config changes
//	commitC             chan<- *commit           // entries committed to log (k,v)
//	errorC              chan<- error             // errors from raft session
//}
