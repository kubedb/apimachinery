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
	"kmodules.xyz/resource-metadata/apis/shared"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindResourceDashboard = "ResourceDashboard"
	ResourceResourceDashboard     = "resourcedashboard"
	ResourceResourceDashboards    = "resourcedashboards"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=resourcedashboards,singular=resourcedashboard,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ResourceDashboard struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceDashboardSpec `json:"spec,omitempty"`
}

// +kubebuilder:validation:Enum=Grafana
type DashboardProvider string

const (
	DashboardProviderGrafana DashboardProvider = "Grafana"
)

type ResourceDashboardSpec struct {
	Resource   kmapi.ResourceID  `json:"resource"`
	Provider   DashboardProvider `json:"provider,omitempty"`
	Dashboards []Dashboard       `json:"dashboards"`
}

type Dashboard struct {
	// +optional
	Title string `json:"title,omitempty"`
	// +optional
	Vars []shared.DashboardVar `json:"vars,omitempty"`
	// +optional
	Panels []PanelLinkRequest `json:"panels,omitempty"`
	// +optional
	If *shared.If `json:"if,omitempty"`
}

type PanelLinkRequest struct {
	Title string `json:"title"`
	// +optional
	Width int `json:"width,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type ResourceDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceDashboard `json:"items,omitempty"`
}
