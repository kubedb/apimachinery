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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindGrafanaDashboard = "GrafanaDashboard"
	ResourceGrafanaDashboard     = "grafanadashboard"
	ResourceGrafanaDashboards    = "grafanadashboards"
)

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	TargetRef     TargetRef `json:"targetRef" protobuf:"bytes,1,opt,name=targetRef"`
	DashboardName string    `json:"dashboardName" protobuf:"bytes,2,opt,name=dashboardName"`
	URL           string    `json:"url" protobuf:"bytes,3,opt,name=url"`
	OrgID         int64     `json:"orgID" protobuf:"varint,4,opt,name=orgID"`
	BoardUID      string    `json:"boardUID" protobuf:"bytes,5,opt,name=boardUID"`
	PanelID       []int64   `json:"panelID" protobuf:"varint,6,rep,name=panelID"`
	// +optional
	From int64 `json:"from" protobuf:"varint,7,opt,name=from"`
	// +kubebuilder:validation:Enum=dark;light;
	Theme GrafanaTheme `json:"theme" protobuf:"bytes,8,opt,name=theme,casttype=GrafanaTheme"`
}

type TargetRef struct {
	APIVersion string `json:"apiVersion" protobuf:"bytes,1,opt,name=apiVersion"`
	Kind       string `json:"kind" protobuf:"bytes,2,opt,name=kind"`
}

// GrafanaDashboardStatus defines the observed state of GrafanaDashboard
type GrafanaDashboardStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// GrafanaDashboard is the Schema for the grafanadashboards API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   GrafanaDashboardSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status GrafanaDashboardStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// GrafanaDashboardList contains a list of GrafanaDashboard

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []GrafanaDashboard `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}
