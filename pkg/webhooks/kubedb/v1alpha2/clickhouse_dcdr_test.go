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
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	psapi "kubeops.dev/petset/apis/apps/v1"
)

func dcdrCH() (*olddbapi.ClickHouse, *psapi.PlacementPolicy) {
	db := &olddbapi.ClickHouse{ObjectMeta: metav1.ObjectMeta{Name: "ch", Namespace: "demo", Annotations: map[string]string{dbapi.DCDREnabledAnnotation: "true"}}}
	db.Spec.Distributed = true
	db.Spec.ClusterTopology = &olddbapi.ClusterTopology{
		Cluster: olddbapi.ClusterSpec{
			Name: "appscode-cluster", Shards: ptr.To[int32](2), Replicas: ptr.To[int32](4),
			PodTemplate: &ofst.PodTemplateSpec{Spec: ofst.PodSpec{PodPlacementPolicy: &core.LocalObjectReference{Name: "ch-dcdr"}}},
		},
		ClickHouseKeeper: &olddbapi.ClickHouseKeeper{Spec: &olddbapi.ClickHouseKeeperSpec{Replicas: ptr.To[int32](3)}},
	}
	pp := &psapi.PlacementPolicy{ObjectMeta: metav1.ObjectMeta{Name: "ch-dcdr"}}
	// the failoverGroupRef shape decodes with an empty trigger into the vendored petset types; it
	// must not be refused (lab 2026-11-15, MySQL/MariaDB)
	pp.Spec.ClusterSpreadConstraint = &psapi.ClusterSpreadConstraint{
		DistributionRules: []psapi.DistributionRule{
			{ClusterName: "spoke-1", ReplicaIndices: []int32{0, 1}, Role: psapi.DCRoleMember},
			{ClusterName: "spoke-2", ReplicaIndices: []int32{2, 3}, Role: psapi.DCRoleMember},
			{ClusterName: "hub", Role: psapi.DCRoleArbiter},
		},
		FailoverPolicy: &psapi.FailoverPolicy{Mode: psapi.FailoverModeTwoDC},
	}
	return db, pp
}

func TestValidateDCDRClickHouse(t *testing.T) {
	db, pp := dcdrCH()
	if w, err := validateDCDRClickHouse(db, pp); err != nil || len(w) != 0 {
		t.Fatalf("valid 2 shards x (2+2) refused: %v %v", w, err)
	}

	db, pp = dcdrCH()
	pp.Spec.ClusterSpreadConstraint.FailoverPolicy = nil
	if _, err := validateDCDRClickHouse(db, pp); err == nil {
		t.Error("a policy without failoverPolicy must be refused")
	}

	db, pp = dcdrCH()
	pp.Spec.ClusterSpreadConstraint.DistributionRules[1].ReplicaIndices = []int32{2}
	if _, err := validateDCDRClickHouse(db, pp); err == nil {
		t.Error("ordinal 3 of 4 unplaced must be refused")
	}

	db, pp = dcdrCH()
	db.Spec.ClusterTopology.Cluster.Replicas = ptr.To[int32](2)
	pp.Spec.ClusterSpreadConstraint.DistributionRules[0].ReplicaIndices = []int32{0}
	pp.Spec.ClusterSpreadConstraint.DistributionRules[1].ReplicaIndices = []int32{1}
	if w, err := validateDCDRClickHouse(db, pp); err != nil || len(w) == 0 {
		t.Errorf("1+1 per shard must be allowed with a warning: %v %v", w, err)
	}

	db, pp = dcdrCH()
	pp.Spec.ClusterSpreadConstraint.DistributionRules = append(pp.Spec.ClusterSpreadConstraint.DistributionRules,
		psapi.DistributionRule{ClusterName: "spoke-3", ReplicaIndices: []int32{4}, Role: psapi.DCRoleMember})
	if _, err := validateDCDRClickHouse(db, pp); err == nil {
		t.Error("three member data centers must be refused")
	}
}

func TestValidateDistributedClickHouseNeedsDCDR(t *testing.T) {
	db, _ := dcdrCH()
	delete(db.Annotations, dbapi.DCDREnabledAnnotation)
	w := ClickHouseCustomWebhook{}
	if _, err := w.validateDistributedClickHouse(db); err == nil {
		t.Error("a distributed ClickHouse without the DC-DR annotation must be refused")
	}
	db.Spec.Distributed = false
	if _, err := w.validateDistributedClickHouse(db); err != nil {
		t.Errorf("a non distributed ClickHouse is not checked: %v", err)
	}
}
