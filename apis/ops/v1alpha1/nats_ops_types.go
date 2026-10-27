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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeNATSOpsRequest     = "natsops"
	ResourceKindNATSOpsRequest     = "NATSOpsRequest"
	ResourceSingularNATSOpsRequest = "natsopsrequest"
	ResourcePluralNATSOpsRequest   = "natsopsrequests"
)

// NATSOpsRequest defines a NATS DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=natsopsrequests,singular=natsopsrequest,shortName=natsops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NATSOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NATSOpsRequestSpec `json:"spec,omitempty"`
	Status OpsRequestStatus   `json:"status,omitempty"`
}

// NATSOpsRequestSpec is the spec for NATSOpsRequest
type NATSOpsRequestSpec struct {
	// Specifies the NATS reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type NATSOpsRequestType `json:"type"`
	// Specifies information necessary for updating NATS version
	UpdateVersion *NATSUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *NATSHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *NATSVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *NATSVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of NATS
	Configuration *NATSCustomConfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *NATSTLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for configuring authSecret of the database
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Specifies information necessary for migrating storageClass of the PVCs
	Migration *StorageMigrationSpec `json:"migration,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;Reconfigure;ReconfigureTLS;RotateAuth;StorageMigration
// ENUM(UpdateVersion, HorizontalScaling, VerticalScaling, VolumeExpansion, Restart, Reconfigure, ReconfigureTLS, RotateAuth, StorageMigration)
type NATSOpsRequestType string

// NATSUpdateVersionSpec contains the update version information of a NATS cluster
type NATSUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// NATSHorizontalScalingSpec contains the horizontal scaling information of a NATS cluster.
// Scaling down evacuates the highest ordinal from the JetStream meta group before the pod is removed.
type NATSHorizontalScalingSpec struct {
	// Number of nats-server pods in the cluster
	Replicas *int32 `json:"replicas,omitempty"`
}

// NATSVerticalScalingSpec is the spec for NATS vertical scaling
type NATSVerticalScalingSpec struct {
	// Resource spec for the nats-server container
	NATS *ContainerResources `json:"nats,omitempty"`
	// Resource spec for the config reloader sidecar container
	Reloader *ContainerResources `json:"reloader,omitempty"`
	// Resource spec for the exporter sidecar container
	Exporter *ContainerResources `json:"exporter,omitempty"`
	// Mode selects how the vertical scaling is actuated. Defaults to Restart.
	// A memory change always needs a restart, because GOMEMLIMIT is read at process start.
	// +optional
	// +kubebuilder:default=Restart
	Mode VerticalScalingMode `json:"mode,omitempty"`
}

// NATSVolumeExpansionSpec is the spec for NATS volume expansion
type NATSVolumeExpansionSpec struct {
	// volume specification for the JetStream file store (/data)
	NATS *resource.Quantity  `json:"nats,omitempty"`
	Mode VolumeExpansionMode `json:"mode"`
}

// NATSCustomConfigurationSpec is the NATS-specific reconfiguration spec.
// It embeds the generic ReconfigurationSpec and adds the typed server settings.
// ApplyConfig keys are nats.conf and accounts.conf, as in spec.configuration.
type NATSCustomConfigurationSpec struct {
	ReconfigurationSpec `json:",inline,omitempty"`

	// ServerConfig holds the typed server limits and timeouts to apply.
	// +optional
	ServerConfig *dbapi.NATSServerConfig `json:"serverConfig,omitempty"`

	// LameDuck changes the lame-duck timing. It is not reloadable and rolls the pods.
	// +optional
	LameDuck *dbapi.NATSLameDuckSpec `json:"lameDuck,omitempty"`
}

// NATSTLSSpec is the spec for reconfiguring the NATS TLS configuration.
// The embedded TLSSpec already carries `rotateCertificates` and `remove`.
type NATSTLSSpec struct {
	TLSSpec `json:",inline,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATSOpsRequestList is a list of NATSOpsRequests
type NATSOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of NATSOpsRequest CRD objects
	Items []NATSOpsRequest `json:"items,omitempty"`
}
