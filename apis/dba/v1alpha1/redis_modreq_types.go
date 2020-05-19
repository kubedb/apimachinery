/*
Copyright The KubeDB Authors.

Licensed under the Apache License, ModificationRequest 2.0 (the "License");
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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceCodeRedisModificationRequest     = "rdmodreq"
	ResourceKindRedisModificationRequest     = "RedisModificationRequest"
	ResourceSingularRedisModificationRequest = "redismodificationrequest"
	ResourcePluralRedisModificationRequest   = "redismodificationrequests"
)

// RedisModificationRequest defines a Redis database version.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=redismodificationrequests,singular=redismodificationrequest,shortName=rdmodreq,scope=Cluster,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RedisModificationRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              RedisModificationRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            RedisModificationRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// RedisModificationRequestSpec is the spec for RedisModificationRequest
type RedisModificationRequestSpec struct {
	// Specifies the Elasticsearch reference
	DatabaseRef v1.LocalObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`
	// Specifies the modification request type: Upgrade, HorizontalScaling, VerticalScaling etc.
	Type ModificationRequestType `json:"type" protobuf:"bytes,2,opt,name=type,casttype=ModificationRequestType"`
	// Specifies the field information that needed to be upgraded
	Upgrade *UpgradeSpec `json:"upgrade,omitempty" protobuf:"bytes,3,opt,name=upgrade"`
}

// RedisModificationRequestStatus is the status for RedisModificationRequest
type RedisModificationRequestStatus struct {
	Phase ModificationRequestPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ModificationRequestPhase"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisModificationRequestList is a list of RedisModificationRequests
type RedisModificationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of RedisModificationRequest CRD objects
	Items []RedisModificationRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
