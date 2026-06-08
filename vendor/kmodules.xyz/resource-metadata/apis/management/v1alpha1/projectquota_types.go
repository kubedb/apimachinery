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
)

const (
	ResourceKindProjectQuota = "ProjectQuota"
	ResourceProjectQuota     = "projectquota"
	ResourceProjectQuotas    = "projectquotas"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=projectquotas,singular=projectquota,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ProjectQuota struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProjectQuotaSpec   `json:"spec,omitempty"`
	Status            ProjectQuotaStatus `json:"status,omitempty"`
}

type ProjectQuotaSpec struct {
	Quotas []ResourceQuotaSpec `json:"quotas"`
}

type ResourceQuotaSpec struct {
	Group string `json:"group,omitempty"`
	Kind  string `json:"kind,omitempty"`
	// Hard is the set of enforced hard limits for each named resource.
	// More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/
	// +optional
	Hard core.ResourceList `json:"hard,omitempty"`
}

// ProjectQuotaStatus defines the observed state of ProjectQuota
type ProjectQuotaStatus struct {
	Quotas []ResourceQuotaStatus `json:"quotas"`
}

type ResourceQuotaStatus struct {
	ResourceQuotaSpec `json:",inline"`
	Result            QuotaResult `json:"result"`
	// +optional
	Reason string `json:"reason,omitempty"`

	// Used is the current observed total usage of the resource in the namespace.
	// +optional
	Used core.ResourceList `json:"used,omitempty"`
}

// +kubebuilder:validation:Enum=Success;Error
type QuotaResult string

const (
	ResultSuccess QuotaResult = "Success"
	ResultError   QuotaResult = "Error"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ProjectQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectQuota `json:"items,omitempty"`
}
