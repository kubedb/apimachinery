/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package qdrant

// PeerState represents the state of a peer in the cluster.
type PeerState struct {
	PeerID uint64 `json:"peer_id"`
	URI    string `json:"uri"`
	State  string `json:"state"`
}

// RaftInfo contains information about the Raft consensus state.
type RaftInfo struct {
	Term              uint64  `json:"term"`
	Commit            uint64  `json:"commit"`
	PendingOperations int     `json:"pending_operations"` // >= 0
	IsVoter           bool    `json:"is_voter"`
	Leader            *uint64 `json:"leader"` // nil if no leader
	Role              string  `json:"role"`
}

// MessageSendError represents a message send failure record.
type MessageSendError struct {
	Count           int     `json:"count"`
	LatestError     *string `json:"latest_error"`
	LatestErrorTime *string `json:"latest_error_timestamp"`
}

// ClusterInfo represents the overall cluster information.
type ClusterInfo struct {
	PeerID                uint64                      `json:"peer_id"`
	Peers                 map[string]PeerState        `json:"peers"`
	ShardTransfers        []ShardTransfer             `json:"shard_transfers"`
	ConsensusThreadStatus map[string]any              `json:"consensus_thread_status"`
	MessageSendFailures   map[string]MessageSendError `json:"message_send_failures"`
	RaftInfo              RaftInfo                    `json:"raft_info"`
}

// GetClusterInfoResponse represents the response from getting cluster info.
type GetClusterInfoResponse struct {
	Time   float64     `json:"time"`
	Status string      `json:"status"`
	Result ClusterInfo `json:"result"`
}

// CollectionClusterInfo represents cluster information for a collection.
type CollectionClusterInfo struct {
	ShardCount     int                  `json:"shard_count"`
	ReplicaCount   int                  `json:"replica_count"`
	PeerID         uint64               `json:"peer_id"`
	Peers          map[string]PeerState `json:"peers"`
	LocalShards    []LocalShardInfo     `json:"local_shards"`
	RemoteShards   []RemoteShardInfo    `json:"remote_shards"`
	ShardTransfers []ShardTransfer      `json:"shard_transfers"`
}

// GetCollectionClusterInfoResponse represents the response from getting collection cluster info
type GetCollectionClusterInfoResponse struct {
	Time   float64               `json:"time"`
	Status string                `json:"status"`
	Result CollectionClusterInfo `json:"result"`
}

// RemovePeerRequest represents the request parameters for removing a peer.
type RemovePeerRequest struct {
	Timeout *int `json:"timeout,omitempty"` // Wait for operation commit timeout in seconds
	Force   bool `json:"force"`             // If true - removes peer even if it has shards/replicas on it
}

// RemovePeerResponse represents the response from removing a peer.
type RemovePeerResponse struct {
	Usage  PeerUsage `json:"usage"`
	Time   float64   `json:"time"`
	Status string    `json:"status"`
	Result bool      `json:"result"`
}

// PeerUsage represents the hardware usage of a peer.
type PeerUsage struct {
	Hardware  HardwareUsage  `json:"hardware"`
	Inference InferenceUsage `json:"inference"`
}

// HardwareUsage represents hardware resource usage.
type HardwareUsage struct {
	CPU                 int `json:"cpu"`
	PayloadIORead       int `json:"payload_io_read"`
	PayloadIOWrite      int `json:"payload_io_write"`
	PayloadIndexIORead  int `json:"payload_index_io_read"`
	PayloadIndexIOWrite int `json:"payload_index_io_write"`
	VectorIORead        int `json:"vector_io_read"`
	VectorIOWrite       int `json:"vector_io_write"`
}

// InferenceUsage represents inference usage.
type InferenceUsage struct {
	Models map[string]any `json:"models"`
}
