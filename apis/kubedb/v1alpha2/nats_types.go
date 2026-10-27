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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeNATS     = "nats"
	ResourceKindNATS     = "NATS"
	ResourceSingularNATS = "nats"
	// ResourcePluralNATS is "natses", not "nats": a plural equal to the singular is not allowed.
	ResourcePluralNATS = "natses"
)

// NATSCertificateAlias is the alias of a certificate in spec.tls.certificates.
// +kubebuilder:validation:Enum=server;client
type NATSCertificateAlias string

const (
	// NATSServerCert is presented by every nats-server on the client, route and monitor listeners.
	NATSServerCert NATSCertificateAlias = "server"
	// NATSClientCert is used by KubeDB (health checker, ops-manager) and by applications to connect.
	NATSClientCert NATSCertificateAlias = "client"
)

// NATSAuthMode selects how clients are authenticated.
// +kubebuilder:validation:Enum=Config;Operator
type NATSAuthMode string

const (
	// NATSAuthModeConfig renders the system account and one application account into the server config.
	NATSAuthModeConfig NATSAuthMode = "Config"
	// NATSAuthModeOperator uses the decentralized JWT trust chain with a full resolver.
	NATSAuthModeOperator NATSAuthMode = "Operator"
)

// NATS is the Schema for the natses API

// +genclient
// +resourceName=natses
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:path=natses,singular=nats,shortName=nats,categories={datastore,kubedb,appscode,all}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NATS struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of NATS
	// +optional
	Spec NATSSpec `json:"spec,omitempty"`

	// status defines the observed state of NATS
	// +optional
	Status NATSStatus `json:"status,omitempty"`
}

// NATSSpec defines the desired state of NATS.
//
// Every nats-server of a NATS object is identical, so there is one PetSet and no per-role topology.
// The fields that an OpsRequest owns (version, replicas, resources, storage size, configuration, TLS,
// auth secret) are rejected by the webhook when edited directly after provisioning.
type NATSSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Version of NATS to be deployed. It refers to the name of a NATSVersion catalog object.
	Version string `json:"version"`

	// Number of nats-server pods. 1, or an odd number >= 3 when JetStream is enabled; any value >= 1
	// when it is not. Defaults to 3. Changed by HorizontalScaling only.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// JetStream configures the persistence layer. Default: enabled, file store on a PVC, no memory store.
	// +optional
	JetStream *NATSJetStreamSpec `json:"jetStream,omitempty"`

	// Configuration holds ordinary NATS server settings. Supported keys are nats.conf (rendered as extra
	// top-level keys in user.conf) and accounts.conf (extra accounts, Config auth mode only).
	// Operator-owned keys (server_name, listen, jetstream, cluster, accounts, ...) are rejected.
	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// ServerConfig holds typed server limits and timeouts.
	// +optional
	ServerConfig *NATSServerConfig `json:"serverConfig,omitempty"`

	// LameDuck configures the lame-duck timing. Defaults come from the NATSVersion (10s / 30s).
	// Not reloadable: a change rolls the pods, since it also changes terminationGracePeriodSeconds.
	// +optional
	LameDuck *NATSLameDuckSpec `json:"lameDuck,omitempty"`

	// DisableSecurity lets clients connect without credentials. They are mapped to the APP account;
	// the system account keeps its credential. Immutable. Not allowed in Operator auth mode.
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// AuthSecret holds the admin credentials (user "admin"). Defaults to <name>-auth.
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Auth selects the authentication mode: Config (default) or Operator. Immutable.
	// +optional
	Auth *NATSAuthSpec `json:"auth,omitempty"`

	// TLS configures certificates issued from spec.tls.issuerRef. Supported aliases: server, client.
	// When TLS is specified, issuerRef must be set.
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// ClientTLS selects whether the client port (4222) uses TLS when spec.tls is set. Default true.
	// +optional
	ClientTLS *bool `json:"clientTLS,omitempty"`

	// GatewayTLS selects whether the gateway port (7222) uses TLS when spec.tls is set. Default true.
	// Gateway TLS material reloads in place.
	// +optional
	GatewayTLS *bool `json:"gatewayTLS,omitempty"`

	// LeafNodeTLS selects whether the leafnodes port (7422) uses TLS when spec.tls is set. Default true.
	// +optional
	LeafNodeTLS *bool `json:"leafNodeTLS,omitempty"`

	// ClusterTLS selects whether the route port (6222) uses TLS when spec.tls is set. Default true.
	// +optional
	ClusterTLS *bool `json:"clusterTLS,omitempty"`

	// MonitorTLS selects whether the monitor port (8222) uses TLS when spec.tls is set. Default true.
	// +optional
	MonitorTLS *bool `json:"monitorTLS,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database.
	// Aliases: primary, stats, leafnodes (the leafnodes port of the client-facing Service when
	// leafNodes.accept is true; default ClusterIP).
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Gateway joins this cluster to a super-cluster of other NATS clusters. A NATS object is always one
	// cluster; clusters in different DCs are joined by gateways. Changing a configured remote is a rolling
	// restart, adding a DC needs none on the existing ones (gateways gossip).
	// +optional
	Gateway *NATSGatewaySpec `json:"gateway,omitempty"`

	// LeafNodes accepts leaf connections (hub side) and/or dials out to a remote NATS system (leaf side).
	// +optional
	LeafNodes *NATSLeafNodesSpec `json:"leafNodes,omitempty"`

	// Distributed makes KubeDB create one NATS cluster per Member rule of the PlacementPolicy referenced by
	// podTemplate.spec.podPlacementPolicy, joined by gateways. The PlacementPolicy needs a
	// clusterSpreadConstraint and no failoverPolicy.
	// +optional
	Distributed bool `json:"distributed,omitempty"`

	// Arbiter configures the vote-only JetStream servers in the Arbiter DC of a distributed NATS.
	// Meaningful only when the PlacementPolicy has an Arbiter rule.
	// +optional
	Arbiter *NATSArbiterSpec `json:"arbiter,omitempty"`

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
}

// NATSGatewaySpec joins this cluster to a super-cluster.
type NATSGatewaySpec struct {
	// Name is the gateway name. It must equal the cluster name (the object name), or the server refuses
	// to start. Defaults to the cluster name. Must be empty when spec.distributed is true: each DC's
	// gateway is named after the DC.
	// +optional
	Name string `json:"name,omitempty"`

	// Remotes seed discovery. Gateways gossip, so a few are enough.
	// +optional
	Remotes []NATSRemoteGateway `json:"remotes,omitempty"`

	// RejectUnknown sets reject_unknown_cluster. The operator never sets it; with it, adding a DC
	// restarts the existing ones.
	// +optional
	RejectUnknown bool `json:"rejectUnknown,omitempty"`
}

// NATSRemoteGateway is a remote cluster of the super-cluster. Exactly one of urls and natsRef is set.
type NATSRemoteGateway struct {
	// Name of the remote cluster (its gateway name).
	Name string `json:"name"`

	// URLs of the remote gateway, nats:// or tls://.
	// +optional
	URLs []string `json:"urls,omitempty"`

	// NATSRef points at another NATS object (through its AppBinding). KubeDB fills in the gateway URLs.
	// +optional
	NATSRef *kmapi.ObjectReference `json:"natsRef,omitempty"`
}

// NATSLeafNodesSpec configures leaf connections.
type NATSLeafNodesSpec struct {
	// Accept (hub side) listens on 7422 and exposes the port on the client Service.
	// +optional
	Accept bool `json:"accept,omitempty"`

	// Remotes (leaf side) dial out to a remote NATS system.
	// +optional
	Remotes []NATSLeafRemote `json:"remotes,omitempty"`

	// JetStreamDomain is required when this NATS runs its own JetStream and has leaf remotes.
	// +optional
	JetStreamDomain string `json:"jetStreamDomain,omitempty"`
}

// NATSLeafRemote is a remote NATS system this NATS connects to as a leaf. Exactly one of urls and
// natsRef is set.
type NATSLeafRemote struct {
	// URLs of the remote leafnodes listener, nats-leaf:// , nats:// or tls://.
	// +optional
	URLs []string `json:"urls,omitempty"`

	// NATSRef points at another NATS object (through its AppBinding). KubeDB fills in the leafnode URLs.
	// +optional
	NATSRef *kmapi.ObjectReference `json:"natsRef,omitempty"`

	// Account binds the leaf connection to a local account. Without it (and without credentialsSecret)
	// the leaf binds the default account and has no subject boundary.
	// +optional
	Account string `json:"account,omitempty"`

	// CredentialsSecret holds a .creds file (key "creds"), used in Operator auth mode.
	// +optional
	CredentialsSecret *core.LocalObjectReference `json:"credentialsSecret,omitempty"`

	// Domain is the JetStream domain of the remote, when it differs from the default.
	// +optional
	Domain string `json:"domain,omitempty"`
}

// NATSArbiterSpec configures the vote-only JetStream servers of the Arbiter DC. They are meta-group
// peers (they vote) but are excluded from stream and consumer placement.
type NATSArbiterSpec struct {
	// Replicas of the arbiter PetSet. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Storage is the small PVC holding the meta state. Defaults to 1Gi.
	// +optional
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`
}

// NATSJetStreamSpec configures JetStream.
type NATSJetStreamSpec struct {
	// Enabled turns JetStream on. Defaults to true. Immutable in v1.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// StorageType of the file store. Durable (default) uses a PVC mounted at /data; Ephemeral uses an
	// emptyDir and is only allowed with replicas == 1. Immutable.
	// +optional
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage is the file store PVC. Defaults to 10Gi.
	// +optional
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// MemoryStoreSize caps the memory store. Unset means 0: memory streams are disabled.
	// Must be <= 50% of limits.memory.
	// +optional
	MemoryStoreSize *resource.Quantity `json:"memoryStoreSize,omitempty"`
}

// NATSLameDuckSpec configures the lame-duck mode a pod enters before it stops.
type NATSLameDuckSpec struct {
	// GracePeriod is the time clients get to leave before they are closed. Must be < Duration.
	// +optional
	GracePeriod metav1.Duration `json:"gracePeriod,omitempty"`

	// Duration is the total lame-duck time (lame_duck_duration). Must be >= 30s, the server minimum.
	// +optional
	Duration metav1.Duration `json:"duration,omitempty"`
}

// NATSServerConfig holds typed server limits and timeouts. Most of them are reloadable;
// maxSubscriptions and maxPending are not and take a rolling restart.
type NATSServerConfig struct {
	// +optional
	MaxConnections *int32 `json:"maxConnections,omitempty"`
	// MaxPayload must be <= maxPending.
	// +optional
	MaxPayload *int32 `json:"maxPayload,omitempty"`
	// MaxPending is not reloadable.
	// +optional
	MaxPending *int64 `json:"maxPending,omitempty"`
	// MaxSubscriptions is not reloadable.
	// +optional
	MaxSubscriptions *int32 `json:"maxSubscriptions,omitempty"`
	// +optional
	MaxControlLine *int32 `json:"maxControlLine,omitempty"`
	// +optional
	WriteDeadline *metav1.Duration `json:"writeDeadline,omitempty"`
	// +optional
	AuthTimeout *metav1.Duration `json:"authTimeout,omitempty"`
	// +optional
	Debug bool `json:"debug,omitempty"`
	// +optional
	Trace bool `json:"trace,omitempty"`
}

// NATSAuthSpec selects the authentication mode.
type NATSAuthSpec struct {
	// Mode is Config (default) or Operator. Immutable.
	// +optional
	Mode NATSAuthMode `json:"mode,omitempty"`

	// Operator configures Operator mode. Absent: KubeDB generates and owns the whole trust chain.
	// +optional
	Operator *NATSOperatorSpec `json:"operator,omitempty"`
}

// NATSOperatorSpec configures the decentralized JWT mode.
type NATSOperatorSpec struct {
	// ExternalSecret brings your own operator (created with nsc). The Secret holds operator.jwt,
	// sys-account.jwt and sys.creds; KubeDB then generates nothing and only renders the server config.
	// +optional
	ExternalSecret *core.LocalObjectReference `json:"externalSecret,omitempty"`

	// ResolverStorage is the resolver PVC. Defaults to 1Gi; the full resolver keeps every account JWT on disk.
	// +optional
	ResolverStorage *core.PersistentVolumeClaimSpec `json:"resolverStorage,omitempty"`

	// AllowDelete is the resolver's allow_delete. Defaults to false.
	// +optional
	AllowDelete bool `json:"allowDelete,omitempty"`
}

// NATSStatus defines the observed state of NATS.
//
// Besides the standard conditions, NATS sets JetStreamQuorum and UnderReplicatedStreams.
// Phase Critical means some pods are not Ready but the JetStream meta group still has quorum.
type NATSStatus struct {
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

	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATSList contains a list of NATS
type NATSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of NATS CRD objects
	Items []NATS `json:"items"`
}

var _ Accessor = &NATS{}

func (n *NATS) GetObjectMeta() metav1.ObjectMeta {
	return n.ObjectMeta
}

func (n *NATS) GetConditions() []kmapi.Condition {
	return n.Status.Conditions
}

func (n *NATS) SetCondition(cond kmapi.Condition) {
	n.Status.Conditions = setCondition(n.Status.Conditions, cond)
}

func (n *NATS) RemoveCondition(typ string) {
	n.Status.Conditions = removeCondition(n.Status.Conditions, typ)
}

func (n *NATS) HasCondition(typ string) bool {
	for _, c := range n.Status.Conditions {
		if string(c.Type) == typ {
			return true
		}
	}
	return false
}
