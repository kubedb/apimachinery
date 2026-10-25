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
	"context"
	"testing"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/utils/ptr"
)

func TestYugabyteDBReadReplica(t *testing.T) {
	w := ybTestWebhook(t)
	withRR := func(db *olddbapi.YugabyteDB) {
		db.Spec.Topology.ReadReplica = &olddbapi.YugabyteDBReadReplica{ReplicationFactor: ptr.To(int32(2))}
		db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(2))
	}
	create := []struct {
		name   string
		mutate func(db *olddbapi.YugabyteDB)
		want   string
	}{
		{name: "read replica", mutate: withRR},
		{name: "defaults", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.ReadReplica = &olddbapi.YugabyteDBReadReplica{} }},
		{name: "zero replicas", mutate: func(db *olddbapi.YugabyteDB) {
			withRR(db)
			db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(0))
		}},
		{name: "negative replicas", mutate: func(db *olddbapi.YugabyteDB) {
			withRR(db)
			db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(-1))
		}, want: "must not be negative"},
		{name: "replicas below RF", mutate: func(db *olddbapi.YugabyteDB) {
			withRR(db)
			db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(1))
		}, want: "at least the read replica replicationFactor (2)"},
		{name: "RF zero", mutate: func(db *olddbapi.YugabyteDB) {
			withRR(db)
			db.Spec.Topology.ReadReplica.ReplicationFactor = ptr.To(int32(0))
		}, want: "at least 1"},
		{name: "bad placement uuid", mutate: func(db *olddbapi.YugabyteDB) {
			withRR(db)
			db.Spec.Topology.ReadReplica.PlacementUUID = "Bad_UUID"
		}, want: "placementUUID"},
	}
	for _, tc := range create {
		t.Run("create/"+tc.name, func(t *testing.T) {
			db := &olddbapi.YugabyteDB{Spec: olddbapi.YugabyteDBSpec{Version: "2025.2.8", Topology: &olddbapi.YugabyteDBTopology{}}}
			db.Name, db.Namespace = "yb", "demo"
			tc.mutate(db)
			if err := w.Default(context.TODO(), db); err != nil {
				t.Fatal(err)
			}
			_, err := w.ValidateCreate(context.TODO(), db)
			checkErr(t, err, tc.want)
		})
	}

	t.Run("defaults", func(t *testing.T) {
		db := &olddbapi.YugabyteDB{Spec: olddbapi.YugabyteDBSpec{Version: "2025.2.8", Topology: &olddbapi.YugabyteDBTopology{ReadReplica: &olddbapi.YugabyteDBReadReplica{}}}}
		db.Name, db.Namespace = "yb", "demo"
		if err := w.Default(context.TODO(), db); err != nil {
			t.Fatal(err)
		}
		rr := db.Spec.Topology.ReadReplica
		if rr.PlacementUUID != "rr" || *rr.Replicas != 1 || db.ReadReplicaReplicationFactor() != 1 {
			t.Fatalf("unexpected defaults %+v", rr)
		}
		if db.PetSetName(olddbapi.YugabyteDBRoleReadReplica) != "yb-readreplica" || db.ContainerName(olddbapi.YugabyteDBRoleReadReplica) != "yb-tserver" {
			t.Fatal("unexpected names")
		}
		if got := db.PodFQDN(olddbapi.YugabyteDBRoleReadReplica, 1); got != "yb-readreplica-1.yb-pods.demo.svc" {
			t.Fatalf("unexpected fqdn %s", got)
		}
		if len(db.Roles()) != 3 {
			t.Fatalf("roles = %v", db.Roles())
		}
		if c := db.Node(olddbapi.YugabyteDBRoleReadReplica).PodTemplate.Spec.Containers; len(c) == 0 || c[0].Name != "yb-tserver" || c[0].Resources.Limits.Memory().IsZero() {
			t.Fatalf("read replica container not defaulted: %+v", c)
		}
	})

	update := []struct {
		name   string
		mutate func(db *olddbapi.YugabyteDB)
		user   string
		want   string
	}{
		{name: "placementUUID", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.ReadReplica.PlacementUUID = "other" }, user: "system:serviceaccount:kubedb:p", want: "placementUUID is immutable"},
		{name: "dataVolumes", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.ReadReplica.DataVolumes = ptr.To(int32(2)) }, user: "system:serviceaccount:kubedb:p", want: "dataVolumes is immutable"},
		{name: "replicas by user", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(4)) }, user: "alice", want: "HorizontalScaling"},
		{name: "replicas by operator", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.ReadReplica.Replicas = ptr.To(int32(4)) }, user: "system:serviceaccount:kubedb:p"},
	}
	for _, tc := range update {
		t.Run("update/"+tc.name, func(t *testing.T) {
			old := ybTestDB(t, w)
			withRR(old)
			if err := w.Default(context.TODO(), old); err != nil {
				t.Fatal(err)
			}
			setProvisioned(old)
			db := old.DeepCopy()
			tc.mutate(db)
			_, err := w.ValidateUpdate(ybCtx(tc.user), old, db)
			checkErr(t, err, tc.want)
		})
	}
}
