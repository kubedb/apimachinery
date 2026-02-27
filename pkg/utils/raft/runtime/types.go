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

package raft

// Commit represents data committed to the raft log
type Commit struct {
	Data       []string
	ApplyDoneC chan<- struct{}
}

// TransferLeadershipConfig holds configuration for leader transfer
type TransferLeadershipConfig struct {
	Transferee *int `json:"transferee" protobuf:"varint,1,opt,name=transferee"`
}

// KeyValue represents a key-value pair for the KV store
type KeyValue struct {
	Key   *string `json:"key" protobuf:"bytes,1,opt,name=key"`
	Value *string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

// NodeInfo represents raft node information
type NodeInfo struct {
	NodeId *int    `json:"id" protobuf:"varint,1,opt,name=id"`
	Url    *string `json:"url" protobuf:"bytes,2,opt,name=url"`
}
