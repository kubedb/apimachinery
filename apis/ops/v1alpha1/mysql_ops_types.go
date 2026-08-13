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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceCodeMySQLOpsRequest     = "myops"
	ResourceKindMySQLOpsRequest     = "MySQLOpsRequest"
	ResourceSingularMySQLOpsRequest = "mysqlopsrequest"
	ResourcePluralMySQLOpsRequest   = "mysqlopsrequests"
)

// MySQLOpsRequest defines a MySQL DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mysqlopsrequests,singular=mysqlopsrequest,shortName=myops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MySQLOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MySQLOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus    `json:"status,omitempty"`
}

// MySQLOpsRequestSpec is the spec for MySQLOpsRequest
type MySQLOpsRequestSpec struct {
	// Specifies the MySQL reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: Upgrade, HorizontalScaling, VerticalScaling etc.
	Type MySQLOpsRequestType `json:"type"`
	// Specifies information necessary for upgrading MySQL
	UpdateVersion *MySQLUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *MySQLHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *MySQLVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *MySQLVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of MySQL
	Configuration *ReconfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *MySQLTLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for configuring authSecret of the database
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Specifies information transform Remote Replica to GroupReplication
	ReplicationModeTransformation *MySQLReplicationModeTransformSpec `json:"replicationModeTransformation,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Specifies information necessary for migrating storageClass or data
	Migration *StorageMigrationSpec `json:"migration,omitempty"`
	// Specifies information necessary for restoring the database in place from its archiver backup
	Archiver *MySQLArchiverRestoreSpec `json:"archiver,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;Reconfigure;ReconfigureTLS;RotateAuth;ReplicationModeTransformation;StorageMigration;ArchiverRestore
// ENUM(UpdateVersion, HorizontalScaling, VerticalScaling, VolumeExpansion, Restart, Reconfigure, ReconfigureTLS, RotateAuth, ReplicationModeTransformation, StorageMigration, ArchiverRestore)
type MySQLOpsRequestType string

// MySQLArchiverRestoreSpec carries the archiver recovery information for an
// ArchiverRestore ops request.
//
// An archiver restore needs an empty data directory, which is why the documented
// flow restores into a *new* MySQL object by setting its spec.init.archiver.
// ArchiverRestore does the same thing in place: the ops request wipes the data
// volumes of the referenced database and then writes this payload into the
// database's own spec.init.archiver, so the provisioner runs the exact same
// restore path it runs for a freshly created restore target.
//
// The fields are therefore literally dbapi.ArchiverRecovery — recoveryTimestamp,
// encryptionSecret, manifestRepository, fullDBRepository, replicationStrategy and
// manifestOptions — embedded inline rather than redeclared, so the two stay in
// step by construction.
type MySQLArchiverRestoreSpec struct {
	dbapi.ArchiverRecovery `json:",inline"`

	// RetainPV controls what happens to the data PersistentVolumes the restore
	// replaces.
	//
	// An archiver restore deletes the database's data PVCs so the restore starts on an
	// empty data directory. Whether the volumes behind them survive that is decided by
	// their persistentVolumeReclaimPolicy, which usually comes from the StorageClass and
	// is usually Delete.
	//
	// When true (the default), the ops request records each volume's original policy and
	// forces every one of them to Retain before deleting anything, so the pre-restore
	// data survives even if the restore fails partway. Volumes that are still Bound at
	// the end get their original policy back; volumes the restore released keep Retain
	// and are named in an ArchiverRestoreManualCleanupRequired condition, because handing
	// Delete back to a Released volume destroys it within seconds -- which is exactly the
	// copy Retain was protecting. Removing them stays an explicit operator action.
	//
	// When false, the volumes are still forced to Retain for the duration of the restore
	// -- a failure after the wipe has to be undoable either way -- and the released ones
	// are deleted once the request succeeds. So the choice is about what survives a
	// successful restore, not about whether a failed one can be rolled back. Set it false
	// when the backup is the copy you trust and you would rather not clean up after every
	// restore.
	//
	// +optional
	// +kubebuilder:default=true
	RetainPV *bool `json:"retainPV,omitempty"`
}

// MySQLReplicaReadinessCriteria is the criteria for checking readiness of a MySQL pod
// after updating, horizontal scaling etc.
type MySQLReplicaReadinessCriteria struct{}

type MySQLUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion     string                         `json:"targetVersion,omitempty"`
	ReadinessCriteria *MySQLReplicaReadinessCriteria `json:"readinessCriteria,omitempty"`
}

type MySQLHorizontalScalingSpec struct {
	// Number of nodes/members of the group
	Member *int32 `json:"member,omitempty"`
}

type MySQLReplicationModeTransformSpec struct {
	// TargetTopologyMode is the clustered topology to transform the database into, i.e. the
	// value that ends up in the database's spec.topology.mode. Supported values are
	// "GroupReplication", "InnoDBCluster" and "SemiSync". This enables promoting a standalone
	// MySQL (or transforming a remote replica) into a clustered topology.
	// +kubebuilder:validation:Enum=GroupReplication;InnoDBCluster;SemiSync
	// +kubebuilder:default=GroupReplication
	// +optional
	TargetTopologyMode *dbapi.MySQLMode `json:"targetTopologyMode,omitempty"`

	// Mode is the Group Replication primary mode, i.e. the primary mode *within* the group:
	// either "Single-Primary" or "Multi-Primary". It applies only when targetTopologyMode is
	// "GroupReplication" or "InnoDBCluster"; it is ignored for "SemiSync", which has no group.
	// +kubebuilder:default=Single-Primary
	Mode *dbapi.MySQLGroupMode `json:"mode"`

	// TLSConfig contains updated tls configurations for client and server.
	// +optional
	kmapi.TLSConfig `json:",inline,omitempty"`

	// Indicates that the database server need to be encrypted connections(ssl)
	// +optional
	RequireSSL *bool `json:"requireSSL,omitempty"`
}

type MySQLVerticalScalingSpec struct {
	MySQL       *PodResources       `json:"mysql,omitempty"`
	Exporter    *ContainerResources `json:"exporter,omitempty"`
	Coordinator *ContainerResources `json:"coordinator,omitempty"`
	// Mode selects how the vertical scaling is actuated. Defaults to Restart.
	// +optional
	// +kubebuilder:default=Restart
	Mode VerticalScalingMode `json:"mode,omitempty"`
}

// MySQLVolumeExpansionSpec is the spec for MySQL volume expansion
type MySQLVolumeExpansionSpec struct {
	MySQL *resource.Quantity  `json:"mysql,omitempty"`
	Mode  VolumeExpansionMode `json:"mode"`
}

type MySQLTLSSpec struct {
	TLSSpec `json:",inline,omitempty"`

	// Indicates that the database server need to be encrypted connections(ssl)
	// +optional
	RequireSSL *bool `json:"requireSSL,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MySQLOpsRequestList is a list of MySQLOpsRequests
type MySQLOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of MySQLOpsRequest CRD objects
	Items []MySQLOpsRequest `json:"items,omitempty"`
}
