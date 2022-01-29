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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
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
	Initialized bool              `json:"initialized"`
	Script      *ScriptSourceSpec `json:"script,omitempty"`

	// Snapshot contains the restore-related details
	Snapshot *SnapshotSourceSpec `json:"snapshot,omitempty"`
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

// +kubebuilder:validation:Enum=Pending;InProgress;Terminating;Current;Failed;Expired
type DatabaseSchemaPhase string

const (
	DatabaseSchemaPhasePending     DatabaseSchemaPhase = "Pending"
	DatabaseSchemaPhaseInProgress  DatabaseSchemaPhase = "InProgress"
	DatabaseSchemaPhaseTerminating DatabaseSchemaPhase = "Terminating"
	DatabaseSchemaPhaseCurrent     DatabaseSchemaPhase = "Current"
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

	DatabaseSchemaConditionTypeDoubleOptInNotPossible DatabaseSchemaConditionType = "DoubleOptInNotPossible"
	DatabaseSchemaMessageDoubleOptInNotPossible       DatabaseSchemaMessage       = "Double OptIn is not possible between the applied Schema & Database server"

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
	DatabaseSchemaMessageRepositoryFound       DatabaseSchemaMessage       = "Repository has been found"

	DatabaseSchemaConditionTypeAppBindingFound DatabaseSchemaConditionType = "AppBindingFound"
	DatabaseSchemaMessageAppBindingNotCreated  DatabaseSchemaMessage       = "AppBinding is not created yet"
	DatabaseSchemaMessageAppBindingFound       DatabaseSchemaMessage       = "AppBinding is Found"

	DatabaseSchemaConditionTypeRestoreCompleted   DatabaseSchemaConditionType = "RestoreSessionCompleted"
	DatabaseSchemaMessageRestoreSessionNotCreated DatabaseSchemaMessage       = "RestoreSession is not created yet"
	DatabaseSchemaMessageRestoreSessionRunning    DatabaseSchemaMessage       = "RestoreSession is running"
	DatabaseSchemaMessageRestoreSessionSucceed    DatabaseSchemaMessage       = "RestoreSession is succeeded"
	DatabaseSchemaMessageRestoreSessionFailed     DatabaseSchemaMessage       = "RestoreSession is failed"
)

func GetFinalizerForSchema() string {
	return SchemeGroupVersion.Group
}

func GetSchemaDoubleOptInLabelKey() string {
	return SchemeGroupVersion.Group + "/doubleoptin"
}
func GetSchemaDoubleOptInLabelValue() string {
	return "enabled"
}

func GetPhase(obj Interface) DatabaseSchemaPhase {
	conditions := obj.GetStatus().Conditions

	if !obj.GetDeletionTimestamp().IsZero() {
		return DatabaseSchemaPhaseTerminating
	}
	if CheckIfSecretExpired(conditions) {
		return DatabaseSchemaPhaseExpired
	}
	if kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSchemaNameConflict)) {
		return DatabaseSchemaPhaseFailed
	}

	// If Database or vault is not in ready state, Phase is 'Pending'
	if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeDBServerReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeVaultReady)) {
		return DatabaseSchemaPhasePending
	}

	if kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeDoubleOptInNotPossible)) {
		return DatabaseSchemaPhaseFailed
	}

	// If SecretEngine or Role is not in ready state, Phase is 'InProgress'
	if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSecretEngineReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRoleReady)) ||
		!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeSecretAccessRequestReady)) {
		return DatabaseSchemaPhaseInProgress
	}
	// we are here means, SecretAccessRequest is approved and not expired. Now handle Init-Restore cases.

	if kmapi.HasCondition(conditions, string(DatabaseSchemaConditionTypeRepositoryFound)) {
		//  ----------------------------- Restore case -----------------------------
		if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRepositoryFound)) ||
			!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeAppBindingFound)) ||
			!kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeRestoreCompleted)) {
			return DatabaseSchemaPhaseInProgress
		}
		if CheckIfRestoreFailed(conditions) {
			return DatabaseSchemaPhaseFailed
		} else {
			return DatabaseSchemaPhaseCurrent
		}
	} else if kmapi.HasCondition(conditions, string(DatabaseSchemaConditionTypeInitScriptCompleted)) {
		//  ----------------------------- Init case -----------------------------
		if !kmapi.IsConditionTrue(conditions, string(DatabaseSchemaConditionTypeInitScriptCompleted)) {
			return DatabaseSchemaPhaseInProgress
		}
		if CheckIfInitScriptFailed(conditions) {
			return DatabaseSchemaPhaseFailed
		} else {
			return DatabaseSchemaPhaseCurrent
		}
	}
	return DatabaseSchemaPhaseCurrent
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

/*
Double OptIn related helpers start here. CheckIfDoubleOptInPossible() is the intended function to be called from operator
*/

func CheckIfDoubleOptInPossible(schemaMeta metav1.ObjectMeta, namespaceMeta metav1.ObjectMeta, consumers *dbapi.AllowedConsumers) (bool, error) {
	if consumers == nil {
		return false, nil
	}
	matchNamespace, err := IsInAllowedNamespaces(schemaMeta, namespaceMeta, consumers)
	if err != nil {
		return false, err
	}
	matchLabels, err := IsMatchByLabels(schemaMeta, consumers)
	if err != nil {
		return false, err
	}
	return matchNamespace && matchLabels, nil
}

func IsInAllowedNamespaces(schemaMeta metav1.ObjectMeta, namespaceMeta metav1.ObjectMeta, consumers *dbapi.AllowedConsumers) (bool, error) {
	if consumers.Namespaces == nil || consumers.Namespaces.From == nil {
		return false, nil
	}

	if *consumers.Namespaces.From == dbapi.NamespacesFromAll {
		return true, nil
	}
	if *consumers.Namespaces.From == dbapi.NamespacesFromSame {
		return schemaMeta.Namespace == namespaceMeta.Name, nil
	}
	if *consumers.Namespaces.From == dbapi.NamespacesFromSelector {
		if consumers.Namespaces.Selector == nil {
			// this says, Select namespace from the Selector, but the Namespace.Selector field is nil. So, no way to select namespace here.
			return false, nil
		}
		ret, err := selectorMatches(consumers.Namespaces.Selector, namespaceMeta.GetLabels())
		if err != nil {
			return false, err
		}
		return ret, nil
	}
	return false, nil
}

func IsMatchByLabels(schemaMeta metav1.ObjectMeta, consumers *dbapi.AllowedConsumers) (bool, error) {
	if consumers.Selector != nil {
		ret, err := selectorMatches(consumers.Selector, schemaMeta.Labels)
		if err != nil {
			return false, err
		}
		return ret, nil
	}
	// if Selector is not given, all the Schemas are allowed of the selected namespace
	return true, nil
}

func selectorMatches(ls *metav1.LabelSelector, srcLabels map[string]string) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		klog.Infoln("invalid selector: ", ls)
		return false, err
	}
	return selector.Matches(labels.Set(srcLabels)), nil
}
