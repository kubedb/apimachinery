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

//go:generate go-enum --mustparse --names --values
package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeDurableWorkflowOpsRequest     = "dwops"
	ResourceKindDurableWorkflowOpsRequest     = "DurableWorkflowOpsRequest"
	ResourceSingularDurableWorkflowOpsRequest = "durableworkflowopsrequest"
	ResourcePluralDurableWorkflowOpsRequest   = "durableworkflowopsrequests"
)

// DurableWorkflowOpsRequest defines a DurableWorkflow DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=durableworkflowopsrequests,singular=durableworkflowopsrequest,shortName=dwops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DurableWorkflowOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DurableWorkflowOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus              `json:"status,omitempty"`
}

// DurableWorkflowOpsRequestSpec is the spec for DurableWorkflowOpsRequest
type DurableWorkflowOpsRequestSpec struct {
	// Specifies the DurableWorkflow reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type DurableWorkflowOpsRequestType `json:"type"`
	// Specifies information necessary for restarting the cluster
	Restart *DurableWorkflowRestartSpec `json:"restart,omitempty"`
	// Specifies information necessary for upgrading the cluster
	UpdateVersion *DurableWorkflowUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *DurableWorkflowHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *DurableWorkflowVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for custom configuration
	Configuration *ReconfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *TLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for rotating the credentials of the persistence stores
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// DurableWorkflowRestartSpec selects the roles to restart. All roles are restarted if none is selected.
type DurableWorkflowRestartSpec struct {
	// +optional
	Roles []DurableWorkflowRole `json:"roles,omitempty"`
}

// DurableWorkflowUpdateVersionSpec contains the update version information
type DurableWorkflowUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
	// AllowDowntime allows a hop whose schema update can't run online, and a non-HA cluster.
	// +optional
	AllowDowntime bool `json:"allowDowntime,omitempty"`
	// Surge lists the roles that are scaled up by one pod while they are rolled. By default, every role
	// except history is surged.
	// +optional
	Surge map[DurableWorkflowRole]bool `json:"surge,omitempty"`
}

// DurableWorkflowHorizontalScalingSpec contains the horizontal scaling information
type DurableWorkflowHorizontalScalingSpec struct {
	// +kubebuilder:validation:Minimum=1
	// +optional
	Frontend *int32 `json:"frontend,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +optional
	History *int32 `json:"history,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +optional
	Matching *int32 `json:"matching,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +optional
	Worker *int32 `json:"worker,omitempty"`
	// +kubebuilder:validation:Minimum=1
	// +optional
	InternalFrontend *int32 `json:"internalFrontend,omitempty"`
}

// DurableWorkflowVerticalScalingSpec contains the vertical scaling information
type DurableWorkflowVerticalScalingSpec struct {
	// +optional
	Frontend *PodResources `json:"frontend,omitempty"`
	// +optional
	History *PodResources `json:"history,omitempty"`
	// +optional
	Matching *PodResources `json:"matching,omitempty"`
	// +optional
	Worker *PodResources `json:"worker,omitempty"`
	// +optional
	InternalFrontend *PodResources `json:"internalFrontend,omitempty"`
}

// +kubebuilder:validation:Enum=frontend;history;matching;worker;internalFrontend
type DurableWorkflowRole string

const (
	DurableWorkflowRoleFrontend         DurableWorkflowRole = "frontend"
	DurableWorkflowRoleHistory          DurableWorkflowRole = "history"
	DurableWorkflowRoleMatching         DurableWorkflowRole = "matching"
	DurableWorkflowRoleWorker           DurableWorkflowRole = "worker"
	DurableWorkflowRoleInternalFrontend DurableWorkflowRole = "internalFrontend"
)

// +kubebuilder:validation:Enum=Restart;ReconfigureTLS;RotateAuth;Reconfigure;HorizontalScaling;VerticalScaling;UpdateVersion
// ENUM(Restart,ReconfigureTLS,RotateAuth,Reconfigure,HorizontalScaling,VerticalScaling,UpdateVersion)
type DurableWorkflowOpsRequestType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DurableWorkflowOpsRequestList is a list of DurableWorkflowOpsRequests
type DurableWorkflowOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DurableWorkflowOpsRequest CRD objects
	Items []DurableWorkflowOpsRequest `json:"items,omitempty"`
}
