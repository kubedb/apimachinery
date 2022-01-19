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
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindMongoDBDatabase = "MongoDBDatabase"
	ResourceMongoDBDatabase     = "mongodbdatabase"
	ResourceMongoDBDatabases    = "mongodbdatabases"
)

// MongoDBDatabaseSpec defines the desired state of MongoDBDatabase
type MongoDBDatabaseSpec struct {
	// DatabaseRef refers to a KubeDB managed database instance
	DatabaseRef kmapi.ObjectReference `json:"databaseRef"`

	// VaultRef refers to a KubeVault managed vault server
	VaultRef kmapi.ObjectReference `json:"vaultRef"`

	// DatabaseConfig defines various configuration options for a database
	DatabaseConfig MongoDBDatabaseConfiguration `json:"databaseConfig"`

	AccessPolicy VaultSecretEngineRole `json:"accessPolicy"`

	// Init contains info about the init script or snapshot info
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	// +kubebuilder:default:="Delete"
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`
}

type MongoDBDatabaseConfiguration struct {
	Name string `json:"name"`
}

// MongoDBDatabase is the Schema for the mongodbdatabases API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="DatabaseName",type="string",JSONPath=".spec.databaseSchema.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MongoDBDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBDatabaseSpec `json:"spec,omitempty"`
	Status DatabaseStatus      `json:"status,omitempty"`
}

// MongoDBDatabaseList contains a list of MongoDBDatabase

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type MongoDBDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBDatabase{}, &MongoDBDatabaseList{})
}
