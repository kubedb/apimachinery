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
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceCodeWorkflowNamespace     = "wfns"
	ResourceKindWorkflowNamespace     = "WorkflowNamespace"
	ResourceSingularWorkflowNamespace = "workflownamespace"
	ResourcePluralWorkflowNamespaces  = "workflownamespaces"

	// DefaultWorkflowNamespaceRetention is the retention of the closed workflows when spec.retention is not set.
	DefaultWorkflowNamespaceRetention = 72 * time.Hour
)

// WorkflowNamespace is a Temporal namespace or a Cadence domain of a DurableWorkflow. The operator creates it in
// the engine, keeps its retention, description and archival in sync, and deletes it according to
// spec.deletionPolicy.
//
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=workflownamespaces,singular=workflownamespace,shortName=wfns,categories={kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseRef.name"
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".status.name"
// +kubebuilder:printcolumn:name="Retention",type="string",JSONPath=".spec.retention"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type WorkflowNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowNamespaceSpec   `json:"spec,omitempty"`
	Status WorkflowNamespaceStatus `json:"status,omitempty"`
}

// WorkflowNamespaceSpec defines the desired state of WorkflowNamespace.
type WorkflowNamespaceSpec struct {
	// DatabaseRef is the DurableWorkflow, in the same namespace, that holds the Temporal namespace or Cadence domain.
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`

	// Name is the name of the Temporal namespace or Cadence domain. It defaults to metadata.name and is immutable.
	// +optional
	Name string `json:"name,omitempty"`

	// Description of the namespace or domain.
	// +optional
	Description string `json:"description,omitempty"`

	// OwnerEmail is the contact of the owner. Cadence requires it; Temporal stores it as given.
	// +optional
	OwnerEmail string `json:"ownerEmail,omitempty"`

	// Retention is how long the closed workflow executions are kept. Defaults to 72h. Cadence keeps whole days
	// (at least 1d): the operator rounds up to the next day.
	// +optional
	Retention *metav1.Duration `json:"retention,omitempty"`

	// Archival turns the archival of the closed workflows on. It needs archival in the spec of the DurableWorkflow.
	// +optional
	Archival *WorkflowNamespaceArchival `json:"archival,omitempty"`

	// TokenSecretRef is a Secret, in the same namespace, whose key "token" holds a JWT that the operator sends
	// to the frontend. Needed only when the DurableWorkflow has authorization enabled: it must carry the admin
	// permission for Temporal (for example "default:admin") and the admin claim for Cadence.
	// +optional
	TokenSecretRef *core.LocalObjectReference `json:"tokenSecretRef,omitempty"`

	// DeletionPolicy controls what happens in the engine when this object is deleted. Delete (the default)
	// deletes the namespace or domain, which drops all its workflows and their histories. DoNotTerminate
	// refuses the deletion of the object.
	// +kubebuilder:validation:Enum=Delete;DoNotTerminate
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`
}

// WorkflowNamespaceArchival is the archival of a namespace or domain.
type WorkflowNamespaceArchival struct {
	// Enabled archives the histories and the visibility records of the closed workflows to the URIs in the
	// archival of the DurableWorkflow.
	// +optional
	Enabled bool `json:"enabled,omitempty"`
}

// +kubebuilder:validation:Enum=Pending;Ready;Failed;Terminating
type WorkflowNamespacePhase string

const (
	WorkflowNamespacePhasePending     WorkflowNamespacePhase = "Pending"
	WorkflowNamespacePhaseReady       WorkflowNamespacePhase = "Ready"
	WorkflowNamespacePhaseFailed      WorkflowNamespacePhase = "Failed"
	WorkflowNamespacePhaseTerminating WorkflowNamespacePhase = "Terminating"
)

const (
	// WorkflowNamespaceConditionReady is true when the namespace exists in the engine with the requested settings.
	WorkflowNamespaceConditionReady = "Ready"
)

// WorkflowNamespaceStatus defines the observed state of WorkflowNamespace.
type WorkflowNamespaceStatus struct {
	// +optional
	Phase WorkflowNamespacePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this resource.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// Name is the name used in the engine.
	// +optional
	Name string `json:"name,omitempty"`

	// ID is the id the engine reported for the namespace or domain.
	// +optional
	ID string `json:"id,omitempty"`

	// Retention is the retention in force in the engine.
	// +optional
	Retention *metav1.Duration `json:"retention,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowNamespaceList contains a list of WorkflowNamespace.
type WorkflowNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkflowNamespace `json:"items"`
}
