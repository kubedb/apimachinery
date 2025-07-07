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
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ResourceKindRabbitMQVersion = "RabbitMQVersion"
	ResourceRabbitMQVersion     = "rabbitmqversion"
	ResourceRabbitMQVersions    = "rabbitmqversions"
)

// RabbitMQVersion is the Schema for the RabbitMQVersions API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RabbitMQVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *runtime.RawExtension `json:"spec,omitempty"`
}

// RabbitMQVersionList contains a list of RabbitMQVersion

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RabbitMQVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQVersion{}, &RabbitMQVersionList{})
}
