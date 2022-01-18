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
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type DeletionPolicy string

const (
	// DeletionPolicyDelete Deletes database pods, service, pvcs and stash backup data.
	DeletionPolicyDelete DeletionPolicy = "Delete"
	// DeletionPolicyDoNotDelete Rejects attempt to delete database using ValidationWebhook.
	DeletionPolicyDoNotDelete DeletionPolicy = "DoNotDelete"
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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgresDatabaseSpec defines the desired state of PostgresDatabase
type PostgresDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	VaultRef       kmapi.ObjectReference  `json:"vaultRef"`
	DatabaseRef    DatabaseRef            `json:"databaseRef"`
	Database       PostgresDatabaseSchema `json:"database"`
	Subjects       []rbac.Subject         `json:"subjects,omitempty"`
	Init           *Init                  `json:"init,omitempty"`
	Restore        *RestoreRef            `json:"restore,omitempty"`
	DeletionPolicy DeletionPolicy         `json:"deletionPolicy"`
	AutoApproval   bool                   `json:"autoApproval"`
	DefaultTTL     string                 `json:"defaultTTL"`
	MaxTTL         string                 `json:"maxTTL"`
}

// PostgresDatabaseStatus defines the observed state of PostgresDatabase
type PostgresDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase      SchemaDatabasePhase    `json:"phase"`
	Conditions []kmapi.Condition      `json:"conditions"`
	LoginCreds *kmapi.ObjectReference `json:"loginCreds,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgresDatabase is the Schema for the postgresdatabases API
type PostgresDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresDatabaseSpec   `json:"spec,omitempty"`
	Status PostgresDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgresDatabaseList contains a list of PostgresDatabase
type PostgresDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresDatabase `json:"items"`
}

type RestoreRef struct {
	Repository      core.ObjectReference `json:"repository,omitempty"`
	Snapshot        string               `json:"snapshot,omitempty"`
	RuntimeSettings ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`
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

type PostgresDatabaseSchema struct {
	DBName     string  `json:"dBName"`
	Tablespace *string `json:"tablespace,omitempty"`
	Params     []Param `json:"params,omitempty"`
}

type Init struct {
	Script      core.VolumeSource    `json:"script"`
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate"`
}

func init() {
	SchemeBuilder.Register(&PostgresDatabase{}, &PostgresDatabaseList{})
}
