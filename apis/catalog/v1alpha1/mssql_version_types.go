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

const (
	ResourceCodeMSSQLVersion     = "msversion"
	ResourceKindMSSQLVersion     = "MSSQLVersion"
	ResourceSingularMSSQLVersion = "mssqlversion"
	ResourcePluralMSSQLVersion   = "mssqlversions"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mssqlversions,singular=mssqlversion,scope=Cluster,shortName=msversion,categories={datastore,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MSSQLVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MSSQLVersionSpec `json:"spec,omitempty"`
}

// MSSQLVersionSpec defines the desired state of MSSQL Version
type MSSQLVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Database Image
	DB MSSQLVersionDatabase `json:"db"`
	// Coordinator Image
	// +optional
	Coordinator MSSQLCoordinator `json:"coordinator,omitempty"`
	// Init container Image
	InitContainer MSSQLInitContainer `json:"initContainer"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`
	// Stash defines backup and restore task definitions.
	// +optional
	Stash appcat.StashAddonSpec `json:"stash,omitempty"`
	// SecurityContext is for the additional config for the DB container
	// +optional
	SecurityContext SecurityContext `json:"securityContext"`
	// update constraints
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// MSSQLVersionDatabase is the MSSQL Database image
type MSSQLVersionDatabase struct {
	Image string `json:"image"`
}

// MSSQLCoordinator is the MSSQL coordinator Container image
type MSSQLCoordinator struct {
	Image string `json:"image"`
}

// MSSQLInitContainer is the MSSQL Container initializer
type MSSQLInitContainer struct {
	Image string `json:"image"`
}

// MSSQLVersionPodSecurityPolicy is the MSSQL pod security policies
type MSSQLVersionPodSecurityPolicy struct {
	DatabasePolicyName string `json:"databasePolicyName"`
}

// MSSQLSecurityContext is for additional configuration for the MSSQL database container
type MSSQLSecurityContext struct {
	RunAsUser  *int64 `json:"runAsUser,omitempty"`
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MSSQLVersionList contains a list of MSSQLVersion
type MSSQLVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MSSQLVersion `json:"items"`
}
