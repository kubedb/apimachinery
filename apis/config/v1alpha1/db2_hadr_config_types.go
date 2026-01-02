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
