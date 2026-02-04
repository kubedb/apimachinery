/*
Copyright 2026.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

// Migrator is the Schema for the migrators API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=migrators,singular=migrator,shortName=mgtr,categories={kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="DBType",type="string",JSONPath=".status.progress.dbType"
// +kubebuilder:printcolumn:name="Stage",type="string",JSONPath=".status.progress.details.Stage"
// +kubebuilder:printcolumn:name="Lag",type="string",JSONPath=".status.progress.details.Lag"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Migrator struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Migrator
	// +required
	Spec MigratorSpec `json:"spec"`

	// status defines the observed state of Migrator
	// +optional
	Status MigratorStatus `json:"status,omitzero"`
}

// MigratorSpec defines the desired state of Migrator
type MigratorSpec struct {
	// Source defines the source database configuration
	Source *Source `json:"source" protobuf:"bytes,1,opt,name=source"`

	// Target defines the target database configuration
	Target *Target `json:"target" protobuf:"bytes,2,opt,name=target"`

	// JobTemplate specifies runtime configurations for the backup/restore Job
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`
}

// MigratorStatus defines the observed state of Migrator.
type MigratorStatus struct {
	// Phase represents the current phase of migration
	// +optional
	Phase MigratorPhase `json:"phase,omitempty"`

	// Progress contains the current progress of migration
	// +optional
	Progress *Progress `json:"progress,omitempty"`

	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// Progress contains the current progress of migration
type Progress struct {
	// DBType indicates the type of database
	// +optional
	DBType string `json:"dbType,omitempty"`

	// Phase indicates the current phase of migration
	// +optional
	Phase string `json:"phase,omitempty"`

	// Details contains additional information about the current phase
	// +optional
	Details map[string]string `json:"details,omitempty"`
}

// MigratorList contains a list of Migrator

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MigratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Migrator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Migrator{}, &MigratorList{})
}
