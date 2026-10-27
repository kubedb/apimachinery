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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeNATSVersion     = "natsversion"
	ResourceKindNATSVersion     = "NATSVersion"
	ResourceSingularNATSVersion = "natsversion"
	ResourcePluralNATSVersion   = "natsversions"
)

// NATSVersion defines a NATS server version.

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=natsversions,singular=natsversion,scope=Cluster,shortName=natsversion,categories={catalog,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NATSVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NATSVersionSpec `json:"spec,omitempty"`
}

// NATSVersionSpec is the spec for NATS version.
//
// There is no init container image: the zone-tag init container runs from db.image.
type NATSVersionSpec struct {
	// Version
	Version string `json:"version"`

	// EndOfLife refers if this version reached into its end of the life or not, based on https://endoflife.date/
	// +optional
	EndOfLife bool `json:"endOfLife"`

	// Database Image (nats-server)
	DB NATSVersionDatabase `json:"db"`

	// Reloader Image (nats-server-config-reloader sidecar)
	Reloader NATSVersionReloader `json:"reloader"`

	// Exporter Image (prometheus-nats-exporter)
	Exporter NATSVersionExporter `json:"exporter"`

	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`

	// SecurityContext is for the additional config for the DB container
	// +optional
	SecurityContext NATSSecurityContext `json:"securityContext"`

	// LameDuck holds the default lame-duck timing, used when spec.lameDuck of a NATS is not set.
	// +optional
	LameDuck NATSVersionLameDuck `json:"lameDuck,omitempty"`

	// update constraints
	// +optional
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// NATSVersionDatabase is the NATS server image
type NATSVersionDatabase struct {
	Image string `json:"image"`
}

// NATSVersionReloader is the image of the config reloader sidecar
type NATSVersionReloader struct {
	Image string `json:"image"`
}

// NATSVersionExporter is the image for the NATS exporter
type NATSVersionExporter struct {
	Image string `json:"image"`
}

// NATSSecurityContext is the additional features for the NATS DB container
type NATSSecurityContext struct {
	// RunAsUser is the default UID for the DB container, also used as the group id and fsGroup.
	// +optional
	RunAsUser *int64 `json:"runAsUser,omitempty"`
}

// NATSVersionLameDuck is the default lame-duck timing.
type NATSVersionLameDuck struct {
	// GracePeriod must be < Duration.
	// +optional
	GracePeriod metav1.Duration `json:"gracePeriod,omitempty"`
	// Duration must be >= 30s, the server minimum.
	// +optional
	Duration metav1.Duration `json:"duration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NATSVersionList is a list of NATSVersions
type NATSVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of NATSVersion CRD objects
	Items []NATSVersion `json:"items,omitempty"`
}
