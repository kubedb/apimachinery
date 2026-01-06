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

type DB2HADRSpec struct {
	Primary int `yaml:"primary,omitempty"`
	Standby int `yaml:"standby,omitempty"`
	//+optional
	ConfigSecret core.LocalObjectReference `yaml:"configSecret,omitempty"`
	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `yaml:"serviceTemplates,omitempty"`
}

// +kubebuilder:validation:Enum=primary;standby
type ServiceAlias string

const (
	PrimaryServiceAlias ServiceAlias = "primary"
	StandbyServiceAlias ServiceAlias = "standby"
)

type NamedServiceTemplateSpec struct {
	// Alias represents the identifier of the service.
	Alias ServiceAlias `yaml:"alias"`

	// ServiceTemplate is an optional configuration for a service used to expose database
	// +optional
	ofst.ServiceTemplateSpec `yaml:",inline,omitempty"`
}
