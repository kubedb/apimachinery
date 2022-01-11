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
	ResourceKindElasticsearchInsight = "ElasticsearchInsight"
	ResourceElasticsearchInsight     = "elasticsearchinsight"
	ResourceElasticsearchInsights    = "elasticsearchinsights"
)

// ElasticsearchInsightSpec defines the desired state of ElasticsearchInsight
type ElasticsearchInsightSpec struct {
	Version string `json:"version" protobuf:"bytes,1,opt,name=version"`
	Status  string `json:"status" protobuf:"bytes,2,opt,name=status"`
	Mode    string `json:"mode" protobuf:"bytes,3,opt,name=mode"`

	ClusterHealth ElasticsearchClusterHealth `json:",inline" protobuf:"bytes,4,opt,name=clusterHealth"`
}

type ElasticsearchClusterHealth struct {
	ActivePrimaryShards               int32  `json:"activePrimaryShards,omitempty" protobuf:"varint,1,opt,name=activePrimaryShards"`
	ActiveShards                      int32  `json:"activeShards,omitempty" protobuf:"varint,2,opt,name=activeShards"`
	ActiveShardsPercentAsNumber       int32  `json:"activeShardsPercentAsNumber,omitempty" protobuf:"varint,3,opt,name=activeShardsPercentAsNumber"`
	ClusterName                       string `json:"clusterName,omitempty" protobuf:"bytes,4,opt,name=clusterName"`
	DelayedUnassignedShards           int32  `json:"delayedUnassignedShards,omitempty" protobuf:"varint,5,opt,name=delayedUnassignedShards"`
	InitializingShards                int32  `json:"initializingShards,omitempty" protobuf:"varint,6,opt,name=initializingShards"`
	NumberOfDataNodes                 int32  `json:"numberOfDataNodes,omitempty" protobuf:"varint,7,opt,name=numberOfDataNodes"`
	NumberOfInFlightFetch             int32  `json:"numberOfInFlightFetch,omitempty" protobuf:"varint,8,opt,name=numberOfInFlightFetch"`
	NumberOfNodes                     int32  `json:"numberOfNodes,omitempty" protobuf:"varint,9,opt,name=numberOfNodes"`
	NumberOfPendingTasks              int32  `json:"numberOfPendingTasks,omitempty" protobuf:"varint,10,opt,name=numberOfPendingTasks"`
	RelocatingShards                  int32  `json:"relocatingShards,omitempty" protobuf:"varint,11,opt,name=relocatingShards"`
	ClusterStatus                     string `json:"clusterStatus,omitempty" protobuf:"bytes,12,opt,name=clusterStatus"`
	UnassignedShards                  int32  `json:"unassignedShards,omitempty" protobuf:"varint,13,opt,name=unassignedShards"`
	TaskMaxWaitingInQueueMilliSeconds int32  `json:"taskMaxWaitingInQueueMilliSeconds,omitempty" protobuf:"varint,14,opt,name=taskMaxWaitingInQueueMilliSeconds"`
}

// ElasticsearchInsight is the Schema for the elasticsearchinsights API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ElasticsearchInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ElasticsearchInsightSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status api.ElasticsearchStatus  `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ElasticsearchInsightList contains a list of ElasticsearchInsight

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ElasticsearchInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ElasticsearchInsight `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchInsight{}, &ElasticsearchInsightList{})
}
