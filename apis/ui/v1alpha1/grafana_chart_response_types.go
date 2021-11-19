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
	ResourceKindGrafanaChartResponse = "GrafanaChartResponse"
	ResourceGrafanaChartResponse     = "grafanachartresponse"
	ResourceGrafanaChartResponses    = "grafanachartresponses"
)

// GrafanaChartResponseSpec defines the desired state of GrafanaChartResponse
type GrafanaChartResponseSpec struct {
	EmbeddedURLs []GrafanaChartURL `json:"embeddedURLs" protobuf:"bytes,1,rep,name=embeddedURLs,casttype=GrafanaChartURL"`
}

// GrafanaChartResponseStatus defines the observed state of GrafanaChartResponse
type GrafanaChartResponseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// GrafanaChartResponse is the Schema for the grafanachartresponses API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaChartResponse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   GrafanaChartResponseSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status GrafanaChartResponseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// GrafanaChartResponseList contains a list of GrafanaChartResponse

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GrafanaChartResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []GrafanaChartResponse `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaChartResponse{}, &GrafanaChartResponseList{})
}
