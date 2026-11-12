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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	psapi "kubeops.dev/petset/apis/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
)

func ybTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{clientgoscheme.AddToScheme, catalog.AddToScheme, olddbapi.AddToScheme, psapi.AddToScheme} {
		if err := add(s); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func ybTestVersion(name string, deprecated bool) *catalog.YugabyteDBVersion {
	return &catalog.YugabyteDBVersion{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: catalog.YugabyteDBVersionSpec{
			Version:          name,
			DB:               catalog.YugabyteDBVersionDatabase{Image: "docker.io/yugabytedb/yugabyte:" + name},
			YSQLMajorVersion: 15,
			Deprecated:       deprecated,
			SecurityContext:  catalog.SecurityContext{RunAsUser: ptr.To(int64(10001))},
		},
	}
}

func ybTestPolicy(name string, zone *psapi.ZoneSpreadConstraint) *psapi.PlacementPolicy {
	return &psapi.PlacementPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       psapi.PlacementPolicySpec{ZoneSpreadConstraint: zone},
	}
}

func ybTestWebhook(t *testing.T, objs ...client.Object) *YugabyteDBCustomWebhook {
	t.Helper()
	base := []client.Object{
		ybTestVersion("2025.2.8", false),
		ybTestVersion("2025.1.4", false),
		ybTestVersion("2024.2.0", true),
		ybTestPolicy("zones", &psapi.ZoneSpreadConstraint{MaxSkew: 1, WhenUnsatisfiable: core.DoNotSchedule}),
		ybTestPolicy("weak", &psapi.ZoneSpreadConstraint{MaxSkew: 1, WhenUnsatisfiable: core.ScheduleAnyway}),
		ybTestPolicy("nozone", nil),
	}
	// objects passed by the caller replace the base objects of the same kind and name
	over := map[string]bool{}
	for _, o := range objs {
		over[fmt.Sprintf("%T/%s", o, o.GetName())] = true
	}
	all := append([]client.Object{}, objs...)
	for _, o := range base {
		if !over[fmt.Sprintf("%T/%s", o, o.GetName())] {
			all = append(all, o)
		}
	}
	kc := fake.NewClientBuilder().WithScheme(ybTestScheme(t)).WithObjects(all...).Build()
	return &YugabyteDBCustomWebhook{DefaultClient: kc}
}

// ybTestDB returns a defaulted, valid YugabyteDB.
func ybTestDB(t *testing.T, w *YugabyteDBCustomWebhook) *olddbapi.YugabyteDB {
	t.Helper()
	db := &olddbapi.YugabyteDB{
		ObjectMeta: metav1.ObjectMeta{Name: "yb", Namespace: "demo"},
		Spec:       olddbapi.YugabyteDBSpec{Version: "2025.2.8"},
	}
	if err := w.Default(context.TODO(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func ybCtx(user string) context.Context {
	return admission.NewContextWithRequest(context.TODO(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{UserInfo: authenticationv1.UserInfo{Username: user}},
	})
}

func setProvisioned(db *olddbapi.YugabyteDB) {
	db.Status.Conditions = []kmapi.Condition{{Type: kubedb.DatabaseProvisioned, Status: metav1.ConditionTrue}}
}

func checkErr(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected an error containing %q, got none", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected an error containing %q, got: %v", want, err)
	}
}

func TestYugabyteDBDefault(t *testing.T) {
	w := ybTestWebhook(t)
	tests := []struct {
		name   string
		mutate func(db *olddbapi.YugabyteDB)
		want   string
		check  func(t *testing.T, db *olddbapi.YugabyteDB)
	}{
		{
			name:   "empty spec gets RF 3 and master replicas 3",
			mutate: func(db *olddbapi.YugabyteDB) {},
			check: func(t *testing.T, db *olddbapi.YugabyteDB) {
				if db.ReplicationFactor() != 3 || *db.Spec.Topology.Master.Replicas != 3 || *db.Spec.Topology.TServer.Replicas != 3 {
					t.Fatalf("unexpected defaults: %+v", db.Spec.Topology)
				}
				if db.Spec.DeletionPolicy != olddbapi.DeletionPolicyDelete {
					t.Fatalf("deletionPolicy = %q", db.Spec.DeletionPolicy)
				}
				if m := db.Spec.Topology.Master.PodTemplate.Spec.Containers; len(m) != 1 || m[0].Resources.Limits.Memory().IsZero() {
					t.Fatalf("master memory limit not defaulted")
				}
			},
		},
		{
			name:   "master replicas follow RF 5",
			mutate: func(db *olddbapi.YugabyteDB) { db.Spec.ReplicationFactor = ptr.To(int32(5)) },
			check: func(t *testing.T, db *olddbapi.YugabyteDB) {
				if *db.Spec.Topology.Master.Replicas != 5 || *db.Spec.Topology.TServer.Replicas != 5 {
					t.Fatalf("unexpected replicas: %+v", db.Spec.Topology)
				}
			},
		},
		{
			name: "master replicas different from RF are rejected",
			mutate: func(db *olddbapi.YugabyteDB) {
				db.Spec.Topology = &olddbapi.YugabyteDBTopology{Master: &olddbapi.YugabyteDBNode{Replicas: ptr.To(int32(1))}}
			},
			want: "must equal spec.replicationFactor",
		},
		{
			name: "master replicas equal to RF are accepted",
			mutate: func(db *olddbapi.YugabyteDB) {
				db.Spec.Topology = &olddbapi.YugabyteDBTopology{Master: &olddbapi.YugabyteDBNode{Replicas: ptr.To(int32(3))}}
			},
		},
		{
			name:   "halted sets deletionPolicy Halt",
			mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Halted = true },
			check: func(t *testing.T, db *olddbapi.YugabyteDB) {
				if db.Spec.DeletionPolicy != olddbapi.DeletionPolicyHalt {
					t.Fatalf("deletionPolicy = %q", db.Spec.DeletionPolicy)
				}
			},
		},
		{
			name: "halted with DoNotTerminate is rejected",
			mutate: func(db *olddbapi.YugabyteDB) {
				db.Spec.Halted = true
				db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
			},
			want: "DoNotTerminate",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &olddbapi.YugabyteDB{
				ObjectMeta: metav1.ObjectMeta{Name: "yb", Namespace: "demo"},
				Spec:       olddbapi.YugabyteDBSpec{Version: "2025.2.8"},
			}
			tc.mutate(db)
			err := w.Default(context.TODO(), db)
			checkErr(t, err, tc.want)
			if err == nil && tc.check != nil {
				tc.check(t, db)
			}
		})
	}
}

func TestYugabyteDBValidateCreate(t *testing.T) {
	w := ybTestWebhook(t, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "good-config", Namespace: "demo"},
		Data:       map[string][]byte{"tserver.flags": []byte("--ysql_max_connections=300\n# comment\n\nmax_clock_skew_usec=1000")},
	}, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "bad-config", Namespace: "demo"},
		Data:       map[string][]byte{"master.flags": []byte("--master_addresses=foo:7100")},
	}, &core.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "badkey-config", Namespace: "demo"},
		Data:       map[string][]byte{"postgresql.conf": []byte("x = 1")},
	})
	setMem := func(db *olddbapi.YugabyteDB, role olddbapi.YugabyteDBRole, mem string) {
		c := &db.Node(role).PodTemplate.Spec.Containers[0]
		if mem == "" {
			delete(c.Resources.Limits, core.ResourceMemory)
		} else {
			c.Resources.Limits[core.ResourceMemory] = resource.MustParse(mem)
		}
	}
	tests := []struct {
		name         string
		mutate       func(db *olddbapi.YugabyteDB)
		want         string
		wantWarnings int
	}{
		{name: "valid defaults"},
		{name: "RF 1 with 1 tserver", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.ReplicationFactor = ptr.To(int32(1))
			db.Spec.Topology.Master.Replicas = ptr.To(int32(1))
			db.Spec.Topology.TServer.Replicas = ptr.To(int32(1))
		}},
		{name: "RF 7", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.ReplicationFactor = ptr.To(int32(7))
			db.Spec.Topology.Master.Replicas = ptr.To(int32(7))
			db.Spec.Topology.TServer.Replicas = ptr.To(int32(7))
		}},
		{name: "RF 2", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.ReplicationFactor = ptr.To(int32(2)) }, want: "replicationFactor"},
		{name: "RF 0", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.ReplicationFactor = ptr.To(int32(0)) }, want: "replicationFactor"},
		{name: "RF 9", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.ReplicationFactor = ptr.To(int32(9)) }, want: "replicationFactor"},
		{name: "master replicas != RF", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.Master.Replicas = ptr.To(int32(1)) }, want: "master replicas must equal"},
		{name: "tserver replicas below RF", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.Replicas = ptr.To(int32(2)) }, want: "at least replicationFactor"},
		{name: "tserver replicas above RF", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.Replicas = ptr.To(int32(6)) }},
		{name: "missing version", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Version = "" }, want: "spec.version is missing"},
		{name: "unknown version", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Version = "1.0.0" }, want: "not available in the catalog"},
		{name: "deprecated version", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Version = "2024.2.0" }, want: "deprecated"},
		{name: "master memory limit missing", mutate: func(db *olddbapi.YugabyteDB) { setMem(db, olddbapi.YugabyteDBRoleMaster, "") }, want: "limits.memory must be set"},
		{name: "tserver memory limit missing", mutate: func(db *olddbapi.YugabyteDB) { setMem(db, olddbapi.YugabyteDBRoleTServer, "") }, want: "limits.memory must be set"},
		{name: "invalid storageType", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.StorageType = "Fast" }, want: "storageType"},
		{name: "ephemeral storage", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.StorageType = olddbapi.StorageTypeEphemeral
			db.Spec.Topology.TServer.Storage = nil
		}},
		{name: "ephemeral storage with Halt", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.StorageType = olddbapi.StorageTypeEphemeral
			db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
		}, want: "Ephemeral"},
		{name: "durable without storage", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.Storage = nil }, want: "storage is required"},
		{name: "dataVolumes 0", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.DataVolumes = ptr.To(int32(0)) }, want: "dataVolumes"},
		{name: "dataVolumes 9", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.DataVolumes = ptr.To(int32(9)) }, want: "dataVolumes"},
		{name: "dataVolumes 4", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.DataVolumes = ptr.To(int32(4)) }},
		{name: "halted with DoNotTerminate", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Halted = true
			db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
		}, want: "DoNotTerminate"},
		{name: "halted with WipeOut", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Halted = true
			db.Spec.DeletionPolicy = olddbapi.DeletionPolicyWipeOut
		}, want: "needs deletionPolicy Halt"},
		{name: "halted with Halt", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Halted = true
			db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
		}},
		{name: "placement policy exists", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "zones"}
		}},
		{name: "placement policy missing", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "nope"}
		}, want: `PlacementPolicy "nope" not found`},
		{name: "placement policy with ScheduleAnyway warns", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "weak"}
		}, wantWarnings: 1},
		{name: "placement policy without zone constraint warns", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.Master.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "nozone"}
		}, wantWarnings: 1},
		{name: "weak placement policy with RF 1 does not warn", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.ReplicationFactor = ptr.To(int32(1))
			db.Spec.Topology.Master.Replicas = ptr.To(int32(1))
			db.Spec.Topology.TServer.Replicas = ptr.To(int32(1))
			db.Spec.Topology.TServer.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "weak"}
		}},
		{name: "inline flags valid", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{
				"tserver.flags": "--ysql_max_connections=300\nmax_clock_skew_usec=1000",
				"master.flags":  "--load_balancer_max_concurrent_adds=2",
				"ysql_pg.conf":  "log_min_duration_statement = 500",
				"ysql_hba.conf": "host all all 10.0.0.0/8 md5",
			}}
		}},
		{name: "inline unknown key", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"postgresql.conf": "a = 1"}}
		}, want: "unknown configuration key"},
		{name: "inline flag without equals", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"tserver.flags": "--ysql_max_connections 300"}}
		}, want: "not in the form"},
		{name: "inline flag bad name", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"tserver.flags": "--a-b=1"}}
		}, want: "not a valid flag name"},
		{name: "inline pg conf bad line", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"ysql_pg.conf": "justonetoken"}}
		}, want: "name = value"},
		{name: "inline hba bad line", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"ysql_hba.conf": "host all"}}
		}, want: "pg_hba.conf"},
		{name: "config secret valid", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{SecretName: "good-config"}
		}},
		{name: "config secret with owned flag", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{SecretName: "bad-config"}
		}, want: "owned by the operator"},
		{name: "config secret with unknown key", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{SecretName: "badkey-config"}
		}, want: "unknown configuration key"},
		{name: "config secret missing", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{SecretName: "missing"}
		}, want: "failed to get the config secret"},
		{name: "tls without issuer", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.TLS = &olddbapi.YugabyteDBTLSConfig{} }, want: "issuerRef is missing"},
		{name: "tls both channels off", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.TLS = &olddbapi.YugabyteDBTLSConfig{
				TLSConfig:    kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "ca", Kind: "Issuer"}},
				NodeToNode:   ptr.To(false),
				ClientToNode: ptr.To(false),
			}
		}, want: "at least one of nodeToNode and clientToNode"},
		{name: "tls valid", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.TLS = &olddbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "ca", Kind: "Issuer"}}}
		}},
		{name: "duplicate preferred zones", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Placement = &olddbapi.YugabyteDBPlacement{PreferredZones: []string{"a", "a"}}
		}, want: "unique"},
		{name: "externally managed auth secret without name", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.AuthSecret = &olddbapi.SecretReference{ExternallyManaged: true}
		}, want: "authSecret"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := ybTestDB(t, w)
			if tc.mutate != nil {
				tc.mutate(db)
			}
			warns, err := w.ValidateCreate(context.TODO(), db)
			checkErr(t, err, tc.want)
			if len(warns) != tc.wantWarnings {
				t.Fatalf("expected %d warnings, got %v", tc.wantWarnings, warns)
			}
		})
	}
}

func TestYugabyteDBOperatorOwnedFlags(t *testing.T) {
	w := ybTestWebhook(t)
	owned := append(olddbapi.YugabyteDBOperatorOwnedFlags(), "placement_cloud", "placement_region", "placement_zone", "placement_uuid")
	if len(owned) < 18 {
		t.Fatalf("expected at least 18 operator owned flags, got %d", len(owned))
	}
	for _, role := range []string{"master.flags", "tserver.flags"} {
		for _, f := range owned {
			for _, line := range []string{"--" + f + "=x", f + "=x"} {
				t.Run(role+"/"+line, func(t *testing.T) {
					db := ybTestDB(t, w)
					db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{role: "--ok_flag=1\n" + line}}
					_, err := w.ValidateCreate(context.TODO(), db)
					checkErr(t, err, "owned by the operator")
				})
			}
		}
	}
	// a flag that only starts like an owned one is fine
	db := ybTestDB(t, w)
	db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"tserver.flags": "--fs_data_dirs_extra=1\n--my_placement_flag=1"}}
	if _, err := w.ValidateCreate(context.TODO(), db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestYugabyteDBValidateUpdateImmutable(t *testing.T) {
	w := ybTestWebhook(t)
	tests := []struct {
		name   string
		mutate func(db *olddbapi.YugabyteDB)
		want   string
	}{
		{name: "no change"},
		{name: "replicationFactor", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.ReplicationFactor = ptr.To(int32(5))
			db.Spec.Topology.Master.Replicas = ptr.To(int32(5))
			db.Spec.Topology.TServer.Replicas = ptr.To(int32(5))
		}, want: "replicationFactor is immutable"},
		{name: "master dataVolumes", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.Master.DataVolumes = ptr.To(int32(2)) }, want: "dataVolumes is immutable"},
		{name: "tserver dataVolumes", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.DataVolumes = ptr.To(int32(2)) }, want: "dataVolumes is immutable"},
		{name: "storageType", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.StorageType = olddbapi.StorageTypeEphemeral }, want: "storageType is immutable"},
		{name: "ysql.enabled", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.YSQL.Enabled = false }, want: "ysql.enabled is immutable"},
		{name: "ycql.enabled", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.YCQL.Enabled = true }, want: "ycql.enabled is immutable"},
		{name: "disableSecurity", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.DisableSecurity = true }, want: "disableSecurity is immutable"},
		{name: "podPlacementPolicy may change", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "zones"}
		}},
		{name: "deletionPolicy may change", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.DeletionPolicy = olddbapi.DeletionPolicyWipeOut }},
	}
	// immutability applies before and after provisioning, for users and for the operator
	for _, provisioned := range []bool{false, true} {
		for _, user := range []string{"alice", "system:serviceaccount:kubedb:kubedb-provisioner"} {
			for _, tc := range tests {
				t.Run(tc.name+"/provisioned="+map[bool]string{false: "no", true: "yes"}[provisioned]+"/"+user, func(t *testing.T) {
					old := ybTestDB(t, w)
					if provisioned {
						setProvisioned(old)
					}
					db := old.DeepCopy()
					if tc.mutate != nil {
						tc.mutate(db)
					}
					_, err := w.ValidateUpdate(ybCtx(user), old, db)
					checkErr(t, err, tc.want)
				})
			}
		}
	}
}

func TestYugabyteDBValidateUpdateOpsOwned(t *testing.T) {
	w := ybTestWebhook(t)
	issuer := &core.TypedLocalObjectReference{Name: "ca", Kind: "Issuer"}
	tests := []struct {
		name   string
		mutate func(db *olddbapi.YugabyteDB)
		want   string
	}{
		{name: "version", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Version = "2025.1.4" }, want: "UpdateVersion"},
		{name: "tserver replicas", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.Topology.TServer.Replicas = ptr.To(int32(5)) }, want: "HorizontalScaling"},
		{name: "master resources", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.Master.PodTemplate.Spec.Containers[0].Resources.Limits[core.ResourceMemory] = resource.MustParse("8Gi")
		}, want: "VerticalScaling"},
		{name: "tserver resources", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.PodTemplate.Spec.Containers[0].Resources.Limits[core.ResourceCPU] = resource.MustParse("16")
		}, want: "VerticalScaling"},
		{name: "master storage size", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.Master.Storage.Resources.Requests[core.ResourceStorage] = resource.MustParse("20Gi")
		}, want: "VolumeExpansion"},
		{name: "tserver storage size", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Topology.TServer.Storage.Resources.Requests[core.ResourceStorage] = resource.MustParse("20Gi")
		}, want: "VolumeExpansion"},
		{name: "configuration added", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"tserver.flags": "--a=1"}}
		}, want: "Reconfigure"},
		{name: "tls added", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.TLS = &olddbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}}
		}, want: "ReconfigureTLS"},
		{name: "authSecret", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.AuthSecret = &olddbapi.SecretReference{TypedLocalObjectReference: appcat.TypedLocalObjectReference{Name: "other"}}
		}, want: "RotateAuth"},
		{name: "halted may change", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Halted = true
			db.Spec.DeletionPolicy = olddbapi.DeletionPolicyHalt
		}},
		{name: "deletionPolicy may change", mutate: func(db *olddbapi.YugabyteDB) { db.Spec.DeletionPolicy = olddbapi.DeletionPolicyWipeOut }},
		{name: "placement preferredZones may change", mutate: func(db *olddbapi.YugabyteDB) {
			db.Spec.Placement = &olddbapi.YugabyteDBPlacement{PreferredZones: []string{"a"}}
		}},
		{name: "no change"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			build := func(provisioned bool) (*olddbapi.YugabyteDB, *olddbapi.YugabyteDB) {
				old := ybTestDB(t, w)
				if provisioned {
					setProvisioned(old)
				}
				db := old.DeepCopy()
				if tc.mutate != nil {
					tc.mutate(db)
				}
				return old, db
			}
			// a user is rejected once the database is provisioned
			old, db := build(true)
			_, err := w.ValidateUpdate(ybCtx("alice"), old, db)
			checkErr(t, err, tc.want)

			// the provisioner and the ops-manager are allowed
			old, db = build(true)
			_, err = w.ValidateUpdate(ybCtx("system:serviceaccount:kubedb:kubedb-ops-manager"), old, db)
			checkErr(t, err, "")

			// before provisioning the user is still writing the spec
			old, db = build(false)
			_, err = w.ValidateUpdate(ybCtx("alice"), old, db)
			checkErr(t, err, "")

			// a request that carries no admission information is treated as a user
			old, db = build(true)
			_, err = w.ValidateUpdate(context.TODO(), old, db)
			checkErr(t, err, tc.want)
		})
	}
}

func TestYugabyteDBValidateUpdateVersionDeprecated(t *testing.T) {
	w := ybTestWebhook(t)
	// an existing database on a version that was deprecated later keeps working
	old := ybTestDB(t, w)
	old.Spec.Version = "2024.2.0"
	db := old.DeepCopy()
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyWipeOut
	if _, err := w.ValidateUpdate(ybCtx("alice"), old, db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// but it cannot be moved to a deprecated version, or an unknown one
	old = ybTestDB(t, w)
	for _, v := range []string{"2024.2.0", "9.9.9"} {
		db = old.DeepCopy()
		db.Spec.Version = v
		_, err := w.ValidateUpdate(ybCtx("system:serviceaccount:kubedb:kubedb-ops-manager"), old, db)
		if err == nil {
			t.Fatalf("expected an error for version %s", v)
		}
	}
}

func TestYugabyteDBValidateDelete(t *testing.T) {
	w := ybTestWebhook(t)
	db := ybTestDB(t, w)
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
	_, err := w.ValidateDelete(context.TODO(), db)
	checkErr(t, err, "DoNotTerminate")
	for _, p := range []olddbapi.DeletionPolicy{olddbapi.DeletionPolicyHalt, olddbapi.DeletionPolicyDelete, olddbapi.DeletionPolicyWipeOut} {
		db.Spec.DeletionPolicy = p
		_, err = w.ValidateDelete(context.TODO(), db)
		checkErr(t, err, "")
	}
}

func TestYugabyteDBDeletionInProgress(t *testing.T) {
	w := ybTestWebhook(t)
	db := ybTestDB(t, w)
	db.Spec.Version = "gone"
	now := metav1.Now()
	db.DeletionTimestamp = &now
	if _, err := w.ValidateCreate(context.TODO(), db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := w.ValidateUpdate(context.TODO(), db, db); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestYugabyteDBExamples decodes the example manifests strictly and runs the databases through the webhooks.
func TestYugabyteDBExamples(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "examples", "yugabytedb")
	versions, err := filepath.Glob(filepath.Join(root, "catalog", "*.yaml"))
	if err != nil || len(versions) != 2 {
		t.Fatalf("expected 2 catalog examples, got %v (%v)", versions, err)
	}
	var objs []client.Object
	for _, f := range versions {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		v := &catalog.YugabyteDBVersion{}
		if err := yaml.UnmarshalStrict(data, v); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if err := v.ValidateSpecs(); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if v.Spec.YSQLMajorVersion != 15 || *v.Spec.SecurityContext.RunAsUser != 10001 || v.Spec.Memory.MasterLimitPercent != 85 || v.Spec.Memory.TServerLimitPercent != 70 {
			t.Fatalf("%s: unexpected spec %+v", f, v.Spec)
		}
		if v.Spec.DefaultFlags.Master["enforce_tablet_replica_limits"] != "true" || v.Spec.DefaultFlags.TServer["graceful_shutdown"] != "true" {
			t.Fatalf("%s: unexpected default flags %+v", f, v.Spec.DefaultFlags)
		}
		wantReload := strings.HasPrefix(v.Name, "2025.2")
		if v.Spec.Features.ClientOnlyCertReload != wantReload {
			t.Fatalf("%s: clientOnlyCertReload = %v", f, v.Spec.Features.ClientOnlyCertReload)
		}
		objs = append(objs, v)
	}
	objs = append(objs, ybTestPolicy("default", &psapi.ZoneSpreadConstraint{MaxSkew: 1, WhenUnsatisfiable: core.DoNotSchedule}))
	w := ybTestWebhook(t, objs...)

	dbs, _ := filepath.Glob(filepath.Join(root, "db", "*.yaml"))
	if len(dbs) == 0 {
		t.Fatal("no database examples")
	}
	for _, f := range dbs {
		t.Run(filepath.Base(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			db := &olddbapi.YugabyteDB{}
			if err := yaml.UnmarshalStrict(data, db); err != nil {
				t.Fatal(err)
			}
			if err := w.Default(context.TODO(), db); err != nil {
				t.Fatal(err)
			}
			if _, err := w.ValidateCreate(context.TODO(), db); err != nil {
				t.Fatal(err)
			}
		})
	}
}
