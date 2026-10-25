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
	"testing"

	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
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
