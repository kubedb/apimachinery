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
	ResourceKindBackupOverview = "BackupOverview"
	ResourceBackupOverview     = "backupoverview"
	ResourceBackupOverviews    = "backupoverviews"
)

// BackupOverviewSpec defines the desired state of BackupOverview
type BackupOverviewSpec struct {
	Schedule           string       `json:"schedule,omitempty" protobuf:"bytes,1,opt,name=schedule"`
	LastBackupTime     *metav1.Time `json:"lastBackupTime,omitempty" protobuf:"bytes,2,opt,name=lastBackupTime"`
	UpcomingBackupTime *metav1.Time `json:"upcomingBackupTime,omitempty" protobuf:"bytes,3,opt,name=upcomingBackupTime"`
	BackupStorage      string       `json:"backupStorage,omitempty" protobuf:"bytes,4,opt,name=backupStorage"`
	DataSize           string       `json:"dataSize" protobuf:"bytes,5,opt,name=dataSize"`
	NumberOfSnapshots  int64        `json:"numberOfSnapshots,omitempty" protobuf:"bytes,6,opt,name=numberOfSnapshots"`
	DataIntegrity      bool         `json:"dataIntegrity,omitempty" protobuf:"bytes,7,opt,name=dataIntegrity"`
	DataDirectory      string       `json:"dataDirectory,omitempty" protobuf:"bytes,8,opt,name=dataDirectory"`
}

// BackupOverview is the Schema for the BackupOverviews API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackupOverview struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec BackupOverviewSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// BackupOverviewList contains a list of BackupOverview

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackupOverviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []BackupOverview `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&BackupOverview{}, &BackupOverviewList{})
}
