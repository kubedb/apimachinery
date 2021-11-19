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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDatabaseNodeView = "DatabaseNodeView"
	ResourceDatabaseNodeView     = "databasenodeview"
	ResourceDatabaseNodeViews    = "databasenodeviews"
)

// DatabaseNodeViewSpec defines the desired state of DatabaseNodeView
type DatabaseNodeViewSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name    string       `json:"name" protobuf:"bytes,1,opt,name=name"`
	Type    string       `json:"type" protobuf:"bytes,2,opt,name=type"`
	Role    DatabaseRole `json:"role,omitempty" protobuf:"bytes,3,opt,name=role,casttype=DatabaseRole"`
	CPU     string       `json:"cpu" protobuf:"bytes,4,opt,name=cpu"`
	Memory  string       `json:"memory" protobuf:"bytes,5,opt,name=memory"`
	Storage string       `json:"storage" protobuf:"bytes,6,opt,name=storage"`
	Status  DBNodeStatus `json:"status" protobuf:"bytes,7,opt,name=status,casttype=DBNodeStatus"`
}

// DatabaseNodeViewStatus defines the observed state of DatabaseNodeView
type DatabaseNodeViewStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// DatabaseNodeView is the Schema for the databasenodeviews API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseNodeView struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DatabaseNodeViewSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DatabaseNodeViewStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DatabaseNodeViewList contains a list of DatabaseNodeView

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseNodeViewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DatabaseNodeView `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseNodeView{}, &DatabaseNodeViewList{})
}
