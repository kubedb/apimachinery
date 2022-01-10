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
	ResourceKindMySQLInsight = "MySQLInsight"
	ResourceMySQLInsight     = "mysqlinsight"
	ResourceMySQLInsights    = "mysqlinsights"
)

// MySQLInsightSpec defines the desired state of MySQLInsight
type MySQLInsightSpec struct {
	Version                       string  `json:"version,omitempty" protobuf:"bytes,1,opt,name=version"`
	Status                        string  `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`
	Mode                          string  `json:"mode,omitempty" protobuf:"bytes,3,opt,name=mode"`
	MaxConnections                int32   `json:"maxConnections,omitempty" protobuf:"varint,4,opt,name=maxConnections"`
	MaxUsedConnections            int32   `json:"maxUsedConnections,omitempty" protobuf:"varint,5,opt,name=maxUsedConnections"`
	Questions                     int32   `json:"questions,omitempty" protobuf:"varint,6,opt,name=questions"`
	LongQueryTimeThresholdSeconds float64 `json:"longQueryTimeThresholdSeconds,omitempty" protobuf:"fixed64,7,opt,name=longQueryTimeThresholdSeconds"`
	NumberOfSlowQueries           int32   `json:"numberOfSlowQueries,omitempty" protobuf:"varint,8,opt,name=numberOfSlowQueries"`
	AbortedClients                int32   `json:"abortedClients,omitempty" protobuf:"varint,9,opt,name=abortedClients"`
	AbortedConnections            int32   `json:"abortedConnections,omitempty" protobuf:"varint,10,opt,name=abortedConnections"`
	ThreadsCached                 int32   `json:"threadsCached,omitempty" protobuf:"varint,11,opt,name=threadsCached"`
	ThreadsConnected              int32   `json:"threadsConnected,omitempty" protobuf:"varint,12,opt,name=threadsConnected"`
	ThreadsCreated                int32   `json:"threadsCreated,omitempty" protobuf:"varint,13,opt,name=threadsCreated"`
	ThreadsRunning                int32   `json:"threadsRunning,omitempty" protobuf:"varint,14,opt,name=threadsRunning"`
}

// MySQLInsight is the Schema for the mysqlinsights API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MySQLInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   MySQLInsightSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status api.MySQLStatus  `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MySQLInsightList contains a list of MySQLInsight

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MySQLInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []MySQLInsight `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&MySQLInsight{}, &MySQLInsightList{})
}
