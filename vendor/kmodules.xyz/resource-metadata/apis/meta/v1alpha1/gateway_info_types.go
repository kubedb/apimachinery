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
	ResourceKindGatewayInfo = "GatewayInfo"
	ResourceGatewayInfo     = "gatewayinfo"
	ResourceGatewayInfos    = "gatewayinfos"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GatewayInfo struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GatewayInfoSpec `json:"spec,omitempty"`
}

type GatewayInfoSpec struct {
	GatewayClassName string `json:"gatewayClassName"`
	ServiceType      string `json:"serviceType"`
	HostName         string `json:"hostName,omitempty"`
	IP               string `json:"ip,omitempty"`
}
