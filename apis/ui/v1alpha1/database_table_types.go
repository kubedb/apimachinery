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
	ResourceKindDatabaseTable = "DatabaseTable"
	ResourceDatabaseTable     = "databasetable"
	ResourceDatabaseTables    = "databasetables"
)

// DatabaseTableSpec defines the desired state of DatabaseTable
type DatabaseTableSpec struct {
	Columns []DatabaseTableColumn `json:"columns" protobuf:"bytes,1,rep,name=columns"`
	Rows    []TableRow            `json:"rows" protobuf:"bytes,2,rep,name=rows"`
}

type TableRow struct {
	Data []string `json:"data" protobuf:"bytes,1,rep,name=data"`
}

// DatabaseTableStatus defines the observed state of DatabaseTable
type DatabaseTableStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// DatabaseTable is the Schema for the databasetables API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DatabaseTableSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DatabaseTableStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DatabaseTableList contains a list of DatabaseTable

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DatabaseTable `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseTable{}, &DatabaseTableList{})
}

type DatabaseTableColumn struct {
	Title      string               `json:"title" protobuf:"bytes,1,opt,name=title"`
	Type       string               `json:"type" protobuf:"bytes,2,opt,name=type"`
	Properties []DatabaseProperties `json:"properties,omitempty" protobuf:"bytes,3,rep,name=properties"`
}

type DatabaseProperties struct {
	Value string `json:"value" protobuf:"bytes,1,opt,name=value"`
	Class string `json:"class" protobuf:"bytes,2,opt,name=class"`
}
