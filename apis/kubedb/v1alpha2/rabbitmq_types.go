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
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeRabbitmq     = "rm"
	ResourceKindRabbitmq     = "RabbitMQ"
	ResourceSingularRabbitmq = "rabbitmq"
	ResourcePluralRabbitmq   = "rabbitmqs"
)

// RabbitMQ is the Schema for the RabbitMQ API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=rabbitmqs,singular=rabbitmq,shortName=rm,categories={datastore,kubedb,appscode,all}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RabbitMQ struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSpec   `json:"spec,omitempty"`
	Status RabbitMQStatus `json:"status,omitempty"`
}

// RabbitMQSpec defines the desired state of RabbitMQ
type RabbitMQSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Version of RabbitMQ to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a RabbitMQ database.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Distributed if set true, manifestwork objects will be created instead of raw resources.
	// A distributed RabbitMQ is expanded by the operator into one self-contained RabbitMQ cluster
	// per Member data center for cross data center disaster recovery (DC-DR). Cross-DC replication
	// is carried by the Federation (or Shovel) plugin, active to standby.
	// +optional
	Distributed bool `json:"distributed,omitempty"`

	// PodPlacementPolicy is the reference of the podPlacementPolicy that spreads the per data
	// center RabbitMQ clusters across data centers for DC-DR.
	// +optional
	PodPlacementPolicy *core.LocalObjectReference `json:"podPlacementPolicy,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// To enable ssl for http layer
	EnableSSL bool `json:"enableSSL,omitempty"`

	// disable security. It disables authentication security of user.
	// If unset, default is false
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for database (i.e config.properties).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// TLS contains tls configurations
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// Indicates that the RabbitMQ Protocols that are required to be disabled on bootstrap.
	// +optional
	DisabledProtocols []RabbitMQProtocol `json:"disabledProtocols,omitempty"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

// RabbitMQStatus defines the observed state of RabbitMQ
type RabbitMQStatus struct {
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
	// DisasterRecovery reports the cross data center (DC-DR) state for a distributed RabbitMQ.
	// +optional
	DisasterRecovery *RabbitMQDisasterRecoveryStatus `json:"disasterRecovery,omitempty"`
}

// RabbitMQDRPhase is the cross data center DR phase of a distributed RabbitMQ.
type RabbitMQDRPhase string

const (
	RabbitMQDRPhaseSteady      RabbitMQDRPhase = "Steady"
	RabbitMQDRPhaseFailingOver RabbitMQDRPhase = "FailingOver"
	RabbitMQDRPhaseFailingBack RabbitMQDRPhase = "FailingBack"
	RabbitMQDRPhaseDegraded    RabbitMQDRPhase = "Degraded"
)

// RabbitMQDisasterRecoveryStatus reports the per data center DC-DR view of a distributed RabbitMQ.
// RabbitMQ DC-DR is active/passive: exactly one data center is the publish cluster, chosen by the
// dr-controlplane primary-DC Lease, and the Federation (or Shovel) plugin asynchronously replicates
// it to the standby. This status reflects that decision on the single RabbitMQ object.
type RabbitMQDisasterRecoveryStatus struct {
	// ActiveDC is the data center that currently holds the primary DC Lease and takes client publishes.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase RabbitMQDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []RabbitMQDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *meta.Time `json:"lastTransitionTime,omitempty"`
}

// RabbitMQDCStatus is one data center's local view inside a distributed RabbitMQ.
type RabbitMQDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same
	// clusterName used in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter. An Arbiter DC holds only the dr-controlplane etcd member and no RabbitMQ.
	// +optional
	Role string `json:"role,omitempty"`

	// Writable is true when this DC is the active publish cluster (its produce fence is open).
	// +optional
	Writable bool `json:"writable,omitempty"`

	// NodesReady is the number of ready nodes in this DC's local RabbitMQ cluster.
	// +optional
	NodesReady *int32 `json:"nodesReady,omitempty"`

	// FederationLagMessages is this DC's cross-DC Federation (or Shovel) replication backlog
	// behind the active DC, in messages (the upstream vs downstream position gap). Nil when this
	// DC is the active publish cluster.
	// +optional
	FederationLagMessages *int64 `json:"federationLagMessages,omitempty"`

	// Healthy reflects whether this DC's health Lease is fresh.
	// +optional
	Healthy bool `json:"healthy,omitempty"`
}

// +kubebuilder:validation:Enum=ca;client;server
type RabbitMQCertificateAlias string

const (
	RabbitmqCACert     RabbitMQCertificateAlias = "ca"
	RabbitmqClientCert RabbitMQCertificateAlias = "client"
	RabbitmqServerCert RabbitMQCertificateAlias = "server"
)

// +kubebuilder:validation:Enum=http;amqp;mqtt;stomp;web_mqtt;web_stomp
type RabbitMQProtocol string

const (
	RabbitmqProtocolHTTP     RabbitMQProtocol = "http"
	RabbitmqProtocolAMQP     RabbitMQProtocol = "amqp"
	RabbitmqProtocolMQTT     RabbitMQProtocol = "mqtt"
	RabbitmqProtocolSTOMP    RabbitMQProtocol = "stomp"
	RabbitmqProtocolWEBMQTT  RabbitMQProtocol = "web_mqtt"
	RabbitmqProtocolWEBSTOMP RabbitMQProtocol = "web_stomp"
)

// RabbitMQList contains a list of RabbitMQ

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RabbitMQList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []RabbitMQ `json:"items"`
}

var _ Accessor = &RabbitMQ{}

func (m *RabbitMQ) GetObjectMeta() meta.ObjectMeta {
	return m.ObjectMeta
}

func (m *RabbitMQ) GetConditions() []kmapi.Condition {
	return m.Status.Conditions
}

func (m *RabbitMQ) SetCondition(cond kmapi.Condition) {
	m.Status.Conditions = setCondition(m.Status.Conditions, cond)
}

func (m *RabbitMQ) RemoveCondition(typ string) {
	m.Status.Conditions = removeCondition(m.Status.Conditions, typ)
}
