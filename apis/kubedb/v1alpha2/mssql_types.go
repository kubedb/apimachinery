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
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeMsSQL     = "ms"
	ResourceKindMsSQL     = "MsSQL"
	ResourceSingularMsSQL = "mssql"
	ResourcePluralMsSQL   = "mssqls"
)

// +kubebuilder:validation:Enum=Standalone;AvailabilityGroup
type MsSQLMode string

const (
	MsSQLModeStandalone        MsSQLMode = "Standalone"
	MsSQLModeAvailabilityGroup MsSQLMode = "AvailabilityGroup"
	MsSQLModeRemoteReplica     MsSQLMode = "RemoteReplica"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MsSQL defines a MsSQL database.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqls,singular=mssql,shortName=ms,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MsSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MsSQLSpec   `json:"spec,omitempty"`
	Status MsSQLStatus `json:"status,omitempty"`
}

// MsSQLSpec defines the desired state of MsSQL
type MsSQLSpec struct {
	// Version of MsSQL to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a MsSQL database. In case of MsSQL Availability Group (default 3).
	Replicas *int32 `json:"replicas,omitempty"`

	// MsSQL cluster topology
	Topology *MsSQLTopology `json:"topology,omitempty"` // ag or standalone

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage spec to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// TerminationPolicy controls the delete operation for database
	// +optional
	TerminationPolicy TerminationPolicy `json:"terminationPolicy,omitempty"`

	// Coordinator defines attributes of the coordinator container
	// +optional
	Coordinator CoordinatorSpec `json:"coordinator,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 1}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

type MsSQLTopology struct {
	// If set to -
	// "AvailabilityGroup", MsSQLAvailabilityGroupSpec is required and MsSQL servers will start an Availability Group
	Mode *MsSQLMode `json:"mode,omitempty"`

	// AvailabilityGroup info for MsSQL
	// +optional
	AvailabilityGroup *MsSQLAvailabilityGroupSpec `json:"availabilityGroup,omitempty"`
}

// MsSQLAvailabilityGroupSpec defines the availability group spec for MsSQL
type MsSQLAvailabilityGroupSpec struct {
	// AvailabilityDatabases is an array of databases to be included in the availability group
	AvailabilityDatabases []string `json:"databases"`
}

// MsSQLStatus defines the observed state of MsSQL
type MsSQLStatus struct {
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MsSQLList contains a list of MsSQL
type MsSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MsSQL `json:"items"`
}
