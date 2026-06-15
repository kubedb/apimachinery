/*
Copyright AppsCode Inc. and Contributors.

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
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/resource-metrics/api"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindResourceCalculator = "ResourceCalculator"
	ResourceResourceCalculator     = "resourcecalculator"
	ResourceResourceCalculators    = "resourcecalculators"
)

// ResourceCalculator is the Schema for any resource supported by resource-metrics library

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ResourceCalculator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:pruning:PreserveUnknownFields
	Request *ResourceCalculatorRequest `json:"request,omitempty"`
	// +optional
	Response *ResourceCalculatorResponse `json:"response,omitempty"`
}

type ResourceCalculatorRequest struct {
	Resource *runtime.RawExtension `json:"resource,omitempty"`
	Edit     bool                  `json:"edit,omitempty"`
}

type ResourceCalculatorResponse struct {
	APIType kmapi.ResourceID `json:"apiType"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	Replicas int64 `json:"replicas,omitempty"`
	// +optional
	RoleReplicas api.ReplicaList `json:"roleReplicas,omitempty"`
	// +optional
	Mode string `json:"mode,omitempty"`
	// +optional
	TotalResource core.ResourceRequirements `json:"totalResource,omitempty"`
	// +optional
	AppResource core.ResourceRequirements `json:"appResource,omitempty"`
	// +optional
	RoleResourceLimits map[api.PodRole]core.ResourceList `json:"roleResourceLimits,omitempty"`
	// +optional
	RoleResourceRequests map[api.PodRole]core.ResourceList `json:"roleResourceRequests,omitempty"`
	Quota                QuotaDecision                     `json:"quota"`
}

// +kubebuilder:validation:Enum=Allow;Deny
type Decision string

const (
	// DecisionNoOpionion means that quota restrictions have no opinion on an action.
	DecisionNoOpinion Decision = ""
	// DecisionAllow means that quota restrictions allow the action.
	DecisionAllow Decision = "Allow"
	// DecisionDeny means that quota restrictions deny the action.
	DecisionDeny Decision = "Deny"
)

type QuotaDecision struct {
	Decision   Decision `json:"decision"`
	Violations []string `json:"violations,omitempty"`
}
