package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeHanaDBVersion     = "hanaversion"
	ResourceKindHanaDBVersion     = "HanaDBVersion"
	ResourceSingularHanaDBVersion = "hanadbversion"
	ResourcePluralHanaDBVersion   = "hanadbversions"
)

// HanaDBVersion defines a HanaDB database version

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hanadbversions,singular=hanadbversion,scope=Cluster,shortName=hanaversion,categories={catalog,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type HanaDBVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HanaDBVersionSpec `json:"spec,omitempty"`
}

type HanaDBVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Database Image
	DB HanaDBVersionDatabase `json:"db"`
	// Init container
	// InitContainer HanaDBInitContainer `json:"initContainer,omitempty"`
	// Deprecated versions usable but considered as obsolete and best avoided typically superseded
	Deprecated bool `json:"deprecated,omitempty"`
	// +optional
	UI []ChartInfo `json:"ui,omitempty"`
	// update constraints
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// HanaDBVersionDatabase is the HanaDB Database image

type HanaDBVersionDatabase struct {
	Image string `json:"image"`
}

//// HanaDBInitContainer is the Qdrant init Container image
//type HanaDBInitContainer struct {
//	Image string `json:"image"`
//}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HanaDBVersionList is a list of HanaDBVersions
type HanaDBVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HanaDBVersion `json:"items,omitempty"`
}
