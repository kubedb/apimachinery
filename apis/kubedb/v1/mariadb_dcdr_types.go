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

// MariaDBDCDRSpec holds the cross data center disaster recovery (DC-DR) options of a
// distributed MariaDB Galera cluster. It only has an effect when spec.distributed is true and
// the DC-DR annotation (dr.kubedb.com/enabled=true) is set. Each Member data center then runs
// its own Galera cluster; one node of a passive data center pulls from the active one through
// an asynchronous GTID channel.
type MariaDBDCDRSpec struct {
	// BestEffortCrossDCLagTransactions is the largest number of transactions a passive data
	// center may be behind the active one for an automatic failover to be allowed. A failover
	// that would exceed it is held back until it is overridden with an annotation. nil means no
	// budget is enforced.
	// +optional
	BestEffortCrossDCLagTransactions *int64 `json:"bestEffortCrossDCLagTransactions,omitempty"`

	// ChannelConnectRetrySeconds is MASTER_CONNECT_RETRY of the cross-DC replication channel.
	// Defaults to 5.
	// +optional
	ChannelConnectRetrySeconds *int32 `json:"channelConnectRetrySeconds,omitempty"`
}

// MariaDBDRPhase is the cross data center DR phase of a distributed MariaDB.
type MariaDBDRPhase string

const (
	MariaDBDRPhaseSteady      MariaDBDRPhase = "Steady"
	MariaDBDRPhaseFailingOver MariaDBDRPhase = "FailingOver"
	MariaDBDRPhaseFailingBack MariaDBDRPhase = "FailingBack"
	MariaDBDRPhaseDegraded    MariaDBDRPhase = "Degraded"
)

// MariaDBDisasterRecoveryStatus reports the per data center DC-DR view of a distributed
// MariaDB. The cross-DC decision is owned by the dr-controlplane primary-DC Lease; this status
// reflects it on the single MariaDB object.
type MariaDBDisasterRecoveryStatus struct {
	// ActiveDC is the data center that currently holds the primary DC Lease and runs the
	// writable Galera cluster.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase MariaDBDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []MariaDBDCStatus `json:"dataCenters,omitempty"`

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

// MariaDBDCStatus is one data center's local view inside a distributed MariaDB.
type MariaDBDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same clusterName
	// used in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter.
	// +optional
	Role string `json:"role,omitempty"`

	// ChannelHolder is the node of this data center that runs the cross-DC channel. Empty in
	// the active data center.
	// +optional
	ChannelHolder string `json:"channelHolder,omitempty"`

	// Source is the node of the active data center the channel pulls from.
	// +optional
	Source string `json:"source,omitempty"`

	// Writable is true when the nodes of this data center have read_only=OFF. It is only
	// evidence when WritableObservedAt is fresh.
	// +optional
	Writable bool `json:"writable,omitempty"`

	// WritableObservedAt is when Writable was last established by probing the nodes.
	// +optional
	WritableObservedAt *metav1.Time `json:"writableObservedAt,omitempty"`

	// SyncedNodes is the number of nodes of this data center that are Synced members of its
	// Primary Galera component.
	// +optional
	SyncedNodes *int32 `json:"syncedNodes,omitempty"`

	// ChannelState is the state of the cross-DC replication channel on the channel holder:
	// "Running" (IO and SQL threads ON), "Stopped", "Connecting", "Error" or "None" (the active
	// DC has no channel).
	// +optional
	ChannelState string `json:"channelState,omitempty"`

	// NotReadyPods names the pods in this data center that are not Synced.
	// +optional
	NotReadyPods []string `json:"notReadyPods,omitempty"`

	// GTIDLag is this DC's cross-DC lag behind the active DC, in transactions of the active
	// DC's GTID domain.
	// +optional
	GTIDLag *int64 `json:"gtidLag,omitempty"`

	// LagObservedAt is when GTIDLag was last measured.
	// +optional
	LagObservedAt *metav1.Time `json:"lagObservedAt,omitempty"`
}

// MariaDBWritableObservationTTL is how long a MariaDBDCStatus.WritableObservedAt stamp stays
// usable as evidence; past it the DC is treated as having no confirmed writable cluster.
const MariaDBWritableObservationTTL = 5 * time.Minute

// WritableConfirmed reports whether this data center is positively observed, recently, to be
// writable.
func (d *MariaDBDCStatus) WritableConfirmed(now time.Time) bool {
	if d == nil || !d.Writable || d.WritableObservedAt == nil {
		return false
	}
	return now.Sub(d.WritableObservedAt.Time) <= MariaDBWritableObservationTTL
}

// IsDCDRDistributed reports whether this MariaDB is a cross data center disaster recovery
// deployment: a distributed database explicitly opted in with the DC-DR annotation. Without
// the annotation spec.distributed keeps its meaning of one Galera cluster stretched over the
// spokes.
func (m MariaDB) IsDCDRDistributed() bool {
	return m.Spec.Distributed && m.Annotations[DCDREnabledAnnotation] == "true"
}

// PetSetNameForDC is the name of the PetSet that runs the Galera cluster of data center dc.
func (m MariaDB) PetSetNameForDC(dc string) string {
	return meta_util.NameWithSuffix(m.OffshootName(), dc)
}

// GoverningServiceNameForDC is the headless Service of the PetSet of data center dc.
func (m MariaDB) GoverningServiceNameForDC(dc string) string {
	return meta_util.NameWithSuffix(m.PetSetNameForDC(dc), "pods")
}

// PrimaryServiceNameForDC selects the writable nodes of data center dc.
func (m MariaDB) PrimaryServiceNameForDC(dc string) string {
	return m.PetSetNameForDC(dc)
}

// GaleraClusterNameForDC is wsrep_cluster_name of data center dc, so the Galera clusters of
// two data centers can never join each other.
func (m MariaDB) GaleraClusterNameForDC(dc string) string {
	return m.PetSetNameForDC(dc)
}

// MariaDBServerIDForDC is server_id of every node of the data center at position dcIndex in the
// placement. wsrep_gtid_mode needs the same server_id on all nodes of one Galera cluster; it
// must differ between data centers.
func MariaDBServerIDForDC(dcIndex int) int {
	return 100 * (dcIndex + 1)
}

// MariaDBGTIDDomainIDForDC is the initial wsrep_gtid_domain_id of the data center at position
// dcIndex. A re-seeded data center moves to a new domain id (RESET MASTER is refused inside
// Galera), which the init script keeps on the data volume.
func MariaDBGTIDDomainIDForDC(dcIndex int) int {
	return 10 * (dcIndex + 1)
}

// ValidateDCDR checks the DC-DR specific constraints of the spec. dataCenters are the Member
// data center names taken from the placement policy, replicas their node counts.
func (m MariaDB) ValidateDCDR(dataCenters []string, replicas []int32) error {
	if !m.IsDCDRDistributed() {
		return nil
	}
	if !m.IsGaleraCluster() {
		return fmt.Errorf("DC-DR requires spec.topology.mode=%s", MariaDBModeGaleraCluster)
	}
	if ppp := m.Spec.PodTemplate.Spec.PodPlacementPolicy; ppp == nil || ppp.Name == "" {
		return fmt.Errorf("DC-DR requires spec.podTemplate.spec.podPlacementPolicy")
	}
	if len(dataCenters) != len(replicas) {
		return fmt.Errorf("got %d data centers but %d replica counts", len(dataCenters), len(replicas))
	}
	seen := map[string]bool{}
	for i, dc := range dataCenters {
		if seen[dc] {
			return fmt.Errorf("data center %q is listed twice in the placement", dc)
		}
		seen[dc] = true
		// a Galera cluster keeps quorum only with an odd number of nodes, and a lone node of a
		// 3-node cluster is non-Primary: every data center needs 3, 5, 7 or 9 nodes
		if n := replicas[i]; n < 3 || n > 9 || n%2 == 0 {
			return fmt.Errorf("data center %q: DC-DR needs an odd number of 3 to 9 Galera nodes per data center, got %d", dc, n)
		}
	}
	if len(dataCenters) < 2 {
		return fmt.Errorf("DC-DR needs at least 2 member data centers, got %d", len(dataCenters))
	}
	return nil
}
