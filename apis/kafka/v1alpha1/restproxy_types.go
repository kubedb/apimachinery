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

package v1alpha1

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeRestProxy     = "krp"
	ResourceKindRestProxy     = "RestProxy"
	ResourceSingularRestProxy = "restproxy"
	ResourcePluralRestProxy   = "restproxies"
)

// RestProxy defines a  runtime server system that stores a specific set of artifacts as files.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=restproxies,singular=restproxy,shortName=krp,categories={kfstore,kubedb,appscode}
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".apiVersion"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Kafka",type="string",JSONPath=".spec.kafkaRef.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RestProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestProxySpec   `json:"spec,omitempty"`
	Status RestProxyStatus `json:"status,omitempty"`
}

// RestProxySpec defines the desired state of RestProxy
type RestProxySpec struct {
	// Version of RestProxy to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a rest proxy.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Kafka app-binding reference
	// KafkaRef is a required field, where RestProxy will connect to Kafka
	KafkaRef *kmapi.ObjectReference `json:"kafkaRef"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// SchemaRegistryRef provides a reference to the Schema Registry configuration.
	// the REST Proxy will connect to the Schema Registry if SchemaRegistryRef is provided.
	// +optional
	SchemaRegistryRef *SchemaRegistryRef `json:"schemaRegistryRef,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []dbapi.NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy dbapi.DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

// SchemaRegistryRef provides a reference to the Schema Registry configuration.
type SchemaRegistryRef struct {
	// Name and namespace of appbinding of schema registry
	// If this is provided, the REST Proxy will connect to the Schema Registry
	// InternallyManaged must be set to false in this case
	// +optional
	*ObjectReference `json:",omitempty"`

	// InternallyManaged true specifies if the schema registry runs internally along with the rest proxy
	// +optional
	InternallyManaged bool `json:"internallyManaged,omitempty"`
}

// RestProxyStatus defines the observed state of RestProxy
type RestProxyStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase RestProxyPhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:validation:Enum=Provisioning;Ready;NotReady;Critical;Unknown
type RestProxyPhase string

const (
	RestProxyPhaseProvisioning RestProxyPhase = "Provisioning"
	RestProxyPhaseReady        RestProxyPhase = "Ready"
	RestProxyPhaseNotReady     RestProxyPhase = "NotReady"
	RestProxyPhaseCritical     RestProxyPhase = "Critical"
	RestProxyPhaseUnknown      RestProxyPhase = "Unknown"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RestProxyList contains a list of RestProxy
type RestProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RestProxy `json:"items"`
}
