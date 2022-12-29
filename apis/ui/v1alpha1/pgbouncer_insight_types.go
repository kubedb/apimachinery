package v1alpha1

import (
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindPgBouncerInsight = "PgBouncerInsight"
	ResourcePgBouncerInsight     = "pgbouncerinsight"
	ResourcePgBouncerInsights    = "pgbouncerinsights"
)

type PgBouncerInsightSpec struct {
	Version        string                `json:"version"`
	Status         string                `json:"status"`
	SSLMode        api.SSLMode           `json:"sslMode,omitempty"`
	MaxConnections *int32                `json:"maxConnections,omitempty"`
	PodInsights    []PgBouncerPodInsight `json:"podInsights,omitempty"`
}

type PgBouncerPodInsight struct {
	PodName           string `json:"podName"`
	Databases         *int32 `json:"databases,omitempty"`
	Users             *int32 `json:"users,omitempty"`
	Pools             *int32 `json:"pools,omitempty"`
	FreeClients       *int32 `json:"freeClients,omitempty"`
	UsedClients       *int32 `json:"usedClients,omitempty"`
	LoginClient       *int32 `json:"loginClient,omitempty"`
	FreeServers       *int32 `json:"freeServers,omitempty"`
	UsedServers       *int32 `json:"usedServers,omitempty"`
	TotalQueryCount   *int32 `json:"totalQueryCount,omitempty"`
	AverageQueryCount *int32 `json:"averageQueryCount,omitempty"`
	TotalQueryTime    *int32 `json:"totalQueryTime,omitempty"`
	AverageQueryTime  *int32 `json:"averageQueryTime,omitempty"`
}

// PgBouncerInsight is the Schema for the pgbouncerinsights API
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PgBouncerInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgBouncerInsightSpec `json:"spec,omitempty"`
	Status api.PgBouncerStatus  `json:"status,omitempty"`
}

// PgBouncerInsightList contains a list of PgBouncerInsight
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PgBouncerInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PgBouncerInsight `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PgBouncerInsight{}, &PgBouncerInsightList{})
}
