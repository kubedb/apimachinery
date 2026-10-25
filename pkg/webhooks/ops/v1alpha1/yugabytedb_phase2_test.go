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

package v1alpha1

import (
	"context"
	"strings"
	"testing"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func ybPendingFinalizeDB() *dbapi.YugabyteDB {
	db := ybOpsDB(nil, false)
	db.Status.Conditions = []kmapi.Condition{{Type: kubedb.YugabyteDBConditionUpgradePendingFinalize, Status: metav1.ConditionTrue}}
	return db
}

func ybUpgradeReq(name string, typ opsapi.YugabyteDBOpsRequestType, phase opsapi.OpsRequestPhase) *opsapi.YugabyteDBOpsRequest {
	r := ybOpsReq(typ)
	r.Name = name
	r.Status.Phase = phase
	switch typ {
	case opsapi.YugabyteDBOpsRequestTypeUpdateVersion:
		r.Spec.UpdateVersion = &opsapi.YugabyteDBUpdateVersionSpec{TargetVersion: "2025.2.8"}
	case opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade:
		r.Spec.FinalizeUpgrade = &opsapi.FinalizeUpgradeSpec{}
	case opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade:
		r.Spec.RollbackUpgrade = &opsapi.RollbackUpgradeSpec{}
	}
	return r
}

func TestYugabyteDBOpsFinalizeAndRollback(t *testing.T) {
	for _, typ := range []opsapi.YugabyteDBOpsRequestType{opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade} {
		newReq := func() *opsapi.YugabyteDBOpsRequest { return ybUpgradeReq("op", typ, "") }
		runYBOpsCases(t, []ybOpsCase{
			{name: string(typ) + " pending upgrade", db: ybPendingFinalizeDB(), req: newReq},
			{name: string(typ) + " without pending upgrade", req: newReq, want: "needs a pending upgrade"},
			{name: string(typ) + " while running without condition", phase: opsapi.OpsRequestPhaseProgressing, req: newReq},
		})
	}
	other := func(typ opsapi.YugabyteDBOpsRequestType, phase opsapi.OpsRequestPhase) *opsapi.YugabyteDBOpsRequest {
		return ybUpgradeReq("other", typ, phase)
	}
	cases := []struct {
		name  string
		db    *dbapi.YugabyteDB
		other *opsapi.YugabyteDBOpsRequest
		req   *opsapi.YugabyteDBOpsRequest
		want  string
	}{
		{
			name: "two update versions", db: ybOpsDB(nil, false), other: other(opsapi.YugabyteDBOpsRequestTypeUpdateVersion, opsapi.OpsRequestPhaseProgressing),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""), want: "only one upgrade related request",
		},
		{
			name: "pending update version", db: ybOpsDB(nil, false), other: other(opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""), want: "only one upgrade related request",
		},
		{
			name: "completed update version", db: ybOpsDB(nil, false), other: other(opsapi.YugabyteDBOpsRequestTypeUpdateVersion, opsapi.OpsRequestPhaseSuccessful),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""),
		},
		{
			name: "update version while finalize runs", db: ybPendingFinalizeDB(), other: other(opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, opsapi.OpsRequestPhaseProgressing),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""), want: "pending finalization",
		},
		{
			name: "update version while upgrade is pending", db: ybPendingFinalizeDB(),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, ""), want: "pending finalization",
		},
		{
			name: "finalize next to the update version", db: ybPendingFinalizeDB(), other: other(opsapi.YugabyteDBOpsRequestTypeUpdateVersion, opsapi.OpsRequestPhaseProgressing),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, ""),
		},
		{
			name: "finalize while rollback runs", db: ybPendingFinalizeDB(), other: other(opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade, opsapi.OpsRequestPhaseProgressing),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, ""), want: "only one upgrade related request",
		},
		{
			name: "rollback while finalize is pending", db: ybPendingFinalizeDB(), other: other(opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, ""),
			req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade, ""), want: "only one upgrade related request",
		},
		{name: "other database", db: ybPendingFinalizeDB(), other: func() *opsapi.YugabyteDBOpsRequest {
			r := other(opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade, opsapi.OpsRequestPhaseProgressing)
			r.Spec.DatabaseRef.Name = "another"
			return r
		}(), req: ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, "")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := ybOpsWebhook(t, tc.db)
			if tc.other != nil {
				if err := w.DefaultClient.Create(context.TODO(), tc.other); err != nil {
					t.Fatal(err)
				}
			}
			_, err := w.ValidateCreate(context.TODO(), tc.req)
			expectErr(t, err, tc.want)
		})
	}
}

func TestYugabyteDBOpsUpdateVersionFinalizeMode(t *testing.T) {
	mode := func(m opsapi.YugabyteDBFinalizeMode) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeUpdateVersion, "")
			r.Spec.UpdateVersion.Finalize = m
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "default", req: mode("")},
		{name: "auto", req: mode(opsapi.YugabyteDBFinalizeModeAuto)},
		{name: "manual", req: mode(opsapi.YugabyteDBFinalizeModeManual)},
		{name: "unknown", req: mode("Later"), want: "is not valid"},
	})
}

func ybReadReplicaDB() *dbapi.YugabyteDB {
	db := ybOpsDB(nil, false)
	db.Spec.Topology.ReadReplica = &dbapi.YugabyteDBReadReplica{
		YugabyteDBNode:    dbapi.YugabyteDBNode{Replicas: ptr.To(int32(2)), Storage: db.Spec.Topology.TServer.Storage.DeepCopy()},
		ReplicationFactor: ptr.To(int32(2)),
	}
	return db
}

func TestYugabyteDBOpsReadReplica(t *testing.T) {
	hs := func(n *int32) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeHorizontalScaling)
			r.Spec.HorizontalScaling = &opsapi.YugabyteDBHorizontalScalingSpec{ReadReplica: n}
			return r
		}
	}
	vs := func() *opsapi.YugabyteDBOpsRequest {
		r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVerticalScaling)
		r.Spec.VerticalScaling = &opsapi.YugabyteDBVerticalScalingSpec{ReadReplica: podResources("4Gi")}
		return r
	}
	ve := func() *opsapi.YugabyteDBOpsRequest {
		r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVolumeExpansion)
		q := resource.MustParse("20Gi")
		r.Spec.VolumeExpansion = &opsapi.YugabyteDBVolumeExpansionSpec{Mode: opsapi.VolumeExpansionModeOffline, ReadReplica: &q}
		return r
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "scale out", db: ybReadReplicaDB(), req: hs(ptr.To(int32(4)))},
		{name: "scale to zero", db: ybReadReplicaDB(), req: hs(ptr.To(int32(0)))},
		{name: "below RF", db: ybReadReplicaDB(), req: hs(ptr.To(int32(1))), want: "read replica replicationFactor (2)"},
		{name: "negative", db: ybReadReplicaDB(), req: hs(ptr.To(int32(-1))), want: "must not be negative"},
		{name: "same", db: ybReadReplicaDB(), req: hs(ptr.To(int32(2))), want: "same as current"},
		{name: "no read replica", req: hs(ptr.To(int32(2))), want: "no spec.topology.readReplica"},
		{name: "vertical", db: ybReadReplicaDB(), req: vs},
		{name: "vertical no read replica", req: vs, want: "no spec.topology.readReplica"},
		{name: "volume expansion", db: ybReadReplicaDB(), req: ve},
		{name: "volume expansion no read replica", req: ve, want: "no spec.topology.readReplica"},
	})
}

func TestYugabyteDBOpsConnectionManager(t *testing.T) {
	cm := func(enabled bool, max *int32) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigure)
			r.Spec.ConnectionManager = &dbapi.YugabyteDBConnectionManager{Enabled: enabled, MaxClientConnections: max}
			return r
		}
	}
	withCM := ybOpsDB(nil, false)
	withCM.Spec.YSQL = &dbapi.YugabyteDBYSQLSpec{Enabled: true, ConnectionManager: &dbapi.YugabyteDBConnectionManager{Enabled: true}}
	noYSQL := ybOpsDB(nil, false)
	noYSQL.Spec.YSQL = &dbapi.YugabyteDBYSQLSpec{Enabled: false}
	runYBOpsCases(t, []ybOpsCase{
		{name: "enable", req: cm(true, nil)},
		{name: "enable with max", req: cm(true, ptr.To(int32(500)))},
		{name: "disable when off", req: cm(false, nil), want: "same as the current"},
		{name: "enable when on", db: withCM, req: cm(true, nil), want: "same as the current"},
		{name: "change max", db: withCM, req: cm(true, ptr.To(int32(500)))},
		{name: "disable", db: withCM, req: cm(false, nil)},
		{name: "zero max", req: cm(true, ptr.To(int32(0))), want: "at least 1"},
		{name: "ysql disabled", db: noYSQL, req: cm(true, nil), want: "YSQL is not enabled"},
		{name: "wrong type", req: func() *opsapi.YugabyteDBOpsRequest {
			r := cm(true, nil)()
			r.Spec.Type = opsapi.YugabyteDBOpsRequestTypeRestart
			return r
		}, want: "only valid in a Reconfigure request"},
		{name: "with configuration", req: func() *opsapi.YugabyteDBOpsRequest {
			r := cm(true, nil)()
			r.Spec.Configuration = &opsapi.ReconfigurationSpec{RemoveCustomConfig: true}
			return r
		}},
	})
}

func TestYugabyteDBOpsWebserverTLS(t *testing.T) {
	issuer := &core.TypedLocalObjectReference{Name: "yb-ca-issuer", Kind: "Issuer"}
	ws := func(v bool) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigureTLS)
			r.Spec.TLS = &opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: dbapi.YugabyteDBTLSConfig{Webserver: ptr.To(v)}}
			return r
		}
	}
	withTLS := func(ws bool) *dbapi.YugabyteDB {
		return ybOpsDB(&dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}, Webserver: ptr.To(ws)}, false)
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "enable", db: withTLS(false), req: ws(true)},
		{name: "disable", db: withTLS(true), req: ws(false)},
		{name: "already on", db: withTLS(true), req: ws(true), want: "same as the current"},
		{name: "no TLS", req: ws(true), want: "issuerRef is required"},
		{name: "with remove", db: withTLS(true), req: func() *opsapi.YugabyteDBOpsRequest {
			r := ws(true)()
			r.Spec.TLS.Remove = true
			return r
		}, want: "one is allowed to run one operation"},
	})
}

func TestYugabyteDBOpsStorageMigration(t *testing.T) {
	sc := func(name string, mode storagev1.VolumeBindingMode) *storagev1.StorageClass {
		return &storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: name}, VolumeBindingMode: &mode}
	}
	mig := func(class string, timeout bool) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeStorageMigration)
			r.Spec.Migration = &opsapi.StorageMigrationSpec{StorageClassName: ptr.To(class)}
			if timeout {
				r.Spec.Timeout = &metav1.Duration{Duration: 3600e9}
			}
			return r
		}
	}
	withClass := func(class string) *dbapi.YugabyteDB {
		db := ybOpsDB(nil, false)
		for _, n := range []*dbapi.YugabyteDBNode{db.Spec.Topology.Master, db.Spec.Topology.TServer} {
			n.Storage.StorageClassName = ptr.To(class)
		}
		return db
	}
	ephemeral := ybOpsDB(nil, false)
	ephemeral.Spec.Topology.TServer.StorageType = dbapi.StorageTypeEphemeral
	cases := []struct {
		name string
		db   *dbapi.YugabyteDB
		req  func() *opsapi.YugabyteDBOpsRequest
		want string
	}{
		{name: "migrate", db: withClass("old"), req: mig("new", true)},
		{name: "no timeout", db: withClass("old"), req: mig("new", false), want: "spec.timeout is required"},
		{name: "unknown class", db: withClass("old"), req: mig("nope", true), want: "not found"},
		{name: "same class", db: withClass("new"), req: mig("new", true), want: "already uses"},
		{name: "no spec", db: withClass("old"), req: func() *opsapi.YugabyteDBOpsRequest {
			return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeStorageMigration)
		}, want: "spec.migration nil"},
		{name: "empty class", db: withClass("old"), req: mig("", true), want: "storageClassName is required"},
		{name: "ephemeral", db: ephemeral, req: mig("new", true), want: "Ephemeral"},
		{name: "bad reclaim policy", db: withClass("old"), req: func() *opsapi.YugabyteDBOpsRequest {
			r := mig("new", true)()
			r.Spec.Migration.OldPVReclaimPolicy = "Recycle"
			return r
		}, want: "oldPVReclaimPolicy"},
		{name: "binding mode", db: withClass("old"), req: mig("immediate", true), want: "WaitForFirstConsumer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := ybOpsWebhook(t, tc.db, sc("new", storagev1.VolumeBindingWaitForFirstConsumer), sc("old", storagev1.VolumeBindingWaitForFirstConsumer), sc("immediate", storagev1.VolumeBindingImmediate))
			_, err := w.ValidateCreate(context.TODO(), tc.req())
			expectErr(t, err, tc.want)
		})
	}

	t.Run("warns with too few tservers", func(t *testing.T) {
		db := withClass("old")
		db.Spec.Topology.TServer.Replicas = ptr.To(int32(3))
		w := ybOpsWebhook(t, db, sc("new", storagev1.VolumeBindingWaitForFirstConsumer), sc("old", storagev1.VolumeBindingWaitForFirstConsumer))
		warns, err := w.ValidateCreate(context.TODO(), mig("new", true)())
		if err != nil || len(warns) != 1 {
			t.Fatalf("want one warning, got %v, %v", warns, err)
		}
	})
}

func TestYugabyteDBOpsVerticalScalingInPlace(t *testing.T) {
	db := ybOpsDB(nil, false)
	db.Spec.Topology.TServer.PodTemplate.Spec.Containers = []core.Container{{Name: "yb-tserver", Resources: core.ResourceRequirements{
		Limits:   core.ResourceList{core.ResourceCPU: resource.MustParse("2"), core.ResourceMemory: resource.MustParse("4Gi")},
		Requests: core.ResourceList{core.ResourceCPU: resource.MustParse("1"), core.ResourceMemory: resource.MustParse("4Gi")},
	}}}
	vs := func(mode opsapi.VerticalScalingMode, cpu, mem, reqMem string) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVerticalScaling)
			res := core.ResourceRequirements{Limits: core.ResourceList{core.ResourceCPU: resource.MustParse(cpu), core.ResourceMemory: resource.MustParse(mem)}}
			if reqMem != "" {
				res.Requests = core.ResourceList{core.ResourceMemory: resource.MustParse(reqMem)}
			}
			r.Spec.VerticalScaling = &opsapi.YugabyteDBVerticalScalingSpec{Mode: mode, TServer: &opsapi.PodResources{Resources: res}}
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "in place cpu only", db: db, req: vs(opsapi.VerticalScalingModeInPlace, "4", "4Gi", "4Gi")},
		{name: "in place memory limit", db: db, req: vs(opsapi.VerticalScalingModeInPlace, "4", "8Gi", ""), want: "only the CPU can be resized in place"},
		{name: "in place memory request", db: db, req: vs(opsapi.VerticalScalingModeInPlace, "4", "4Gi", "2Gi"), want: "only the CPU can be resized in place"},
		{name: "restart mode memory", db: db, req: vs(opsapi.VerticalScalingModeRestart, "4", "8Gi", "")},
		{name: "default mode memory", db: db, req: vs("", "4", "8Gi", "")},
		{name: "in place memory while running", db: db, phase: opsapi.OpsRequestPhaseProgressing, req: vs(opsapi.VerticalScalingModeInPlace, "4", "8Gi", "")},
		{name: "unknown mode", db: db, req: vs("Hot", "4", "4Gi", ""), want: "is not valid"},
	})
}

func ybPITRDB() *dbapi.YugabyteDB {
	db := ybOpsDB(nil, false)
	db.Spec.PITR = []dbapi.YugabyteDBPITR{{Database: "app", Kind: dbapi.YugabyteDBPITRKindYSQL, IntervalMinutes: 60, RetentionMinutes: 120}}
	return db
}

func TestYugabyteDBOpsRewind(t *testing.T) {
	rewind := func(mut func(r *opsapi.YugabyteDBRewindSpec)) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRewind)
			r.Spec.Rewind = &opsapi.YugabyteDBRewindSpec{Database: "app", Minus: &metav1.Duration{Duration: 15 * time.Minute}}
			if mut != nil {
				mut(r.Spec.Rewind)
			}
			return r
		}
	}
	at := func(d time.Duration) *metav1.Time { return &metav1.Time{Time: time.Now().Add(d)} }
	runYBOpsCases(t, []ybOpsCase{
		{name: "minus", db: ybPITRDB(), req: rewind(nil)},
		{name: "restore time", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus, r.RestoreTime = nil, at(-30*time.Minute) })},
		{name: "both", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.RestoreTime = at(-time.Minute) }), want: "exactly one of"},
		{name: "neither", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus = nil }), want: "exactly one of"},
		{name: "minus too long", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus = &metav1.Duration{Duration: 3 * time.Hour} }), want: "longer than the retention"},
		{name: "minus negative", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus = &metav1.Duration{Duration: -time.Minute} }), want: "must be positive"},
		{name: "future", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus, r.RestoreTime = nil, at(time.Hour) }), want: "in the future"},
		{name: "too old", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus, r.RestoreTime = nil, at(-3*time.Hour) }), want: "older than the retention"},
		{name: "too old while running", db: ybPITRDB(), phase: opsapi.OpsRequestPhaseProgressing, req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Minus, r.RestoreTime = nil, at(-3*time.Hour) })},
		{name: "no schedule", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Database = "other" }), want: "has no schedule"},
		{name: "no database", db: ybPITRDB(), req: rewind(func(r *opsapi.YugabyteDBRewindSpec) { r.Database = "" }), want: "database is required"},
		{name: "no spec", db: ybPITRDB(), req: func() *opsapi.YugabyteDBOpsRequest { return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRewind) }, want: "spec.rewind nil"},
	})
	t.Run("warns about the limits", func(t *testing.T) {
		w := ybOpsWebhook(t, ybPITRDB())
		warns, err := w.ValidateCreate(context.TODO(), rewind(nil)())
		if err != nil || len(warns) != 1 || !strings.Contains(warns[0], "upgrade_ysql") {
			t.Fatalf("unexpected %v %v", warns, err)
		}
	})
}

func TestYugabyteDBOpsClone(t *testing.T) {
	clone := func(mut func(c *opsapi.YugabyteDBCloneSpec)) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeClone)
			r.Spec.Clone = &opsapi.YugabyteDBCloneSpec{SourceDatabase: "app", TargetDatabase: "app_copy"}
			if mut != nil {
				mut(r.Spec.Clone)
			}
			return r
		}
	}
	at := func(d time.Duration) *metav1.Time { return &metav1.Time{Time: time.Now().Add(d)} }
	noYSQL := ybPITRDB()
	noYSQL.Spec.YSQL = &dbapi.YugabyteDBYSQLSpec{Enabled: false}
	old := ybPITRDB()
	old.Spec.Version = "2024.1.0"
	runYBOpsCases(t, []ybOpsCase{
		{name: "now", req: clone(nil)},
		{name: "as of", db: ybPITRDB(), req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.AsOf = at(-30 * time.Minute) })},
		{name: "as of without schedule", req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.AsOf = at(-30 * time.Minute) }), want: "needs a schedule"},
		{name: "as of too old", db: ybPITRDB(), req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.AsOf = at(-3 * time.Hour) }), want: "older than the retention"},
		{name: "as of future", db: ybPITRDB(), req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.AsOf = at(time.Hour) }), want: "in the future"},
		{name: "same name", req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.TargetDatabase = "app" }), want: "must differ"},
		{name: "no target", req: clone(func(c *opsapi.YugabyteDBCloneSpec) { c.TargetDatabase = "" }), want: "are required"},
		{name: "no ysql", db: noYSQL, req: clone(nil), want: "needs YSQL"},
		{name: "old version", db: old, req: clone(nil), want: "2024.2.0 or later"},
		{name: "no spec", req: func() *opsapi.YugabyteDBOpsRequest { return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeClone) }, want: "spec.clone nil"},
	})
}
