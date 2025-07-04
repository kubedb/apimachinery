/*
Copyright 2023.

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
	ResourceCodeMSSQLServer     = "ms"
	ResourceKindMSSQLServer     = "MSSQLServer"
	ResourceSingularMSSQLServer = "mssqlserver"
	ResourcePluralMSSQLServer   = "mssqlservers"
)

// +kubebuilder:validation:Enum=AvailabilityGroup;DistributedAG
type MSSQLServerMode string

const (
	MSSQLServerModeAvailabilityGroup MSSQLServerMode = "AvailabilityGroup"
	MSSQLServerModeDistributedAG     MSSQLServerMode = "DistributedAG"
)

// +kubebuilder:validation:Enum=server;client;endpoint
type MSSQLServerCertificateAlias string

const (
	MSSQLServerServerCert   MSSQLServerCertificateAlias = "server"
	MSSQLServerClientCert   MSSQLServerCertificateAlias = "client"
	MSSQLServerEndpointCert MSSQLServerCertificateAlias = "endpoint"
)

// +kubebuilder:validation:Enum=Passive;ReadOnly;All
type SecondaryAccessMode string

const (
	// Passive  = secondary is passive, no connections allowed
	SecondaryAccessModePassive SecondaryAccessMode = "Passive"
	// ReadOnly = secondary allows read-intent only
	SecondaryAccessModeReadOnly SecondaryAccessMode = "ReadOnly"
	// All = secondary allows any connections
	SecondaryAccessModeAll SecondaryAccessMode = "All"
)

// +kubebuilder:validation:Enum=Primary;Secondary
type DistributedAGRole string

const (
	DistributedAGRolePrimary   DistributedAGRole = "Primary"
	DistributedAGRoleSecondary DistributedAGRole = "Secondary"
)

// MSSQLServer defines a MSSQLServer database.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqlservers,singular=mssqlserver,shortName=ms,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MSSQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MSSQLServerSpec   `json:"spec,omitempty"`
	Status MSSQLServerStatus `json:"status,omitempty"`
}

// MSSQLServerSpec defines the desired state of MSSQLServer
type MSSQLServerSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Version of MSSQLServer to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a MSSQLServer database. In case of MSSQLServer Availability Group.
	Replicas *int32 `json:"replicas,omitempty"`

	// MSSQLServer cluster topology
	// +optional
	Topology *MSSQLServerTopology `json:"topology,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// ConfigSecret is an optional field to provide a custom configuration file for the database (i.e., mssql.conf).
	// If specified, this file will be used as a configuration file, otherwise a default configuration file will be used.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// Init is used to initialize a database
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// TLS contains tls configurations for client and server.
	TLS *MSSQLServerTLSConfig `json:"tls,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 1}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// Archiver controls database backup using Archiver CR
	// +optional
	Archiver *Archiver `json:"archiver,omitempty"`

	// Arbiter controls spec for arbiter pods
	// +optional
	Arbiter *ArbiterSpec `json:"arbiter,omitempty"`
}

type MSSQLServerTLSConfig struct {
	kmapi.TLSConfig `json:",inline"`

	// +optional
	ClientTLS *bool `json:"clientTLS"`
}

type MSSQLServerTopology struct {
	// If set to -
	// "AvailabilityGroup", MSSQLAvailabilityGroupSpec is required and MSSQLServer servers will start an Availability Group
	// "DistributedAG", MSSQLServerDistributedAGSpec is required, and MSSQLServer servers will start a Distributed Availability Group
	Mode *MSSQLServerMode `json:"mode,omitempty"`

	// AvailabilityGroup info for MSSQLServer (used when Mode is "AvailabilityGroup" or "DistributedAG").
	// +optional
	AvailabilityGroup *MSSQLServerAvailabilityGroupSpec `json:"availabilityGroup,omitempty"`

	// DistributedAG contains information of the DAG (Distributed Availability Group) configuration.
	// Used when Mode is "DistributedAG".
	// +optional
	DistributedAG *MSSQLServerDistributedAGSpec `json:"distributedAG,omitempty"`
}

// MSSQLServerAvailabilityGroupSpec defines the availability group spec for MSSQLServer
type MSSQLServerAvailabilityGroupSpec struct {
	// AvailabilityDatabases is an array of databases to be included in the availability group
	// +optional
	Databases []string `json:"databases,omitempty"`

	// Leader election configuration
	// +optional
	LeaderElection *MSSQLServerLeaderElectionConfig `json:"leaderElection,omitempty"`

	// SecondaryAccessMode controls which connections are allowed to secondary replicas.
	// https://learn.microsoft.com/en-us/sql/t-sql/statements/create-availability-group-transact-sql?view=sql-server-ver16#secondary_role---
	// +optional
	// +kubebuilder:default=Passive
	SecondaryAccessMode SecondaryAccessMode `json:"secondaryAccessMode,omitempty"`

	// LoginSecretName is the name of the secret containing the password for the 'dbm_login' user.
	// For a Distributed AG, both the primary and secondary AGs must use the same login and password.
	// This secret must be created by the user.
	// +optional
	LoginSecretName string `json:"loginSecretName,omitempty"`

	// MasterKeySecretName is the name of the secret containing the password for the database master key.
	// For a Distributed AG, both sides must use the same master key password.
	// This secret must be created by the user.
	// +optional
	MasterKeySecretName string `json:"masterKeySecretName,omitempty"`

	// EndpointCertSecretName is the name of the secret containing the certificate and private key for the database mirroring endpoint.
	// For a Distributed AG, both sides must use the same certificate. The secret should contain `tls.crt` and `tls.key`.
	// This secret must be created by the user.
	// +optional
	EndpointCertSecretName string `json:"endpointCertSecretName,omitempty"`
}

// MSSQLServerDistributedAGSpec defines the configuration for a Distributed Availability Group.
type MSSQLServerDistributedAGSpec struct {
	// Self defines the configuration for the local Availability Group's participation in the DAG.
	// +kubebuilder:validation:Required
	Self MSSQLServerDistributedAGSelfSpec `json:"self"`

	// Remote defines the connection details and name of the remote Availability Group.
	// +kubebuilder:validation:Required
	Remote MSSQLServerDistributedAGRemoteSpec `json:"remote"`
}

// MSSQLServerDistributedAGSelfSpec defines the configuration for the local AG's role in the DAG.
type MSSQLServerDistributedAGSelfSpec struct {
	// Role indicates if the local Availability Group (defined in spec.topology.availabilityGroup)
	// is acting as Primary or Secondary in this Distributed Availability Group (DAG).
	// +kubebuilder:validation:Required
	Role DistributedAGRole `json:"role"`

	// URL is the listener endpoint URL of the *local* Availability Group that will participate in this DAG.
	// This must be reachable by the SQL Server instance.
	// Example: "ag1-listener.my-namespace.svc:5022" or an externally reachable IP:Port.
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// MSSQLServerDistributedAGRemoteSpec defines the connection details for the remote AG.
type MSSQLServerDistributedAGRemoteSpec struct {
	// Name is the actual name of the Availability Group on the remote cluster.
	// +kubebuilder:validation:Required
	Name string `json:"name"` // Name of the remote cluster's local AG.

	// URL is the listener endpoint URL of the *remote* Availability Group that will be the other member of this DAG.
	// This URL must be reachable from the SQL Server instances in this cluster.
	// Example: Use the external LoadBalancer IP or hostname
	// e.g., "external-ip-of-remote-ag-listener:5022" or 10.2.0.64:5022 (instead of an internal cluster DNS name)
	// +kubebuilder:validation:Required
	URL string `json:"url"`
}

// MSSQLServerStatus defines the observed state of MSSQLServer
type MSSQLServerStatus struct {
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
}

// MSSQLServerLeaderElectionConfig contains essential attributes of leader election.
type MSSQLServerLeaderElectionConfig struct {
	// Period between Node.Tick invocations
	// +kubebuilder:default="100ms"
	// +optional
	Period metav1.Duration `json:"period,omitempty"`

	// ElectionTick is the number of Node.Tick invocations that must pass between
	//	elections. That is, if a follower does not receive any message from the
	//  leader of current term before ElectionTick has elapsed, it will become
	//	candidate and start an election. ElectionTick must be greater than
	//  HeartbeatTick. We suggest ElectionTick = 10 * HeartbeatTick to avoid
	//  unnecessary leader switching. default value is 10.
	// +default=10
	// +kubebuilder:default=10
	// +optional
	ElectionTick int32 `json:"electionTick,omitempty"`

	// HeartbeatTick is the number of Node.Tick invocations that must pass between
	// heartbeats. That is, a leader sends heartbeat messages to maintain its
	// leadership every HeartbeatTick ticks. default value is 1.
	// +default=1
	// +kubebuilder:default=1
	// +optional
	HeartbeatTick int32 `json:"heartbeatTick,omitempty"`

	// TransferLeadershipInterval retry interval for transfer leadership
	// to the healthiest node
	// +kubebuilder:default="1s"
	// +optional
	TransferLeadershipInterval *metav1.Duration `json:"transferLeadershipInterval,omitempty"`

	// TransferLeadershipTimeout retry timeout for transfer leadership
	// to the healthiest node
	// +kubebuilder:default="60s"
	// +optional
	TransferLeadershipTimeout *metav1.Duration `json:"transferLeadershipTimeout,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MSSQLServerList contains a list of MSSQLServer
type MSSQLServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MSSQLServer `json:"items"`
}
