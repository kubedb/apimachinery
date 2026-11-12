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

package v1alpha2

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeYugabyteDB     = "yb"
	ResourceKindYugabyteDB     = "YugabyteDB"
	ResourceSingularYugabyteDB = "yugabytedb"
	ResourcePluralYugabyteDB   = "yugabytedbs"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=yugabytedbs,singular=yugabytedb,shortName=yb,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type YugabyteDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YugabyteDBSpec   `json:"spec,omitempty"`
	Status YugabyteDBStatus `json:"status,omitempty"`
}

// YugabyteDBSpec defines the desired state of YugabyteDB.
type YugabyteDBSpec struct {
	// Version of YugabyteDB to deploy. Must correspond to a YugabyteDBVersion in the catalog.
	Version string `json:"version"`

	// ReplicationFactor is the number of masters and the default number of replicas of every tablet.
	// Defaults to 3. It cannot be changed after the database is created.
	// +kubebuilder:validation:Enum=1;3;5;7
	// +optional
	ReplicationFactor *int32 `json:"replicationFactor,omitempty"`

	// Topology configures the yb-master and yb-tserver roles.
	// +optional
	Topology *YugabyteDBTopology `json:"topology,omitempty"`

	// Placement names the cloud and region reported to YugabyteDB and orders the zones that hold the tablet leaders.
	// The zone of each pod is read from the node it runs on.
	// +optional
	Placement *YugabyteDBPlacement `json:"placement,omitempty"`

	// YSQL configures the PostgreSQL compatible API. It is enabled by default.
	// +optional
	YSQL *YugabyteDBYSQLSpec `json:"ysql,omitempty"`

	// YCQL configures the Cassandra compatible API. It is disabled by default.
	// +optional
	YCQL *YugabyteDBYCQLSpec `json:"ycql,omitempty"`

	// DisableSecurity disables authentication when set to true. Not recommended for production.
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// AuthSecret references the Secret that holds the credentials of the yugabyte user (keys username and password).
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Configuration holds custom flags. The keys are master.flags, tserver.flags, ysql_pg.conf and ysql_hba.conf.
	// Flags that the operator owns, such as addresses, placement and TLS, are rejected.
	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// TLS contains the cert-manager based TLS configuration.
	// +optional
	TLS *YugabyteDBTLSConfig `json:"tls,omitempty"`

	// ServiceTemplates customizes the Services that expose YugabyteDB (aliases primary and stats).
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Halted is a shortcut for deletionPolicy Halt. The pods are removed and the volumes and Secrets are kept.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// Monitor is used to monitor the database instance.
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// DeletionPolicy controls what happens to the data when this resource is deleted.
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker configures the health checks performed by the operator.
	// +optional
	// +kubebuilder:default={periodSeconds: 20, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`

	// Init is used to initialize the database from a script.
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// PITR lists the point in time recovery schedules, one for each database. Every entry becomes a YugabyteDB
	// snapshot schedule (create_snapshot_schedule). A Rewind YugabyteDBOpsRequest restores a database to a time
	// inside the retention of its schedule, and a Clone request copies a database as of such a time.
	// YugabyteDB allows at most one schedule for each database. A rewind can not go back past a YSQL upgrade
	// (upgrade_ysql), and roles and tablespaces are not rewound.
	// +listType=map
	// +listMapKey=database
	// +optional
	PITR []YugabyteDBPITR `json:"pitr,omitempty"`
}

// YugabyteDBPITRKind is the API of the database or keyspace that a PITR schedule protects.
// +kubebuilder:validation:Enum=YSQL;YCQL
type YugabyteDBPITRKind string

const (
	YugabyteDBPITRKindYSQL YugabyteDBPITRKind = "YSQL"
	YugabyteDBPITRKindYCQL YugabyteDBPITRKind = "YCQL"
)

// YugabyteDBPITR is the point in time recovery schedule of one database.
type YugabyteDBPITR struct {
	// Database is the YSQL database or the YCQL keyspace to protect.
	Database string `json:"database"`
	// Kind is YSQL (default) or YCQL.
	// +kubebuilder:default=YSQL
	// +optional
	Kind YugabyteDBPITRKind `json:"kind,omitempty"`
	// IntervalMinutes is the time between two snapshots.
	// +kubebuilder:validation:Minimum=1
	IntervalMinutes int32 `json:"intervalMinutes"`
	// RetentionMinutes is how far back the database can be rewound. It must be at least IntervalMinutes.
	// +kubebuilder:validation:Minimum=1
	RetentionMinutes int32 `json:"retentionMinutes"`
}

// YugabyteDBPITRStatus is the observed state of the PITR schedule of one database.
type YugabyteDBPITRStatus struct {
	// Database is the database or keyspace of the schedule.
	Database string `json:"database"`
	// ScheduleID is the ID of the snapshot schedule in YugabyteDB.
	// +optional
	ScheduleID string `json:"scheduleID,omitempty"`
	// LastSnapshotTime is the time of the latest snapshot of the schedule.
	// +optional
	LastSnapshotTime *metav1.Time `json:"lastSnapshotTime,omitempty"`
}

// YugabyteDBTopology configures the two roles of a YugabyteDB universe.
type YugabyteDBTopology struct {
	// Master is the yb-master role. Its replicas always equal the replication factor.
	// +optional
	Master *YugabyteDBNode `json:"master,omitempty"`
	// TServer is the yb-tserver role. It stores the tablets and runs the YSQL and YCQL query layers.
	// +optional
	TServer *YugabyteDBNode `json:"tserver,omitempty"`
	// ReadReplica is an optional second group of tservers, a read replica cluster. Its tservers hold
	// non-voting copies of every tablet in their own placement (placement_uuid), so they never take part in writes.
	// +optional
	ReadReplica *YugabyteDBReadReplica `json:"readReplica,omitempty"`
}

// YugabyteDBReadReplica configures the read replica tservers. The node fields (replicas, podTemplate, storageType,
// storage, dataVolumes) behave as they do for the tservers. The pods run in the PetSet <name>-readreplica.
type YugabyteDBReadReplica struct {
	YugabyteDBNode `json:",inline"`

	// PlacementUUID is the placement_uuid of the read replica tservers and of the read replica placement
	// registered on the masters. Defaults to "rr". It cannot be changed once the read replica is created.
	// +optional
	PlacementUUID string `json:"placementUUID,omitempty"`

	// Region overrides the region of the read replica placement. Defaults to spec.placement.region.
	// +optional
	Region string `json:"region,omitempty"`

	// Zone overrides the zone of the read replica placement. By default the zone of every pod is read from its node.
	// YugabyteDB needs the read replica zones to differ from the zones of the primary tservers.
	// +optional
	Zone string `json:"zone,omitempty"`

	// ReplicationFactor is the total number of read replica copies of every tablet. Defaults to 1.
	// It must not exceed the number of read replica pods.
	// +kubebuilder:validation:Minimum=1
	// +optional
	ReplicationFactor *int32 `json:"replicationFactor,omitempty"`
}

// YugabyteDBNode configures one role.
type YugabyteDBNode struct {
	// Replicas is the number of pods of the role.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// PodTemplate customizes the pods of the role. Set spec.podPlacementPolicy to spread the pods across zones and nodes.
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// StorageType is Durable (default, PVCs) or Ephemeral (emptyDir).
	// +optional
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage is the PVC template of every data volume when StorageType is Durable.
	// +optional
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// DataVolumes is the number of data volumes of each pod. They are mounted at /mnt/disk0 and up and listed in --fs_data_dirs.
	// It cannot be changed after the database is created.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8
	// +optional
	DataVolumes *int32 `json:"dataVolumes,omitempty"`
}

// YugabyteDBPlacement names the placement of the universe.
type YugabyteDBPlacement struct {
	// Cloud is reported as --placement_cloud. Defaults to "kubernetes".
	// +optional
	Cloud string `json:"cloud,omitempty"`
	// Region is reported as --placement_region. Defaults to the topology.kubernetes.io/region label of the node, else "region1".
	// +optional
	Region string `json:"region,omitempty"`
	// PreferredZones lists, in order of preference, the zones that should hold the tablet leaders.
	// +optional
	PreferredZones []string `json:"preferredZones,omitempty"`
}

// YugabyteDBYSQLSpec configures the YSQL API.
type YugabyteDBYSQLSpec struct {
	// Enabled starts the YSQL API. It cannot be changed after the database is created.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// ConnectionManager configures the YSQL Connection Manager, a built-in connection pooler. When it is enabled
	// the tservers move PostgreSQL to port 6433 and the Connection Manager listens on 5433.
	// +optional
	ConnectionManager *YugabyteDBConnectionManager `json:"connectionManager,omitempty"`
}

// YugabyteDBConnectionManager configures the YSQL Connection Manager (--enable_ysql_conn_mgr).
// It can only be changed by a Reconfigure request.
type YugabyteDBConnectionManager struct {
	// Enabled starts the Connection Manager.
	Enabled bool `json:"enabled"`

	// MaxClientConnections is --ysql_conn_mgr_max_client_connections, the number of client connections the
	// Connection Manager accepts. YugabyteDB defaults to 10000.
	// +kubebuilder:validation:Minimum=1
	// +optional
	MaxClientConnections *int32 `json:"maxClientConnections,omitempty"`
}

// YugabyteDBYCQLSpec configures the YCQL API.
type YugabyteDBYCQLSpec struct {
	// Enabled starts the YCQL API. It cannot be changed after the database is created.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`
}

// YugabyteDBTLSConfig configures TLS for node to node and client to node traffic.
type YugabyteDBTLSConfig struct {
	kmapi.TLSConfig `json:",inline"`
	// NodeToNode encrypts the traffic between the masters and the tservers. Defaults to true.
	// +optional
	NodeToNode *bool `json:"nodeToNode,omitempty"`
	// ClientToNode encrypts the traffic between the clients and the tservers (YSQL and YCQL). Defaults to true.
	// +optional
	ClientToNode *bool `json:"clientToNode,omitempty"`
	// Webserver serves the web UIs and the metrics endpoint of the masters (7000) and the tservers (9000) over HTTPS
	// with the node certificate (--webserver_certificate_file and related flags). Defaults to false.
	// It can only be changed by a ReconfigureTLS request.
	// +optional
	Webserver *bool `json:"webserver,omitempty"`
}

// YugabyteDBStatus defines the observed state of YugabyteDB.
type YugabyteDBStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabasePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this resource.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// PITR is the state of the snapshot schedules of spec.pitr.
	// +listType=map
	// +listMapKey=database
	// +optional
	PITR []YugabyteDBPITRStatus `json:"pitr,omitempty"`
}

// +kubebuilder:validation:Enum=master;tserver;readreplica
type YugabyteDBRole string

const (
	YugabyteDBRoleMaster      YugabyteDBRole = "master"
	YugabyteDBRoleTServer     YugabyteDBRole = "tserver"
	YugabyteDBRoleReadReplica YugabyteDBRole = "readreplica"
)

// +kubebuilder:validation:Enum=server;client
type YugabyteDBCertificateType string

const (
	YugabyteDBCertificateTypeServer YugabyteDBCertificateType = "server"
	YugabyteDBCertificateTypeClient YugabyteDBCertificateType = "client"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// YugabyteDBList contains a list of YugabyteDB.
type YugabyteDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YugabyteDB `json:"items"`
}
