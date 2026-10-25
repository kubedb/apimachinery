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
	"os"
	"path/filepath"
	"strings"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func ybOpsScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{clientgoscheme.AddToScheme, catalog.AddToScheme, dbapi.AddToScheme, opsapi.AddToScheme} {
		if err := add(s); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func ybOpsVersion(name string, ysql int32, deprecated bool, allow, deny []string) *catalog.YugabyteDBVersion {
	return &catalog.YugabyteDBVersion{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: catalog.YugabyteDBVersionSpec{
			Version:           name,
			DB:                catalog.YugabyteDBVersionDatabase{Image: "yugabytedb/yugabyte:" + name},
			YSQLMajorVersion:  ysql,
			Deprecated:        deprecated,
			UpdateConstraints: catalog.UpdateConstraints{Allowlist: allow, Denylist: deny},
		},
	}
}

func ybOpsDB(tls *dbapi.YugabyteDBTLSConfig, disableSecurity bool) *dbapi.YugabyteDB {
	storage := func() *core.PersistentVolumeClaimSpec {
		return &core.PersistentVolumeClaimSpec{Resources: core.VolumeResourceRequirements{
			Requests: core.ResourceList{core.ResourceStorage: resource.MustParse("10Gi")},
		}}
	}
	return &dbapi.YugabyteDB{
		ObjectMeta: metav1.ObjectMeta{Name: "yb", Namespace: "demo"},
		Spec: dbapi.YugabyteDBSpec{
			Version:           "2025.1.4",
			ReplicationFactor: ptr.To(int32(3)),
			Topology: &dbapi.YugabyteDBTopology{
				Master:  &dbapi.YugabyteDBNode{Replicas: ptr.To(int32(3)), Storage: storage()},
				TServer: &dbapi.YugabyteDBNode{Replicas: ptr.To(int32(4)), Storage: storage()},
			},
			TLS:             tls,
			DisableSecurity: disableSecurity,
		},
	}
}

func ybOpsWebhook(t *testing.T, objs ...client.Object) *YugabyteDBOpsRequestCustomWebhook {
	t.Helper()
	base := []client.Object{
		ybOpsVersion("2025.1.0", 15, false, nil, nil),
		ybOpsVersion("2025.1.4", 15, false, []string{">= 2025.1.0, < 2025.1.4"}, nil),
		ybOpsVersion("2025.2.8", 15, false, []string{">= 2025.1.4, < 2025.2.8"}, nil),
		ybOpsVersion("2025.2.9", 15, true, []string{">= 2025.1.4, < 2025.2.9"}, nil),
		ybOpsVersion("2026.1.0", 17, false, []string{">= 2025.1.4"}, nil),
		ybOpsVersion("2025.2.7", 15, false, []string{">= 2025.1.0"}, []string{"2025.1.4"}),
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "new-auth", Namespace: "demo"},
			Data:       map[string][]byte{"username": []byte("yugabyte"), "password": []byte("s3cret")},
		},
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "wrong-user", Namespace: "demo"},
			Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("s3cret")},
		},
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "no-password", Namespace: "demo"},
			Data:       map[string][]byte{"username": []byte("yugabyte")},
		},
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "good-config", Namespace: "demo"},
			Data:       map[string][]byte{"tserver.flags": []byte("--ysql_max_connections=300")},
		},
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "owned-config", Namespace: "demo"},
			Data:       map[string][]byte{"tserver.flags": []byte("--placement_zone=a")},
		},
		&core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "badkey-config", Namespace: "demo"},
			Data:       map[string][]byte{"other.conf": []byte("a=b")},
		},
	}
	kc := fake.NewClientBuilder().WithScheme(ybOpsScheme(t)).WithStatusSubresource(&dbapi.YugabyteDB{}).WithObjects(append(base, objs...)...).Build()
	return &YugabyteDBOpsRequestCustomWebhook{DefaultClient: kc}
}

func ybOpsReq(typ opsapi.YugabyteDBOpsRequestType) *opsapi.YugabyteDBOpsRequest {
	return &opsapi.YugabyteDBOpsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "op", Namespace: "demo"},
		Spec: opsapi.YugabyteDBOpsRequestSpec{
			DatabaseRef: core.LocalObjectReference{Name: "yb"},
			Type:        typ,
		},
	}
}

func podResources(mem string) *opsapi.PodResources {
	r := core.ResourceRequirements{Limits: core.ResourceList{core.ResourceCPU: resource.MustParse("2")}}
	if mem != "" {
		r.Limits[core.ResourceMemory] = resource.MustParse(mem)
	}
	return &opsapi.PodResources{Resources: r}
}

func expectErr(t *testing.T, err error, want string) {
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

type ybOpsCase struct {
	name  string
	db    *dbapi.YugabyteDB // nil means the default database
	req   func() *opsapi.YugabyteDBOpsRequest
	phase opsapi.OpsRequestPhase
	want  string
}

func runYBOpsCases(t *testing.T, cases []ybOpsCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.db
			if db == nil {
				db = ybOpsDB(nil, false)
			}
			w := ybOpsWebhook(t, db)
			req := tc.req()
			req.Status.Phase = tc.phase
			_, err := w.ValidateCreate(context.TODO(), req)
			expectErr(t, err, tc.want)
		})
	}
}

func TestYugabyteDBOpsCommon(t *testing.T) {
	runYBOpsCases(t, []ybOpsCase{
		{name: "restart", req: func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart)
			r.Spec.Restart = &opsapi.RestartSpec{}
			return r
		}},
		{name: "restart without spec", req: func() *opsapi.YugabyteDBOpsRequest { return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart) }},
		{name: "unknown type", req: func() *opsapi.YugabyteDBOpsRequest { return ybOpsReq("Scale") }, want: "is not supported"},
		{name: "missing database", req: func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart)
			r.Spec.DatabaseRef.Name = "nope"
			return r
		}, want: "invalid or not found"},
	})
}

func TestYugabyteDBOpsUpdateVersion(t *testing.T) {
	update := func(target string) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeUpdateVersion)
			r.Spec.UpdateVersion = &opsapi.YugabyteDBUpdateVersionSpec{TargetVersion: target}
			return r
		}
	}
	dbAt := func(v string) *dbapi.YugabyteDB {
		db := ybOpsDB(nil, false)
		db.Spec.Version = v
		return db
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "1.4 to 2.8 is in the allowlist", req: update("2025.2.8")},
		{name: "2.8 reached from 2025.1.0 is not", db: dbAt("2025.1.0"), req: update("2025.2.8"), want: "not allowed"},
		{name: "1.0 to 1.4", db: dbAt("2025.1.0"), req: update("2025.1.4")},
		{name: "downgrade is not allowed", db: dbAt("2025.2.8"), req: update("2025.1.4"), want: "not allowed"},
		{name: "same version", req: update("2025.1.4"), want: "is the current version"},
		{name: "unknown target", req: update("1.2.3"), want: "does not exist"},
		{name: "deprecated target", req: update("2025.2.9"), want: "deprecated"},
		{name: "different ysql major", req: update("2026.1.0"), want: "YSQL major"},
		{name: "denylist", db: dbAt("2025.1.4"), req: update("2025.2.7"), want: "not allowed"},
		{name: "no updateVersion", req: func() *opsapi.YugabyteDBOpsRequest { return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeUpdateVersion) }, want: "spec.updateVersion nil"},
		{name: "empty target", req: update(""), want: "targetVersion is required"},
		{name: "negative soak", req: func() *opsapi.YugabyteDBOpsRequest {
			r := update("2025.2.8")()
			r.Spec.UpdateVersion.SoakDuration = &metav1.Duration{Duration: -1}
			return r
		}, want: "soakDuration"},
		{name: "running request skips the current state checks", phase: opsapi.OpsRequestPhaseProgressing, db: dbAt("2025.2.8"), req: update("2025.2.8")},
	})
}

func TestYugabyteDBOpsHorizontalScaling(t *testing.T) {
	scale := func(n *int32) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeHorizontalScaling)
			r.Spec.HorizontalScaling = &opsapi.YugabyteDBHorizontalScalingSpec{TServer: n}
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "scale out", req: scale(ptr.To(int32(6)))},
		{name: "scale in to RF", req: scale(ptr.To(int32(3)))},
		{name: "below RF", req: scale(ptr.To(int32(2))), want: "at least the replication factor"},
		{name: "zero", req: scale(ptr.To(int32(0))), want: "at least the replication factor"},
		{name: "same as current", req: scale(ptr.To(int32(4))), want: "same as current"},
		{name: "same as current while running", phase: opsapi.OpsRequestPhaseProgressing, req: scale(ptr.To(int32(4)))},
		{name: "no tserver", req: scale(nil), want: "at least one of spec.horizontalScaling.tserver"},
		{name: "no spec", req: func() *opsapi.YugabyteDBOpsRequest {
			return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeHorizontalScaling)
		}, want: "spec.horizontalScaling nil"},
		{name: "RF 5 needs 5", db: func() *dbapi.YugabyteDB {
			db := ybOpsDB(nil, false)
			db.Spec.ReplicationFactor = ptr.To(int32(5))
			return db
		}(), req: scale(ptr.To(int32(3))), want: "at least the replication factor (5)"},
	})
}

func TestYugabyteDBOpsVerticalScaling(t *testing.T) {
	vs := func(m, ts *opsapi.PodResources) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVerticalScaling)
			r.Spec.VerticalScaling = &opsapi.YugabyteDBVerticalScalingSpec{Master: m, TServer: ts}
			return r
		}
	}
	over := podResources("4Gi")
	over.Resources.Requests = core.ResourceList{core.ResourceMemory: resource.MustParse("8Gi")}
	runYBOpsCases(t, []ybOpsCase{
		{name: "both roles", req: vs(podResources("2Gi"), podResources("8Gi"))},
		{name: "tserver only", req: vs(nil, podResources("8Gi"))},
		{name: "master only", req: vs(podResources("4Gi"), nil)},
		{name: "none", req: vs(nil, nil), want: "at least one of"},
		{name: "master without memory limit", req: vs(podResources(""), nil), want: "master.resources.limits.memory"},
		{name: "tserver without memory limit", req: vs(podResources("2Gi"), podResources("")), want: "tserver.resources.limits.memory"},
		{name: "request above limit", req: vs(nil, over), want: "must not exceed the limit"},
		{name: "no spec", req: func() *opsapi.YugabyteDBOpsRequest {
			return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVerticalScaling)
		}, want: "spec.verticalScaling nil"},
	})
}

func TestYugabyteDBOpsVolumeExpansion(t *testing.T) {
	ve := func(mode opsapi.VolumeExpansionMode, m, ts string) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVolumeExpansion)
			r.Spec.VolumeExpansion = &opsapi.YugabyteDBVolumeExpansionSpec{Mode: mode}
			if m != "" {
				q := resource.MustParse(m)
				r.Spec.VolumeExpansion.Master = &q
			}
			if ts != "" {
				q := resource.MustParse(ts)
				r.Spec.VolumeExpansion.TServer = &q
			}
			return r
		}
	}
	ephemeral := ybOpsDB(nil, false)
	ephemeral.Spec.Topology.TServer.StorageType = dbapi.StorageTypeEphemeral
	ephemeral.Spec.Topology.TServer.Storage = nil
	runYBOpsCases(t, []ybOpsCase{
		{name: "online both", req: ve(opsapi.VolumeExpansionModeOnline, "20Gi", "50Gi")},
		{name: "offline tserver", req: ve(opsapi.VolumeExpansionModeOffline, "", "50Gi")},
		{name: "no mode", req: ve("", "", "50Gi"), want: "mode must be"},
		{name: "bad mode", req: ve("Maybe", "", "50Gi"), want: "mode must be"},
		{name: "no size", req: ve(opsapi.VolumeExpansionModeOnline, "", ""), want: "at least one of"},
		{name: "same size", req: ve(opsapi.VolumeExpansionModeOnline, "", "10Gi"), want: "must be greater than current"},
		{name: "smaller", req: ve(opsapi.VolumeExpansionModeOnline, "5Gi", ""), want: "must be greater than current"},
		{name: "same size while running", phase: opsapi.OpsRequestPhaseProgressing, req: ve(opsapi.VolumeExpansionModeOnline, "", "10Gi")},
		{name: "ephemeral", db: ephemeral, req: ve(opsapi.VolumeExpansionModeOnline, "", "50Gi"), want: "Ephemeral"},
		{name: "no spec", req: func() *opsapi.YugabyteDBOpsRequest {
			return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeVolumeExpansion)
		}, want: "spec.volumeExpansion nil"},
	})
}

func TestYugabyteDBOpsReconfigure(t *testing.T) {
	rc := func(c *opsapi.ReconfigurationSpec) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigure)
			r.Spec.Configuration = c
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "applyConfig", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{
			"tserver.flags": "--ysql_max_connections=400", "ysql_hba.conf": "host all all 0.0.0.0/0 md5", "ysql_pg.conf": "work_mem = 8MB", "master.flags": "load_balancer_max_concurrent_adds=3",
		}})},
		{name: "configSecret", req: rc(&opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{Name: "good-config"}})},
		{name: "removeCustomConfig", req: rc(&opsapi.ReconfigurationSpec{RemoveCustomConfig: true})},
		{name: "restart true", req: rc(&opsapi.ReconfigurationSpec{RemoveCustomConfig: true, Restart: opsapi.ReconfigureRestartTrue})},
		{name: "restart bad", req: rc(&opsapi.ReconfigurationSpec{RemoveCustomConfig: true, Restart: "sometimes"}), want: "restart"},
		{name: "nothing", req: rc(&opsapi.ReconfigurationSpec{}), want: "at least one of"},
		{name: "no spec", req: rc(nil), want: "spec.configuration nil"},
		{name: "applyConfig unknown key", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"postgresql.conf": "a = 1"}}), want: "unknown configuration key"},
		{name: "applyConfig owned flag", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"master.flags": "--master_addresses=a:7100"}}), want: "owned by the operator"},
		{name: "applyConfig placement flag", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"tserver.flags": "placement_zone=z"}}), want: "owned by the operator"},
		{name: "applyConfig bad syntax", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"tserver.flags": "--flag"}}), want: "not in the form"},
		{name: "applyConfig bad hba", req: rc(&opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"ysql_hba.conf": "host"}}), want: "pg_hba.conf"},
		{name: "configSecret missing", req: rc(&opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{Name: "nope"}}), want: "not found"},
		{name: "configSecret empty name", req: rc(&opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{}}), want: "can not be empty"},
		{name: "configSecret owned flag", req: rc(&opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{Name: "owned-config"}}), want: "owned by the operator"},
		{name: "configSecret unknown key", req: rc(&opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{Name: "badkey-config"}}), want: "unknown configuration key"},
	})
}

func TestYugabyteDBOpsReconfigureTLS(t *testing.T) {
	issuer := &core.TypedLocalObjectReference{Name: "ca", Kind: "Issuer"}
	other := &core.TypedLocalObjectReference{Name: "other-ca", Kind: "Issuer"}
	withTLS := ybOpsDB(&dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}}, false)
	tls := func(s opsapi.YugabyteDBTLSSpec) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigureTLS)
			r.Spec.TLS = &s
			return r
		}
	}
	cfg := func(ref *core.TypedLocalObjectReference) dbapi.YugabyteDBTLSConfig {
		return dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: ref}}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "add", req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: cfg(issuer)})},
		{name: "add node to node only", req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}, ClientToNode: ptr.To(false)}})},
		{name: "add with both channels off", req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}, NodeToNode: ptr.To(false), ClientToNode: ptr.To(false)}}), want: "at least one of"},
		{name: "add without issuer", req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{Certificates: []kmapi.CertificateSpec{{Alias: "server"}}}}}), want: "issuerRef is required"},
		{name: "remove", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{Remove: true})},
		{name: "remove without tls", req: tls(opsapi.YugabyteDBTLSSpec{Remove: true}), want: "no TLS configured"},
		{name: "rotate", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{RotateCertificates: true})},
		{name: "rotate without tls", req: tls(opsapi.YugabyteDBTLSSpec{RotateCertificates: true}), want: "no TLS configured"},
		{name: "issuer change", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: cfg(other)})},
		{name: "same issuer", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: cfg(issuer)}), want: "same as the current issuer"},
		{name: "certificates update", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{YugabyteDBTLSConfig: dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer, Certificates: []kmapi.CertificateSpec{{Alias: "server"}}}}})},
		{name: "remove and rotate", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{Remove: true, RotateCertificates: true}), want: "more than 1 field"},
		{name: "remove and issuer", db: withTLS, req: tls(opsapi.YugabyteDBTLSSpec{Remove: true, YugabyteDBTLSConfig: cfg(other)}), want: "more than 1 field"},
		{name: "nothing", req: tls(opsapi.YugabyteDBTLSSpec{}), want: "no reconfiguration"},
		{name: "no spec", req: func() *opsapi.YugabyteDBOpsRequest {
			return ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigureTLS)
		}, want: "spec.tls nil"},
		{name: "running request skips the state checks", phase: opsapi.OpsRequestPhaseProgressing, req: tls(opsapi.YugabyteDBTLSSpec{Remove: true})},
	})
}

func TestYugabyteDBOpsRotateAuth(t *testing.T) {
	ra := func(name string) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRotateAuth)
			if name != "-" {
				r.Spec.Authentication = &opsapi.AuthSpec{}
				if name != "" {
					r.Spec.Authentication.SecretRef = &appcat.TypedLocalObjectReference{Kind: "Secret", Name: name}
				}
			}
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "generated password, no authentication", req: ra("-")},
		{name: "generated password", req: ra("")},
		{name: "new secret", req: ra("new-auth")},
		{name: "missing secret", req: ra("nope"), want: "not found"},
		{name: "wrong username", req: ra("wrong-user"), want: "username in the secret must be"},
		{name: "no password", req: ra("no-password"), want: "does not contain username or password"},
		{name: "current secret", req: ra("yb-auth"), want: "current auth secret"},
		{name: "security disabled", db: ybOpsDB(nil, true), req: ra(""), want: "disableSecurity"},
	})
}

func TestYugabyteDBOpsValidateUpdate(t *testing.T) {
	w := ybOpsWebhook(t, ybOpsDB(nil, false))
	old := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeHorizontalScaling)
	old.Spec.HorizontalScaling = &opsapi.YugabyteDBHorizontalScalingSpec{TServer: ptr.To(int32(6))}

	// the status can change
	upd := old.DeepCopy()
	upd.Status.Phase = opsapi.OpsRequestPhaseProgressing
	if _, err := w.ValidateUpdate(context.TODO(), old, upd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// the spec cannot
	upd = old.DeepCopy()
	upd.Spec.HorizontalScaling.TServer = ptr.To(int32(7))
	_, err := w.ValidateUpdate(context.TODO(), old, upd)
	expectErr(t, err, "spec")
	upd = old.DeepCopy()
	upd.Spec.Type = opsapi.YugabyteDBOpsRequestTypeRestart
	_, err = w.ValidateUpdate(context.TODO(), old, upd)
	expectErr(t, err, "spec")

	if _, err := w.ValidateDelete(context.TODO(), old); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestYugabyteDBOpsResumeOnCompletion(t *testing.T) {
	db := ybOpsDB(nil, false)
	db.Status.Conditions = []kmapi.Condition{{Type: "Paused", Status: metav1.ConditionTrue}}
	w := ybOpsWebhook(t, db)
	old := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart)
	old.Status.Phase = opsapi.OpsRequestPhaseProgressing
	upd := old.DeepCopy()
	upd.Status.Phase = opsapi.OpsRequestPhaseSuccessful
	if _, err := w.ValidateUpdate(context.TODO(), old, upd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := &dbapi.YugabyteDB{}
	if err := w.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(db), got); err != nil {
		t.Fatal(err)
	}
	for _, c := range got.Status.Conditions {
		if c.Type == "Paused" {
			t.Fatalf("the Paused condition was not removed: %+v", got.Status.Conditions)
		}
	}
}

// TestYugabyteDBOpsExamples decodes the example manifests strictly and validates them against a database
// that allows each of them.
func TestYugabyteDBOpsExamples(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "..", "..", "..", "examples", "yugabytedb", "ops", "*.yaml"))
	if len(files) < len(opsapi.YugabyteDBOpsRequestTypeNames()) {
		t.Fatalf("expected at least one ops example per type, got %d files", len(files))
	}
	issuer := &core.TypedLocalObjectReference{Name: "yb-ca-issuer", Kind: "Issuer"}
	types := map[opsapi.YugabyteDBOpsRequestType]bool{}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			req := &opsapi.YugabyteDBOpsRequest{}
			if err := yaml.UnmarshalStrict(data, req); err != nil {
				t.Fatal(err)
			}
			types[req.Spec.Type] = true
			db := ybOpsDB(nil, false)
			db.Name = req.Spec.DatabaseRef.Name
			if db.Name == "yb-rr" {
				db.Spec.Topology.ReadReplica = &dbapi.YugabyteDBReadReplica{YugabyteDBNode: dbapi.YugabyteDBNode{Replicas: ptr.To(int32(2)), Storage: db.Spec.Topology.TServer.Storage}, ReplicationFactor: ptr.To(int32(2))}
			}
			// tls remove and rotate need a database that has TLS
			if req.Spec.TLS != nil && (req.Spec.TLS.Remove || req.Spec.TLS.RotateCertificates) {
				db.Spec.TLS = &dbapi.YugabyteDBTLSConfig{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}}
			}
			// finalize and rollback need an upgrade that is pending finalization
			if req.Spec.Type == opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade || req.Spec.Type == opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade {
				db.Status.Conditions = []kmapi.Condition{{Type: kubedb.YugabyteDBConditionUpgradePendingFinalize, Status: metav1.ConditionTrue}}
			}
			w := ybOpsWebhook(
				t, db,
				&core.Secret{ObjectMeta: metav1.ObjectMeta{Name: "yb-custom-config", Namespace: "demo"}, Data: map[string][]byte{"tserver.flags": []byte("--ysql_max_connections=300")}},
				&core.Secret{ObjectMeta: metav1.ObjectMeta{Name: "yb-new-auth", Namespace: "demo"}, Data: map[string][]byte{"username": []byte("yugabyte"), "password": []byte("x")}},
			)
			if _, err := w.ValidateCreate(context.TODO(), req); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, n := range opsapi.YugabyteDBOpsRequestTypeNames() {
		if !types[opsapi.YugabyteDBOpsRequestType(n)] {
			t.Errorf("no example for type %s", n)
		}
	}
}
