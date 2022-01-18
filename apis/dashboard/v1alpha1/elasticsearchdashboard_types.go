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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchDashboardSpec defines the desired state of ElasticsearchDashboard

type ElasticsearchDashboardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version of ElasticsearchDashboard to be deployed.
	Version string `json:"version"`

	// Number of instances to deploy for a ElasticsearchDashboard Dashboard.

	Replicas *int32 `json:"replicas,omitempty"`

	// To enable ssl for http layer
	EnableSSL bool `json:"enableSSL,omitempty"`

	// disable security of authPlugin (ie, xpack or searchguard). It disables authentication security of user.
	// If unset, default is false
	// +optional
	DisableSecurity bool `json:"disableSecurity,omitempty"`

	// host elasticsearch name and namespace
	DatabaseRef *core.ObjectReference `json:"databaseRef,omitempty"`

	// Dashboard authentication secret
	// +optional
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty"`

	// Monitor is used monitor Dashboard instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for dashboard.
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose Dashboard
	// +optional
	PodTemplate ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose Dashboard
	// +optional
	ServiceTemplates []api.NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// TLS contains tls configurations
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// TerminationPolicy controls the delete operation for Dashboard
	// +optional
	TerminationPolicy api.TerminationPolicy `json:"terminationPolicy,omitempty"`
}

// ElasticsearchDashboardStatus defines the observed state of ElasticsearchDashboard
type ElasticsearchDashboardStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies the current phase of the database
	// +optional
	Phase DashboardPhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ed,scope=Namespaced
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
//+kubebuilder:printcolumn:name="Type",type="string",JSONPath=".apiVersion"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"

// ElasticsearchDashboard is the Schema for the elasticsearchdashboards API
type ElasticsearchDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchDashboardSpec   `json:"spec,omitempty"`
	Status ElasticsearchDashboardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchDashboardList contains a list of ElasticsearchDashboard
type ElasticsearchDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticsearchDashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticsearchDashboard{}, &ElasticsearchDashboardList{})
}
