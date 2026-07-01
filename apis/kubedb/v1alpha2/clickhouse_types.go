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
	ResourceKindClickHouse     = "ClickHouse"
	ResourceSingularClickHouse = "clickhouse"
	ResourcePluralClickHouse   = "clickhouses"
	ResourceCodeClickHouse     = "ch"

	ClickHouseCACert     ClickHouseCertificateAlias = "ca"
	ClickHouseClientCert ClickHouseCertificateAlias = "client"
	ClickHouseServerCert ClickHouseCertificateAlias = "server"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clickhouses,singular=clickhouse,shortName=ch,categories={datastore,kubedb,appscode,all}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ClickHouse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseSpec   `json:"spec,omitempty"`
	Status ClickHouseStatus `json:"status,omitempty"`
}

// ClickHouseSpec defines the desired state of ClickHouse
type ClickHouseSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Distributed if set true, the operator expands this ClickHouse across data centers
	// (DC-DR) by creating ManifestWork objects instead of raw resources. The single CR
	// materializes per-DC ReplicatedMergeTree replicas of every shard, a 3-site
	// ClickHouse Keeper ensemble, and a Lease-routed write endpoint.
	// +optional
	Distributed bool `json:"distributed,omitempty"`

	// Version of ClickHouse to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a ClickHouse database.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Cluster
	// +optional
	ClusterTopology *ClusterTopology `json:"clusterTopology,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// disable security. It disables authentication security of user.
	// If unset, default is false
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Configuration is an optional field to provide custom configuration file for database (i.e config.properties).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// You can provide custom configurations using Secret or ApplyConfig.
	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// Indicates how SSL/TLS certificate verification will be handled for both the server and client sides.
	// +optional
	SSLVerificationMode SSLVerificationMode `json:"sslVerificationMode,omitempty"`

	// TLS contains tls configurations for client and server.
	// +optional
	TLS *ClickHouseTLSConfig `json:"tls,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 20, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`

	// Archiver controls database backup using Archiver CR
	// +optional
	Archiver *Archiver `json:"archiver,omitempty"`

	// PodPlacementPolicy is the reference of the podPlacementPolicy. For a distributed
	// (DC-DR) ClickHouse it selects the PlacementPolicy whose clusterSpreadConstraint
	// spreads the per-DC replicas and the 3-site Keeper ensemble across data centers.
	// +kubebuilder:default={name:"default"}
	// +optional
	PodPlacementPolicy *core.LocalObjectReference `json:"podPlacementPolicy,omitempty"`
}

type ClusterTopology struct {
	// Clickhouse Cluster Structure
	Cluster ClusterSpec `json:"cluster,omitempty"`

	// ClickHouse Keeper server name
	ClickHouseKeeper *ClickHouseKeeper `json:"clickHouseKeeper,omitempty"`
}

type ClusterSpec struct {
	// Cluster Name
	Name string `json:"name,omitempty"`
	// Number of replica for each shard to deploy for a cluster.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Number of shard to deploy for a cluster.
	// +optional
	Shards *int32 `json:"shards,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`
}

type ClickHouseKeeper struct {
	ExternallyManaged bool `json:"externallyManaged,omitempty"`

	Node *ClickHouseKeeperNode `json:"node,omitempty"`

	Spec *ClickHouseKeeperSpec `json:"spec,omitempty"`
}

type ClickHouseKeeperSpec struct {
	// Number of replica for each shard to deploy for a cluster.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`
}

// ClickHouseKeeperNode defines item of nodes section of .spec.clusterTopology.
type ClickHouseKeeperNode struct {
	Host string `json:"host,omitempty"`

	// +optional
	Port *int32 `json:"port,omitempty"`
}

// ClickHouseStatus defines the observed state of ClickHouse
type ClickHouseStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabasePhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
	// DisasterRecovery reports the cross data center (DC-DR) state for a distributed ClickHouse.
	// +optional
	DisasterRecovery *ClickHouseDisasterRecoveryStatus `json:"disasterRecovery,omitempty"`
}

// ClickHouseDRPhase is the cross data center DR phase of a distributed ClickHouse.
type ClickHouseDRPhase string

const (
	ClickHouseDRPhaseSteady      ClickHouseDRPhase = "Steady"
	ClickHouseDRPhaseFailingOver ClickHouseDRPhase = "FailingOver"
	ClickHouseDRPhaseFailingBack ClickHouseDRPhase = "FailingBack"
	ClickHouseDRPhaseDegraded    ClickHouseDRPhase = "Degraded"
)

// ClickHouseDisasterRecoveryStatus reports the per data center DC-DR view of a
// distributed ClickHouse. ClickHouse is multi-master (ReplicatedMergeTree over a
// shared Keeper ensemble), so the cross-DC safety is the Keeper Raft quorum, not a
// promotion. The primary-DC Lease only routes the single write endpoint; this status
// reflects that routing and the per-DC replica health on the single Database object.
type ClickHouseDisasterRecoveryStatus struct {
	// ActiveDC is the write-routed data center: the DC the primary-DC Lease points at and
	// where the single write endpoint resolves. Because ClickHouse is multi-master this is
	// a routing choice for a stable single writer, not an engine-enforced primary.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase ClickHouseDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC plus the Arbiter DC.
	// +optional
	DataCenters []ClickHouseDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// ClickHouseDCStatus is one data center's local view inside a distributed ClickHouse.
type ClickHouseDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same
	// clusterName used in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member (holds ReplicatedMergeTree replicas of every shard plus a Keeper
	// voter) or Arbiter (holds only a data-less Keeper voter and the dr-controlplane
	// etcd member, no ClickHouse data).
	// +optional
	Role string `json:"role,omitempty"`

	// KeeperVoter is true when this DC hosts a ClickHouse Keeper voter in the 3-site ensemble.
	// +optional
	KeeperVoter bool `json:"keeperVoter,omitempty"`

	// KeeperQuorum is true when this DC observes the Keeper ensemble holding a Raft
	// quorum. A partitioned minority DC loses quorum and cannot register parts, so it
	// cannot commit writes: that is the split-brain guarantee.
	// +optional
	KeeperQuorum bool `json:"keeperQuorum,omitempty"`

	// Writable is true when this DC is the write-routed active DC.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// Shards is the per-shard ReplicatedMergeTree replica health inside this DC.
	// +optional
	Shards []ClickHouseDCShardStatus `json:"shards,omitempty"`

	// AbsoluteDelaySeconds is the maximum cross-DC ReplicatedMergeTree replication delay
	// (system.replicas.absolute_delay) across this DC's replicas, in seconds.
	// +optional
	AbsoluteDelaySeconds *int64 `json:"absoluteDelaySeconds,omitempty"`

	// QueueSize is the maximum replication queue size (system.replicas.queue_size) across
	// this DC's replicas.
	// +optional
	QueueSize *int32 `json:"queueSize,omitempty"`

	// Healthy reflects whether this DC's health Lease is fresh.
	// +optional
	Healthy bool `json:"healthy,omitempty"`
}

// ClickHouseDCShardStatus is one shard's ReplicatedMergeTree replica health inside a data center.
type ClickHouseDCShardStatus struct {
	// Shard is the shard ordinal.
	Shard int32 `json:"shard"`

	// TotalReplicas mirrors system.replicas.total_replicas for this shard's replica in this DC.
	// +optional
	TotalReplicas int32 `json:"totalReplicas,omitempty"`

	// ActiveReplicas mirrors system.replicas.active_replicas for this shard's replica in this DC.
	// +optional
	ActiveReplicas int32 `json:"activeReplicas,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClickHouseList contains a list of ClickHouse
type ClickHouseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClickHouse `json:"items"`
}

type ClickHouseTLSConfig struct {
	// TLS contains tls configurations for client and server.
	// +optional
	kmapi.TLSConfig `json:",omitempty"`

	// Specifies the external ca certificate secrets, which clickhouse will use as a client.
	// +optional
	ClientCACertificateRefs []core.SecretKeySelector `json:"clientCaCertificateRefs,omitempty"`
}

// +kubebuilder:validation:Enum=none;relaxed;strict;once
type SSLVerificationMode string

const (
	SSLVerificationModeNone    SSLVerificationMode = "none"
	SSLVerificationModeRelaxed SSLVerificationMode = "relaxed"
	SSLVerificationModeStrict  SSLVerificationMode = "strict"
	SSLVerificationModeOnce    SSLVerificationMode = "once"
)

var _ Accessor = &ClickHouse{}

func (m *ClickHouse) GetObjectMeta() metav1.ObjectMeta {
	return m.ObjectMeta
}

func (m *ClickHouse) GetConditions() []kmapi.Condition {
	return m.Status.Conditions
}

func (m *ClickHouse) SetCondition(cond kmapi.Condition) {
	m.Status.Conditions = setCondition(m.Status.Conditions, cond)
}

func (m *ClickHouse) RemoveCondition(typ string) {
	m.Status.Conditions = removeCondition(m.Status.Conditions, typ)
}
