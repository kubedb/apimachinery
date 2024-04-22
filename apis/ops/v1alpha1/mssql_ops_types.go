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
	ResourceCodeMSSQLOpsRequest     = "msops"
	ResourceKindMSSQLOpsRequest     = "MSSQLOpsRequest"
	ResourceSingularMSSQLOpsRequest = "mssqlopsrequest"
	ResourcePluralMSSQLOpsRequest   = "mssqlopsrequests"
)

// MsSDBOpsRequest defines a MsS DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqlopsrequests,singular=mssqlopsrequest,shortName=msops,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MSSQLOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MSSQLOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus    `json:"status,omitempty"`
}

// MSSQLOpsRequestSpec is the spec for MSSQLOpsRequest
type MSSQLOpsRequestSpec struct {
	// Specifies the MSSQL reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type MSSQLOpsRequestType `json:"type"`
	// Specifies information necessary for upgrading MSSQL
	UpdateVersion *MSSQLUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *MSSQLHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *MSSQLVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *MSSQLVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of MSSQL
	Configuration *MSSQLCustomConfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *TLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;Reconfigure;ReconfigureTLS
// ENUM(UpdateVersion, HorizontalScaling, VerticalScaling, VolumeExpansion, Restart, Reconfigure, ReconfigureTLS)
type MSSQLOpsRequestType string

// MSSQLReplicaReadinessCriteria is the criteria for checking readiness of a MSSQL pod
// after updating, horizontal scaling etc.
type MSSQLReplicaReadinessCriteria struct{}

// MSSQLUpdateVersionSpec contains the update version information of a MSSQL cluster
type MSSQLUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// MSSQLHorizontalScalingSpec contains the horizontal scaling information of a MSSQL cluster
type MSSQLHorizontalScalingSpec struct{}

// MSSQLVerticalScalingSpec contains the vertical scaling information of a MSSQL cluster
type MSSQLVerticalScalingSpec struct{}

// MSSQLVolumeExpansionSpec is the spec for MSSQL volume expansion
type MSSQLVolumeExpansionSpec struct{}

// MSSQLCustomConfigurationSpec is the spec for Reconfiguring the MSSQL Settings
type MSSQLCustomConfigurationSpec struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MSSQLOpsRequestList is a list of MSSQLOpsRequests
type MSSQLOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of MSSQLOpsRequest CRD objects
	Items []MSSQLOpsRequest `json:"items,omitempty"`
}
