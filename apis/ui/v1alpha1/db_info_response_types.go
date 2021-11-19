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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDBInfoResponse = "DBInfoResponse"
	ResourceDBInfoResponse     = "dbinforesponse"
	ResourceDBInfoResponses    = "dbinforesponses"
)

// DBInfoResponseSpec defines the desired state of DBInfoResponse
type DBInfoResponseSpec struct {
	Resources        corev1.ResourceList `json:"resources" protobuf:"bytes,1,rep,name=resources,casttype=k8s.io/api/core/v1.ResourceList,castkey=k8s.io/api/core/v1.ResourceName"`
	Cluster          string              `json:"cluster" protobuf:"bytes,2,opt,name=cluster"`
	Database         string              `json:"database" protobuf:"bytes,3,opt,name=database"`
	NumberOfInstance int64               `json:"numberOfInstance" protobuf:"varint,4,opt,name=numberOfInstance"`
	ProdDatabases    int64               `json:"prodDatabases" protobuf:"varint,5,opt,name=prodDatabases"`
	DevDatabases     int64               `json:"devDatabases" protobuf:"varint,6,opt,name=devDatabases"`
	QADatabases      int64               `json:"qaDatabases" protobuf:"varint,7,opt,name=qaDatabases"`
	Cost             int64               `json:"cost" protobuf:"varint,8,opt,name=cost"`
	Status           DBStatus            `json:"status" protobuf:"bytes,9,opt,name=status,casttype=DBStatus"`
	Age              int64               `json:"age" protobuf:"varint,10,opt,name=age"`
}

// DBInfoResponseStatus defines the observed state of DBInfoResponse
type DBInfoResponseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// DBInfoResponse is the Schema for the dbinforesponses API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DBInfoResponse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DBInfoResponseSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DBInfoResponseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DBInfoResponseList contains a list of DBInfoResponse

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DBInfoResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DBInfoResponse `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type NodeMetrics struct {
	CPULimit     string `json:"cpuLimit" protobuf:"bytes,1,opt,name=cpuLimit"`
	MemoryLimit  string `json:"memoryLimit" protobuf:"bytes,2,opt,name=memoryLimit"`
	StorageLimit string `json:"storageCapacity" protobuf:"bytes,3,opt,name=storageCapacity"`
}

func init() {
	SchemeBuilder.Register(&DBInfoResponse{}, &DBInfoResponseList{})
}
