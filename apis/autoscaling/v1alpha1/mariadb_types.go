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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceCodeMariaDBAutoscaler     = "mdautoscaler"
	ResourceKindMariaDBAutoscaler     = "MariaDBAutoscaler"
	ResourceSingularMariaDBAutoscaler = "mariadbautoscaler"
	ResourcePluralMariaDBAutoscaler   = "mariadbautoscalers"
)

// MariaDBAutoscaler is the configuration for a mariadb database
// autoscaler, which automatically manages pod resources based on historical and
// real time resource utilization.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mariadbautoscalers,singular=mariadbautoscaler,shortName=mdautoscaler,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
type MariaDBAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	Spec MariaDBAutoscalerSpec `json:"spec"`

	// Current information about the autoscaler.
	// +optional
	Status MariaDBAutoscalerStatus `json:"status,omitempty"`
}

// MariaDBAutoscalerSpec is the specification of the behavior of the autoscaler.
type MariaDBAutoscalerSpec struct {
	DatabaseRef *core.LocalObjectReference `json:"databaseRef"`

	Compute *MariaDBComputeAutoscalerSpec `json:"compute,omitempty"`
	Storage *MariaDBStorageAutoscalerSpec `json:"storage,omitempty"`
}

type MariaDBComputeAutoscalerSpec struct {
	MariaDB          *ComputeAutoscalerSpec `json:"mariadb,omitempty"`
	DisableScaleDown bool                   `json:"disableScaleDown,omitempty"`
}

type MariaDBStorageAutoscalerSpec struct {
	MariaDB *StorageAutoscalerSpec `json:"mariadb,omitempty"`
}

// MariaDBAutoscalerStatus describes the runtime state of the autoscaler.
type MariaDBAutoscalerStatus struct {
	// observedGeneration is the most recent generation observed by this autoscaler.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions is the set of conditions required for this autoscaler to scale its target,
	// and indicates whether or not those conditions are met.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []kmapi.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// This field is equivalent to this one:
	// https://github.com/kubernetes/autoscaler/blob/273e35b88cb50c5aac383c5eceb88fb337cb31b6/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L218-L230
	VPAs []VPAStatus `json:"vpas,omitempty"`

	// Checkpoints hold all the Checkpoint those are associated
	// with this Autoscaler object. Equivalent to :
	// https://github.com/kubernetes/autoscaler/blob/273e35b88cb50c5aac383c5eceb88fb337cb31b6/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1/types.go#L354-L378
	Checkpoints []Checkpoint `json:"checkpoints,omitempty"`
}

// MariaDBAutoscalerConditionType are the valid conditions of
// a MariaDBAutoscaler.
type MariaDBAutoscalerConditionType string

var (
	// ConfigDeprecated indicates that this VPA configuration is deprecated
	// and will stop being supported soon.
	MariaDBAutoscalerConfigDeprecated MariaDBAutoscalerConditionType = "ConfigDeprecated"
	// ConfigUnsupported indicates that this VPA configuration is unsupported
	// and recommendations will not be provided for it.
	MariaDBAutoscalerConfigUnsupported MariaDBAutoscalerConditionType = "ConfigUnsupported"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MariaDBAutoscalerList is a list of MariaDBAutoscaler objects.
type MariaDBAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata"`

	// items is the list of mariadb database autoscaler objects.
	Items []MariaDBAutoscaler `json:"items"`
}
