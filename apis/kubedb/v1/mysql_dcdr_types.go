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

package v1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// MySQLDCDRSpec holds the cross data center disaster recovery (DC-DR) options of a
// distributed MySQL. It only has an effect when spec.distributed is true and the
// DC-DR annotation (dr.kubedb.com/enabled=true) is set.
type MySQLDCDRSpec struct {
	// BestEffortCrossDCLagTransactions is the largest number of transactions (GTIDs) a passive
	// data center may be behind the active one for an automatic failover to be allowed. A
	// failover that would exceed it is held back until the operator overrides it with an
	// annotation. nil means no budget is enforced.
	// +optional
	BestEffortCrossDCLagTransactions *int64 `json:"bestEffortCrossDCLagTransactions,omitempty"`

	// ChannelRetryCount is SOURCE_RETRY_COUNT of the cross-DC replication channel. The channel
	// source is the active DC's primary Service, so a primary change on the active side is
	// followed by retrying, not by reconfiguring. Defaults to 60.
	// +optional
	ChannelRetryCount *int32 `json:"channelRetryCount,omitempty"`

	// ChannelConnectRetrySeconds is SOURCE_CONNECT_RETRY of the cross-DC replication channel.
	// Defaults to 5.
	// +optional
	ChannelConnectRetrySeconds *int32 `json:"channelConnectRetrySeconds,omitempty"`
}

// MySQLDRPhase is the cross data center DR phase of a distributed MySQL.
type MySQLDRPhase string

const (
	MySQLDRPhaseSteady      MySQLDRPhase = "Steady"
	MySQLDRPhaseFailingOver MySQLDRPhase = "FailingOver"
	MySQLDRPhaseFailingBack MySQLDRPhase = "FailingBack"
	MySQLDRPhaseDegraded    MySQLDRPhase = "Degraded"
)

// MySQLDisasterRecoveryStatus reports the per data center DC-DR view of a distributed
// MySQL. The cross-DC decision is owned by the dr-controlplane primary-DC Lease; this status
// reflects it on the single MySQL object.
type MySQLDisasterRecoveryStatus struct {
	// ActiveDC is the data center that currently holds the primary DC Lease and runs the
	// writable Group Replication primary.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase MySQLDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []MySQLDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Protected reports whether at least one Member data center other than the active one has
	// its replication channel running (IO and SQL threads ON) within the lag budget, as of a
	// fresh observation. nil means unknown, which must be treated as unprotected.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// ProtectionMessage explains the current value of Protected.
	// +optional
	ProtectionMessage string `json:"protectionMessage,omitempty"`
}

// MySQLDCStatus is one data center's local view inside a distributed MySQL.
type MySQLDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same clusterName
	// used in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter.
	// +optional
	Role string `json:"role,omitempty"`

	// Primary is this DC's Group Replication primary pod.
	// +optional
	Primary string `json:"primary,omitempty"`

	// Writable is true when this DC's primary has super_read_only=OFF. It is only evidence
	// when WritableObservedAt is fresh.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// WritableObservedAt is when Writable was last established by probing the primary.
	// +optional
	WritableObservedAt *metav1.Time `json:"writableObservedAt,omitempty"`

	// OnlineMembers is the number of ONLINE members of this DC's group.
	// +optional
	OnlineMembers *int32 `json:"onlineMembers,omitempty"`

	// ChannelState is the state of the cross-DC replication channel on this DC's primary:
	// "Running" (IO and SQL threads ON), "Stopped", "Connecting", "Error" or "None" (the active
	// DC has no channel).
	// +optional
	ChannelState string `json:"channelState,omitempty"`

	// NotReadyPods names the pods in this data center that are not participating.
	// +optional
	NotReadyPods []string `json:"notReadyPods,omitempty"`

	// GTIDLag is this DC's cross-DC lag behind the active DC, in transactions.
	// +optional
	GTIDLag *int64 `json:"gtidLag,omitempty"`

	// LagObservedAt is when GTIDLag was last measured.
	// +optional
	LagObservedAt *metav1.Time `json:"lagObservedAt,omitempty"`
}

// MySQLWritableObservationTTL is how long a MySQLDCStatus.WritableObservedAt stamp stays
// usable as evidence; past it the DC is treated as having no confirmed writable primary.
const MySQLWritableObservationTTL = 5 * time.Minute

// WritablePrimaryConfirmed reports whether this data center is positively observed, recently,
// to hold a writable primary.
func (d *MySQLDCStatus) WritablePrimaryConfirmed(now time.Time) bool {
	if d == nil || !d.Writable || d.Primary == "" || d.WritableObservedAt == nil {
		return false
	}
	return now.Sub(d.WritableObservedAt.Time) <= MySQLWritableObservationTTL
}

// IsDCDRDistributed reports whether this MySQL is a cross data center disaster recovery
// deployment: a distributed database explicitly opted in with the DC-DR annotation.
func (m MySQL) IsDCDRDistributed() bool {
	return m.Spec.Distributed && m.Annotations[DCDREnabledAnnotation] == "true"
}

// PetSetNameForDC is the name of the PetSet that runs the group of data center dc.
func (m MySQL) PetSetNameForDC(dc string) string {
	return meta_util.NameWithSuffix(m.OffshootName(), dc)
}

// GoverningServiceNameForDC is the headless Service of the PetSet of data center dc.
func (m MySQL) GoverningServiceNameForDC(dc string) string {
	return meta_util.NameWithSuffix(m.PetSetNameForDC(dc), "pods")
}

// PrimaryServiceNameForDC selects only the Group Replication primary of data center dc. The
// passive DC uses the active DC's one as the source of its replication channel.
func (m MySQL) PrimaryServiceNameForDC(dc string) string {
	return m.PetSetNameForDC(dc)
}

// ServerIDOffsetForDC is added to the pod ordinal based server_id so that the PetSets of two
// data centers never collide. dcIndex is the position of the DC in the placement.
func ServerIDOffsetForDC(dcIndex int) int {
	return 100 * dcIndex
}

// ValidateDCDR checks the DC-DR specific constraints of the spec. dataCenters are the Member
// data center names taken from the placement policy.
func (m MySQL) ValidateDCDR(dataCenters []string) error {
	if !m.IsDCDRDistributed() {
		return nil
	}
	if m.Spec.Topology == nil || m.Spec.Topology.Mode == nil || *m.Spec.Topology.Mode != MySQLModeGroupReplication {
		return fmt.Errorf("DC-DR requires spec.topology.mode=%s", MySQLModeGroupReplication)
	}
	if m.Spec.Topology.Group != nil && m.Spec.Topology.Group.Mode != nil && *m.Spec.Topology.Group.Mode != MySQLGroupModeSinglePrimary {
		return fmt.Errorf("DC-DR requires spec.topology.group.mode=%s", MySQLGroupModeSinglePrimary)
	}
	if ppp := m.Spec.PodTemplate.Spec.PodPlacementPolicy; ppp == nil || ppp.Name == "" {
		return fmt.Errorf("DC-DR requires spec.podTemplate.spec.podPlacementPolicy")
	}
	seen := map[string]bool{}
	for _, dc := range dataCenters {
		if seen[dc] {
			return fmt.Errorf("data center %q is listed twice in the placement", dc)
		}
		seen[dc] = true
	}
	if len(dataCenters) < 2 {
		return fmt.Errorf("DC-DR needs at least 2 member data centers, got %d", len(dataCenters))
	}
	return nil
}
