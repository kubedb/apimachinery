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

package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeRedis     = "rd"
	ResourceKindRedis     = "Redis"
	ResourceSingularRedis = "redis"
	ResourcePluralRedis   = "redises"
)

// +kubebuilder:validation:Enum=Standalone;Cluster;Sentinel
type RedisMode string

const (
	RedisModeStandalone RedisMode = "Standalone"
	RedisModeCluster    RedisMode = "Cluster"
	RedisModeSentinel   RedisMode = "Sentinel"
)

// Redis defines a Redis database.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=redises,singular=redis,shortName=rd,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Redis struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisSpec   `json:"spec,omitempty"`
	Status            RedisStatus `json:"status,omitempty"`
}

type RedisSpec struct {
	// Redis ACL Configuration
	// +optional
	Acl *RedisAclSpec `json:"acl,omitempty"`

	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Version of Redis to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a Redis database.
	Replicas *int32 `json:"replicas,omitempty"`

	// Default is "Standalone". If set to "Cluster", ClusterSpec is required and redis servers will
	// start in cluster mode
	Mode RedisMode `json:"mode,omitempty"`

	SentinelRef *RedisSentinelRef `json:"sentinelRef,omitempty"`

	// Redis cluster configuration for running redis servers in cluster mode. Required if Mode is set to "Cluster"
	Cluster *RedisClusterSpec `json:"cluster,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// If disable Auth true then don't create any auth secret
	// +optional
	DisableAuth bool `json:"disableAuth,omitempty"`

	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for database (i.e redis.conf).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// +optional
	Configuration *RedisConfiguration `json:"configuration,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// TLS contains tls configurations for client and server.
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// AllowedSchemas defines the types of database schemas that MAY refer to
	// a database instance and the trusted namespaces where those schema resources MAY be
	// present.
	//
	// +kubebuilder:default={namespaces:{from: Same}}
	// +optional
	AllowedSchemas *AllowedConsumers `json:"allowedSchemas,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 1}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

// +kubebuilder:validation:Enum=server;client;metrics-exporter
type RedisCertificateAlias string

const (
	RedisServerCert          RedisCertificateAlias = "server"
	RedisClientCert          RedisCertificateAlias = "client"
	RedisMetricsExporterCert RedisCertificateAlias = "metrics-exporter"
)

type RedisClusterSpec struct {
	// Number of shards. It must be >= 3. If not specified, defaults to 3.
	Shards *int32 `json:"shards,omitempty"`

	// Number of replica(s) per shard. If not specified, defaults to 2.
	Replicas *int32 `json:"replicas,omitempty"`

	// Announce is used to announce the redis cluster endpoints.
	// It is used to set
	// cluster-announce-ip, cluster-announce-port, cluster-announce-bus-port, cluster-announce-tls-port
	// +optional
	Announce *Announce `json:"announce,omitempty"`
}

type RedisConfiguration struct {
	ConfigurationSpec `json:",inline,omitempty"`

	// Redis ACL Configuration
	// +optional
	Acl *RedisAclSpec `json:"acl,omitempty"`
}

// +kubebuilder:validation:Enum=ip;hostname
type PreferredEndpointType string

const (
	PreferredEndpointTypeIP       PreferredEndpointType = "ip"
	PreferredEndpointTypeHostname PreferredEndpointType = "hostname"
)

type Announce struct {
	// +kubebuilder:default=hostname
	Type PreferredEndpointType `json:"type,omitempty"`
	// This field is used to set cluster-announce information for redis cluster of each shard.
	Shards []Shards `json:"shards,omitempty"`
}

type RedisAclSpec struct {
	// SecretRef holds the password against which ACLs will be created if Rules are given.
	// +optional
	SecretRef *core.LocalObjectReference `json:"secretRef,omitempty"`

	// Rules specifies the ACL rules to be applied to the user associated with the provided SecretRef.
	// If provided, the system will update the ACLs for this user to ensure they are in sync with the new authentication settings.
	Rules []string `json:"rules,omitempty"`
}

type Shards struct {
	// Endpoints contains the cluster-announce information for all the replicas in a shard.
	// This will be used to set cluster-announce-ip/hostname, cluster-announce-port/cluster-announce-tls-port
	// and cluster-announce-bus-port
	// format cluster-announce (host:port@busport)
	Endpoints []string `json:"endpoints,omitempty"`
}

type RedisSentinelRef struct {
	// Name of the refereed sentinel
	Name string `json:"name,omitempty"`

	// Namespace where refereed sentinel has been deployed
	Namespace string `json:"namespace,omitempty"`
}

type RedisStatus struct {
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
	AuthSecret *Age `json:"authSecret,omitempty"`
	// DisasterRecovery reports the cross data center (DC-DR) state for a distributed Redis.
	// +optional
	DisasterRecovery *RedisDisasterRecoveryStatus `json:"disasterRecovery,omitempty"`
}

// RedisDRPhase is the cross data center DR phase of a distributed Redis.
type RedisDRPhase string

const (
	RedisDRPhaseSteady      RedisDRPhase = "Steady"
	RedisDRPhaseFailingOver RedisDRPhase = "FailingOver"
	RedisDRPhaseFailingBack RedisDRPhase = "FailingBack"
	RedisDRPhaseDegraded    RedisDRPhase = "Degraded"
)

// RedisDisasterRecoveryStatus reports the per data center DC-DR view of a
// distributed Redis. The cross-DC decision is owned by the dr-controlplane
// primary-DC Lease; this status reflects it on the single Database object. Each
// Member DC runs a self-contained Redis (its own gossip ring or Sentinel quorum);
// the standby DC's master async-replicates from the active DC's master.
type RedisDisasterRecoveryStatus struct {
	// ActiveDC is the data center that currently holds the primary DC Lease and runs the writable master.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase RedisDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []RedisDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

// RedisDCStatus is one data center's local view inside a distributed Redis.
type RedisDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same
	// clusterName used in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter. An Arbiter DC holds only the dr-controlplane etcd
	// member and runs no Redis.
	// +optional
	Role string `json:"role,omitempty"`

	// Master is this DC's local Redis master pod.
	// +optional
	Master string `json:"master,omitempty"`

	// Writable is true when this DC's master is the cluster's writable primary.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// LinkStatus is this DC master's cross-DC replication link health
	// (master_link_status from INFO replication), for example up or down. It is empty
	// on the active DC, which has no upstream.
	// +optional
	LinkStatus string `json:"linkStatus,omitempty"`

	// LagBytes is this DC's cross-DC replication lag behind the active DC, measured as
	// the active master's master_repl_offset minus this DC master's replicated offset.
	// +optional
	LagBytes *int64 `json:"lagBytes,omitempty"`

	// Healthy reflects whether this DC's health Lease is fresh.
	// +optional
	Healthy bool `json:"healthy,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of Redis TPR objects
	Items []Redis `json:"items,omitempty"`
}

var _ Accessor = &Redis{}

func (r *Redis) GetObjectMeta() metav1.ObjectMeta {
	return r.ObjectMeta
}

func (r *Redis) GetConditions() []kmapi.Condition {
	return r.Status.Conditions
}

func (r *Redis) SetCondition(cond kmapi.Condition) {
	r.Status.Conditions = setCondition(r.Status.Conditions, cond)
}

func (r *Redis) RemoveCondition(typ string) {
	r.Status.Conditions = removeCondition(r.Status.Conditions, typ)
}
