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

// DBNodeResponseSpec defines the desired state of DBNodeResponse
type DBNodeResponseSpec struct {
	Name    string       `json:"name" protobuf:"bytes,1,opt,name=name"`
	Type    string       `json:"type" protobuf:"bytes,2,opt,name=type"`
	Role    DatabaseRole `json:"role" protobuf:"bytes,3,opt,name=role,casttype=DatabaseRole"`
	CPU     string       `json:"cpu" protobuf:"bytes,4,opt,name=cpu"`
	Memory  string       `json:"memory" protobuf:"bytes,5,opt,name=memory"`
	Storage string       `json:"storage" protobuf:"bytes,6,opt,name=storage"`
	Status  string       `json:"status" protobuf:"bytes,7,opt,name=status"`
}

// DBNodeResponseStatus defines the observed state of DBNodeResponse
type DBNodeResponseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DBNodeResponse is the Schema for the dbnoderesponses API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DBNodeResponse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   DBNodeResponseSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status DBNodeResponseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// DBNodeResponseList contains a list of DBNodeResponse

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DBNodeResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []DBNodeResponse `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&DBNodeResponse{}, &DBNodeResponseList{})
}
