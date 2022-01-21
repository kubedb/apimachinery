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
)

type Interface interface {
	metav1.Object
	GetInit() *InitSpec
	GetStatus() DatabaseStatus
}

type InitSpec struct {
	// Initialized indicates that this database has been initialized.
	// This will be set by the operator to ensure
	// that database is not mistakenly reset when recovered using disaster recovery tools.
	Initialized bool                `json:"initialized"`
	Script      *ScriptSourceSpec   `json:"script,omitempty"`
	Snapshot    *SnapshotSourceSpec `json:"snapshot,omitempty"`
}

type ScriptSourceSpec struct {
	ScriptPath        string `json:"scriptPath,omitempty"`
	core.VolumeSource `json:",inline,omitempty"`
	// This will take some database related config from the user
	PodTemplate *core.PodTemplateSpec `json:"podTemplate,omitempty"`
}

type SnapshotSourceSpec struct {
	Repository kmapi.TypedObjectReference `json:"repository,omitempty"`
	// +kubebuilder:default:="latest"
	SnapshotID string `json:"snapshotID,omitempty"`
}

// DatabaseStatus defines the observed state of schema api types
type DatabaseStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabaseSchemaPhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
	// Database authentication secret
	// +optional
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty"`
}

// +kubebuilder:validation:Enum=Delete;DoNotDelete
type DeletionPolicy string

const (
	// DeletionPolicyDelete Deletes database pods, service, pvcs and stash backup data.
	DeletionPolicyDelete DeletionPolicy = "Delete"
	// DeletionPolicyDoNotDelete Rejects attempt to delete database using ValidationWebhook.
	DeletionPolicyDoNotDelete DeletionPolicy = "DoNotDelete"
)

type VaultSecretEngineRole struct {
	Subjects []rbac.Subject `json:"subjects"`
	// +optional
	DefaultTTL string `json:"defaultTTL,omitempty"`
	// +optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Progressing;Terminating;Successful;Failed;Expired
type DatabaseSchemaPhase string

const (
	DatabaseSchemaPhasePending     DatabaseSchemaPhase = "Pending"
	DatabaseSchemaPhaseProgressing DatabaseSchemaPhase = "Progressing"
	DatabaseSchemaPhaseTerminating DatabaseSchemaPhase = "Terminating"
	DatabaseSchemaPhaseSuccessful  DatabaseSchemaPhase = "Successful"
	DatabaseSchemaPhaseFailed      DatabaseSchemaPhase = "Failed"
	DatabaseSchemaPhaseExpired     DatabaseSchemaPhase = "Expired"
)

type DatabaseSchemaConditionType string
type DatabaseSchemaMessage string

const (
	DatabaseSchemaConditionTypeDBServerReady  DatabaseSchemaConditionType = "DatabaseServerReady"
	DatabaseSchemaMessageDBServerNotCreated   DatabaseSchemaMessage       = "Database Server is not created yet"
	DatabaseSchemaMessageDBServerProvisioning DatabaseSchemaMessage       = "Database Server is provisioning"
	DatabaseSchemaMessageDBServerReady        DatabaseSchemaMessage       = "Database Server is Ready"

	DatabaseSchemaConditionTypeVaultReady  DatabaseSchemaConditionType = "VaultReady"
	DatabaseSchemaMessageVaultNotCreated   DatabaseSchemaMessage       = "VaultServer is not created yet"
	DatabaseSchemaMessageVaultProvisioning DatabaseSchemaMessage       = "VaultServer is provisioning"
	DatabaseSchemaMessageVaultReady        DatabaseSchemaMessage       = "VaultServer is Ready"

	DatabaseSchemaConditionTypeSecretEngineReady DatabaseSchemaConditionType = "SecretEngineReady"
	DatabaseSchemaMessageSecretEngineNotCreated  DatabaseSchemaMessage       = "SecretEngine is not created yet"
	DatabaseSchemaMessageSecretEngineCreating    DatabaseSchemaMessage       = "SecretEngine is being creating"
	DatabaseSchemaMessageSecretEngineSuccess     DatabaseSchemaMessage       = "SecretEngine phase is success"

	DatabaseSchemaConditionTypeRoleReady        DatabaseSchemaConditionType = "RoleReady"
	DatabaseSchemaMessageDatabaseRoleNotCreated DatabaseSchemaMessage       = "Database Role is not created yet"
	DatabaseSchemaMessageDatabaseRoleCreating   DatabaseSchemaMessage       = "Database Role is being creating"
	DatabaseSchemaMessageDatabaseRoleSuccess    DatabaseSchemaMessage       = "Database Role is success"

	DatabaseSchemaConditionTypeSecretAccessRequestReady DatabaseSchemaConditionType = "SecretAccessRequestReady"
	DatabaseSchemaMessageSecretAccessRequestNotCreated  DatabaseSchemaMessage       = "SecretAccessRequest is not created yet"
	DatabaseSchemaMessageSecretAccessRequestWaiting     DatabaseSchemaMessage       = "SecretAccessRequest is waiting for approval"
	DatabaseSchemaMessageSecretAccessRequestApproved    DatabaseSchemaMessage       = "SecretAccessRequest has been approved"
	DatabaseSchemaMessageSecretAccessRequestExpired     DatabaseSchemaMessage       = "SecretAccessRequest has been expired"

	DatabaseSchemaConditionTypeSchemaNameConflict DatabaseSchemaConditionType = "SchemaNameConflict"
	DatabaseSchemaMessageSchemaNameConflicted     DatabaseSchemaMessage       = "Schema name is conflicted"

	DatabaseSchemaConditionTypeInitScriptCompleted DatabaseSchemaConditionType = "InitScriptCompleted"
	DatabaseSchemaMessageInitScriptNotApplied      DatabaseSchemaMessage       = "InitScript is not applied yet"
	DatabaseSchemaMessageInitScriptRunning         DatabaseSchemaMessage       = "InitScript is running"
	DatabaseSchemaMessageInitScriptCompleted       DatabaseSchemaMessage       = "InitScript is completed"
	DatabaseSchemaMessageInitScriptSucceeded       DatabaseSchemaMessage       = "InitScript is succeeded"
	DatabaseSchemaMessageInitScriptFailed          DatabaseSchemaMessage       = "InitScript is failed"

	DatabaseSchemaConditionTypeRepositoryFound DatabaseSchemaConditionType = "RepositoryFound"
	DatabaseSchemaMessageRepositoryNotCreated  DatabaseSchemaMessage       = "Repository is not created yet"
	DatabaseSchemaMessageRepositoryFound       DatabaseSchemaMessage       = "Repository has been successfully copied"

	DatabaseSchemaConditionTypeAppBindingFound DatabaseSchemaConditionType = "AppBindingFound"
	DatabaseSchemaMessageAppBindingNotCreated  DatabaseSchemaMessage       = "AppBinding is not created yet"
	DatabaseSchemaMessageAppBindingFound       DatabaseSchemaMessage       = "AppBinding is Found"

	DatabaseSchemaConditionTypeRestoreCompleted   DatabaseSchemaConditionType = "RestoreSessionCompleted"
	DatabaseSchemaMessageRestoreSessionNotCreated DatabaseSchemaMessage       = "RestoreSession is not created yet"
	DatabaseSchemaMessageRestoreSessionRunning    DatabaseSchemaMessage       = "RestoreSession is running"
	DatabaseSchemaMessageRestoreSessionSucceed    DatabaseSchemaMessage       = "RestoreSession is succeeded"
	DatabaseSchemaMessageRestoreSessionFailed     DatabaseSchemaMessage       = "RestoreSession is failed"
)

func GetPhase(obj Interface) DatabaseSchemaPhase {
	conditions := obj.GetStatus().Conditions

	if !obj.GetDeletionTimestamp().IsZero() {
		return DatabaseSchemaPhaseTerminating
	}
	if CheckIfSecretExpired(conditions) {
		return DatabaseSchemaPhaseExpired
	}
	if obj.GetStatus().Phase == DatabaseSchemaPhaseSuccessful {
		return DatabaseSchemaPhaseSuccessful
	}
	if kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSchemaNameConflict)) {
		return DatabaseSchemaPhaseFailed
	}

	// If Database or vault is not in ready state, Phase is 'Pending'
	if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeDBServerReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeVaultReady)) {
		return DatabaseSchemaPhasePending
	}

	// If SecretEngine or Role is not in ready state, Phase is 'Progressing'
	if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSecretEngineReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRoleReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSecretAccessRequestReady)) {
		return DatabaseSchemaPhaseProgressing
	}
	// we are here means, SecretAccessRequest is approved and not expired. Now handle Init-Restore cases.

	if kmapi.HasCondition(conditions, string(DatabaseSchemaConditionTypeRepositoryFound)) {
		//  ----------------------------- Restore case -----------------------------
		if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRepositoryFound)) ||
			!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeAppBindingFound)) ||
			!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRestoreCompleted)) {
			return DatabaseSchemaPhaseProgressing
		}
		if CheckIfRestoreFailed(conditions) {
			return DatabaseSchemaPhaseFailed
		} else {
			return DatabaseSchemaPhaseSuccessful
		}
	} else if kmapi.HasCondition(conditions, string(DatabaseSchemaConditionTypeInitScriptCompleted)) {
		//  ----------------------------- Init case -----------------------------
		if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeInitScriptCompleted)) {
			return DatabaseSchemaPhaseProgressing
		}
		if CheckIfInitScriptFailed(conditions) {
			return DatabaseSchemaPhaseFailed
		} else {
			return DatabaseSchemaPhaseSuccessful
		}
	}
	return DatabaseSchemaPhaseSuccessful
}

func CheckIfInitScriptFailed(conditions []kmapi.Condition) bool {
	_, cond := kmapi.GetCondition(conditions, string(DatabaseSchemaConditionTypeInitScriptCompleted))
	return cond.Message == string(DatabaseSchemaMessageInitScriptFailed)
}

func CheckIfRestoreFailed(conditions []kmapi.Condition) bool {
	_, cond := kmapi.GetCondition(conditions, string(DatabaseSchemaConditionTypeRestoreCompleted))
	return cond.Message == string(DatabaseSchemaMessageRestoreSessionFailed)
}

func CheckIfSecretExpired(conditions []kmapi.Condition) bool {
	i, cond := kmapi.GetCondition(conditions, string(DatabaseSchemaConditionTypeSecretAccessRequestReady))
	if i == -1 {
		return false
	}
	return cond.Message == string(DatabaseSchemaMessageSecretAccessRequestExpired)
}
