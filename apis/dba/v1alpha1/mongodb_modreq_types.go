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
)

const (
	ResourceCodeMongoDBModificationRequest     = "mgmodreq"
	ResourceKindMongoDBModificationRequest     = "MongoDBModificationRequest"
	ResourceSingularMongoDBModificationRequest = "mongodbmodificationrequest"
	ResourcePluralMongoDBModificationRequest   = "mongodbmodificationrequests"
)

// MongoDBModificationRequest defines a MongoDB database version.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mongodbmodificationrequests,singular=mongodbmodificationrequest,shortName=mgmodreq,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MongoDBModificationRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MongoDBModificationRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            MongoDBModificationRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MongoDBModificationRequestSpec is the spec for mongodb modification request
type MongoDBModificationRequestSpec struct {
	// Specifies the Elasticsearch reference
	DatabaseRef v1.LocalObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`
	// Specifies the modification request type; ScaleUp, ScaleDown, Upgrade etc.
	Type ModificationRequestType `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Specifies the field information that needed to be updated
	Update *UpdateSpec `json:"update,omitempty" protobuf:"bytes,3,opt,name=update"`
}

type UpdateSpec struct {
	// Specifies the ElasticsearchVersion object name
	TargetVersion string `json:"targetVersion,omitempty" protobuf:"bytes,1,opt,name=targetVersion"`
}

// MongoDBModificationRequestStatus is the status for mongodb modification request
type MongoDBModificationRequestStatus struct {
	Phase  ModificationRequestPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ModificationRequestPhase"`
	Reason string                   `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []MongoDBModificationRequestCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,4,opt,name=observedGeneration"`
}

type MongoDBModificationRequestCondition struct {
	// request approval state, currently Approved or Denied.
	Type RequestConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RequestConditionType"`

	// brief reason for the request state
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`

	// human readable message with details about the request state
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,4,opt,name=lastUpdateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongoDBModificationRequestList is a list of MongoDBModificationRequests
type MongoDBModificationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of MongoDBModificationRequest CRD objects
	Items []MongoDBModificationRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
