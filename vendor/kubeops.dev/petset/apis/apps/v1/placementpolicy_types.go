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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

const (
	ResourceKindPlacementPolicy = "PlacementPolicy"
	ResourcePlacementPolicy     = "placementpolicy"
	ResourcePlacementPolicies   = "placementpolicies"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster

// PlacementPolicy represents a set of pods with consistent identities.
// Identities are defined as:
//   - Network: A single stable DNS and hostname.
//   - Storage: As many VolumeClaims as requested.
//
// The PlacementPolicy guarantees that a given network identity will always
// map to the same storage identity.
type PlacementPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired identities of pods in this set.
	// +optional
	Spec PlacementPolicySpec `json:"spec,omitempty"`
}

// A PlacementPolicySpec is the specification of a PlacementPolicy.
type PlacementPolicySpec struct {
	// +optional
	ZoneSpreadConstraint *ZoneSpreadConstraint `json:"zoneSpreadConstraint,omitempty"`

	// +optional
	NodeSpreadConstraint *NodeSpreadConstraint `json:"nodeSpreadConstraint,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *Affinity `json:"affinity,omitempty"`

	// ClusterSpreadConstraint provides spec for distributed pod placements
	// +optional
	ClusterSpreadConstraint *ClusterSpreadConstraint `json:"clusterSpreadConstraint,omitempty"`
}

// https://kubeslice.io/documentation/open-source/1.4.0/install-kubeslice/yaml/yaml-controller-install#create-project-namespace
// https://kubeslice.io/documentation/open-source/1.4.0/install-kubeslice/yaml/slice-operations/slice-operations-slice-creation#serviceexport-dns

/*
A cluster cannot be registered to multiple projects.
The KubeSlice worker controller must be deployed in the kubeslice-system namespace.
When a SliceConfig resource is created, KubeSlice generates a Slice resource in the kubeslice-system namespace across all clusters specified in the SliceConfig resource.
When a sliceName is specified in a ServiceExport, the KubeSlice worker controller retrieves the Slice resource from the kubeslice-system namespace and verifies whether the namespace associated with the ServiceExport is onboarded.

So, the project field is not required in the placementPolicy API . But if want to validate whether the clusters specified in the placementPolicy are included in the SliceConfig's cluster list then we need to include the project name as well.
*/
type ClusterSpreadConstraint struct {
	DistributionRules []DistributionRule `json:"distributionRules"`
	Slice             KubeSliceConfig    `json:"slice"`

	// FailoverPolicy, when set, marks this as a cross data center (DC/DR) deployment
	// and selects how the common DC failover service drives failover for it.
	// When nil, the placement is not managed for DC/DR.
	// +optional
	FailoverPolicy *FailoverPolicy `json:"failoverPolicy,omitempty"`
}

type DistributionRule struct {
	ClusterName string `json:"clusterName"`

	// ReplicaIndices are the pod ordinals this cluster runs. For a database with
	// node roles (topology mode) leave it empty and set Topology instead.
	// +optional
	ReplicaIndices []int32 `json:"replicaIndices,omitempty"`

	// Topology places a database's node roles, for databases that run several
	// kinds of nodes (for example Elasticsearch master, ingest and data nodes).
	// The key is the role name as used in the database's spec.topology, the value
	// lists the ordinals of that role this cluster runs, numbered across all
	// clusters like ReplicaIndices. A rule sets either ReplicaIndices or Topology.
	// +optional
	Topology map[string][]int32 `json:"topology,omitempty"`

	StorageClassName string      `json:"storageClassName,omitempty"`
	Monitoring       *Monitoring `json:"monitoring,omitempty"`

	// Role describes how this data center participates in DC/DR failover.
	// Member: data bearing and primary eligible (a candidate for the primary DC).
	// Arbiter: votes only, holds no data, never primary.
	// Witness: data bearing but never primary (for engines like MongoDB whose
	// witness must carry data to satisfy majority writes, yet must not be elected).
	// Defaults to Member when empty.
	// +kubebuilder:validation:Enum=Member;Arbiter;Witness
	// +optional
	Role DCRole `json:"role,omitempty"`
}

// DCRole is the role a data center plays in DC/DR failover.
type DCRole string

const (
	// DCRoleMember is data bearing and primary eligible.
	DCRoleMember DCRole = "Member"
	// DCRoleArbiter votes only, holds no data, and is never primary.
	DCRoleArbiter DCRole = "Arbiter"
	// DCRoleWitness is data bearing but never primary.
	DCRoleWitness DCRole = "Witness"
)

// FailoverMode is the data center topology of a DC/DR deployment.
type FailoverMode string

const (
	// FailoverModeTwoDC is two Members plus one Arbiter or Witness.
	FailoverModeTwoDC FailoverMode = "TwoDC"
	// FailoverModeThreeDC is three Members, all primary eligible.
	FailoverModeThreeDC FailoverMode = "ThreeDC"
)

// FailoverPolicy selects the mode and the failover group of a DC/DR deployment.
type FailoverPolicy struct {
	// Mode is the DC topology. It can be derived from the per rule roles
	// (two Members plus an Arbiter or Witness is TwoDC, three Members is ThreeDC);
	// when set it is validated against the roles.
	// +kubebuilder:validation:Enum=TwoDC;ThreeDC
	// +optional
	Mode FailoverMode `json:"mode,omitempty"`

	// FailoverGroupRef names the FailoverGroup this workload fails over with. Every
	// group is backed by its own primary-dc-<group> Lease. When nil, the workload
	// follows the global primary-dc Lease. The referenced FailoverGroup object is not
	// required for failover: the Lease is derived from this reference alone, the
	// object only adds ordering (dependsOn) and status.
	// +optional
	FailoverGroupRef *v1.LocalObjectReference `json:"failoverGroupRef,omitempty"`
}

type Monitoring struct {
	Prometheus *mona.Prometheus `json:"prometheus,omitempty"`
}

type KubeSliceConfig struct {
	ProjectNamespace string `json:"projectNamespace"`
	SliceName        string `json:"sliceName"`
}

type ZoneSpreadConstraint struct {
	// +kubebuilder:default=1
	MaxSkew int32 `json:"maxSkew"`
	// +kubebuilder:default=DoNotSchedule
	WhenUnsatisfiable v1.UnsatisfiableConstraintAction `json:"whenUnsatisfiable"`
}

type NodeSpreadConstraint struct {
	// +kubebuilder:default=1
	MaxSkew int32 `json:"maxSkew"`
	// +kubebuilder:default=DoNotSchedule
	WhenUnsatisfiable v1.UnsatisfiableConstraintAction `json:"whenUnsatisfiable"`
}

// Affinity is a group of affinity scheduling rules.
type Affinity struct {
	// Describes node affinity scheduling rules for the pod.
	// +optional
	NodeAffinity []NodeAffinityRule `json:"nodeAffinity,omitempty"`
}

type NodeAffinityRule struct {
	TopologyKey string           `json:"topologyKey"`
	Domains     []TopologyDomain `json:"domains"`
	// +kubebuilder:default=DoNotSchedule
	WhenUnsatisfiable v1.UnsatisfiableConstraintAction `json:"whenUnsatisfiable"`
	// Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100.
	// +kubebuilder:default=50
	Weight int32 `json:"weight"`
}

type TopologyDomain struct {
	Values   []string `json:"values"`
	Replicas string   `json:"replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PlacementPolicyList is a collection of PlacementPolicys.
type PlacementPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of stateful sets.
	Items []PlacementPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PlacementPolicy{}, &PlacementPolicyList{})
}
