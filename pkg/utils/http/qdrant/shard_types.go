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

// LocalShardInfo represents information about a local shard.
type LocalShardInfo struct {
	ShardID uint64 `json:"shard_id"`
	State   string `json:"state"`
}

// RemoteShardInfo represents information about a remote shard.
type RemoteShardInfo struct {
	ShardID uint64 `json:"shard_id"`
	State   string `json:"state"`
	PeerID  uint64 `json:"peer_id"`
	PeerURI string `json:"peer_uri"`
}

// ShardTransfer represents information about a shard transfer.
type ShardTransfer struct {
	ShardID     uint64 `json:"shard_id"`
	FromPeerID  uint64 `json:"from_peer_id"`
	ToPeerID    uint64 `json:"to_peer_id"`
	FromPeerURI string `json:"from_peer_uri,omitempty"`
	ToPeerURI   string `json:"to_peer_uri,omitempty"`
	SyncState   string `json:"sync_state,omitempty"`
}

// MoveShardRequest represents a request to move a shard between peers.
type MoveShardRequest struct {
	MoveShard MoveShardOperation `json:"move_shard"`
}

// MoveShardOperation contains the details of a shard move operation.
type MoveShardOperation struct {
	ShardID    uint64 `json:"shard_id"`
	FromPeerID uint64 `json:"from_peer_id"`
	ToPeerID   uint64 `json:"to_peer_id"`
}

// DropReplicaRequest represents a request to drop a shard replica.
type DropReplicaRequest struct {
	DropReplica DropReplicaOperation `json:"drop_replica"`
}

// DropReplicaOperation contains the details of a drop replica operation.
type DropReplicaOperation struct {
	ShardID uint64 `json:"shard_id"`
	PeerID  uint64 `json:"peer_id"`
}
