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
	ResourceKindPostgresDatabase = "PostgresDatabase"
	ResourcePostgresDatabase     = "postgresdatabase"
	ResourcePostgresDatabases    = "postgresdatabases"
)

// PostgresDatabaseSpec defines the desired state of PostgresDatabase
type PostgresDatabaseSpec struct {
	// DatabaseRef refers to a KubeDB managed database instance
	DatabaseRef kmapi.ObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`

	// VaultRef refers to a KubeVault managed vault server
	VaultRef kmapi.ObjectReference `json:"vaultRef" protobuf:"bytes,2,opt,name=vaultRef"`

	// DatabaseConfig defines various configuration options for a database
	DatabaseConfig PostgresDatabaseConfiguration `json:"databaseConfig" protobuf:"bytes,3,opt,name=databaseConfig"`

	AccessPolicy VaultSecretEngineRole `json:"accessPolicy" protobuf:"bytes,4,opt,name=accessPolicy"`

	// Init contains info about the init script or snapshot info
	// +optional
	Init *InitSpec `json:"init,omitempty" protobuf:"bytes,5,opt,name=init"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	// +kubebuilder:default:="Delete"
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty" protobuf:"bytes,6,opt,name=deletionPolicy,casttype=DeletionPolicy"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PostgresDatabase is the Schema for the postgresdatabases API
type PostgresDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   PostgresDatabaseSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DatabaseStatus       `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +kubebuilder:object:root=true

// PostgresDatabaseList contains a list of PostgresDatabase
type PostgresDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []PostgresDatabase `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type Param struct {
	ConfigParameter string  `json:"configParameter" protobuf:"bytes,1,opt,name=configParameter"`
	Value           *string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type PostgresDatabaseConfiguration struct {
	DBName     string  `json:"dBName" protobuf:"bytes,1,opt,name=dBName"`
	Tablespace *string `json:"tablespace,omitempty" protobuf:"bytes,2,opt,name=tablespace"`
	Params     []Param `json:"params,omitempty" protobuf:"bytes,3,rep,name=params"`
}

func init() {
	SchemeBuilder.Register(&PostgresDatabase{}, &PostgresDatabaseList{})
}
