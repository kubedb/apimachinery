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
	ResourceCodeRabbitmqVersion     = "rmversion"
	ResourceKindRabbitmqVersion     = "RabbitmqVersion"
	ResourceSingularRabbitmqVersion = "Rabbitmqversion"
	ResourcePluralRabbitmqVersion   = "Rabbitmqversions"
)

// RabbitmqVersion defines a Rabbitmq database version.

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=rabbitmqversions,singular=rabbitmqversion,scope=Cluster,shortName=rmversion,categories={datastore,kubedb,appscode}
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="DB_IMAGE",type="string",JSONPath=".spec.db.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RabbitmqVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RabbitmqVersionSpec `json:"spec,omitempty"`
}

// RabbitmqVersionSpec is the spec for Rabbitmq version
type RabbitmqVersionSpec struct {
	// Version
	Version string `json:"version"`
	// Database Image
	DB RabbitmqVersionDatabase `json:"db"`
	// Database Image
	InitContainer RabbitmqInitContainer `json:"initContainer"`
	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`
	// PSP names
	// +optional
	PodSecurityPolicies RabbitmqVersionPodSecurityPolicy `json:"podSecurityPolicies"`
	// Stash defines backup and restore task definitions.
	// +optional
	Stash appcat.StashAddonSpec `json:"stash,omitempty"`
	// update constraints
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// RabbitmqVersionDatabase is the Rabbitmq Database image
type RabbitmqVersionDatabase struct {
	Image string `json:"image"`
}

// RabbitmqInitContainer is the Rabbitmq init Container image
type RabbitmqInitContainer struct {
	Image string `json:"image"`
}

// RabbitmqVersionPodSecurityPolicy is the Rabbitmq pod security policies
type RabbitmqVersionPodSecurityPolicy struct {
	DatabasePolicyName string `json:"databasePolicyName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqVersionList is a list of RabbitmqVersions
type RabbitmqVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of RedisVersion CRD objects
	Items []RabbitmqVersion `json:"items,omitempty"`
}
