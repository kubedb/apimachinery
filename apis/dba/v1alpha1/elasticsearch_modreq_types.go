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
	ResourceCodeElasticsearchModificationRequest     = "esmodreq"
	ResourceKindElasticsearchModificationRequest     = "ElasticsearchModificationRequest"
	ResourceSingularElasticsearchModificationRequest = "elasticsearchmodificationrequest"
	ResourcePluralElasticsearchModificationRequest   = "elasticsearchmodificationrequests"
)

// ElasticsearchModificationRequest defines a Elasticsearch Modification Request object.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=elasticsearchmodificationrequests,singular=elasticsearchmodificationrequest,shortName=esmodreq,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ElasticsearchModificationRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              ElasticsearchModificationRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            ElasticsearchModificationRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ElasticsearchModificationRequestSpec is the spec for ElasticsearchModificationRequest object
type ElasticsearchModificationRequestSpec struct {
	// Specifies the Elasticsearch reference
	DatabaseRef v1.LocalObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`
	// Specifies the modification request type; ScaleUp, ScaleDown, Upgrade etc.
	Type ModificationRequestType `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Specifies the field information that needed to be updated
	Update *UpdateSpec `json:"update,omitempty" protobuf:"bytes,3,opt,name=update"`
	//Specifies the scaling info of Elasticsearch Object
	Scale *ScaleSpec `json:"scale,omitempty" protobuf:"bytes,4,opt,name=scale"`
}

// ScaleSpec contains the scaling information of the Elasticsearch
type ScaleSpec struct {
	// Number of master nodes
	Master *int32 `json:"master,omitempty" protobuf:"bytes,1,opt,name=master"`
	// Number of data nodes
	Data *int32 `json:"data,omitempty" protobuf:"bytes,2,opt,name=data"`
	// Number of client nodes
	Client *int32 `json:"client,omitempty" protobuf:"bytes,3,opt,name=client"`
}

// ElasticsearchModificationRequestStatus is the status for elasticsearch version
type ElasticsearchModificationRequestStatus struct {
	// Specifies the current phase of the modification request
	// +optional
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

// ElasticsearchModificationRequestList is a list of ElasticsearchModificationRequests
type ElasticsearchModificationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of ElasticsearchModificationRequest CRD objects
	Items []ElasticsearchModificationRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
