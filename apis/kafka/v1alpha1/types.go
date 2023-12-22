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
	core "k8s.io/api/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

// +kubebuilder:validation:Enum=Delete;DoNotTerminate
type TerminationPolicy string

const (
	// TerminationPolicyDelete deletes cr, resources it creates
	TerminationPolicyDelete TerminationPolicy = "Delete"
	// TerminationPolicyDoNotTerminate Rejects attempt to delete using ValidationWebhook.
	TerminationPolicyDoNotTerminate TerminationPolicy = "DoNotTerminate"
)

type SecretReference struct {
	core.LocalObjectReference `json:",inline,omitempty"`
	ExternallyManaged         bool `json:"externallyManaged,omitempty"`
}

// +kubebuilder:validation:Enum=primary;standby;stats
type ServiceAlias string

const (
	PrimaryServiceAlias ServiceAlias = "primary"
	StandbyServiceAlias ServiceAlias = "standby"
	StatsServiceAlias   ServiceAlias = "stats"
)

type NamedServiceTemplateSpec struct {
	// Alias represents the identifier of the service.
	Alias ServiceAlias `json:"alias"`

	// ServiceTemplate is an optional configuration for a service used to expose database
	// +optional
	ofst.ServiceTemplateSpec `json:",inline,omitempty"`
}
