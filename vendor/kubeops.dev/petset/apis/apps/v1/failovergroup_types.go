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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

const (
	ResourceKindFailoverGroup = "FailoverGroup"
	ResourceFailoverGroup     = "failovergroup"
	ResourceFailoverGroups    = "failovergroups"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Lease Holder",type="string",JSONPath=".status.leaseHolder"
// +kubebuilder:printcolumn:name="Active DC",type="string",JSONPath=".status.activeDC"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// FailoverGroup is a set of DC/DR workloads that fail over together, backed by the
// primary-dc-<name> Lease. Workloads join a group through their PlacementPolicy's
// clusterSpreadConstraint.failoverPolicy.failoverGroupRef.
//
// The Lease, not this object, decides which data center is primary. A FailoverGroup
// only orders groups relative to each other (dependsOn) and publishes where each
// group is active. Deleting it never stops a failover.
type FailoverGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec FailoverGroupSpec `json:"spec,omitempty"`
	// +optional
	Status FailoverGroupStatus `json:"status,omitempty"`
}

// FailoverGroupSpec is the specification of a FailoverGroup.
type FailoverGroupSpec struct {
	// DependsOn names the FailoverGroups that must be active and ready on a data
	// center before this group may become active there. This applies to both
	// unplanned failover and planned switchover. The dependencies must all be
	// active on the same data center.
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`
}

// FailoverGroupPhase is the phase of a FailoverGroup.
type FailoverGroupPhase string

const (
	// FailoverGroupPhaseSteady means every member is ready on the active DC.
	FailoverGroupPhaseSteady FailoverGroupPhase = "Steady"
	// FailoverGroupPhaseWaitingForDependency means a dependency is not yet ready on
	// the DC this group has to follow.
	FailoverGroupPhaseWaitingForDependency FailoverGroupPhase = "WaitingForDependency"
	// FailoverGroupPhaseWaitingForMembers means the group's Lease or target DC moved
	// and not every member has reported ready there yet.
	FailoverGroupPhaseWaitingForMembers FailoverGroupPhase = "WaitingForMembers"
	// FailoverGroupPhaseBlocked means the group cannot converge without a human, for
	// example a dependency cycle, a missing dependency, or dependencies active on
	// different data centers.
	FailoverGroupPhaseBlocked FailoverGroupPhase = "Blocked"
)

// FailoverGroupStatus is the observed state of a FailoverGroup.
type FailoverGroupStatus struct {
	// ObservedGeneration is the most recent generation observed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LeaseName is the primary DC Lease backing this group.
	// +optional
	LeaseName string `json:"leaseName,omitempty"`

	// LeaseHolder is the data center currently holding the group's Lease.
	// +optional
	LeaseHolder string `json:"leaseHolder,omitempty"`

	// ActiveDC is the data center where every member of this group is ready, and,
	// for a group with dependencies, where every dependency is active too. It only
	// moves once that holds, so it can lag LeaseHolder during a failover.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the FailoverGroup phase.
	// +optional
	Phase FailoverGroupPhase `json:"phase,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Members are the workloads of this group. Each engine operator reports its own
	// workloads here, using server side apply with its own field manager.
	// +optional
	// +listType=map
	// +listMapKey=apiGroup
	// +listMapKey=kind
	// +listMapKey=namespace
	// +listMapKey=name
	Members []FailoverGroupMember `json:"members,omitempty"`

	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// FailoverGroupMember is one workload's readiness, as reported by its operator.
type FailoverGroupMember struct {
	APIGroup  string `json:"apiGroup"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`

	// DC is the data center this member currently treats as active.
	// +optional
	DC string `json:"dc,omitempty"`

	// Ready is true when the member is positively observed serving on DC, for
	// example a confirmed writable primary.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// ObservedAt is when Ready was last established. A stale report is not
	// treated as ready.
	// +optional
	ObservedAt *metav1.Time `json:"observedAt,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FailoverGroupList is a collection of FailoverGroups.
type FailoverGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FailoverGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FailoverGroup{}, &FailoverGroupList{})
}
