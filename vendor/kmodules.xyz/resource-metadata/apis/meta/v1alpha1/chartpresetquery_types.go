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
	"k8s.io/apimachinery/pkg/util/json"
	chartsapi "x-helm.dev/apimachinery/apis/charts/v1alpha1"
)

const (
	ResourceKindChartPresetQuery = "ChartPresetQuery"
	ResourceChartPresetQuery     = "chartpresetquery"
	ResourceChartPresetQueries   = "chartpresetqueries"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=chartpresetquery,singular=chartpresetquery,scope=Cluster
type ChartPresetQuery struct {
	metav1.TypeMeta `json:",inline"`
	// Request describes the attributes for the preset query.
	// +optional
	Request *chartsapi.ChartPresetFlatRef `json:"request,omitempty"`
	// Response describes the attributes for the preset response.
	// +optional
	Response *ChartPresetQueryResponse `json:"response,omitempty"`
}

type ChartPresetQueryResponse struct {
	Presets []chartsapi.ChartPresetValues `json:"presets,omitempty"`
	State   string                        `json:"state,omitempty"`
}

func (q *ChartPresetQuery) UnmarshalJSON(data []byte) error {
	// https://stackoverflow.com/a/43178272/244009
	type nq ChartPresetQuery
	if err := json.Unmarshal(data, (*nq)(q)); err == nil {
		return nil
	}

	var oq struct {
		metav1.TypeMeta `json:",inline"`
		Request         *chartsapi.ChartPresetFlatRef `json:"request,omitempty"`
		Response        []chartsapi.ChartPresetValues `json:"response,omitempty"`
	}
	if err := json.Unmarshal(data, &oq); err != nil {
		return err
	}
	*q = ChartPresetQuery{
		TypeMeta: oq.TypeMeta,
		Request:  oq.Request,
		Response: &ChartPresetQueryResponse{
			Presets: oq.Response,
		},
	}
	return nil
}
