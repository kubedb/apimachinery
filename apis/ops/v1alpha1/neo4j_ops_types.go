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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeNeo4jOpsRequest     = "neoops"
	ResourceKindNeo4jOpsRequest     = "Neo4jOpsRequest"
	ResourceSingularNeo4jOpsRequest = "neo4jopsrequest"
	ResourcePluralNeo4jOpsRequest   = "neo4jopsrequests"
)

// Neo4jDBOpsRequest defines a Neo4j DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=neo4jopsrequests,singular=neo4jopsrequest,shortName=neoops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Neo4jOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Neo4jOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus    `json:"status,omitempty"`
}

// Neo4jOpsRequestSpec is the spec for Neo4jOpsRequest
type Neo4jOpsRequestSpec struct {
	// Specifies the Neo4j reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type Neo4jOpsRequestType `json:"type"`
	// Specifies information necessary for upgrading Neo4j
	UpdateVersion *Neo4jUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *Neo4jHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *Neo4jVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *Neo4jVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of Neo4j
	Configuration *ReconfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *TLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for configuring authSecret of the database
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;Reconfigure;ReconfigureTLS;RotateAuth
// ENUM(UpdateVersion, HorizontalScaling, VerticalScaling, VolumeExpansion, Restart, Reconfigure, ReconfigureTLS, RotateAuth)
type Neo4jOpsRequestType string

// Neo4jReplicaReadinessCriteria is the criteria for checking readiness of a Neo4j pod
// after updating, horizontal scaling etc.
type Neo4jReplicaReadinessCriteria struct{}

// Neo4jUpdateVersionSpec contains the update version information of a Neo4j cluster
type Neo4jUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// Neo4jHorizontalScalingSpec contains the horizontal scaling information of a Neo4j cluster
type Neo4jHorizontalScalingSpec struct {
	// Number of servers
	Server *int32 `json:"server,omitempty"`
}

// Neo4jVerticalScalingSpec contains the vertical scaling information of a Neo4j cluster
type Neo4jVerticalScalingSpec struct {
	// Resource spec for servers
	Server *PodResources `json:"server,omitempty"`
}

// Neo4jVolumeExpansionSpec is the spec for Neo4j volume expansion
type Neo4jVolumeExpansionSpec struct {
	Mode VolumeExpansionMode `json:"mode"`
	// volume specification for servers
	Server *resource.Quantity `json:"server,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Neo4jOpsRequestList is a list of Neo4jOpsRequests
type Neo4jOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of Neo4jOpsRequest CRD objects
	Items []Neo4jOpsRequest `json:"items,omitempty"`
}
