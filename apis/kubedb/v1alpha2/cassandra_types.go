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
	ResourceKindCassandra     = "Cassandra"
	ResourceSingularCassandra = "cassandra"
	ResourcePluralCassandra   = "cassandras"
	ResourceCodeCassandra     = "cas"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=cassandras,singular=cassandra,shortName=cas,categories={datastore,kubedb,appscode,all}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
}

// CassandraSpec defines the desired state of Cassandra
type CassandraSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Distributed if set true, the operator expands this Cassandra across data centers
	// (DC-DR) by creating ManifestWork objects instead of raw resources. The single CR
	// materializes one Cassandra datacenter per Member DC (its racks placed within it) on
	// a single ring, wires NetworkTopologyStrategy and cross-DC seeds/snitch over
	// KubeSlice, and routes the single user-facing endpoint via the Lease.
	// +optional
	Distributed bool `json:"distributed,omitempty"`

	// Version of Cassandra to be deployed.
	Version string `json:"version"`

	// Number of replicas for  Cassandra database.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Cassandra Topology for Racks
	// +optional
	Topology *Topology `json:"topology,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// disable security. It disables authentication security of user.
	// If unset, default is false
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Configuration is an optional field to provide custom configuration file for database (i.e. config.properties).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// You can provide custom configurations using Secret or ApplyConfig.
	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// Keystore encryption secret
	// +optional
	KeystoreCredSecret *SecretReference `json:"keystoreCredSecret,omitempty"`

	// To enable ssl for http layer
	EnableSSL bool `json:"enableSSL,omitempty"`

	// TLS contains tls configurations
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

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

	// Init is used to initialize the database from a script or git repo.
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// PodPlacementPolicy is the reference of the podPlacementPolicy. For a distributed
	// (DC-DR) Cassandra it selects the PlacementPolicy whose clusterSpreadConstraint
	// spreads the per-DC datacenters (and, for an even data-DC count, the engine-free
	// Arbiter DC) across data centers.
	// +kubebuilder:default={name:"default"}
	// +optional
	PodPlacementPolicy *core.LocalObjectReference `json:"podPlacementPolicy,omitempty"`
}

type Topology struct {
	// cassandra rack structure
	Rack []RackSpec `json:"rack,omitempty"`
}

type RackSpec struct {
	// rack Name
	Name string `json:"name,omitempty"`
	// Number of replica for each shard to deploy for a rack.
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

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
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
	// DisasterRecovery reports the cross data center (DC-DR) state for a distributed Cassandra.
	// +optional
	DisasterRecovery *CassandraDisasterRecoveryStatus `json:"disasterRecovery,omitempty"`
}

// CassandraDRPhase is the cross data center DR phase of a distributed Cassandra.
type CassandraDRPhase string

const (
	CassandraDRPhaseSteady      CassandraDRPhase = "Steady"
	CassandraDRPhaseFailingOver CassandraDRPhase = "FailingOver"
	CassandraDRPhaseFailingBack CassandraDRPhase = "FailingBack"
	CassandraDRPhaseDegraded    CassandraDRPhase = "Degraded"
)

// CassandraDisasterRecoveryStatus reports the per data center DC-DR view of a
// distributed Cassandra. Cassandra is masterless (one Dynamo-style ring spans the
// Member DCs, each a Cassandra datacenter, with native continuous NetworkTopologyStrategy
// replication), so there is no primary and no cross-DC election: the cross-DC safety is
// per-DC LOCAL_QUORUM, not a promotion. The primary-DC Lease only routes the single
// user-facing write endpoint; this status reflects that routing and the per-DC ring
// health on the single Database object.
type CassandraDisasterRecoveryStatus struct {
	// ActiveDC is the write-routed data center: the DC the primary-DC Lease points at and
	// where the single user-facing endpoint resolves. Because Cassandra is masterless this
	// is a routing choice for a stable single-writer posture, not an engine-enforced
	// primary; writing in multiple DCs at once (active-active) is also legitimate.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase CassandraDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC plus the Arbiter DC.
	// +optional
	DataCenters []CassandraDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// CassandraDCStatus is one data center's local view inside a distributed Cassandra.
type CassandraDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same
	// clusterName used in the PlacementPolicy distributionRule, and the Cassandra
	// datacenter name in cassandra-rackdc.properties).
	ClusterName string `json:"clusterName"`

	// Role is Member (a full Cassandra datacenter holding NetworkTopologyStrategy replicas)
	// or Arbiter (engine-free, holds only the dr-controlplane etcd vote, no Cassandra).
	// +optional
	Role string `json:"role,omitempty"`

	// ReplicationFactor is this DC's NetworkTopologyStrategy replication factor for the
	// managed keyspaces.
	// +optional
	ReplicationFactor int32 `json:"replicationFactor,omitempty"`

	// Writable is true when this DC is the write-routed active DC.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// UpNodes is the number of nodes reported Up/Normal (UN) by nodetool status for this DC.
	// +optional
	UpNodes int32 `json:"upNodes,omitempty"`

	// TotalNodes is the number of nodes nodetool status lists for this DC.
	// +optional
	TotalNodes int32 `json:"totalNodes,omitempty"`

	// HintBacklogBytes is the maximum cross-DC hinted-handoff backlog observed for this DC
	// (hints queued for delivery to it), a proxy for cross-DC replication delay.
	// +optional
	HintBacklogBytes *int64 `json:"hintBacklogBytes,omitempty"`

	// PendingRanges is the number of streaming/pending ranges (nodetool netstats) for this
	// DC, non-zero while it catches up after a rejoin or repair.
	// +optional
	PendingRanges *int32 `json:"pendingRanges,omitempty"`

	// Healthy reflects whether this DC's health Lease is fresh.
	// +optional
	Healthy bool `json:"healthy,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraList contains a list of Cassandra
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

// +kubebuilder:validation:Enum=server;client
type CassandraCertificateAlias string

const (
	CassandraServerCert CassandraCertificateAlias = "server"
	CassandraClientCert CassandraCertificateAlias = "client"
)

var _ Accessor = &Cassandra{}

func (m *Cassandra) GetObjectMeta() metav1.ObjectMeta {
	return m.ObjectMeta
}

func (m *Cassandra) GetConditions() []kmapi.Condition {
	return m.Status.Conditions
}

func (m *Cassandra) SetCondition(cond kmapi.Condition) {
	m.Status.Conditions = setCondition(m.Status.Conditions, cond)
}

func (m *Cassandra) RemoveCondition(typ string) {
	m.Status.Conditions = removeCondition(m.Status.Conditions, typ)
}
