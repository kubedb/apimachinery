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

	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

func dcdrClickHouse() *ClickHouse {
	db := &ClickHouse{ObjectMeta: metav1.ObjectMeta{Name: "ch", Namespace: "demo", Annotations: map[string]string{v1.DCDREnabledAnnotation: "true"}}}
	db.Spec.Distributed = true
	db.Spec.ClusterTopology = &ClusterTopology{
		Cluster: ClusterSpec{
			Name:        "appscode-cluster",
			Shards:      ptr.To[int32](2),
			Replicas:    ptr.To[int32](4),
			PodTemplate: &ofst.PodTemplateSpec{Spec: ofst.PodSpec{PodPlacementPolicy: &core.LocalObjectReference{Name: "ch-dcdr"}}},
		},
		ClickHouseKeeper: &ClickHouseKeeper{Spec: &ClickHouseKeeperSpec{Replicas: ptr.To[int32](3)}},
	}
	return db
}

func TestClickHouseIsDCDRDistributed(t *testing.T) {
	db := dcdrClickHouse()
	if !db.IsDCDRDistributed() {
		t.Fatal("distributed + annotation must be DC-DR")
	}
	delete(db.Annotations, v1.DCDREnabledAnnotation)
	if db.IsDCDRDistributed() {
		t.Error("without the annotation it is not DC-DR")
	}
	db = dcdrClickHouse()
	db.Spec.Distributed = false
	if db.IsDCDRDistributed() {
		t.Error("without spec.distributed it is not DC-DR")
	}
}

func TestClickHouseValidateDCDR(t *testing.T) {
	two := []string{"spoke-1", "spoke-2"}
	cases := []struct {
		name    string
		mutate  func(*ClickHouse)
		dcs     []string
		wantErr bool
		warn    bool
	}{
		{"valid 2x2", func(*ClickHouse) {}, two, false, false},
		{"one replica per DC warns", func(c *ClickHouse) { c.Spec.ClusterTopology.Cluster.Replicas = ptr.To[int32](2) }, two, false, true},
		{"one replica total", func(c *ClickHouse) { c.Spec.ClusterTopology.Cluster.Replicas = ptr.To[int32](1) }, two, true, false},
		{"no cluster topology", func(c *ClickHouse) { c.Spec.ClusterTopology = nil }, two, true, false},
		{"no placement policy", func(c *ClickHouse) { c.Spec.ClusterTopology.Cluster.PodTemplate = nil }, two, true, false},
		{"security disabled", func(c *ClickHouse) { c.Spec.DisableSecurity = true }, two, true, false},
		{"tls", func(c *ClickHouse) { c.Spec.TLS = &ClickHouseTLSConfig{} }, two, true, false},
		{"ephemeral cluster", func(c *ClickHouse) { c.Spec.ClusterTopology.Cluster.StorageType = StorageTypeEphemeral }, two, true, false},
		{"ephemeral keeper", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper.Spec.StorageType = StorageTypeEphemeral }, two, true, false},
		{"external keeper", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper.ExternallyManaged = true }, two, true, false},
		{"no keeper", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper = nil }, two, true, false},
		{"keeper 1", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper.Spec.Replicas = ptr.To[int32](1) }, two, true, false},
		{"keeper 4", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper.Spec.Replicas = ptr.To[int32](4) }, two, true, false},
		{"keeper 5", func(c *ClickHouse) { c.Spec.ClusterTopology.ClickHouseKeeper.Spec.Replicas = ptr.To[int32](5) }, two, false, false},
		{"one DC", func(*ClickHouse) {}, []string{"spoke-1"}, true, false},
		{"three DCs", func(*ClickHouse) {}, []string{"spoke-1", "spoke-2", "spoke-3"}, true, false},
		{"duplicate DC", func(*ClickHouse) {}, []string{"spoke-1", "spoke-1"}, true, false},
		{"not DC-DR is not checked", func(c *ClickHouse) { c.Spec.Distributed = false; c.Spec.ClusterTopology = nil }, nil, false, false},
	}
	for _, tc := range cases {
		db := dcdrClickHouse()
		tc.mutate(db)
		w, err := db.ValidateDCDR(tc.dcs)
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: err=%v, wantErr=%v", tc.name, err, tc.wantErr)
		}
		if (len(w) > 0) != tc.warn {
			t.Errorf("%s: warnings=%v, want warning=%v", tc.name, w, tc.warn)
		}
	}
}

func TestClickHouseDCDRReplicasPerDC(t *testing.T) {
	dcs := []ClickHouseDCDRDataCenter{{Name: "spoke-1", Ordinals: []int32{0, 1}}, {Name: "spoke-2", Ordinals: []int32{2, 3}}}
	got, err := ClickHouseDCDRReplicasPerDC(dcs, 4)
	if err != nil || got["spoke-1"] != 2 || got["spoke-2"] != 2 {
		t.Fatalf("2+2: got %v, %v", got, err)
	}
	// ordinals listed for a later scale up do not count yet, but every DC needs one below replicas
	if got, err := ClickHouseDCDRReplicasPerDC([]ClickHouseDCDRDataCenter{{Name: "spoke-1", Ordinals: []int32{0, 2}}, {Name: "spoke-2", Ordinals: []int32{1, 3}}}, 2); err != nil || got["spoke-1"] != 1 || got["spoke-2"] != 1 {
		t.Errorf("1+1 with spare ordinals: got %v, %v", got, err)
	}
	if _, err := ClickHouseDCDRReplicasPerDC([]ClickHouseDCDRDataCenter{{Name: "spoke-1", Ordinals: []int32{0, 1}}, {Name: "spoke-2", Ordinals: []int32{1, 2}}}, 3); err == nil {
		t.Error("an ordinal in two DCs must be refused")
	}
	if _, err := ClickHouseDCDRReplicasPerDC([]ClickHouseDCDRDataCenter{{Name: "spoke-1", Ordinals: []int32{0}}, {Name: "spoke-2", Ordinals: []int32{2}}}, 3); err == nil {
		t.Error("an unplaced ordinal must be refused")
	}
	if _, err := ClickHouseDCDRReplicasPerDC([]ClickHouseDCDRDataCenter{{Name: "spoke-1", Ordinals: []int32{0, 1}}, {Name: "spoke-2", Ordinals: []int32{5}}}, 2); err == nil {
		t.Error("a DC without a replica below replicas must be refused")
	}
}

func TestClickHouseDCDRNames(t *testing.T) {
	db := dcdrClickHouse()
	if got := db.KeeperPetSetNameForDC("spoke-1"); got != db.OffshootKeeperName()+"-spoke-1" {
		t.Errorf("keeper petset: %s", got)
	}
	if got := db.ShardPetSetNameForDC("appscode-cluster", 1, "spoke-2"); got != db.OffshootClusterPetSetName("appscode-cluster", 1)+"-spoke-2" {
		t.Errorf("shard petset: %s", got)
	}
	if got := DCDRReplicaName("spoke-2", 3); got != "spoke-2-3" {
		t.Errorf("replica name: %s", got)
	}
	if db.FenceMarginSeconds() != ClickHouseDCDRDefaultFenceMarginSeconds {
		t.Error("default fence margin")
	}
	db.Spec.DisasterRecovery = &ClickHouseDCDRSpec{FenceMarginSeconds: ptr.To[int32](30)}
	if db.FenceMarginSeconds() != 30 {
		t.Error("fence margin override")
	}
}
