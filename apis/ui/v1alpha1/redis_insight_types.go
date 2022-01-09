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

package v1alpha1

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindRedisInsight = "RedisInsight"
	ResourceRedisInsight     = "redisinsight"
	ResourceRedisInsights    = "redisinsights"
)

// RedisInsightSpec defines the desired state of RedisInsight
type RedisInsightSpec struct {
	Version                      string `json:"version" protobuf:"bytes,1,opt,name=version"`
	Status                       string `json:"status" protobuf:"bytes,2,opt,name=status"`
	Mode                         string `json:"mode" protobuf:"bytes,3,opt,name=mode"`
	EvictionPolicy               string `json:"evictionPolicy" protobuf:"bytes,4,opt,name=evictionPolicy"`
	MaxClients                   int64  `json:"maxClients" protobuf:"varint,5,opt,name=maxClients"`
	ConnectedClients             int64  `json:"connectedClients" protobuf:"varint,6,opt,name=connectedClients"`
	BlockedClients               int64  `json:"blockedClients" protobuf:"varint,7,opt,name=blockedClients"`
	TotalKeys                    int64  `json:"totalKeys" protobuf:"varint,8,opt,name=totalKeys"`
	ExpiredKeys                  int64  `json:"expiredKeys" protobuf:"varint,9,opt,name=expiredKeys"`
	EvictedKeys                  int64  `json:"evictedKeys" protobuf:"varint,10,opt,name=evictedKeys"`
	ReceivedConnections          int64  `json:"receivedConnections" protobuf:"varint,11,opt,name=receivedConnections"`
	RejectedConnections          int64  `json:"rejectedConnections" protobuf:"varint,12,opt,name=rejectedConnections"`
	SlowLogThresholdMircoSeconds int64  `json:"slowLogThresholdMircoSeconds" protobuf:"varint,13,opt,name=slowLogThresholdMircoSeconds"`
	SlowLogMaxLen                int64  `json:"slowLogMaxLen" protobuf:"varint,14,opt,name=slowLogMaxLen"`
}

// RedisInsight is the Schema for the redisinsights API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedisInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   RedisInsightSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status api.RedisStatus  `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// RedisInsightList contains a list of RedisInsight

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedisInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []RedisInsight `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&RedisInsight{}, &RedisInsightList{})
}
