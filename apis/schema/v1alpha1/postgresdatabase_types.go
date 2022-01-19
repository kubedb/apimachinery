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

type SchemaDatabasePhase string

const (
	// used for SchemaDatabases that are currently running
	SchemaDatabasePhaseRunning SchemaDatabasePhase = "Running"
	// used for SchemaDatabases that are Successfull
	SchemaDatabasePhaseSuccessfull SchemaDatabasePhase = "Succeeded"
	// used for SchemaDatabases that are Failed
	SchemaDatabasePhaseFailed SchemaDatabasePhase = "Failed"
)

type SchemaDatabaseCondition string

const (
	SchemaDatabaseConditionDBReady                  SchemaDatabaseCondition = "DatabaseReady"
	SchemaDatabaseConditionVaultReady               SchemaDatabaseCondition = "VaultReady"
	SchemaDatabaseConditionSecretEngineReady        SchemaDatabaseCondition = "SecretEngineReady"
	SchemaDatabaseConditionMongoDBRoleReady         SchemaDatabaseCondition = "MongoRoleDBReady"
	SchemaDatabaseConditionPostgresRoleReady        SchemaDatabaseCondition = "PostgresRoleReady"
	SchemaDatabaseConditionMysqlRoleReady           SchemaDatabaseCondition = "MysqlRoleReady"
	SchemaDatabaseConditionMariaDBRoleReady         SchemaDatabaseCondition = "MariaDBRoleReady"
	SchemaDatabaseConditionSecretAccessRequestReady SchemaDatabaseCondition = "SecretAccessRequestReady"
	SchemaDatabaseConditionJobCompleted             SchemaDatabaseCondition = "JobCompleted"
	SchemaDatabaseConditionRepositoryReady          SchemaDatabaseCondition = "RepositoryReady"
	SchemaDatabaseConditionRestoreSecretReady       SchemaDatabaseCondition = "RestoreSecretReady"
	SchemaDatabaseConditionAppBindingReady          SchemaDatabaseCondition = "AppBindingReady"
	SchemaDatabaseConditionRestoreSessionReady      SchemaDatabaseCondition = "RestoreSessionReady"
)

type SchemaDatabaseReason string

const (
	SchemaDatabaseReasonDBReady                     SchemaDatabaseReason = "CheckDBIsReady"
	SchemaDatabaseReasonDBNotReady                  SchemaDatabaseReason = "CheckDBIsNotReady"
	SchemaDatabaseReasonVaultReady                  SchemaDatabaseReason = "CheckVaultIsReady"
	SchemaDatabaseReasonVaultNotReady               SchemaDatabaseReason = "CheckVaultIsNotReady"
	SchemaDatabaseReasonSecretEngineReady           SchemaDatabaseReason = "CheckSecretEngineIsReady"
	SchemaDatabaseReasonSecretEngineNotReady        SchemaDatabaseReason = "CheckSecretEngineIsNotReady"
	SchemaDatabaseReasonPostgresRoleReady           SchemaDatabaseReason = "CheckPostgresRoleIsReady"
	SchemaDatabaseReasonPostgresRoleNotReady        SchemaDatabaseReason = "CheckPostgresRoleIsNotReady"
	SchemaDatabaseReasonSecretAccessRequestReady    SchemaDatabaseReason = "CheckSecretAccessRequestIsReady"
	SchemaDatabaseReasonSecretAccessRequestNotReady SchemaDatabaseReason = "CheckSecretAccessRequestIsNotReady"
	SchemaDatabaseReasonJobNotCompleted             SchemaDatabaseReason = "CheckJobIsNotCompleted"
	SchemaDatabaseReasonJobCompleted                SchemaDatabaseReason = "CheckJobIsCompleted"
	SchemaDatabaseReasonRepositoryReady             SchemaDatabaseReason = "CheckRepositoryIsReady"
	SchemaDatabaseReasonRepositoryNotReady          SchemaDatabaseReason = "CheckRepositoryIsNotReady"
	SchemaDatabaseReasonRestoreSecretReady          SchemaDatabaseReason = "CheckRestoreSecretIsReady"
	SchemaDatabaseReasonRestoreSecretNotReady       SchemaDatabaseReason = "CheckRestoreSecretIsNotReady"
	SchemaDatabaseReasonAppBindingReady             SchemaDatabaseReason = "CheckAppBindingIsReady"
	SchemaDatabaseReasonAppBindingNotReady          SchemaDatabaseReason = "CheckAppBindingIsNotReady"
	SchemaDatabaseReasonRestoreSessionReady         SchemaDatabaseReason = "CheckRestoreSessionIsReady"
	SchemaDatabaseReasonRestoreSessionNotReady      SchemaDatabaseReason = "CheckRestoreSessionIsNotReady"
)

// PostgresDatabaseSpec defines the desired state of PostgresDatabase
type PostgresDatabaseSpec struct {
	// DatabaseRef refers to a KubeDB managed database instance
	DatabaseRef kmapi.ObjectReference `json:"databaseRef"`

	// VaultRef refers to a KubeVault managed vault server
	VaultRef kmapi.ObjectReference `json:"vaultRef"`

	// DatabaseConfig defines various configuration options for a database
	DatabaseConfig PostgresDatabaseConfiguration `json:"databaseConfig"`

	AccessPolicy VaultSecretEngineRole `json:"accessPolicy"`

	// Init contains info about the init script or snapshot info
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	// +kubebuilder:default:="Delete"
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PostgresDatabase is the Schema for the postgresdatabases API
type PostgresDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresDatabaseSpec `json:"spec,omitempty"`
	Status DatabaseStatus       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgresDatabaseList contains a list of PostgresDatabase
type PostgresDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresDatabase `json:"items"`
}

type DatabaseRef struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	PluginName string `json:"pluginName"`
}

type Param struct {
	ConfigParameter string  `json:"configParameter"`
	Value           *string `json:"value"`
}

type PostgresDatabaseConfiguration struct {
	DBName     string  `json:"dBName"`
	Tablespace *string `json:"tablespace,omitempty"`
	Params     []Param `json:"params,omitempty"`
}

func init() {
	SchemeBuilder.Register(&PostgresDatabase{}, &PostgresDatabaseList{})
}
