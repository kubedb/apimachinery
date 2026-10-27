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
	ResourceCodeYugabyteDBVersion     = "ybversion"
	ResourceKindYugabyteDBVersion     = "YugabyteDBVersion"
	ResourceSingularYugabyteDBVersion = "yugabytedbversion"
	ResourcePluralYugabyteDBVersion   = "yugabytedbversions"
)

// YugabyteDBVersion defines a YugabyteDB database version.

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=yugabytedbversions,singular=yugabytedbversion,scope=Cluster,shortName=ybversion,categories={catalog,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type YugabyteDBVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              YugabyteDBVersionSpec `json:"spec,omitempty"`
}

// YugabyteDBVersionSpec is the spec for a YugabyteDB version.
type YugabyteDBVersionSpec struct {
	// Version
	Version string `json:"version"`

	// Build is the upstream build of the image, for example 2025.2.8.0-b60.
	// +optional
	Build string `json:"build,omitempty"`

	// EndOfLife refers if this version reached into its end of the life or not, based on https://endoflife.date/
	// +optional
	EndOfLife bool `json:"endOfLife"`

	// Database Image
	DB YugabyteDBVersionDatabase `json:"db"`

	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`

	// SecurityContext is for the additional config for the DB container
	// +optional
	SecurityContext SecurityContext `json:"securityContext"`

	// YSQLMajorVersion is the major version of PostgreSQL that the YSQL API is built on, for example 15.
	// An update between versions with a different major version needs the YSQL major upgrade flow.
	YSQLMajorVersion int32 `json:"ysqlMajorVersion"`

	// Memory sizes --memory_limit_hard_bytes from the container memory limit of each role.
	// +optional
	Memory YugabyteDBVersionMemory `json:"memory,omitempty"`

	// DefaultFlags are written to the flag files before the user configuration.
	// +optional
	DefaultFlags YugabyteDBVersionFlags `json:"defaultFlags,omitempty"`

	// Features lists what this version supports.
	// +optional
	Features YugabyteDBVersionFeatures `json:"features,omitempty"`

	// +optional
	UI []ChartInfo `json:"ui,omitempty"`

	// update constraints
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// YugabyteDBVersionDatabase is the YugabyteDB Database image
type YugabyteDBVersionDatabase struct {
	Image string `json:"image"`
}

// YugabyteDBVersionMemory sets the share of the container memory limit given to each role.
type YugabyteDBVersionMemory struct {
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=85
	MasterLimitPercent int32 `json:"masterLimitPercent,omitempty"`
	// The tserver keeps the rest for the PostgreSQL backends, which run outside its memory tracker.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=70
	TServerLimitPercent int32 `json:"tserverLimitPercent,omitempty"`
}

// YugabyteDBVersionFlags are the default flags of each role.
type YugabyteDBVersionFlags struct {
	// +optional
	Master map[string]string `json:"master,omitempty"`
	// +optional
	TServer map[string]string `json:"tserver,omitempty"`
}

// YugabyteDBVersionFeatures lists the capabilities of a version.
type YugabyteDBVersionFeatures struct {
	// ClientOnlyCertReload is true when certificates can be reloaded without a restart even if only
	// the client to node TLS is on (2025.2.1 and later).
	// +optional
	ClientOnlyCertReload bool `json:"clientOnlyCertReload,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// YugabyteDBVersionList is a list of YugabyteDBVersions
type YugabyteDBVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of YugabyteDBVersion CRD objects
	Items []YugabyteDBVersion `json:"items,omitempty"`
}
