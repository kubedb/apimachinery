package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
)

const (
	ResourceCodeHazelcastAutoscaler     = "hzscaler"
	ResourceKindHazelcastAutoscaler     = "HazelcastAutoscaler"
	ResourceSingularHazelcastAutoscaler = "hazelcastautoscaler"
	ResourcePluralHazelcastAutoscaler   = "hazelcastautoscalers"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=hazelcastautoscalers,singular=hazelcastautoscaler,shortName=hzscaler,categories={autoscaler,kubedb,appscode}
// +kubebuilder:subresource:status

type HazelcastAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	Spec HazelcastComputeAutoscalerSpec `json:"spec"`

	// Current information about the autoscaler.
	// +optional
	Status AutoscalerStatus `json:"status,omitempty"`
}

type HazelcastComputeAutoscalerSpec struct {
	// +optional
	NodeTopology *NodeTopology `json:"nodeTopology,omitempty"`

	Hazelcast *ComputeAutoscalerSpec `json:"rabbitmq,omitempty"`
}

type HazelcastStorageAutoscalerSpec struct {
	Hazelcast *StorageAutoscalerSpec `json:"rabbitmq,omitempty"`
}

type HazelcastOpsrequestOptions struct {
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply opsapi.ApplyOption `json:"apply,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HazelcastAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HazelcastAutoscaler `json:"items"`
}
