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
	meta_util "kmodules.xyz/client-go/meta"
)

// RabbitMQ cross data center disaster recovery (DC-DR), shovel outbox design.
//
// One RabbitMQ object (spec.distributed=true plus the dr.kubedb.com/enabled=true annotation, which
// the operator sets when the PlacementPolicy has a failoverPolicy) runs one self-contained RabbitMQ
// cluster per Member data center. Quorum queues and the metadata store never cross a data center.
// The active data center (decided by the dr-controlplane primary-DC Lease) takes client traffic and
// copies every message of an application exchange into an outbox queue; a dynamic shovel in the
// standby data center drains the outbox into the same exchange there. The rabbitmq-coordinator
// sidecar fences a data center whose marker does not name it: it revokes the vhost permissions of
// every application user and closes their connections. Delivery across a data center change is
// at least once (acknowledgements are not replicated) and an unplanned failover loses the outbox
// backlog.

// RabbitMQDCDRSpec holds the DC-DR options of a distributed RabbitMQ. It only has an effect when
// spec.distributed is true and the DC-DR annotation (dr.kubedb.com/enabled=true) is set.
type RabbitMQDCDRSpec struct {
	// MaxLagMessages is the largest outbox backlog (messages not yet copied to the standby data
	// center) for the database to report itself protected and for an automatic failover to be
	// allowed without the dr.kubedb.com/accept-failover-data-loss annotation. nil means 1000.
	// +optional
	MaxLagMessages *int64 `json:"maxLagMessages,omitempty"`
}

const (
	RabbitMQDCDRDefaultMaxLagMessages int64 = 1000

	// RabbitMQDCDRLinkUser is the internal user the standby data center's shovels use to read the
	// active data center's outbox queues. It is never fenced.
	RabbitMQDCDRLinkUser = "kubedb-dcdr-link"
	// RabbitMQDCDROutboxPrefix prefixes the outbox queue of every replicated exchange.
	RabbitMQDCDROutboxPrefix = "dr.outbox."
	// RabbitMQDCDRShovelPrefix prefixes the dynamic shovel of every replicated exchange.
	RabbitMQDCDRShovelPrefix = "dr-link-"
	// RabbitMQDCDRPermitParameter is the global parameter the hub writes on the active data center.
	RabbitMQDCDRPermitParameter = "kubedb-dcdr-permit"
	// RabbitMQDCDRPermissionsParameter holds the saved vhost permissions of a fenced data center.
	RabbitMQDCDRPermissionsParameter = "kubedb-dcdr-permissions"
)

// RabbitMQDRPhase is the cross data center DR phase of a distributed RabbitMQ.
type RabbitMQDRPhase string

const (
	// RabbitMQDRPhaseSteady: the active data center is open and the standby follows it.
	RabbitMQDRPhaseSteady RabbitMQDRPhase = "Steady"
	// RabbitMQDRPhaseSwitchingOver: a planned switchover is draining the link.
	RabbitMQDRPhaseSwitchingOver RabbitMQDRPhase = "SwitchingOver"
	// RabbitMQDRPhaseFailingOver: the Lease moved and the new active data center is being opened.
	RabbitMQDRPhaseFailingOver RabbitMQDRPhase = "FailingOver"
	// RabbitMQDRPhaseFailingBack: a returning data center is being reseeded as the standby.
	RabbitMQDRPhaseFailingBack RabbitMQDRPhase = "FailingBack"
	// RabbitMQDRPhaseDegraded: only one data center is usable.
	RabbitMQDRPhaseDegraded RabbitMQDRPhase = "Degraded"
)

// RabbitMQDisasterRecoveryStatus reports the per data center DC-DR view of a distributed RabbitMQ.
// The cross-DC decision is owned by the dr-controlplane primary-DC Lease; this status reflects it on
// the single RabbitMQ object.
type RabbitMQDisasterRecoveryStatus struct {
	// ActiveDC is the data center that currently holds the primary DC Lease and takes client traffic.
	// +optional
	ActiveDC string `json:"activeDC,omitempty"`

	// Phase is the DC-DR phase.
	// +optional
	Phase RabbitMQDRPhase `json:"phase,omitempty"`

	// DataCenters is the per data center view, one entry per Member DC.
	// +optional
	DataCenters []RabbitMQDCStatus `json:"dataCenters,omitempty"`

	// LastTransitionTime is when ActiveDC last changed.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Protected reports whether a standby data center has a running shovel for every replicated
	// exchange and the outbox backlog is within spec.disasterRecovery.maxLagMessages, as of a fresh
	// observation. nil means unknown, which must be treated as unprotected.
	// +optional
	Protected *bool `json:"protected,omitempty"`

	// ProtectionMessage explains the current value of Protected.
	// +optional
	ProtectionMessage string `json:"protectionMessage,omitempty"`

	// PermittedDC is the data center the hub last wrote the permit for, with the Lease it was
	// written under. The coordinator opens a data center only for a permit that matches the Lease.
	// +optional
	PermittedDC string `json:"permittedDC,omitempty"`

	// Message explains a blocked failover or the last error.
	// +optional
	Message string `json:"message,omitempty"`
}

// RabbitMQDCStatus is one data center's local view inside a distributed RabbitMQ.
type RabbitMQDCStatus struct {
	// ClusterName is the data center, named by its OCM managed cluster (the same clusterName used
	// in the PlacementPolicy distributionRule).
	ClusterName string `json:"clusterName"`

	// Role is Member or Arbiter.
	// +optional
	Role string `json:"role,omitempty"`

	// Writable is true when the application users of this data center have their permissions
	// (the coordinator reports it open).
	// +optional
	Writable bool `json:"writable,omitempty"`

	// Fenced is true when the coordinator of this data center has revoked the application users.
	// +optional
	Fenced *bool `json:"fenced,omitempty"`

	// NodesReady is the number of running nodes of this data center's RabbitMQ cluster.
	// +optional
	NodesReady *int32 `json:"nodesReady,omitempty"`

	// OutboxMessages is the number of messages in this data center's outbox queues that the
	// standby has not copied yet. Only set for the active data center.
	// +optional
	OutboxMessages *int64 `json:"outboxMessages,omitempty"`

	// LinkState is the state of this data center's shovels: "Running", "Starting", "Terminated",
	// "Partial" (some exchanges have no running shovel) or "None" (the active data center).
	// +optional
	LinkState string `json:"linkState,omitempty"`

	// NeedsReseed is true when this data center returned after an unplanned failover and keeps its
	// own messages: it stays fenced until the dr.kubedb.com/reseed annotation approves a reseed.
	// +optional
	NeedsReseed bool `json:"needsReseed,omitempty"`

	// Healthy is true when every node of this data center is running and its management API answers.
	// +optional
	Healthy bool `json:"healthy,omitempty"`

	// Message explains an unhealthy or unknown entry.
	// +optional
	Message string `json:"message,omitempty"`

	// ObservedAt is when this entry was last measured.
	// +optional
	ObservedAt *metav1.Time `json:"observedAt,omitempty"`
}

// IsDCDRDistributed reports whether this RabbitMQ is a cross data center disaster recovery
// deployment: a distributed database opted in with the DC-DR annotation.
func (r *RabbitMQ) IsDCDRDistributed() bool {
	return r.Spec.Distributed && r.Annotations[v1.DCDREnabledAnnotation] == "true"
}

// DCDRPlacementPolicyName is the PlacementPolicy that places a DC-DR RabbitMQ: the one named in
// its pod template.
func (r *RabbitMQ) DCDRPlacementPolicyName() string {
	if ppp := r.Spec.PodTemplate.Spec.PodPlacementPolicy; ppp != nil {
		return ppp.Name
	}
	return ""
}

// PetSetNameForDC is the PetSet that runs the RabbitMQ cluster of data center dc.
func (r *RabbitMQ) PetSetNameForDC(dc string) string {
	return meta_util.NameWithSuffix(r.OffshootName(), dc)
}

// GoverningServiceNameForDC is the headless Service of the PetSet of data center dc. RabbitMQ
// peer discovery lists its endpoints, so the clusters of two data centers never see each other.
func (r *RabbitMQ) GoverningServiceNameForDC(dc string) string {
	return meta_util.NameWithSuffix(r.PetSetNameForDC(dc), "pods")
}

// ServiceNameForDC is the client Service (AMQP and management) of data center dc. It is exported
// to the slice, so the hub and the other data center reach it by its slice DNS name.
func (r *RabbitMQ) ServiceNameForDC(dc string) string {
	return r.PetSetNameForDC(dc)
}

// ClusterNameForDC is cluster_name of the RabbitMQ cluster of data center dc.
func (r *RabbitMQ) ClusterNameForDC(dc string) string {
	return r.PetSetNameForDC(dc)
}

// ErlangCookieSecretNameForDC is the Erlang cookie of data center dc. A different cookie per data
// center makes an accidental cross data center join fail at the distribution handshake.
func (r *RabbitMQ) ErlangCookieSecretNameForDC(dc string) string {
	return meta_util.NameWithSuffix(r.PetSetNameForDC(dc), "erlang-cookie")
}

// ConfigSecretNameForDC is the generated configuration of data center dc.
func (r *RabbitMQ) ConfigSecretNameForDC(dc string) string {
	return meta_util.NameWithSuffix(r.PetSetNameForDC(dc), "config")
}

// DCDRLinkSecretName holds the credentials of the link user, the same in every data center.
func (r *RabbitMQ) DCDRLinkSecretName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "dcdr-link")
}

// CoordinatorLeaseNameForDC is the Lease the coordinator sidecars of data center dc elect a
// leader with.
func (r *RabbitMQ) CoordinatorLeaseNameForDC(dc string) string {
	return meta_util.NameWithSuffix(r.PetSetNameForDC(dc), "rmq-coordinator")
}

// MaxLagMessages returns spec.disasterRecovery.maxLagMessages or its default.
func (r *RabbitMQ) MaxLagMessages() int64 {
	if r.Spec.DisasterRecovery != nil && r.Spec.DisasterRecovery.MaxLagMessages != nil {
		return *r.Spec.DisasterRecovery.MaxLagMessages
	}
	return RabbitMQDCDRDefaultMaxLagMessages
}

// RabbitMQDCDRDataCenter is a Member data center of a DC-DR placement and the node ordinals
// (PlacementPolicy replicaIndices) it holds.
// +k8s:deepcopy-gen=false
// +k8s:openapi-gen=false
type RabbitMQDCDRDataCenter struct {
	Name     string
	Ordinals []int32
}

// RabbitMQDCDRNodesPerDC returns the cluster size of every Member data center for a database of
// `replicas` nodes: the number of its ordinals below replicas. Every ordinal below replicas must be
// placed exactly once and every data center needs at least one node.
func RabbitMQDCDRNodesPerDC(dataCenters []RabbitMQDCDRDataCenter, replicas int32) (map[string]int32, error) {
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
			return nil, fmt.Errorf("replica index %d (of spec.replicas=%d) is not placed in any data center", o, replicas)
		}
	}
	for dc, n := range sizes {
		if n == 0 {
			return nil, fmt.Errorf("data center %q has no node below spec.replicas=%d", dc, replicas)
		}
	}
	return sizes, nil
}

// RabbitMQDCDRMaxNameLength keeps <name>-<dc>-pods and the slice service names within the
// KubeSlice label limits.
const RabbitMQDCDRMaxNameLength = 30

// ValidateDCDR checks the DC-DR specific constraints of the spec. dataCenters are the Member data
// center names taken from the placement policy. It returns warnings for allowed but risky layouts.
func (r *RabbitMQ) ValidateDCDR(dataCenters []string, nodes map[string]int32) ([]string, error) {
	if !r.Spec.Distributed {
		return nil, nil
	}
	var warnings []string
	if r.DCDRPlacementPolicyName() == "" {
		return nil, fmt.Errorf("DC-DR requires spec.podTemplate.spec.podPlacementPolicy")
	}
	if len(r.Name) > RabbitMQDCDRMaxNameLength {
		return nil, fmt.Errorf("DC-DR needs a name of at most %d characters, got %d", RabbitMQDCDRMaxNameLength, len(r.Name))
	}
	if r.Spec.DisableSecurity {
		return nil, fmt.Errorf("DC-DR requires authentication: the fence revokes user permissions, so spec.disableSecurity must be false")
	}
	if r.Spec.TLS != nil || r.Spec.EnableSSL {
		return nil, fmt.Errorf("DC-DR does not support spec.tls yet")
	}
	if r.Spec.StorageType == StorageTypeEphemeral {
		return nil, fmt.Errorf("DC-DR needs durable storage: the saved permissions of a fenced data center live on the data volume")
	}
	seen := map[string]bool{}
	for _, dc := range dataCenters {
		if seen[dc] {
			return nil, fmt.Errorf("data center %q is listed twice in the placement", dc)
		}
		seen[dc] = true
	}
	if len(dataCenters) != 2 {
		return nil, fmt.Errorf("DC-DR needs exactly 2 member data centers, got %d", len(dataCenters))
	}
	for _, dc := range dataCenters {
		n := nodes[dc]
		if n > 0 && n%2 == 0 {
			// quorum queues and the metadata store need a majority inside the data center
			warnings = append(warnings, fmt.Sprintf("data center %q runs %d nodes: an even cluster tolerates no more failures than one node fewer", dc, n))
		}
		if n == 1 {
			warnings = append(warnings, fmt.Sprintf("data center %q runs a single node: it has no high availability inside the data center", dc))
		}
	}
	return warnings, nil
}
