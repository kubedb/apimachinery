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
	"strings"
	"testing"
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func natsVersion(name string, deprecated bool, allow ...string) *catalog.NATSVersion {
	return &catalog.NATSVersion{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: catalog.NATSVersionSpec{
			Version:           name,
			Deprecated:        deprecated,
			DB:                catalog.NATSVersionDatabase{Image: "nats:" + name},
			Reloader:          catalog.NATSVersionReloader{Image: "reloader"},
			Exporter:          catalog.NATSVersionExporter{Image: "exporter"},
			UpdateConstraints: catalog.UpdateConstraints{Allowlist: allow},
		},
	}
}

func natsOpsTestWebhook(t *testing.T, objs ...runtime.Object) *NATSOpsRequestCustomWebhook {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{catalog.AddToScheme, olddbapi.AddToScheme, core.AddToScheme, storagev1.AddToScheme, opsapi.AddToScheme} {
		if err := add(scheme); err != nil {
			t.Fatal(err)
		}
	}
	objs = append(
		objs,
		natsVersion("2.14.8", false, ">= 2.14.0, < 2.16.0"),
		natsVersion("2.15.1", false, ">= 2.14.0, < 2.16.0"),
		natsVersion("2.16.0", false, ">= 2.15.0, < 2.17.0"),
		natsVersion("2.15.9", true, ">= 2.14.0, < 2.16.0"),
		&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "fast"}},
		&storagev1.StorageClass{ObjectMeta: metav1.ObjectMeta{Name: "standard"}},
	)
	return &NATSOpsRequestCustomWebhook{DefaultClient: fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()}
}

func opsNATS(version string) *olddbapi.NATS {
	db := &olddbapi.NATS{
		ObjectMeta: metav1.ObjectMeta{Name: "nats", Namespace: "demo"},
		Spec:       olddbapi.NATSSpec{Version: version},
	}
	db.SetDefaults(nil)
	return db
}

func natsOps(typ opsapi.NATSOpsRequestType) *opsapi.NATSOpsRequest {
	return &opsapi.NATSOpsRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "op", Namespace: "demo"},
		Spec:       opsapi.NATSOpsRequestSpec{DatabaseRef: core.LocalObjectReference{Name: "nats"}, Type: typ},
	}
}

type natsOpsCase struct {
	name    string
	db      *olddbapi.NATS
	req     *opsapi.NATSOpsRequest
	wantErr string
}

func runNATSOpsCases(t *testing.T, cases []natsOpsCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := c.db
			if db == nil {
				db = opsNATS("2.15.1")
			}
			w := natsOpsTestWebhook(t, db)
			err := w.validateCreateOrUpdate(c.req)
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestNATSOpsRequestTypeAndDatabaseRef(t *testing.T) {
	runNATSOpsCases(t, []natsOpsCase{
		{"unknown type", nil, natsOps("Defragment"), "is not supported"},
		{"missing database", nil, func() *opsapi.NATSOpsRequest {
			r := natsOps(opsapi.NATSOpsRequestTypeRestart)
			r.Spec.DatabaseRef.Name = "missing"
			return r
		}(), "invalid or not found"},
		{"restart", nil, natsOps(opsapi.NATSOpsRequestTypeRestart), ""},
	})
	if got := len(opsapi.NATSOpsRequestTypeNames()); got != 9 {
		t.Fatalf("NATSOpsRequest must have the nine standard types, got %d", got)
	}
}

func TestNATSOpsUpdateVersion(t *testing.T) {
	uv := func(target string) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeUpdateVersion)
		r.Spec.UpdateVersion = &opsapi.NATSUpdateVersionSpec{TargetVersion: target}
		return r
	}
	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeUpdateVersion), "spec.updateVersion is nil"},
		{"empty target", nil, uv(""), "targetVersion can not be empty"},
		{"unknown target", nil, uv("9.9.9"), "not found in the catalog"},
		{"deprecated target", nil, uv("2.15.9"), "is deprecated"},
		{"one minor up is allowed", opsNATS("2.14.8"), uv("2.15.1"), ""},
		{"downgrade is refused", opsNATS("2.15.1"), uv("2.14.8"), "downgrade"},
		{"same version is refused", opsNATS("2.15.1"), uv("2.15.1"), "downgrade or no-op"},
		{"outside updateConstraints is refused", opsNATS("2.14.8"), uv("2.16.0"), "is not supported"},
	})
}

func TestNATSOpsHorizontalScaling(t *testing.T) {
	hs := func(n *int32) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeHorizontalScaling)
		r.Spec.HorizontalScaling = &opsapi.NATSHorizontalScalingSpec{Replicas: n}
		return r
	}
	single := opsNATS("2.15.1")
	single.Spec.Replicas = ptr.To(int32(1))
	noJS := opsNATS("2.15.1")
	noJS.Spec.JetStream.Enabled = ptr.To(false)
	noJS.Spec.Replicas = ptr.To(int32(3))
	ephemeral := opsNATS("2.15.1")
	ephemeral.Spec.JetStream.StorageType = olddbapi.StorageTypeEphemeral
	ephemeral.Spec.Replicas = ptr.To(int32(1))

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeHorizontalScaling), "spec.horizontalScaling is nil"},
		{"nil replicas", nil, hs(nil), "replicas can not be empty"},
		{"zero", nil, hs(ptr.To(int32(0))), "at least 1"},
		{"up to 5", nil, hs(ptr.To(int32(5))), ""},
		{"two with jetstream", single, hs(ptr.To(int32(2))), "1 or at least 3"},
		{"1 to 3 is allowed", single, hs(ptr.To(int32(3))), ""},
		{"3 down to 1 is refused", nil, hs(ptr.To(int32(1))), "below 3 replicas"},
		{"5 down to 3 is allowed", func() *olddbapi.NATS { d := opsNATS("2.15.1"); d.Spec.Replicas = ptr.To(int32(5)); return d }(), hs(ptr.To(int32(3))), ""},
		{"without jetstream anything goes", noJS, hs(ptr.To(int32(2))), ""},
		{"ephemeral stays at one", ephemeral, hs(ptr.To(int32(3))), "only allows 1 replica"},
	})
}

func TestNATSOpsVerticalScaling(t *testing.T) {
	mem := func(limit, request string) *opsapi.ContainerResources {
		r := core.ResourceRequirements{Limits: core.ResourceList{core.ResourceMemory: resource.MustParse(limit)}}
		if request != "" {
			r.Requests = core.ResourceList{core.ResourceMemory: resource.MustParse(request)}
		}
		return &opsapi.ContainerResources{Resources: r}
	}
	vs := func(spec *opsapi.NATSVerticalScalingSpec) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeVerticalScaling)
		r.Spec.VerticalScaling = spec
		return r
	}
	withStore := opsNATS("2.15.1")
	withStore.Spec.JetStream.MemoryStoreSize = ptr.To(resource.MustParse("512Mi"))

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeVerticalScaling), "spec.verticalScaling is nil"},
		{"nothing to scale", nil, vs(&opsapi.NATSVerticalScalingSpec{}), "at least one of"},
		{"nats memory", nil, vs(&opsapi.NATSVerticalScalingSpec{NATS: mem("2Gi", "")}), ""},
		{"sidecar only", nil, vs(&opsapi.NATSVerticalScalingSpec{Exporter: mem("256Mi", "")}), ""},
		{"nats without memory limit", nil, vs(&opsapi.NATSVerticalScalingSpec{NATS: &opsapi.ContainerResources{}}), "limits.memory is required"},
		{"request above limit", nil, vs(&opsapi.NATSVerticalScalingSpec{NATS: mem("1Gi", "2Gi")}), "requests.memory can not exceed"},
		{"limit exactly twice the memory store", withStore, vs(&opsapi.NATSVerticalScalingSpec{NATS: mem("1Gi", "")}), ""},
		{"limit below twice the memory store", withStore, vs(&opsapi.NATSVerticalScalingSpec{NATS: mem("768Mi", "")}), "at least twice the JetStream memoryStoreSize"},
	})
}

func TestNATSOpsVolumeExpansion(t *testing.T) {
	ve := func(size string, mode opsapi.VolumeExpansionMode) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeVolumeExpansion)
		q := resource.MustParse(size)
		r.Spec.VolumeExpansion = &opsapi.NATSVolumeExpansionSpec{NATS: &q, Mode: mode}
		return r
	}
	noJS := opsNATS("2.15.1")
	noJS.Spec.JetStream.Enabled = ptr.To(false)
	ephemeral := opsNATS("2.15.1")
	ephemeral.Spec.JetStream.StorageType = olddbapi.StorageTypeEphemeral

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeVolumeExpansion), "spec.volumeExpansion is nil"},
		{"nil quantity", nil, func() *opsapi.NATSOpsRequest {
			r := natsOps(opsapi.NATSOpsRequestTypeVolumeExpansion)
			r.Spec.VolumeExpansion = &opsapi.NATSVolumeExpansionSpec{Mode: opsapi.VolumeExpansionModeOnline}
			return r
		}(), "spec.volumeExpansion.nats can not be empty"},
		{"missing mode", nil, ve("20Gi", ""), "mode is required"},
		{"grow", nil, ve("20Gi", opsapi.VolumeExpansionModeOnline), ""},
		{"not larger", nil, ve("10Gi", opsapi.VolumeExpansionModeOffline), "must be greater than current storage"},
		{"smaller", nil, ve("5Gi", opsapi.VolumeExpansionModeOnline), "must be greater than current storage"},
		{"jetstream off", noJS, ve("20Gi", opsapi.VolumeExpansionModeOnline), "durable JetStream file store"},
		{"ephemeral", ephemeral, ve("20Gi", opsapi.VolumeExpansionModeOnline), "durable JetStream file store"},
	})
}

func TestNATSOpsReconfigure(t *testing.T) {
	rc := func(c *opsapi.NATSCustomConfigurationSpec) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeReconfigure)
		r.Spec.Configuration = c
		return r
	}
	apply := func(key, content string) *opsapi.NATSCustomConfigurationSpec {
		return &opsapi.NATSCustomConfigurationSpec{ReconfigurationSpec: opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{key: content}}}
	}
	operatorMode := opsNATS("2.15.1")
	operatorMode.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeReconfigure), "spec.configuration is nil"},
		{"nothing to apply", nil, rc(&opsapi.NATSCustomConfigurationSpec{}), "nothing to apply"},
		{"ordinary key", nil, rc(apply(kubedb.NATSConfKey, "max_payload: 2097152")), ""},
		{"operator-owned key", nil, rc(apply(kubedb.NATSConfKey, "jetstream { max_file_store: 1 }")), "operator-owned keys: jetstream"},
		{"authorization key", nil, rc(apply(kubedb.NATSConfKey, "authorization { token: x }")), "operator-owned keys: authorization"},
		{"unsupported file key", nil, rc(apply("other.conf", "debug: true")), "unsupported configuration key"},
		{"accounts.conf in Config mode", nil, rc(apply(kubedb.NATSAccountsConfKey, "TEAM_A { }")), ""},
		{"accounts.conf in Operator mode", operatorMode, rc(apply(kubedb.NATSAccountsConfKey, "TEAM_A { }")), "not allowed in Operator auth mode"},
		{"accounts.conf reserved account", nil, rc(apply(kubedb.NATSAccountsConfKey, "APP { }")), "reserved"},
		{"empty configSecret name", nil, rc(&opsapi.NATSCustomConfigurationSpec{ReconfigurationSpec: opsapi.ReconfigurationSpec{ConfigSecret: &core.LocalObjectReference{}}}), "configSecret.name can not be empty"},
		{"remove custom config", nil, rc(&opsapi.NATSCustomConfigurationSpec{ReconfigurationSpec: opsapi.ReconfigurationSpec{RemoveCustomConfig: true}}), ""},
		{"serverConfig payload above pending", nil, rc(&opsapi.NATSCustomConfigurationSpec{ServerConfig: &olddbapi.NATSServerConfig{MaxPayload: ptr.To(int32(10)), MaxPending: ptr.To(int64(5))}}), "maxPayload must be <= maxPending"},
		{"serverConfig ok", nil, rc(&opsapi.NATSCustomConfigurationSpec{ServerConfig: &olddbapi.NATSServerConfig{MaxPayload: ptr.To(int32(5)), MaxPending: ptr.To(int64(10))}}), ""},
		{"lameDuck grace not below duration", nil, rc(&opsapi.NATSCustomConfigurationSpec{LameDuck: &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: time.Minute}, Duration: metav1.Duration{Duration: time.Minute}}}), "gracePeriod must be less than"},
		{"lameDuck duration below 30s", nil, rc(&opsapi.NATSCustomConfigurationSpec{LameDuck: &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: time.Second}, Duration: metav1.Duration{Duration: 10 * time.Second}}}), "at least 30s"},
		{"lameDuck ok", nil, rc(&opsapi.NATSCustomConfigurationSpec{LameDuck: &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: 10 * time.Second}, Duration: metav1.Duration{Duration: time.Minute}}}), ""},
	})
}

func TestNATSOpsReconfigureTLS(t *testing.T) {
	tlsOps := func(s opsapi.TLSSpec) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeReconfigureTLS)
		r.Spec.TLS = &opsapi.NATSTLSSpec{TLSSpec: s}
		return r
	}
	issuer := &core.TypedLocalObjectReference{Name: "issuer", Kind: "Issuer"}
	withTLS := opsNATS("2.15.1")
	withTLS.Spec.TLS = &kmapi.TLSConfig{IssuerRef: issuer}
	withTLS.SetDefaults(nil)

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, natsOps(opsapi.NATSOpsRequestTypeReconfigureTLS), "spec.tls is nil"},
		{"nothing requested", nil, tlsOps(opsapi.TLSSpec{}), "at least one of"},
		{"two operations", withTLS, tlsOps(opsapi.TLSSpec{Remove: true, RotateCertificates: true}), "only one TLS reconfiguration"},
		{"add tls", nil, tlsOps(opsapi.TLSSpec{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer}}), ""},
		{"certificates without issuer", nil, tlsOps(opsapi.TLSSpec{TLSConfig: kmapi.TLSConfig{Certificates: []kmapi.CertificateSpec{{Alias: "server"}}}}), "issuerRef is required"},
		{"certificates reuse existing issuer", withTLS, tlsOps(opsapi.TLSSpec{TLSConfig: kmapi.TLSConfig{Certificates: []kmapi.CertificateSpec{{Alias: "server"}}}}), ""},
		{"unsupported alias", nil, tlsOps(opsapi.TLSSpec{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer, Certificates: []kmapi.CertificateSpec{{Alias: "peer"}}}}), "unsupported certificate alias"},
		{"rotate without tls", nil, tlsOps(opsapi.TLSSpec{RotateCertificates: true}), "requires TLS to already be enabled"},
		{"rotate", withTLS, tlsOps(opsapi.TLSSpec{RotateCertificates: true}), ""},
		{"remove without tls", nil, tlsOps(opsapi.TLSSpec{Remove: true}), "nothing to remove"},
		{"remove", withTLS, tlsOps(opsapi.TLSSpec{Remove: true}), ""},
	})
}

func TestNATSOpsRotateAuth(t *testing.T) {
	ra := func(secret string) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeRotateAuth)
		if secret != "" {
			r.Spec.Authentication = &opsapi.AuthSpec{SecretRef: &appcat.TypedLocalObjectReference{Name: secret}}
		}
		return r
	}
	noSec := opsNATS("2.15.1")
	noSec.Spec.DisableSecurity = true
	operatorMode := opsNATS("2.15.1")
	operatorMode.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator

	good := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: "good", Namespace: "demo"}, Data: map[string][]byte{"username": []byte("admin"), "password": []byte("pw")}}
	wrongUser := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: "wrong-user", Namespace: "demo"}, Data: map[string][]byte{"username": []byte("bob"), "password": []byte("pw")}}
	noPass := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: "no-pass", Namespace: "demo"}, Data: map[string][]byte{"username": []byte("admin")}}

	cases := []struct {
		name    string
		db      *olddbapi.NATS
		secret  string
		wantErr string
	}{
		{"generated password", nil, "", ""},
		{"disableSecurity has nothing to rotate", noSec, "", "no credential to rotate"},
		{"operator mode refuses a user supplied secret", operatorMode, "good", "re-issued"},
		{"operator mode without secret", operatorMode, "", ""},
		{"good secret", nil, "good", ""},
		{"missing secret", nil, "absent", "not found"},
		{"wrong username", nil, "wrong-user", `must be "admin"`},
		{"no password", nil, "no-pass", "non-empty"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := c.db
			if db == nil {
				db = opsNATS("2.15.1")
			}
			w := natsOpsTestWebhook(t, db, good, wrongUser, noPass)
			err := w.validateCreateOrUpdate(ra(c.secret))
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestNATSOpsStorageMigration(t *testing.T) {
	sm := func(class string, timeout bool) *opsapi.NATSOpsRequest {
		r := natsOps(opsapi.NATSOpsRequestTypeStorageMigration)
		if class != "" {
			r.Spec.Migration = &opsapi.StorageMigrationSpec{StorageClassName: ptr.To(class)}
		}
		if timeout {
			r.Spec.Timeout = &metav1.Duration{Duration: time.Hour}
		}
		return r
	}
	single := opsNATS("2.15.1")
	single.Spec.Replicas = ptr.To(int32(1))
	ephemeral := opsNATS("2.15.1")
	ephemeral.Spec.JetStream.StorageType = olddbapi.StorageTypeEphemeral
	ephemeral.Spec.Replicas = ptr.To(int32(1))
	noJS := opsNATS("2.15.1")
	noJS.Spec.JetStream.Enabled = ptr.To(false)

	runNATSOpsCases(t, []natsOpsCase{
		{"nil spec", nil, sm("", true), "spec.migration is required"},
		{"nil class", nil, func() *opsapi.NATSOpsRequest {
			r := sm("x", true)
			r.Spec.Migration.StorageClassName = nil
			return r
		}(), "storageClassName is required"},
		{"timeout required", nil, sm("fast", false), "spec.timeout is required"},
		{"three replicas", nil, sm("fast", true), ""},
		{"single replica would lose the only copy", single, sm("fast", true), "needs at least 3 replicas"},
		{"ephemeral", ephemeral, sm("fast", true), "durable JetStream file store"},
		{"jetstream off", noJS, sm("fast", true), "durable JetStream file store"},
	})

	withClass := opsNATS("2.15.1")
	withClass.Spec.JetStream.Storage.StorageClassName = ptr.To("standard")
	runNATSOpsCases(t, []natsOpsCase{
		{"new class exists", withClass, sm("fast", true), ""},
		{"new class missing", withClass, sm("nope", true), "storage class nope not found"},
	})
}
