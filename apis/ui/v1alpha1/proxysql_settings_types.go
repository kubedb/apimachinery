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
	ResourceKindProxySQLSettings = "ProxySQLSettings"
	ResourceProxySQLSettings     = "proxysqlsettings"
	ResourceProxySQLSettingss    = "proxysqlsettings"
)

// ProxySQLSettingsSpec defines the desired state of ProxySQLSettings
type ProxySQLSettingsSpec struct {
	Settings []PGSetting `json:"settings"`
}

type PRXSetting struct {
	Name         string `json:"name"`
	CurrentValue string `json:"currentValue"`
	DefaultValue string `json:"defaultValue"`
	Unit         string `json:"unit,omitempty"`
	Source       string `json:"source,omitempty"`
}

// ProxySQLSettings is the Schema for the ProxySQLSettingss API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProxySQLSettings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProxySQLSettingsSpec `json:"spec,omitempty"`
}

// ProxySQLSettingsList contains a list of ProxySQLSettings

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProxySQLSettingsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxySQLSettings `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxySQLSettings{}, &ProxySQLSettingsList{})
}
