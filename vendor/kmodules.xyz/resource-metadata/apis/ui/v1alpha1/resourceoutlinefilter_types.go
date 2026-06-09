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
	kmapi "kmodules.xyz/client-go/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindResourceOutlineFilter = "ResourceOutlineFilter"
	ResourceResourceOutlineFilter     = "resourceoutlinefilter"
	ResourceResourceOutlineFilters    = "resourceoutlinefilters"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=get
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=resourceoutlinefilters,singular=resourceoutlinefilter,scope=Cluster
type ResourceOutlineFilter struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceOutlineFilterSpec `json:"spec,omitempty"`
}

type ResourceOutlineFilterSpec struct {
	Resource kmapi.ResourceID            `json:"resource"`
	Header   bool                        `json:"header"`
	TabBar   bool                        `json:"tabBar"`
	Pages    []ResourcePageOutlineFilter `json:"pages,omitempty"`
	// +optional
	Actions []ActionTemplateGroupFilter `json:"actions,omitempty"`
}

type ResourcePageOutlineFilter struct {
	Name     string                 `json:"name"`
	Sections []SectionOutlineFilter `json:"sections,omitempty"`
	Show     bool                   `json:"show"`
}

type ActionTemplateGroupFilter struct {
	Name  string          `json:"name"`
	Items map[string]bool `json:"items,omitempty"`
	Show  bool            `json:"show"`
}

type SectionOutlineFilter struct {
	Name    string          `json:"name,omitempty"`
	Show    bool            `json:"show"`
	Info    map[string]bool `json:"info"`
	Insight bool            `json:"insight"`
	Blocks  map[string]bool `json:"blocks,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ResourceOutlineFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceOutlineFilter `json:"items,omitempty"`
}
