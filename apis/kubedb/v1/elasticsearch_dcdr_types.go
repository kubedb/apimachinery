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

import "strings"

// DCDRExternalMembersAnnotation, when "true", tells the operator not to run the per data center
// clusters itself: each Member data center already runs its own Elasticsearch cluster, and the
// hub only orchestrates CCR, promotion and switchover across them. The per-DC REST endpoint,
// remote-cluster proxy address and auth Secret are then supplied per data center by the
// dr.kubedb.com/endpoint., dr.kubedb.com/remote-proxy-address. and dr.kubedb.com/auth-secret.
// annotations the operator reads.
const DCDRExternalMembersAnnotation = "dr.kubedb.com/external-members"

// IsDCDRDistributed reports whether this Elasticsearch is a cross data center disaster recovery
// deployment: a distributed database explicitly opted in with the DC-DR annotation
// (DCDREnabledAnnotation). Without the annotation spec.distributed keeps its meaning of one
// cluster delivered as ManifestWorks.
func (e Elasticsearch) IsDCDRDistributed() bool {
	return e.Spec.Distributed && e.Annotations[DCDREnabledAnnotation] == "true"
}

// HasExternalMembers reports whether the per data center clusters of a DC-DR Elasticsearch are
// managed outside this operator (see DCDRExternalMembersAnnotation).
func (e Elasticsearch) HasExternalMembers() bool {
	return e.Annotations[DCDRExternalMembersAnnotation] == "true"
}

// PlacementPolicyName returns the PlacementPolicy that makes the database DC-DR and names its
// data centers and failover group. Each PetSet names its own in its pod template: the combined
// PetSet's spec.podTemplate.spec.podPlacementPolicy, or in topology mode each role's
// spec.topology.<role>.podTemplate.spec.podPlacementPolicy, all naming the same data centers,
// Arbiter and failover group; the master role's is returned. "" when none is set.
func (e Elasticsearch) PlacementPolicyName() string {
	ref := e.Spec.PodTemplate.Spec.PodPlacementPolicy
	if t := e.Spec.Topology; t != nil {
		ref = t.Master.PodTemplate.Spec.PodPlacementPolicy
	}
	if ref == nil {
		return ""
	}
	return ref.Name
}

// PerDCPlacementPolicyName names the single data center PlacementPolicy the operator writes on
// the hub for one of the database's PlacementPolicies: <policy>-<dc>, one per PetSet and data
// center.
func PerDCPlacementPolicyName(policy, dc string) string { return policy + "-" + dc }

// DCScopedName returns the per data center form of a name derived from the database's offshoot
// name: <db>[suffix] becomes <db>-<dc>[suffix]. It names every per-DC object of a DC-DR
// distributed Elasticsearch: PetSets (<db>-<dc>, <db>-<dc>-master), their pods, the governing and
// master-discovery Services (<db>-<dc>-pods, <db>-<dc>-master) and the cluster.name.
func (e Elasticsearch) DCScopedName(dc, name string) string {
	o := e.OffshootName()
	return o + "-" + dc + strings.TrimPrefix(name, o)
}

// DeliveryManifestWorkName names the ManifestWork that carries a DC-DR database's Secrets, RBAC,
// Services and exports to each data center (in that data center's hub namespace).
func (e Elasticsearch) DeliveryManifestWorkName() string {
	return e.Namespace + "-" + e.OffshootName() + "-dcdr"
}

// CoordinatorLeaseName names the Lease the es-coordinator replicas of one data center's cluster
// elect their single acting fence with.
func (e Elasticsearch) CoordinatorLeaseName(dc string) string {
	return e.DCScopedName(dc, e.OffshootName()) + "-es-coordinator"
}

// MarkerReaderName names the marker-reader Role and RoleBinding in the shared marker namespace.
// Namespace-qualified so same-named databases in different namespaces do not share one
// RoleBinding.
func (e Elasticsearch) MarkerReaderName() string {
	return e.Namespace + "-" + e.OffshootName() + "-marker-reader"
}
