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

//go:generate go-enum --mustparse --names --values
package v1alpha1

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeYugabyteDBOpsRequest     = "ybops"
	ResourceKindYugabyteDBOpsRequest     = "YugabyteDBOpsRequest"
	ResourceSingularYugabyteDBOpsRequest = "yugabytedbopsrequest"
	ResourcePluralYugabyteDBOpsRequest   = "yugabytedbopsrequests"
)

// YugabyteDBOpsRequest defines a YugabyteDB DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=yugabytedbopsrequests,singular=yugabytedbopsrequest,shortName=ybops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type YugabyteDBOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              YugabyteDBOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus         `json:"status,omitempty"`
}

// YugabyteDBOpsRequestSpec is the spec for YugabyteDBOpsRequest
type YugabyteDBOpsRequestSpec struct {
	// Specifies the YugabyteDB reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type YugabyteDBOpsRequestType `json:"type"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *YugabyteDBTLSSpec `json:"tls,omitempty"`
	// ConnectionManager changes the YSQL Connection Manager of the database. It is only valid in a Reconfigure
	// request, next to or instead of spec.configuration. The pods restart one at a time.
	ConnectionManager *dbapi.YugabyteDBConnectionManager `json:"connectionManager,omitempty"`
	// Specifies information necessary for configuring authSecret of the database
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Specifies information necessary for custom configuration of YugabyteDB
	Configuration *ReconfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *YugabyteDBHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *YugabyteDBVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *YugabyteDBVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for upgrading YugabyteDB
	UpdateVersion *YugabyteDBUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for migrating the data volumes to another StorageClass. It applies to the
	// masters, the tservers and the read replicas. The pods are migrated one at a time, which moves all the data of a
	// tserver for every pod.
	Migration *StorageMigrationSpec `json:"migration,omitempty"`
	// Specifies information necessary for finalizing an upgrade that was started with finalize: Manual.
	// The upgrade is finalized and can not be rolled back after it.
	FinalizeUpgrade *FinalizeUpgradeSpec `json:"finalizeUpgrade,omitempty"`
	// Specifies information necessary for rolling an upgrade back to the previous version.
	// It is only possible while the upgrade is not finalized.
	RollbackUpgrade *RollbackUpgradeSpec `json:"rollbackUpgrade,omitempty"`
	// Rolling tunes the operations that restart the pods one at a time.
	Rolling *YugabyteDBRollingSpec `json:"rolling,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// YugabyteDBRollingSpec tunes a rolling restart. Pods restart one at a time, masters first.
type YugabyteDBRollingSpec struct {
	// MasterWait is the pause after a master is back and caught up. Defaults to 60s.
	// +optional
	MasterWait *metav1.Duration `json:"masterWait,omitempty"`
	// TServerWait is the pause after a tserver is back and caught up. Defaults to 60s.
	// +optional
	TServerWait *metav1.Duration `json:"tserverWait,omitempty"`
	// LeaderDrainTimeout is how long to wait for the tablet leaders to leave a tserver before it is stopped.
	// The operation continues after the timeout and records a LeaderDrainTimedOut condition. Defaults to 120s.
	// +optional
	LeaderDrainTimeout *metav1.Duration `json:"leaderDrainTimeout,omitempty"`
}

// YugabyteDBVolumeExpansionSpec is the spec for YugabyteDB volume expansion
type YugabyteDBVolumeExpansionSpec struct {
	Mode VolumeExpansionMode `json:"mode"`
	// volume specification for the masters
	Master *resource.Quantity `json:"master,omitempty"`
	// volume specification for the tservers
	TServer *resource.Quantity `json:"tserver,omitempty"`
	// volume specification for the read replicas
	ReadReplica *resource.Quantity `json:"readReplica,omitempty"`
}

// YugabyteDBUpdateVersionSpec contains the update version information of a YugabyteDB universe
type YugabyteDBUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
	// SoakDuration is how long the universe is watched, before the upgrade is finalized, after every pod runs the new version.
	// The upgrade can still be rolled back until it is finalized. Defaults to 10m.
	// +optional
	SoakDuration *metav1.Duration `json:"soakDuration,omitempty"`
	// Finalize selects what happens after the soak. Auto finalizes the upgrade. Manual stops with the
	// UpgradePendingFinalize condition on the database, and a FinalizeUpgrade or RollbackUpgrade
	// YugabyteDBOpsRequest completes the upgrade either way. Defaults to Auto.
	// +optional
	// +kubebuilder:default=Auto
	Finalize YugabyteDBFinalizeMode `json:"finalize,omitempty"`
}

// YugabyteDBFinalizeMode selects whether an upgrade is finalized by the UpdateVersion request itself.
// +kubebuilder:validation:Enum=Auto;Manual
type YugabyteDBFinalizeMode string

const (
	// YugabyteDBFinalizeModeAuto finalizes the upgrade when the soak ends.
	YugabyteDBFinalizeModeAuto YugabyteDBFinalizeMode = "Auto"
	// YugabyteDBFinalizeModeManual leaves the upgrade pending until a FinalizeUpgrade or RollbackUpgrade request.
	YugabyteDBFinalizeModeManual YugabyteDBFinalizeMode = "Manual"
)

// FinalizeUpgradeSpec is the spec of a FinalizeUpgrade request. It has no fields: the upgrade to finalize is
// the one that is pending on the database.
type FinalizeUpgradeSpec struct{}

// RollbackUpgradeSpec is the spec of a RollbackUpgrade request. It has no fields: the upgrade to roll back is
// the one that is pending on the database.
type RollbackUpgradeSpec struct{}

// YugabyteDBHorizontalScalingSpec contains the horizontal scaling information of a YugabyteDB universe.
// The masters cannot be scaled, their number is the replication factor.
type YugabyteDBHorizontalScalingSpec struct {
	// Number of tservers
	// +kubebuilder:validation:Minimum=1
	TServer *int32 `json:"tserver,omitempty"`
	// Number of read replica tservers. It needs spec.topology.readReplica on the database.
	// +kubebuilder:validation:Minimum=0
	ReadReplica *int32 `json:"readReplica,omitempty"`
}

type YugabyteDBVerticalScalingSpec struct {
	// Resource spec for the masters
	Master *PodResources `json:"master,omitempty"`
	// Resource spec for the tservers
	TServer *PodResources `json:"tserver,omitempty"`
	// Resource spec for the read replicas
	ReadReplica *PodResources `json:"readReplica,omitempty"`
	// Mode selects how the vertical scaling is actuated. Defaults to Restart.
	// InPlace resizes the CPU of the running pods without a restart. The memory limit of YugabyteDB is fixed by
	// flags when the process starts, so an InPlace request that changes the memory is rejected.
	// +optional
	// +kubebuilder:default=Restart
	Mode VerticalScalingMode `json:"mode,omitempty"`
}

type YugabyteDBTLSSpec struct {
	// YugabyteDBTLSConfig contains the updated tls configurations.
	// +optional
	dbapi.YugabyteDBTLSConfig `json:",inline,omitempty"`

	// RotateCertificates tells operator to initiate certificate rotation
	// +optional
	RotateCertificates bool `json:"rotateCertificates,omitempty"`

	// Remove tells operator to remove TLS configuration
	// +optional
	Remove bool `json:"remove,omitempty"`
}

// +kubebuilder:validation:Enum=Restart;ReconfigureTLS;RotateAuth;Reconfigure;HorizontalScaling;VerticalScaling;VolumeExpansion;UpdateVersion;FinalizeUpgrade;RollbackUpgrade;StorageMigration
// ENUM(Restart,ReconfigureTLS,RotateAuth,Reconfigure,HorizontalScaling,VerticalScaling,VolumeExpansion,UpdateVersion,FinalizeUpgrade,RollbackUpgrade,StorageMigration)
type YugabyteDBOpsRequestType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// YugabyteDBOpsRequestList is a list of YugabyteDBOpsRequests
type YugabyteDBOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of YugabyteDBOpsRequest CRD objects
	Items []YugabyteDBOpsRequest `json:"items,omitempty"`
}
