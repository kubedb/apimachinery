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
	ResourceKindDatabaseSummary = "DatabaseSummary"
	ResourceDatabaseSummary     = "databasesummary"
	ResourceDatabaseSummaries   = "databasesummaries"
)

// DatabaseSummarySpec defines the desired state of DatabaseSummary
type DatabaseSummarySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	NumberOfClusters int64               `json:"numberOfClusters" protobuf:"varint,1,opt,name=numberOfClusters"`
	TotalDatabases   int64               `json:"totalDatabases" protobuf:"varint,2,opt,name=totalDatabases"`
	ProdDatabases    int64               `json:"prodDatabases" protobuf:"varint,3,opt,name=prodDatabases"`
	DevDatabases     int64               `json:"devDatabases" protobuf:"varint,4,opt,name=devDatabases"`
	QADatabases      int64               `json:"qaDatabases" protobuf:"varint,5,opt,name=qaDatabases"`
	Resources        corev1.ResourceList `json:"resources" protobuf:"bytes,6,rep,name=resources,casttype=k8s.io/api/core/v1.ResourceList,castkey=k8s.io/api/core/v1.ResourceName"`
	TotalCost        string              `json:"totalCost" protobuf:"bytes,7,opt,name=totalCost"`
}

// DatabaseSummaryStatus defines the observed state of DatabaseSummary
type DatabaseSummaryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// DatabaseSummary is the Schema for the databasesummaries API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseSummary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DatabaseSummarySpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DatabaseSummaryStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DatabaseSummaryList contains a list of DatabaseSummary

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DatabaseSummaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DatabaseSummary `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseSummary{}, &DatabaseSummaryList{})
}
