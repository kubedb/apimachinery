package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindPgBouncerPools = "PgBouncerPools"
	ResourcePgBouncerPools     = "pgbouncerpools"
	ResourcePgBouncerPoolss    = "pgbouncerpools"
)

type PgBouncerPoolsSpec struct {
	Pools []PBPools `json:"pools"`
}

type PBPools struct {
	PodName                  string `json:"podName"`
	Database                 string `json:"database"`
	User                     string `json:"user"`
	ClientConnectionsActive  *int32 `json:"clientConnectionsActive"`
	ClientConnectionsWaiting *int32 `json:"clientConnectionsWaiting"`
	ClientsActiveCancelReq   *int32 `json:"clientsActiveCancelReq"`
	ClientsWaitingCancelReq  *int32 `json:"clientsWaitingCancelReq"`
	ServerConnectionsActive  *int32 `json:"serverConnectionsActive"`
	ServersActiveCancel      *int32 `json:"serversActiveCancel"`
	ServersBeingCanceled     *int32 `json:"serversBeingCanceled"`
	ServersIdle              *int32 `json:"serversIdle"`
	ServersUsed              *int32 `json:"serversUsed"`
	ServersTested            *int32 `json:"serversTested"`
	ServersLogin             *int32 `json:"serversLogin"`
	MaxWait                  *int32 `json:"maxWait"`
	Mode                     string `json:"mode"`
}

// PgBouncerPools is the Schema for the PgBouncerPoolss API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PgBouncerPools struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PgBouncerPoolsSpec `json:"spec,omitempty"`
}

// PgBouncerSettingsList contains a list of PgBouncerSettings

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PgBouncerPoolsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PgBouncerPools `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PgBouncerPools{}, &PgBouncerPoolsList{})
}
