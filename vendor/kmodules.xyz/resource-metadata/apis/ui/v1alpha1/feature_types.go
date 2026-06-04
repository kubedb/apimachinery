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
	v1 "kmodules.xyz/client-go/api/v1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	helmshared "x-helm.dev/apimachinery/apis/shared"
)

const (
	ResourceKindFeature = "Feature"
	ResourceFeature     = "feature"
	ResourceFeatures    = "features"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=features,singular=feature,scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Enabled",type="boolean",JSONPath=".status.enabled"
// +kubebuilder:printcolumn:name="Managed",type="boolean",JSONPath=".status.managed"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Feature struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              FeatureSpec   `json:"spec,omitempty"`
	Status            FeatureStatus `json:"status,omitempty"`
}

type FeatureSpec struct {
	// Title specifies the title of this feature.
	Title string `json:"title"`
	// Description specifies a short description of the service this feature provides.
	Description string `json:"description"`
	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type. These icons will be used in UI.
	Icons []helmshared.ImageSpec `json:"icons,omitempty"`
	// FeatureSet specifies the name of the FeatureSet where this feature belong to.
	FeatureSet string `json:"featureSet"`
	// FeatureBlock specifies the ui block name of this feature.
	// +optional
	FeatureBlock string `json:"featureBlock,omitempty"`
	// FeatureExclusionGroup specifies the name of the exclusion group for features
	// Only one feature in a feature exclusion group can be installed
	// +optional
	FeatureExclusionGroup string `json:"featureExclusionGroup,omitempty"`
	// Required specifies whether this feature is mandatory or not for enabling the respecting FeatureSet.
	// +optional
	Recommended bool `json:"recommended,omitempty"`
	// Disabled specify whether this feature set is disabled.
	// +optional
	Disabled bool `json:"disabled,omitempty"`
	// Requirements specifies the requirements to enable this feature.
	// +optional
	Requirements Requirements `json:"requirements,omitempty"`
	// ReadinessChecks specifies the conditions for this feature to be considered enabled.
	// +optional
	ReadinessChecks ReadinessChecks `json:"readinessChecks,omitempty"`
	// Chart specifies the chart information that will be used by the FluxCD to install the respective feature
	// +optional
	Chart ChartInfo `json:"chart,omitempty"`
	// ValuesFrom holds references to resources containing Helm values for this HelmRelease,
	// and information about how they should be merged.
	ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`
	// Values holds the values for this Helm release.
	// +optional
	Values *apiextensionsv1.JSON `json:"values,omitempty"`
}

type Requirements struct {
	// Features specifies a list of Feature names that must be enabled for using this feature.
	// +optional
	Features []string `json:"features,omitempty"`
}

type ReadinessChecks struct {
	// Resources specifies the resources that should be registered to consider this feature as enabled.
	// +optional
	Resources []metav1.GroupVersionKind `json:"resources,omitempty"`
	// Workloads specifies the workloads that should exist to consider this feature as enabled.
	// +optional
	Workloads []WorkloadInfo `json:"workloads,omitempty"`
}

type WorkloadInfo struct {
	metav1.GroupVersionKind `json:",inline"`
	// Selector specifies label selector that should be used to select this workload
	Selector map[string]string `json:"selector"`
	Optional string            `json:"optional,omitempty"`
}

type ChartInfo struct {
	// Name specifies the name of the chart
	Name string `json:"name"`
	// Namespace where the respective feature resources will be deployed.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	CreateNamespace bool `json:"createNamespace,omitempty"`
	// Version specifies the version of the chart.
	// +optional
	Version string `json:"version,omitempty"`
	// SourceRef specifies the source of the chart
	SourceRef v1.TypedObjectReference `json:"sourceRef"`
	// Alternative list of values files to use as the chart values (values.yaml
	// is not included by default), expected to be a relative path in the SourceRef.
	// Values files are merged in the order of this list with the last file overriding
	// the first. Ignored when omitted.
	// +optional
	ValuesFiles []string `json:"valuesFiles,omitempty"`
}

// copied from: https://github.com/fluxcd/helm-controller/blob/v0.37.4/api/v2beta2/reference_types.go#L45-L80
// ValuesReference contains a reference to a resource containing Helm values,
// and optionally the key they can be found at.
type ValuesReference struct {
	// Kind of the values referent, valid values are ('Secret', 'ConfigMap').
	// +kubebuilder:validation:Enum=Secret;ConfigMap
	// +required
	Kind string `json:"kind"`

	// Name of the values referent. Should reside in the same namespace as the
	// referring resource.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name"`

	// ValuesKey is the data key where the values.yaml or a specific value can be
	// found at. Defaults to 'values.yaml'.
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[\-._a-zA-Z0-9]+$`
	// +optional
	ValuesKey string `json:"valuesKey,omitempty"`

	// TargetPath is the YAML dot notation path the value should be merged at. When
	// set, the ValuesKey is expected to be a single flat value. Defaults to 'None',
	// which results in the values getting merged at the root.
	// +kubebuilder:validation:MaxLength=250
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9_\-.\\\/]|\[[0-9]{1,5}\])+$`
	// +optional
	TargetPath string `json:"targetPath,omitempty"`

	// Optional marks this ValuesReference as optional. When set, a not found error
	// for the values reference is ignored, but any ValuesKey, TargetPath or
	// transient error will still result in a reconciliation failure.
	// +optional
	Optional bool `json:"optional,omitempty"`
}

type FeatureStatus struct {
	// Enabled specifies whether this feature is enabled or not.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// Managed specifies whether this feature is managed by AppsCode Inc. or not.
	// +optional
	Managed *bool `json:"managed,omitempty"`
	// Ready specifies whether this feature is ready to user or not. This field will be present only for the
	// features that are managed by AppsCode Inc.
	// +optional
	Ready *bool `json:"ready,omitempty"`
	// Note specifies the respective reason if the feature does not meet the requirements or is not ready.
	// +optional
	Note string `json:"note,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

type FeatureList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Feature `json:"items,omitempty"`
}
