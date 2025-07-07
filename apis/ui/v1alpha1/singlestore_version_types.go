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
	ResourceKindSinglestoreVersion = "SinglestoreVersion"
	ResourceSinglestoreVersion     = "singlestoreversion"
	ResourceSinglestoreVersions    = "singlestoreversions"
)

// SinglestoreVersion is the Schema for the SinglestoreVersions API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SinglestoreVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *runtime.RawExtension `json:"spec,omitempty"`
}

// SinglestoreVersionList contains a list of SinglestoreVersion

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SinglestoreVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SinglestoreVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SinglestoreVersion{}, &SinglestoreVersionList{})
}
