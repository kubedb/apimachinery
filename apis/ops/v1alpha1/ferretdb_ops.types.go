//go:generate go-enum --mustparse --names --values
package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apis "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
)

const (
	ResourceCodeFerretDBOpsRequest     = "frops"
	ResourceKindFerretDBOpsRequest     = "FerretDBOpsRequest"
	ResourceSingularFerretDBOpsRequest = "ferretdbopsrequest"
	ResourcePluralFerretDBOpsRequest   = "ferretdbopsrequests"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=ferretdbopsrequests,singular=ferretdbopsrequest,shortName=frops,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type FerretDBOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FerretDBOpsRequestSpec `json:"spec,omitempty"`
	Status            OpsRequestStatus       `json:"status,omitempty"`
}

// FerretDBOpsRequestSpec is the spec for FerretDBOpsRequest
type FerretDBOpsRequestSpec struct {
	// Specifies the FerretDB reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type FerretDBOpsRequestType `json:"type"`
	// Specifies information necessary for upgrading ferretdb
	UpdateVersion *FerretDBUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *FerretDBHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *FerretDBVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *FerretDBTLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
}

type FerretDBTLSSpec struct {
	TLSSpec `json:",inline,omitempty"`

	// SSLMode for both standalone and clusters. [disabled;requireSSL]
	// +optional
	SSLMode apis.SSLMode `json:"sslMode,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;VerticalScaling;Restart;HorizontalScaling;ReconfigureTLS
// ENUM(UpdateVersion, Restart, VerticalScaling, HorizontalScaling, ReconfigureTLS)
type FerretDBOpsRequestType string

// FerretDBUpdateVersionSpec contains the update version information of a ferretdb cluster
type FerretDBUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// FerretDBHorizontalScalingSpec contains the horizontal scaling information of a ferretdb cluster
type FerretDBHorizontalScalingSpec struct {
	// Number of node
	Node *int32 `json:"node,omitempty"`
}

// FerretDBVerticalScalingSpec contains the vertical scaling information of a ferretdb cluster
type FerretDBVerticalScalingSpec struct {
	// Resource spec for nodes
	Node *PodResources `json:"node,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FerretDBOpsRequestList is a list of FerretDBOpsRequests
type FerretDBOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of FerretDBOpsRequest CRD objects
	Items []FerretDBOpsRequest `json:"items,omitempty"`
}
