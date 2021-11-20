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
	ResourceKindDatabasesIndex = "DatabasesIndex"
	ResourceDatabasesIndex     = "databasesindex"
	ResourceDatabasesIndices   = "databasesindices"
)

// DatabasesIndexSpec defines the desired state of DatabasesIndex
type DatabasesIndexSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Databases []DBDisplay `json:"databases" protobuf:"bytes,1,rep,name=databases"`
}

type DBDisplay struct {
	Title string `json:"title" protobuf:"bytes,1,opt,name=title"`
	ID    string `json:"id" protobuf:"bytes,2,opt,name=id"`
	URL   string `json:"url" protobuf:"bytes,3,opt,name=url"`
	Icon  string `json:"icon" protobuf:"bytes,4,opt,name=icon"`
}

// DatabasesIndexStatus defines the observed state of DatabasesIndex
type DatabasesIndexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// DatabasesIndex is the Schema for the databasesindices API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabasesIndex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DatabasesIndexSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DatabasesIndexStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DatabasesIndexList contains a list of DatabasesIndex

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabasesIndexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DatabasesIndex `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&DatabasesIndex{}, &DatabasesIndexList{})
}
