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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindClusterSummary = "ClusterSummary"
	ResourceClusterSummary     = "clustersummary"
	ResourceClusterSummaries   = "clustersummarier"
)

// ClusterSummarySpec defines the desired state of ClusterSummary
type ClusterSummarySpec struct {
	Summary   []ClusterSummaryInfo `json:"summary" protobuf:"bytes,1,rep,name=summary"`
	Resources []ClusterResource    `json:"resources" protobuf:"bytes,2,rep,name=resources"`
}

type ClusterResource struct {
	ClusterName  string     `json:"clusterName" protobuf:"bytes,1,opt,name=clusterName"`
	ResourceInfo []Resource `json:"resourceInfo" protobuf:"bytes,2,rep,name=resourceInfo"`
}

type Resource struct {
	Label corev1.ResourceName `json:"label" protobuf:"bytes,1,opt,name=label,casttype=k8s.io/api/core/v1.ResourceName"`
	Value string              `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type ClusterSummaryInfo struct {
	Label string `json:"label" protobuf:"bytes,1,opt,name=label"`
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
	Icon  string `json:"icon" protobuf:"bytes,3,opt,name=icon"`
}

// ClusterSummaryStatus defines the observed state of ClusterSummary
type ClusterSummaryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// ClusterSummary is the Schema for the clustersummaries API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterSummary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ClusterSummarySpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ClusterSummaryStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ClusterSummaryList contains a list of ClusterSummary

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterSummaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ClusterSummary `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSummary{}, &ClusterSummaryList{})
}
