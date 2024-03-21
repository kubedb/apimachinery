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
	ResourceCodeMsSQLOpsRequest     = "msops"
	ResourceKindMsSQLOpsRequest     = "MsSQLOpsRequest"
	ResourceSingularMsSQLOpsRequest = "mssqlopsrequest"
	ResourcePluralMsSQLOpsRequest   = "mssqlopsrequests"
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
type MsSQLOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MsSQLOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus    `json:"status,omitempty"`
}

// MsSQLOpsRequestSpec is the spec for MsSQLOpsRequest
type MsSQLOpsRequestSpec struct {
	// Specifies the MsSQL reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type MsSQLOpsRequestType `json:"type"`
	// Specifies information necessary for upgrading MsSQL
	UpdateVersion *MsSQLUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *MsSQLHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *MsSQLVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *MsSQLVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of MsSQL
	Configuration *MsSQLCustomConfigurationSpec `json:"configuration,omitempty"`
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
type MsSQLOpsRequestType string

// MsSQLReplicaReadinessCriteria is the criteria for checking readiness of a MsSQL pod
// after updating, horizontal scaling etc.
type MsSQLReplicaReadinessCriteria struct{}

// MsSQLUpdateVersionSpec contains the update version information of a MsSQL cluster
type MsSQLUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// MsSQLHorizontalScalingSpec contains the horizontal scaling information of a MsSQL cluster
type MsSQLHorizontalScalingSpec struct{}

// MsSQLVerticalScalingSpec contains the vertical scaling information of a MsSQL cluster
type MsSQLVerticalScalingSpec struct{}

// MsSQLVolumeExpansionSpec is the spec for MsSQL volume expansion
type MsSQLVolumeExpansionSpec struct{}

// MsSQLCustomConfigurationSpec is the spec for Reconfiguring the MsSQL Settings
type MsSQLCustomConfigurationSpec struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MsSQLOpsRequestList is a list of MsSQLOpsRequests
type MsSQLOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of MsSQLOpsRequest CRD objects
	Items []MsSQLOpsRequest `json:"items,omitempty"`
}
