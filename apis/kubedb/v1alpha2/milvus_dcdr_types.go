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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
)

// MilvusDCDRSpec holds the cross data center disaster recovery (DC-DR) options of a Milvus
// with spec.distributed set. Every Member data center of the PlacementPolicy named by
// spec.podTemplate.spec.podPlacementPolicy runs its own complete Milvus with its own meta etcd
// and object storage. The data center that holds the dr-controlplane primary-DC Lease is the
// primary; the others are standbys that receive its write-ahead log through Milvus CDC.
type MilvusDCDRSpec struct {
	// DataCenters holds the storage of each Member data center, one entry per data center.
	// Each data center needs its own meta etcd and its own bucket or root path.
	// +kubebuilder:validation:MinItems=2
	DataCenters []MilvusDataCenterSpec `json:"dataCenters"`

	// CDC configures the CDC pod that runs in every data center. It always runs one replica.
	// +optional
	CDC *MilvusCDCSpec `json:"cdc,omitempty"`

	// MaxFailoverLagSeconds is the largest replication lag, in seconds, a standby may have been
	// observed with for an unplanned failover to promote it without the
	// dr.kubedb.com/accept-failover-data-loss annotation. Defaults to 30.
	// +optional
	MaxFailoverLagSeconds *int64 `json:"maxFailoverLagSeconds,omitempty"`

	// MaxRebuildLossSeconds is the largest span, in seconds, of writes that a returning old
	// primary holds and the new primary never received, which an automatic rebuild may discard.
	// A larger span waits for the dr.kubedb.com/accept-rebuild-data-loss annotation. Defaults to 0:
	// any unreplicated write blocks the automatic rebuild.
	// +optional
	MaxRebuildLossSeconds *int64 `json:"maxRebuildLossSeconds,omitempty"`
}

// MilvusDataCenterSpec is the storage of one Member data center of a DC-DR Milvus.
type MilvusDataCenterSpec struct {
	// Name is the data center, named by its OCM managed cluster (the clusterName of the
	// PlacementPolicy distribution rule).
	Name string `json:"name"`

	// MetaStorage is the meta etcd of this data center. It must be externally managed and
	// reachable from this data center's pods and from the hub.
	MetaStorage *MetaStorageSpec `json:"metaStorage"`

	// ObjectStorage is the object storage of this data center. Its bucket or root path must
	// differ from every other data center's.
	ObjectStorage *ObjectStorageSpec `json:"objectStorage"`
}

// MilvusCDCSpec configures the CDC pod of a DC-DR Milvus.
type MilvusCDCSpec struct {
	// PodTemplate is an optional configuration for the CDC pod.
	// +optional
	PodTemplate *ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// MilvusDRPhase is the cross data center DR phase of a DC-DR Milvus.
type MilvusDRPhase string

const (
	MilvusDRPhaseSteady      MilvusDRPhase = "Steady"
	MilvusDRPhaseFailingOver MilvusDRPhase = "FailingOver"
	MilvusDRPhaseDegraded    MilvusDRPhase = "Degraded"
	MilvusDRPhaseRebuilding  MilvusDRPhase = "Rebuilding"
)

// MilvusReplicationRole is the replication role a data center's Milvus reports.
type MilvusReplicationRole string

const (
	// MilvusReplicationRolePrimary accepts client writes and is the CDC source.
	MilvusReplicationRolePrimary MilvusReplicationRole = "Primary"
	// MilvusReplicationRoleStandby rejects client writes and receives the primary's WAL.
	MilvusReplicationRoleStandby MilvusReplicationRole = "Standby"
	// MilvusReplicationRoleUnconfigured has no replication configuration yet.
	MilvusReplicationRoleUnconfigured MilvusReplicationRole = "Unconfigured"
	// MilvusReplicationRoleUnknown could not be observed.
	MilvusReplicationRoleUnknown MilvusReplicationRole = "Unknown"
)

// MilvusDisasterRecoveryStatus reports the per data center DC-DR view of a Milvus. The cross-DC
// decision is owned by the dr-controlplane primary-DC Lease; this status reflects it.
type MilvusDisasterRecoveryStatus struct {
	// ActiveDC is the data center that holds the primary-DC Lease.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase MilvusDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member data center.
	// +optional
	DataCenters []MilvusDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Protected reports whether a standby data center receives the primary's WAL within the lag
	// budget, as of a fresh observation. nil means unknown, which must be treated as unprotected.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// ProtectionMessage explains the current value of Protected.
	// +optional
	ProtectionMessage string `json:"protectionMessage,omitempty"`

	// LastProtectedAt is when a standby was last observed caught up within the lag budget. An
	// unplanned failover relies on it when the old primary can no longer be reached.
	// +optional
	LastProtectedAt *metav1.Time `json:"lastProtectedAt,omitempty"`

	// Promotion records the last promotion of a data center. It is kept after the promotion so
	// that a rebuild of the old primary knows what was lost.
	// +optional
	Promotion *MilvusDRPromotion `json:"promotion,omitempty"`

	// Rebuild reports the rebuild of an old primary, while it runs.
	// +optional
	Rebuild *MilvusDRRebuild `json:"rebuild,omitempty"`
}

// MilvusDCStatus is one data center's view inside a DC-DR Milvus.
type MilvusDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster.
	ClusterName string `json:"clusterName"`

	// ClusterID is the Milvus cluster id of this data center (msgChannel.chanNamePrefix.cluster).
	// +optional
	ClusterID string `json:"clusterID,omitempty"`

	// ReplicationRole is the replication role this data center's Milvus reports.
	// +optional
	ReplicationRole MilvusReplicationRole `json:"replicationRole,omitempty"`

	// Reachable is true when the hub reached this data center's Milvus in the last pass.
	// +optional
	Reachable bool `json:"reachable,omitempty"`

	// LastSeenAt is when the hub last reached this data center's Milvus.
	// +optional
	LastSeenAt *metav1.Time `json:"lastSeenAt,omitempty"`

	// LagSeconds is how far this standby's replicated WAL is behind the current time, using the
	// oldest channel. Unset on the primary.
	// +optional
	LagSeconds *int64 `json:"lagSeconds,omitempty"`

	// LagObservedAt is when LagSeconds was measured.
	// +optional
	LagObservedAt *metav1.Time `json:"lagObservedAt,omitempty"`

	// Healthy reflects whether this data center's dr-controlplane health Lease is fresh.
	// +optional
	Healthy bool `json:"healthy,omitempty"`
}

// MilvusDRPromotion records a promotion of a data center to primary.
type MilvusDRPromotion struct {
	// DC is the promoted data center.
	DC string `json:"dc"`

	// From is the data center that was primary before.
	// +optional
	From string `json:"from,omitempty"`

	// Forced is true for an unplanned failover (force promote). The old primary then needs a
	// rebuild before it can follow the new primary.
	// +optional
	Forced bool `json:"forced,omitempty"`

	// Time is when the promotion completed.
	// +optional
	Time *metav1.Time `json:"time,omitempty"`

	// FromLastSeenAt is when the hub last reached the old primary before a forced promotion.
	// Writes it took after the salvage checkpoints and before this time were lost.
	// +optional
	FromLastSeenAt *metav1.Time `json:"fromLastSeenAt,omitempty"`

	// SalvageCheckpoints are the last positions the promoted data center received from the old
	// primary, per channel. Writes after them were lost by the failover.
	// +optional
	SalvageCheckpoints []MilvusReplicateCheckpoint `json:"salvageCheckpoints,omitempty"`
}

// MilvusReplicateCheckpoint is the last replicated position of one channel.
type MilvusReplicateCheckpoint struct {
	// Pchannel is the source cluster's physical channel.
	Pchannel string `json:"pchannel"`

	// TimeTick is the hybrid timestamp of the last replicated message.
	TimeTick int64 `json:"timeTick"`
}

// MilvusDRRebuildStep is a step of the rebuild of an old primary.
type MilvusDRRebuildStep string

const (
	MilvusDRRebuildStepPending   MilvusDRRebuildStep = "Pending"
	MilvusDRRebuildStepStopping  MilvusDRRebuildStep = "Stopping"
	MilvusDRRebuildStepWiping    MilvusDRRebuildStep = "Wiping"
	MilvusDRRebuildStepStarting  MilvusDRRebuildStep = "Starting"
	MilvusDRRebuildStepLinking   MilvusDRRebuildStep = "Linking"
	MilvusDRRebuildStepSeeding   MilvusDRRebuildStep = "Seeding"
	MilvusDRRebuildStepResuming  MilvusDRRebuildStep = "Resuming"
	MilvusDRRebuildStepCompleted MilvusDRRebuildStep = "Completed"
)

// MilvusDRRebuild reports the rebuild of an old primary as a standby of the active data center.
type MilvusDRRebuild struct {
	// DC is the data center being rebuilt.
	DC string `json:"dc"`

	// Step is the current step.
	Step MilvusDRRebuildStep `json:"step"`

	// StartedAt is when the rebuild started.
	// +optional
	StartedAt *metav1.Time `json:"startedAt,omitempty"`

	// Message explains the current step, for example why it waits.
	// +optional
	Message string `json:"message,omitempty"`
}
