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
	ResourceKindRecentDatabase = "RecentDatabase"
	ResourceRecentDatabase     = "recentdatabase"
	ResourceRecentDatabases    = "recentdatabases"
)

// RecentDatabaseSpec defines the desired state of RecentDatabase
type RecentDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name        string              `json:"name" protobuf:"bytes,1,opt,name=name"`
	Type        string              `json:"type" protobuf:"bytes,2,opt,name=type"`
	ClusterID   string              `json:"clusterID" protobuf:"bytes,3,opt,name=clusterID"`
	ClusterName string              `json:"clusterName" protobuf:"bytes,4,opt,name=clusterName"`
	Environment DBEnvironment       `json:"environment" protobuf:"bytes,5,opt,name=environment,casttype=DBEnvironment"`
	Status      string              `json:"status" protobuf:"bytes,6,opt,name=status"`
	Age         string              `json:"age" protobuf:"bytes,7,opt,name=age"`
	Resources   corev1.ResourceList `json:"resources" protobuf:"bytes,8,rep,name=resources,casttype=k8s.io/api/core/v1.ResourceList,castkey=k8s.io/api/core/v1.ResourceName"`
}

// RecentDatabaseStatus defines the observed state of RecentDatabase
type RecentDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// RecentDatabase is the Schema for the recentdatabases API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RecentDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   RecentDatabaseSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status RecentDatabaseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// RecentDatabaseList contains a list of RecentDatabase

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RecentDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []RecentDatabase `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&RecentDatabase{}, &RecentDatabaseList{})
}
