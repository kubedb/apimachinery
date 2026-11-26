/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"fmt"

	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	meta_util "kmodules.xyz/client-go/meta"
)

// ClickHouse cross data center disaster recovery (DC-DR), Keeper learner design.
//
// One ClickHouse object (spec.distributed=true plus the dr.kubedb.com/enabled=true annotation)
// runs, in each Member data center of its PlacementPolicy, one Keeper PetSet and one PetSet per
// shard. Every shard has replicas in both data centers. The active data center (decided by the
// dr-controlplane primary-DC Lease) runs the Keeper voters; the standby data center runs one Keeper
// learner of that ensemble plus parked Keeper members, and its replicas fetch parts asynchronously
// while staying read-only. The clickhouse-coordinator sidecar fences a data center whose marker is
// stale, promotes the standby after a failover, and moves the Keeper quorum for a planned switchover.

// ClickHouseDCDRSpec holds the DC-DR options of a distributed ClickHouse. It only has an effect
// when spec.distributed is true and the DC-DR annotation (dr.kubedb.com/enabled=true) is set.
type ClickHouseDCDRSpec struct {
	// FenceMarginSeconds is how long the standby data center waits, after it holds the primary-DC
	// Lease, before it promotes itself. It must cover the time the old active data center needs to
	// fence itself once its marker is stale. Defaults to 15.
	// +optional
	FenceMarginSeconds *int32 `json:"fenceMarginSeconds,omitempty"`

	// MaxLearnerLagZxid is the largest number of Keeper transactions the standby learner may be
	// behind the leader for the database to report itself protected. nil means 1000.
	// +optional
	MaxLearnerLagZxid *int64 `json:"maxLearnerLagZxid,omitempty"`

	// MaxFetchLagSeconds is the largest replication delay (system.replicas.absolute_delay) of a
	// standby replica for the database to report itself protected. nil means 60.
	// +optional
	MaxFetchLagSeconds *int32 `json:"maxFetchLagSeconds,omitempty"`
}

const (
	ClickHouseDCDRDefaultFenceMarginSeconds int32 = 15
	ClickHouseDCDRDefaultMaxLearnerLagZxid  int64 = 1000
	ClickHouseDCDRDefaultMaxFetchLagSeconds int32 = 60
)

// ClickHouseDRPhase is the cross data center DR phase of a distributed ClickHouse.
type ClickHouseDRPhase string

const (
	// ClickHouseDRPhaseSteady: the active data center has its Keeper voters, the standby learner and
	// replicas follow it.
	ClickHouseDRPhaseSteady ClickHouseDRPhase = "Steady"
	// ClickHouseDRPhaseSwitchingOver: a planned switchover is moving the Keeper quorum.
	ClickHouseDRPhaseSwitchingOver ClickHouseDRPhase = "SwitchingOver"
	// ClickHouseDRPhaseFailingOver: the standby data center is promoting itself after the Lease moved.
	ClickHouseDRPhaseFailingOver ClickHouseDRPhase = "FailingOver"
	// ClickHouseDRPhaseFailingBack: the old active data center is rejoining as the standby.
	ClickHouseDRPhaseFailingBack ClickHouseDRPhase = "FailingBack"
	// ClickHouseDRPhaseDegraded: only one data center is usable.
	ClickHouseDRPhaseDegraded ClickHouseDRPhase = "Degraded"
)

// ClickHouseDCKeeperRole is what the Keeper members of a data center are in the ensemble.
type ClickHouseDCKeeperRole string

const (
	// ClickHouseDCKeeperRoleVoters: the data center runs the voting Keeper ensemble (active DC).
	ClickHouseDCKeeperRoleVoters ClickHouseDCKeeperRole = "Voters"
	// ClickHouseDCKeeperRoleLearner: the data center runs one learner of the active ensemble plus
	// parked members (standby DC).
	ClickHouseDCKeeperRoleLearner ClickHouseDCKeeperRole = "Learner"
)

// ClickHouseKeeperMemberType is a Keeper raft member type, as reported by /keeper/config.
type ClickHouseKeeperMemberType string

const (
	ClickHouseKeeperMemberParticipant ClickHouseKeeperMemberType = "participant"
	ClickHouseKeeperMemberLearner     ClickHouseKeeperMemberType = "learner"
	// ClickHouseKeeperMemberParked is a Keeper pod that is configured but not part of any ensemble.
	ClickHouseKeeperMemberParked ClickHouseKeeperMemberType = "parked"
)

// ClickHouseDisasterRecoveryStatus reports the DC-DR view of a distributed ClickHouse. The cross-DC
// decision is owned by the dr-controlplane primary-DC Lease; this status reflects it on the single
// ClickHouse object.
type ClickHouseDisasterRecoveryStatus struct {
	// ActiveDC is the data center that holds the primary-DC Lease and runs the Keeper voters.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase ClickHouseDRPhase `json:"phase,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Protected reports whether the standby data center can take over within the lag budgets, as of
	// a fresh observation: its learner and replicas follow the active data center. nil means unknown,
	// which must be treated as unprotected.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// ProtectionMessage explains the current value of Protected.
	// +optional
	ProtectionMessage string `json:"protectionMessage,omitempty"`

	// KeeperLeader is the Keeper pod that currently leads the ensemble.
	// +optional
	KeeperLeader string `json:"keeperLeader,omitempty"`

	// KeeperMembers is the Keeper membership as read from /keeper/config, plus parked members.
	// +optional
	KeeperMembers []ClickHouseKeeperMemberStatus `json:"keeperMembers,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []ClickHouseDCStatus `json:"dataCenters,omitempty"`

	// DataAtRiskParts is the number of parts that existed only in the old active data center when the
	// standby was promoted (dropped replica queue entries). It is the measured data loss of the last
	// unplanned failover; 0 or unset after a planned switchover.
	// +optional
	DataAtRiskParts *int64 `json:"dataAtRiskParts,omitempty"`
}

// ClickHouseKeeperMemberStatus is one Keeper member.
type ClickHouseKeeperMemberStatus struct {
	// ID is the Keeper server id. Ids are never reused.
	ID int32 `json:"id"`
	// Pod is the Keeper pod name.
	Pod string `json:"pod"`
	// DataCenter is the data center the pod runs in.
	// +optional
	DataCenter string `json:"dataCenter,omitempty"`
	// Type is participant, learner or parked.
	Type ClickHouseKeeperMemberType `json:"type"`
}

// ClickHouseDCStatus is one data center's view inside a distributed ClickHouse.
type ClickHouseDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the clusterName used in the
	// PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter.
	// +optional
	Role string `json:"role,omitempty"`

	// KeeperRole is Voters (active DC) or Learner (standby DC).
	// +optional
	KeeperRole ClickHouseDCKeeperRole `json:"keeperRole,omitempty"`

	// Writable is true when application users of this data center can write (the coordinator has
	// removed the fence). It is only evidence when ObservedAt is fresh.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// Fenced is true when the coordinator holds this data center fenced (Keeper voters stopped
	// and/or the read-only application profile in place).
	// +optional
	Fenced bool `json:"fenced,omitempty"`

	// OnlineReplicas is the number of replicas of this data center that are up and connected to
	// Keeper.
	// +optional
	OnlineReplicas *int32 `json:"onlineReplicas,omitempty"`

	// ReadonlyReplicas is the number of replicated tables reporting is_readonly in this data center.
	// +optional
	ReadonlyReplicas *int32 `json:"readonlyReplicas,omitempty"`

	// NotReadyPods names the pods of this data center that are not participating.
	// +optional
	NotReadyPods []string `json:"notReadyPods,omitempty"`

	// LearnerLagZxid is how many Keeper transactions this data center's learner is behind the leader
	// (standby DC only).
	// +optional
	LearnerLagZxid *int64 `json:"learnerLagZxid,omitempty"`

	// MaxAbsoluteDelaySeconds is the largest system.replicas.absolute_delay in this data center.
	// +optional
	MaxAbsoluteDelaySeconds *int64 `json:"maxAbsoluteDelaySeconds,omitempty"`

	// FetchBacklog is the number of GET_PART entries in this data center's replication queues: parts
	// it has not fetched yet (standby DC: data that would be lost if the active DC died now).
	// +optional
	FetchBacklog *int64 `json:"fetchBacklog,omitempty"`

	// DetachedParts is the number of parts in system.detached_parts of this data center's replicas.
	// They are reported, never re-attached automatically.
	// +optional
	DetachedParts *int64 `json:"detachedParts,omitempty"`

	// ObservedAt is when this entry was last measured.
	// +optional
	ObservedAt *metav1.Time `json:"observedAt,omitempty"`
}

// IsDCDRDistributed reports whether this ClickHouse is a cross data center disaster recovery
// deployment: a distributed database explicitly opted in with the DC-DR annotation.
func (c *ClickHouse) IsDCDRDistributed() bool {
	return c.Spec.Distributed && c.Annotations[v1.DCDREnabledAnnotation] == "true"
}

// DCDRPlacementPolicyName is the PlacementPolicy that places a DC-DR ClickHouse: the one named in the
// cluster pod template. The Keeper PetSets are placed into the same Member data centers.
func (c *ClickHouse) DCDRPlacementPolicyName() string {
	if c.Spec.ClusterTopology == nil || c.Spec.ClusterTopology.Cluster.PodTemplate == nil {
		return ""
	}
	if ppp := c.Spec.ClusterTopology.Cluster.PodTemplate.Spec.PodPlacementPolicy; ppp != nil {
		return ppp.Name
	}
	return ""
}

// KeeperPetSetNameForDC is the Keeper PetSet of data center dc.
func (c *ClickHouse) KeeperPetSetNameForDC(dc string) string {
	return meta_util.NameWithSuffix(c.OffshootKeeperName(), dc)
}

// ShardPetSetNameForDC is the PetSet of shard shardNo of cluster clusterName in data center dc.
func (c *ClickHouse) ShardPetSetNameForDC(clusterName string, shardNo int, dc string) string {
	return meta_util.NameWithSuffix(c.OffshootClusterPetSetName(clusterName, shardNo), dc)
}

// DCDRReplicaName is the table replica name ({replica} macro) of ordinal in data center dc. It is
// part of the Keeper path of every replicated table, so it must be unique across data centers and
// never change once a table exists.
func DCDRReplicaName(dc string, ordinal int32) string {
	return fmt.Sprintf("%s-%d", dc, ordinal)
}

// FenceMarginSeconds returns spec.disasterRecovery.fenceMarginSeconds or its default.
func (c *ClickHouse) FenceMarginSeconds() int32 {
	if c.Spec.DisasterRecovery != nil && c.Spec.DisasterRecovery.FenceMarginSeconds != nil {
		return *c.Spec.DisasterRecovery.FenceMarginSeconds
	}
	return ClickHouseDCDRDefaultFenceMarginSeconds
}

// ClickHouseDCDRDataCenter is a Member data center of a DC-DR placement and the replica ordinals
// (PlacementPolicy replicaIndices, per shard) it holds.
// +k8s:deepcopy-gen=false
// +k8s:openapi-gen=false
type ClickHouseDCDRDataCenter struct {
	Name     string
	Ordinals []int32
}

// ClickHouseDCDRReplicasPerDC returns how many replicas of every shard each Member data center runs,
// for shards of `replicas` replicas. A data center runs one replica per ordinal below replicas, so a
// PlacementPolicy may already list ordinals for a later scale up. Every ordinal below replicas must
// belong to exactly one data center.
func ClickHouseDCDRReplicasPerDC(dataCenters []ClickHouseDCDRDataCenter, replicas int32) (map[string]int32, error) {
	owner := map[int32]string{}
	sizes := map[string]int32{}
	for _, dc := range dataCenters {
		sizes[dc.Name] += 0
		for _, o := range dc.Ordinals {
			if prev, ok := owner[o]; ok {
				return nil, fmt.Errorf("replica index %d is placed in both %q and %q", o, prev, dc.Name)
			}
			owner[o] = dc.Name
			if o < replicas {
				sizes[dc.Name]++
			}
		}
	}
	for o := int32(0); o < replicas; o++ {
		if _, ok := owner[o]; !ok {
			return nil, fmt.Errorf("replica index %d (of spec.clusterTopology.cluster.replicas=%d) is not placed in any data center", o, replicas)
		}
	}
	for dc, n := range sizes {
		if n == 0 {
			return nil, fmt.Errorf("data center %q has no replica below spec.clusterTopology.cluster.replicas=%d", dc, replicas)
		}
	}
	return sizes, nil
}

// ValidateDCDR checks the DC-DR specific constraints of the spec. dataCenters are the Member data
// center names taken from the placement policy. It returns warnings for allowed but risky layouts.
func (c *ClickHouse) ValidateDCDR(dataCenters []string) ([]string, error) {
	if !c.IsDCDRDistributed() {
		return nil, nil
	}
	var warnings []string
	ct := c.Spec.ClusterTopology
	if ct == nil {
		return nil, fmt.Errorf("DC-DR requires spec.clusterTopology")
	}
	if c.DCDRPlacementPolicyName() == "" {
		return nil, fmt.Errorf("DC-DR requires spec.clusterTopology.cluster.podTemplate.spec.podPlacementPolicy")
	}
	if c.Spec.DisableSecurity {
		return nil, fmt.Errorf("DC-DR requires authentication: spec.disableSecurity must be false")
	}
	// TLS: the hub issues the certificates (cert-manager on the hub), the operator copies their Secrets
	// to every member data center. A spec.tls without an issuer (client CA refs only) has no server
	// certificate for the cross-DC replication and Distributed links.
	if c.Spec.TLS != nil && c.Spec.TLS.IssuerRef == nil {
		return nil, fmt.Errorf("DC-DR with spec.tls requires spec.tls.issuerRef: the cross data center links need operator issued certificates")
	}
	if ct.Cluster.StorageType == StorageTypeEphemeral || (ct.ClickHouseKeeper != nil && ct.ClickHouseKeeper.Spec != nil && ct.ClickHouseKeeper.Spec.StorageType == StorageTypeEphemeral) {
		return nil, fmt.Errorf("DC-DR needs durable storage: the coordinator state and the Keeper membership live on the data volumes")
	}
	if ct.ClickHouseKeeper == nil || ct.ClickHouseKeeper.ExternallyManaged || ct.ClickHouseKeeper.Spec == nil {
		return nil, fmt.Errorf("DC-DR requires an operator managed Keeper: spec.clusterTopology.clickHouseKeeper.spec, not externallyManaged")
	}
	switch k := ptr.Deref(ct.ClickHouseKeeper.Spec.Replicas, 0); k {
	case 3, 5:
	default:
		return nil, fmt.Errorf("DC-DR runs spec.clusterTopology.clickHouseKeeper.spec.replicas Keeper pods in EACH data center: it must be 3 or 5, got %d", k)
	}
	seen := map[string]bool{}
	for _, dc := range dataCenters {
		if seen[dc] {
			return nil, fmt.Errorf("data center %q is listed twice in the placement", dc)
		}
		seen[dc] = true
	}
	if len(dataCenters) != 2 {
		return nil, fmt.Errorf("ClickHouse DC-DR needs exactly 2 member data centers (an active and a standby), got %d", len(dataCenters))
	}
	if ptr.Deref(ct.Cluster.Replicas, 0) < 2 {
		return nil, fmt.Errorf("DC-DR needs at least 2 replicas per shard (spec.clusterTopology.cluster.replicas), one in each data center")
	}
	if ptr.Deref(ct.Cluster.Replicas, 0) < 4 {
		warnings = append(warnings, "with fewer than 2 replicas per shard in each data center, losing one pod of the active data center stops writes to its shard until the pod returns")
	}
	return warnings, nil
}
