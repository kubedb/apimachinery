/*
Copyright AppsCode Inc. and Contributors.

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
	kmapi "kmodules.xyz/client-go/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindClusterStatus = "ClusterStatus"
	ResourceClusterStatus     = "clusterstatus"
	ResourceClusterStatuses   = "clusterstatuses"
)

const (
	KeyACEManaged = "ace.appscode.com/managed"
)

// ClusterStatus is the Schema for any resource supported by resource-metrics library

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterStatus struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	Response *ClusterStatusResponse `json:"response,omitempty"`
}

type ClusterStatusResponse struct {
	// Phase represents current status of the cluster
	// +optional
	Phase ClusterPhase `json:"phase,omitempty"`
	// Reason explains the reason behind the cluster current phase
	// +optional
	Reason ClusterPhaseReason `json:"reason,omitempty"`
	// Message specifies additional information regarding the possible actions for the user
	// +optional
	Message string `json:"message,omitempty"`
	// +optional
	ClusterManagers []string `json:"clusterManagers,omitempty"`
	// ClusterAPI contains capi cluster information if the cluster is created by cluster-api
	// +optional
	ClusterAPI *kmapi.CAPIClusterInfo `json:"clusterAPI,omitempty"`
	// +optional
	ClusterMetadata *kmapi.ClusterMetadata `json:"clusterMetadata,omitempty"`
}

// +kubebuilder:validation:Enum=Active;Inactive;NotReady;NotConnected;Registered;NotImported;Lost
type ClusterPhase string

const (
	ClusterPhaseActive       ClusterPhase = "Active"
	ClusterPhaseInactive     ClusterPhase = "Inactive"
	ClusterPhaseNotReady     ClusterPhase = "NotReady"
	ClusterPhaseNotConnected ClusterPhase = "NotConnected"
	ClusterPhaseRegistered   ClusterPhase = "Registered"
	ClusterPhaseNotImported  ClusterPhase = "NotImported"
	ClusterPhaseLost         ClusterPhase = "Lost"
)

// +kubebuilder:validation:Enum=Unknown;ClusterNotFound;AuthIssue;MissingComponent
type ClusterPhaseReason string

const (
	ClusterPhaseReasonReasonUnknown    ClusterPhaseReason = "Unknown"
	ClusterPhaseReasonClusterNotFound  ClusterPhaseReason = "ClusterNotFound"
	ClusterPhaseReasonAuthIssue        ClusterPhaseReason = "AuthIssue"
	ClusterPhaseReasonMissingComponent ClusterPhaseReason = "MissingComponent"
)
