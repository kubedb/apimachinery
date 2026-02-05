//go:generate go-enum --mustparse --names --values
package v1alpha1

import (
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
	// Specifies information necessary for configuring TLS
	TLS *TLSSpec `json:"tls,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Neo4jOpsRequestList is a list of Neo4jOpsRequests
type Neo4jOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of Neo4jOpsRequest CRD objects
	Items []Neo4jOpsRequest `json:"items,omitempty"`
}
