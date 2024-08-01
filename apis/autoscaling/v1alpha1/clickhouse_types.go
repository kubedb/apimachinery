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
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeClickHouseAutoscaler     = "chscaler"
	ResourceKindClickHouseAutoscaler     = "ClickHouseAutoscaler"
	ResourceSingularClickHouseAutoscaler = "clickhouseautoscaler"
	ResourcePluralClickHouseAutoscaler   = "clickhouseautoscalers"
)

// ClickHouseAutoscaler is the configuration for a clickhouse database
// autoscaler, which automatically manages pod resources based on historical and
// real time resource utilization.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=clickhouseautoscalers,singular=clickhouseautoscaler,shortName=chscaler,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
type ClickHouseAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	Spec ClickHouseAutoscalerSpec `json:"spec"`

	// Current information about the autoscaler.
	// +optional
	Status AutoscalerStatus `json:"status,omitempty"`
}

// ClickHouseAutoscalerSpec is the specification of the behavior of the autoscaler.
type ClickHouseAutoscalerSpec struct {
	DatabaseRef *core.LocalObjectReference `json:"databaseRef"`

	// This field will be used to control the behaviour of ops-manager
	OpsRequestOptions *ClickHouseOpsRequestOptions `json:"opsRequestOptions,omitempty"`

	Compute *ClickHouseComputeAutoscalerSpec `json:"compute,omitempty"`
	Storage *ClickHouseStorageAutoscalerSpec `json:"storage,omitempty"`
}

type ClickHouseComputeAutoscalerSpec struct {
	// +optional
	NodeTopology *NodeTopology `json:"nodeTopology,omitempty"`

	ClickHouse *ComputeAutoscalerSpec `json:"clickhouse,omitempty"`
}

type ClickHouseStorageAutoscalerSpec struct {
	ClickHouse *StorageAutoscalerSpec `json:"clickhouse,omitempty"`
}

type ClickHouseOpsRequestOptions struct {
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply opsapi.ApplyOption `json:"apply,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClickHouseAutoscalerList is a list of ClickHouseAutoscaler objects.
type ClickHouseAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata"`

	// items is the list of clickhouse database autoscaler objects.
	Items []ClickHouseAutoscaler `json:"items"`
}
