/*
Copyright 2023.

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
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TODO:
// set scope=Cluster for  MsSQLVersion struct

const (
	ResourceCodeMsSQLVersion     = "msversion"
	ResourceKindMsSQLVersion     = "MsSQLVersion"
	ResourceSingularMsSQLVersion = "mssqlversion"
	ResourcePluralMsSQLVersion   = "mssqlversions"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqlversions,singular=mssqlversion,shortName=msversion,categories={datastore,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MsSQLVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MsSQLVersionSpec `json:"spec,omitempty"`
}

// MsSQLVersionSpec defines the desired state of MsSQL Version
type MsSQLVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Database Image
	DB MsSQLVersionDatabase `json:"db"`
	// Coordinator Image
	// +optional
	Coordinator MsSQLCoordinator `json:"coordinator,omitempty"`
	// Init container Image
	InitContainer MsSQLInitContainer `json:"initContainer"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`
	// Stash defines backup and restore task definitions.
	// +optional
	Stash appcat.StashAddonSpec `json:"stash,omitempty"`
	// SecurityContext is for the additional config for the DB container
	// +optional
	SecurityContext MsSQLSecurityContext `json:"securityContext"`
	// PSP names
	// +optional
	PodSecurityPolicies MsSQLVersionPodSecurityPolicy `json:"podSecurityPolicies"`
	// update constraints
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// MsSQLVersionDatabase is the MsSQL Database image
type MsSQLVersionDatabase struct {
	Image string `json:"image"`
}

// MsSQLCoordinator is the MSSQL coordinator Container image
type MsSQLCoordinator struct {
	Image string `json:"image"`
}

// MsSQLInitContainer is the MsSQL Container initializer
type MsSQLInitContainer struct {
	Image string `json:"image"`
}

// MsSQLVersionPodSecurityPolicy is the MsSQL pod security policies
type MsSQLVersionPodSecurityPolicy struct {
	DatabasePolicyName string `json:"databasePolicyName"`
}

// MsSQLSecurityContext is for additional configuration for the MSSQL database container
type MsSQLSecurityContext struct {
	RunAsUser  *int64 `json:"runAsUser,omitempty"`
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MsSQLVersionList contains a list of MsSQLVersion
type MsSQLVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MsSQLVersion `json:"items"`
}
