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
	Initialized bool                `json:"initialized" protobuf:"varint,1,opt,name=initialized"`
	Script      *ScriptSourceSpec   `json:"script,omitempty" protobuf:"bytes,2,opt,name=script"`
	Snapshot    *SnapshotSourceSpec `json:"snapshot,omitempty" protobuf:"bytes,3,opt,name=snapshot"`
}

type ScriptSourceSpec struct {
	ScriptPath        string `json:"scriptPath,omitempty" protobuf:"bytes,1,opt,name=scriptPath"`
	core.VolumeSource `json:",inline,omitempty" protobuf:"bytes,2,opt,name=volumeSource"`
	// This will take some database related config from the user
	PodTemplate *core.PodTemplateSpec `json:"podTemplate,omitempty" protobuf:"bytes,3,opt,name=podTemplate"`
}

type SnapshotSourceSpec struct {
	Repository kmapi.TypedObjectReference `json:"repository,omitempty" protobuf:"bytes,1,opt,name=repository"`
	// +kubebuilder:default:="latest"
	SnapshotID string `json:"snapshotID,omitempty" protobuf:"bytes,2,opt,name=snapshotID"`
}

// DatabaseStatus defines the observed state of schema api types
type DatabaseStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabaseSchemaPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=DatabaseSchemaPhase"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// Database authentication secret
	// +optional
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty" protobuf:"bytes,4,opt,name=authSecret"`
}

type DatabaseSchemaPhase string

const (
	Success     DatabaseSchemaPhase = "Success"
	Running     DatabaseSchemaPhase = "Running"
	Waiting     DatabaseSchemaPhase = "Waiting"
	Ignored     DatabaseSchemaPhase = "Ignored"
	Failed      DatabaseSchemaPhase = "Failed"
	Expired     DatabaseSchemaPhase = "Expired"
	Terminating DatabaseSchemaPhase = "Terminating"
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

type MySQLDatabaseCondition string
type MySQLDatabaseVerbs string

const (
	AddCondition    MySQLDatabaseVerbs = "AddCondition"
	RemoveCondition MySQLDatabaseVerbs = "RemoveCondition"

	MySQLRoleCreated           MySQLDatabaseCondition = "MySQLRoleCreated"
	VaultSecretEngineCreated   MySQLDatabaseCondition = "VaultSecretEngineCreated"
	SecretAccessRequestCreated MySQLDatabaseCondition = "SecretAccessRequestCreated"

	MySQLNotReady               MySQLDatabaseCondition = "MySQLNotReady"
	VaultNotReady               MySQLDatabaseCondition = "VaultNotReady"
	SecretAccessRequestApproved MySQLDatabaseCondition = "SecretAccessRequestApproved"
	SecretAccessRequestDenied   MySQLDatabaseCondition = "SecretAccessRequestDenied"
	SecretAccessRequestExpired  MySQLDatabaseCondition = "SecretAccessRequestExpired"

	SchemaIgnored          MySQLDatabaseCondition = "SchemaIgnored"
	DatabaseCreated        MySQLDatabaseCondition = "DatabaseCreated"
	DatabaseDeleted        MySQLDatabaseCondition = "DatabaseDeleted"
	DatabaseAltered        MySQLDatabaseCondition = "DatabaseAltered"
	ScriptApplied          MySQLDatabaseCondition = "ScriptApplied"
	RestoredFromRepository MySQLDatabaseCondition = "RestoredFromRepository"
	FailedInitializing     MySQLDatabaseCondition = "FailedInitializing"
	FailedRestoring        MySQLDatabaseCondition = "FailedRestoring"
	TerminationHalted      MySQLDatabaseCondition = "TerminationHalted"
	UserDisconnected       MySQLDatabaseCondition = "UserDisconnected"
)

//todo do phase works, update phase of success schema aftel altering databases, check more cases
func GetPhase(obj Interface) DatabaseSchemaPhase {
	conditions := obj.GetStatus().Conditions

	if !obj.GetDeletionTimestamp().IsZero() {
		return Terminating
	}
	if kmapi.IsConditionTrue(conditions, string(SecretAccessRequestExpired)) {
		return Expired
	}
	if obj.GetStatus().Phase == Success {
		return Success
	}
	if kmapi.IsConditionTrue(conditions, string(SchemaIgnored)) {
		return Failed
	}
	if kmapi.IsConditionTrue(conditions, string(MySQLNotReady)) {
		return Waiting
	}
	if kmapi.IsConditionTrue(conditions, string(VaultNotReady)) {
		return Waiting
	}

	if kmapi.IsConditionTrue(conditions, string(SecretAccessRequestCreated)) && !kmapi.IsConditionTrue(conditions, string(SecretAccessRequestApproved)) {
		if kmapi.IsConditionTrue(conditions, string(SecretAccessRequestDenied)) {
			return Failed
		}
		return Waiting
	}
	if kmapi.IsConditionTrue(conditions, string(DatabaseCreated)) {
		if obj.GetInit() != nil {
			if kmapi.IsConditionTrue(conditions, string(FailedInitializing)) {
				return Failed
			} else if !kmapi.IsConditionTrue(conditions, string(ScriptApplied)) {
				return Running
			}
		}
		if obj.GetInit().Snapshot != nil {
			if kmapi.IsConditionTrue(conditions, string(FailedRestoring)) {
				return Failed
			} else if !kmapi.IsConditionTrue(conditions, string(RestoredFromRepository)) {
				return Running
			}
		}
		return Success
	}
	return Waiting
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
	Subjects []rbac.Subject `json:"subjects" protobuf:"bytes,1,rep,name=subjects"`
	// +optional
	DefaultTTL string `json:"defaultTTL,omitempty" protobuf:"bytes,2,opt,name=defaultTTL"`
	// +optional
	MaxTTL string `json:"maxTTL,omitempty" protobuf:"bytes,3,opt,name=maxTTL"`
}
