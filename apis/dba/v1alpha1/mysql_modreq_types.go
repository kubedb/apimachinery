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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceCodeMySQLModificationRequest     = "mymodreq"
	ResourceKindMySQLModificationRequest     = "MySQLModificationRequest"
	ResourceSingularMySQLModificationRequest = "mysqlmodificationrequest"
	ResourcePluralMySQLModificationRequest   = "mysqlmodificationrequests"
)

// MySQLModificationRequest defines a MySQL Modification Request object.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mysqlmodificationrequests,singular=mysqlmodificationrequest,shortName=mymodreq,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MySQLModificationRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MySQLModificationRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            MySQLModificationRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MySQLModificationRequestSpec is the spec for MySQLModificationRequest version
type MySQLModificationRequestSpec struct {
	// Specifies the database reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`
	// Specifies the modification request type; ScaleUp, ScaleDown, Upgrade etc.
	Type ModificationRequestType `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Specifies the field information that needed to be updated
	Update *MySQLUpdateSpec `json:"update,omitempty" protobuf:"bytes,3,opt,name=update"`
	//Specifies the scaling info of database
	Scale *MySQLScaleSpec `json:"scale,omitempty" protobuf:"bytes,4,opt,name=scale"`
}

type MySQLUpdateSpec struct {
	// Specifies the MySQLVersion object name
	TargetVersion string `json:"targetVersion,omitempty" protobuf:"bytes,1,opt,name=targetVersion"`
	// Specifies the current ordinal of the StatefulSet
	CurrentStatefulSetOrdinal *int32 `json:"CurrentStatefulSetOrdinal,omitempty" protobuf:"varint,2,opt,name=CurrentStatefulSetOrdinal"`
}

// MySQLScaleSpec contains the scaling information of the MySQL
type MySQLScaleSpec struct {
	// Horizontal specifies the horizontal scaling.
	Horizontal *HorizontalScale `json:"horizontal,omitempty" protobuf:"bytes,1,opt,name=horizontal"`
	// Vertical specifies the vertical scaling.
	Vertical *VerticalScale `json:"vertical,omitempty" protobuf:"bytes,2,opt,name=vertical"`
	// specifies the weight of the current member/Node
	MemberWeight int32 `json:"memberWeight,omitempty" protobuf:"varint,3,opt,name=memberWeight"`
}

type HorizontalScale struct {
	// Number of nodes/members of the group
	Member *int32 `json:"member,omitempty" protobuf:"varint,1,opt,name=member"`
}

type VerticalScale struct {
	// Containers represents the containers specification for scaling the requested resources.
	Containers []ContainerResources `json:"containers,omitempty" protobuf:"bytes,1,opt,name=containers"`
}

// MySQLModificationRequestStatus is the status for MySQLModificationRequest object
type MySQLModificationRequestStatus struct {
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

// MySQLModificationRequestList is a list of MySQLModificationRequests
type MySQLModificationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of MySQLModificationRequest CRD objects
	Items []MySQLModificationRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
