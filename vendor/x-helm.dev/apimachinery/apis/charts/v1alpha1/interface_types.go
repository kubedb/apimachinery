/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	kmapi "kmodules.xyz/client-go/api/v1"
	releasesv1alpha1 "x-helm.dev/apimachinery/apis/releases/v1alpha1"
)

// +kubebuilder:object:generate:=false
type Preset interface {
	GetObjectKind() schema.ObjectKind
	GetName() string
	GetLabels() map[string]string
	GetSpec() ClusterChartPresetSpec
}

var _ Preset = &ClusterChartPreset{}

func (in ClusterChartPreset) GetSpec() ClusterChartPresetSpec {
	return in.Spec
}

var _ Preset = &ChartPreset{}

func (in ChartPreset) GetSpec() ClusterChartPresetSpec {
	return in.Spec
}

type ChartPresetFlatRef struct {
	releasesv1alpha1.ChartSourceFlatRef `json:",inline"`

	// Editor GVR
	Group     string `json:"group,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Variant   string `json:"variant,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type ChartPresetValues struct {
	Source SourceLocator         `json:"source"`
	Values *runtime.RawExtension `json:"values"`
}

type SourceLocator struct {
	// +optional
	Resource kmapi.ResourceID `json:"resource"`
	// +optional
	Ref kmapi.ObjectReference `json:"ref"`
	// +optional
	UID types.UID `json:"-"`
	// +optional
	Generation int64 `json:"-"`
}
