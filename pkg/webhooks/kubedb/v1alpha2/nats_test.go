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
	"strings"
	"testing"
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func natsTestWebhook(t *testing.T, objs ...runtime.Object) *NATSCustomWebhook {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := catalog.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := olddbapi.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := core.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	objs = append(objs, &catalog.NATSVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "2.15.1"},
		Spec: catalog.NATSVersionSpec{
			Version:  "2.15.1",
			DB:       catalog.NATSVersionDatabase{Image: "nats:2.15.1-alpine"},
			Reloader: catalog.NATSVersionReloader{Image: "reloader:0.24.0"},
			Exporter: catalog.NATSVersionExporter{Image: "exporter:0.20.2"},
		},
	}, &catalog.NATSVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "2.12.0"},
		Spec: catalog.NATSVersionSpec{
			Version:    "2.12.0",
			Deprecated: true,
			DB:         catalog.NATSVersionDatabase{Image: "nats:2.12.0-alpine"},
			Reloader:   catalog.NATSVersionReloader{Image: "reloader:0.24.0"},
			Exporter:   catalog.NATSVersionExporter{Image: "exporter:0.20.2"},
		},
	}, &catalog.NATSVersion{
		// images missing
		ObjectMeta: metav1.ObjectMeta{Name: "broken"},
		Spec:       catalog.NATSVersionSpec{Version: "2.15.2", DB: catalog.NATSVersionDatabase{Image: "nats"}},
	})
	return &NATSCustomWebhook{
		DefaultClient:    fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build(),
		StrictValidation: true,
	}
}

// newNATS returns a valid, defaulted 3 replica NATS.
func newNATS() *olddbapi.NATS {
	db := &olddbapi.NATS{
		ObjectMeta: metav1.ObjectMeta{Name: "nats", Namespace: "demo"},
		Spec: olddbapi.NATSSpec{
			Version: "2.15.1",
		},
	}
	db.SetDefaults(nil)
	return db
}

func natsContainer(db *olddbapi.NATS) *core.Container {
	for i := range db.Spec.PodTemplate.Spec.Containers {
		if db.Spec.PodTemplate.Spec.Containers[i].Name == kubedb.NATSContainerName {
			return &db.Spec.PodTemplate.Spec.Containers[i]
		}
	}
	return nil
}

func TestNATSDefaultsPass(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	warnings, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
	if len(errs) != 0 || len(warnings) != 0 {
		t.Fatalf("defaulted NATS must be valid, errs=%v warnings=%v", errs, warnings)
	}
	if *db.Spec.Replicas != 3 || !db.IsJetStreamEnabled() || db.Spec.JetStream.StorageType != olddbapi.StorageTypeDurable {
		t.Fatalf("unexpected defaults: %+v", db.Spec)
	}
	if db.AuthMode() != olddbapi.NATSAuthModeConfig {
		t.Fatalf("auth mode should default to Config")
	}
}

func TestNATSReplicasRule(t *testing.T) {
	w := natsTestWebhook(t)
	cases := []struct {
		name      string
		replicas  int32
		jsOff     bool
		wantErr   string
		wantWarns bool
	}{
		{"zero", 0, false, "less than 1", false},
		{"two with jetstream", 2, false, "1 or at least 3", false},
		{"one", 1, false, "", false},
		{"three", 3, false, "", false},
		{"five", 5, false, "", false},
		{"four warns", 4, false, "", true},
		{"two without jetstream", 2, true, "", false},
		{"four without jetstream does not warn", 4, true, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := newNATS()
			db.Spec.Replicas = ptr.To(c.replicas)
			if c.jsOff {
				db.Spec.JetStream.Enabled = ptr.To(false)
			}
			warnings, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
			got := errs.ToAggregate()
			switch {
			case c.wantErr == "" && got != nil:
				t.Fatalf("unexpected error: %v", got)
			case c.wantErr != "" && (got == nil || !strings.Contains(got.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, got)
			}
			if c.wantWarns != (len(warnings) > 0) {
				t.Fatalf("wantWarns=%v, got %v", c.wantWarns, warnings)
			}
		})
	}
}

func TestNATSEphemeralRequiresSingleReplica(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	db.Spec.JetStream.StorageType = olddbapi.StorageTypeEphemeral
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); !strings.Contains(errs.ToAggregate().Error(), "Ephemeral JetStream storage") {
		t.Fatalf("ephemeral with 3 replicas must be rejected, got %v", errs)
	}
	db.Spec.Replicas = ptr.To(int32(1))
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDelete
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
		t.Fatalf("ephemeral with 1 replica must pass, got %v", errs)
	}
}

func TestNATSJetStreamImmutable(t *testing.T) {
	w := natsTestWebhook(t)
	ctx := context.TODO()

	old := newNATS()
	enabledFlip := newNATS()
	enabledFlip.Spec.JetStream.Enabled = ptr.To(false)
	if _, err := w.ValidateUpdate(ctx, old, enabledFlip); err == nil || !strings.Contains(err.Error(), "spec.jetStream.enabled is immutable") {
		t.Fatalf("flipping jetStream.enabled must be rejected, got %v", err)
	}

	typeFlip := newNATS()
	typeFlip.Spec.JetStream.StorageType = olddbapi.StorageTypeEphemeral
	typeFlip.Spec.Replicas = ptr.To(int32(1))
	if _, err := w.ValidateUpdate(ctx, old, typeFlip); err == nil || !strings.Contains(err.Error(), "spec.jetStream.storageType is immutable") {
		t.Fatalf("flipping jetStream.storageType must be rejected, got %v", err)
	}

	if _, err := w.ValidateUpdate(ctx, old, newNATS()); err != nil {
		t.Fatalf("unchanged update must pass, got %v", err)
	}
}

func TestNATSLameDuckRules(t *testing.T) {
	w := natsTestWebhook(t)
	cases := []struct {
		name    string
		mutate  func(db *olddbapi.NATS)
		wantErr string
	}{
		{"defaults", func(db *olddbapi.NATS) {}, ""},
		{"grace equal to duration", func(db *olddbapi.NATS) {
			db.Spec.LameDuck = &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: time.Minute}, Duration: metav1.Duration{Duration: time.Minute}}
		}, "gracePeriod must be less than lameDuck.duration"},
		{"grace above duration", func(db *olddbapi.NATS) {
			db.Spec.LameDuck = &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: 2 * time.Minute}, Duration: metav1.Duration{Duration: time.Minute}}
		}, "gracePeriod must be less than lameDuck.duration"},
		{"duration below server minimum", func(db *olddbapi.NATS) {
			db.Spec.LameDuck = &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: 5 * time.Second}, Duration: metav1.Duration{Duration: 20 * time.Second}}
		}, "at least 30s"},
		{"valid longer drain", func(db *olddbapi.NATS) {
			db.Spec.LameDuck = &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: 30 * time.Second}, Duration: metav1.Duration{Duration: 2 * time.Minute}}
			db.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds = ptr.To(int64(150))
		}, ""},
		{"termination grace below duration plus 30s", func(db *olddbapi.NATS) {
			db.Spec.LameDuck = &olddbapi.NATSLameDuckSpec{GracePeriod: metav1.Duration{Duration: 30 * time.Second}, Duration: metav1.Duration{Duration: 2 * time.Minute}}
			db.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds = ptr.To(int64(120))
		}, "terminationGracePeriodSeconds must be at least 150"},
		{"termination grace below the 60s floor", func(db *olddbapi.NATS) {
			db.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds = ptr.To(int64(30))
		}, "terminationGracePeriodSeconds must be at least 60"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := newNATS()
			c.mutate(db)
			_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
			got := errs.ToAggregate()
			switch {
			case c.wantErr == "" && got != nil:
				t.Fatalf("unexpected error: %v", got)
			case c.wantErr != "" && (got == nil || !strings.Contains(got.Error(), c.wantErr)):
				t.Fatalf("want error containing %q, got %v", c.wantErr, got)
			}
		})
	}
}

func TestNATSTerminationGracePeriodFormula(t *testing.T) {
	cases := []struct {
		duration time.Duration
		want     int64
	}{
		{30 * time.Second, 60}, // catalog default: 30 + 30
		{time.Minute, 90},
		{2 * time.Minute, 150},
		{10 * time.Second, 60}, // floor
	}
	for _, c := range cases {
		if got := olddbapi.NATSTerminationGracePeriodSeconds(c.duration); got != c.want {
			t.Errorf("duration %v: got %d, want %d", c.duration, got, c.want)
		}
	}
	db := newNATS()
	if got := *db.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds; got != 60 {
		t.Errorf("defaulted terminationGracePeriodSeconds: got %d, want 60", got)
	}
}

func TestNATSMaxFileStoreIsNinetyPercentOfPVC(t *testing.T) {
	db := newNATS()
	if got, want := db.MaxFileStoreBytes(), int64(10)*1024*1024*1024/100*90; got != want {
		t.Fatalf("max_file_store for 10Gi: got %d, want %d", got, want)
	}
	db.Spec.JetStream.Storage.Resources.Requests[core.ResourceStorage] = resource.MustParse("100Gi")
	if got, want := db.MaxFileStoreBytes(), int64(100)*1024*1024*1024/100*90; got != want {
		t.Fatalf("max_file_store for 100Gi: got %d, want %d", got, want)
	}
	db.Spec.JetStream.Enabled = ptr.To(false)
	if db.MaxFileStoreBytes() != 0 {
		t.Fatalf("max_file_store must be 0 with JetStream off")
	}
}

func TestNATSMemoryRules(t *testing.T) {
	w := natsTestWebhook(t)

	db := newNATS()
	natsContainer(db).Resources.Limits = nil
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); !strings.Contains(errs.ToAggregate().Error(), "limits.memory of the nats container is required") {
		t.Fatalf("missing limits.memory must be rejected, got %v", errs)
	}

	// default limit is 1Gi, so up to 512Mi is allowed
	db = newNATS()
	db.Spec.JetStream.MemoryStoreSize = ptr.To(resource.MustParse("512Mi"))
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
		t.Fatalf("memoryStoreSize equal to 50%% must pass, got %v", errs)
	}
	db.Spec.JetStream.MemoryStoreSize = ptr.To(resource.MustParse("513Mi"))
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); !strings.Contains(errs.ToAggregate().Error(), "must not exceed 50% of limits.memory") {
		t.Fatalf("memoryStoreSize above 50%% must be rejected, got %v", errs)
	}
}

func TestNATSOperatorOwnedKeysRejected(t *testing.T) {
	w := natsTestWebhook(t)

	for _, key := range []string{
		"server_name", "listen", "port", "host", "http", "http_port", "https_port", "http_base_path",
		"pid_file", "lame_duck_duration", "lame_duck_grace_period", "cluster", "gateway", "leafnodes",
		"jetstream", "store_dir", "system_account", "accounts", "authorization", "tls", "include",
		// keys the operator writes in some mode, beyond the design list
		"server_tags", "no_auth_user", "operator", "resolver", "resolver_preload",
		"SERVER_NAME", // config keys are case-insensitive
	} {
		t.Run(key, func(t *testing.T) {
			db := newNATS()
			db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSConfKey: key + ": x\n"}}
			_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
			if !strings.Contains(errs.ToAggregate().Error(), "operator-owned keys") {
				t.Fatalf("key %q must be rejected, got %v", key, errs)
			}
		})
	}

	allowed := `
# limits are ordinary settings
max_payload: 2097152
max_connections = 1000
ping_interval: "30s"
debug: true
logtime: true
// a comment mentioning jetstream { } must not count
websocket_note "server_name is only text here"
log_file: '/tmp/jetstream.log'
`
	db := newNATS()
	db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSConfKey: allowed}}
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
		t.Fatalf("ordinary settings must pass, got %v", errs)
	}

	t.Run("owned key behind a block is found", func(t *testing.T) {
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSConfKey: "max_payload: 1\njetstream {\n  max_file_store: 5\n}\n"}}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "jetstream") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("nested owned keys are not top-level", func(t *testing.T) {
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSConfKey: "websocket {\n  port: 8080\n  tls { cert_file: x }\n}\n"}}
		if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
			t.Fatalf("nested keys must pass, got %v", errs)
		}
	})
	t.Run("unbalanced braces", func(t *testing.T) {
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSConfKey: "websocket {\n  port: 8080\n"}}
		if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) == 0 {
			t.Fatalf("unbalanced braces must be rejected")
		}
	})
	t.Run("unsupported file key", func(t *testing.T) {
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{"other.conf": "debug: true"}}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "unsupported configuration key") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("accounts.conf with reserved account", func(t *testing.T) {
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSAccountsConfKey: "SYS { users: [] }"}}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), `account name "SYS" is reserved`) {
			t.Fatalf("got %v", errs)
		}
		db.Spec.Configuration.Inline[kubedb.NATSAccountsConfKey] = "TEAM_A { users: [{user: a, password: b}] }"
		if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
			t.Fatalf("extra account must pass, got %v", errs)
		}
	})
	t.Run("referenced secret is checked too", func(t *testing.T) {
		sec := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "user-conf", Namespace: "demo"},
			Data:       map[string][]byte{kubedb.NATSConfKey: []byte("jetstream { }")},
		}
		w := natsTestWebhook(t, sec)
		db := newNATS()
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{SecretName: "user-conf"}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "operator-owned keys") {
			t.Fatalf("got %v", errs)
		}
	})
}

func TestNATSServerConfigRules(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	db.Spec.ServerConfig = &olddbapi.NATSServerConfig{MaxPayload: ptr.To(int32(2000)), MaxPending: ptr.To(int64(1000))}
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); !strings.Contains(errs.ToAggregate().Error(), "maxPayload must be <= maxPending") {
		t.Fatalf("got %v", errs)
	}
	db.Spec.ServerConfig = &olddbapi.NATSServerConfig{
		MaxPayload: ptr.To(int32(1000)), MaxPending: ptr.To(int64(1000)), MaxConnections: ptr.To(int32(100)),
		WriteDeadline: &metav1.Duration{Duration: 10 * time.Second},
	}
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
		t.Fatalf("got %v", errs)
	}
	db.Spec.ServerConfig = &olddbapi.NATSServerConfig{MaxConnections: ptr.To(int32(-1))}
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); !strings.Contains(errs.ToAggregate().Error(), "maxConnections must be positive") {
		t.Fatalf("got %v", errs)
	}
}

func TestNATSTLSRules(t *testing.T) {
	w := natsTestWebhook(t)

	t.Run("clusterTLS without tls", func(t *testing.T) {
		db := newNATS()
		db.Spec.ClusterTLS = ptr.To(true)
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "spec.clusterTLS can only be set together with spec.tls") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("tls without issuerRef", func(t *testing.T) {
		db := newNATS()
		db.Spec.TLS = &kmapi.TLSConfig{}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "spec.tls.issuerRef is missing") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("unknown alias", func(t *testing.T) {
		db := newNATS()
		db.Spec.TLS = &kmapi.TLSConfig{
			IssuerRef:    &core.TypedLocalObjectReference{Name: "issuer", Kind: "Issuer"},
			Certificates: []kmapi.CertificateSpec{{Alias: "peer"}},
		}
		_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "supported aliases") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("tls with all channels defaults to on and passes", func(t *testing.T) {
		db := newNATS()
		db.Spec.TLS = &kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "issuer", Kind: "Issuer"}}
		db.SetDefaults(nil)
		if !db.IsClientTLS() || !db.IsClusterTLS() || !db.IsMonitorTLS() || db.MonitorScheme() != "https" {
			t.Fatalf("all channels must default to TLS")
		}
		if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) != 0 {
			t.Fatalf("got %v", errs)
		}
		db.Spec.ClusterTLS = ptr.To(false)
		if db.IsClusterTLS() {
			t.Fatalf("clusterTLS=false must switch routes to plaintext")
		}
	})
}

func TestNATSAuthRules(t *testing.T) {
	w := natsTestWebhook(t)
	ctx := context.TODO()

	t.Run("mode is immutable", func(t *testing.T) {
		old := newNATS()
		op := newNATS()
		op.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator
		if _, err := w.ValidateUpdate(ctx, old, op); err == nil || !strings.Contains(err.Error(), "spec.auth.mode is immutable") {
			t.Fatalf("got %v", err)
		}
		if _, err := w.ValidateUpdate(ctx, op, old); err == nil || !strings.Contains(err.Error(), "spec.auth.mode is immutable") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("disableSecurity is immutable", func(t *testing.T) {
		old := newNATS()
		nw := newNATS()
		nw.Spec.DisableSecurity = true
		if _, err := w.ValidateUpdate(ctx, old, nw); err == nil || !strings.Contains(err.Error(), "spec.disableSecurity is immutable") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("operator mode rejects disableSecurity", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator
		db.Spec.DisableSecurity = true
		_, errs := w.ValidateCreateOrUpdate(ctx, db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "disableSecurity is not allowed with spec.auth.mode: Operator") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("operator mode always gets the resolver PVC", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator
		db.Spec.JetStream.Enabled = ptr.To(false)
		db.SetDefaults(nil)
		if !db.HasResolverPVC() {
			t.Fatalf("Operator mode must have a resolver PVC")
		}
		q := db.Spec.Auth.Operator.ResolverStorage.Resources.Requests[core.ResourceStorage]
		if q.String() != "1Gi" {
			t.Fatalf("resolver storage should default to 1Gi, got %s", q.String())
		}
		if _, errs := w.ValidateCreateOrUpdate(ctx, db, true); len(errs) != 0 {
			t.Fatalf("operator mode without JetStream must pass, got %v", errs)
		}
	})
	t.Run("operator mode rejects accounts.conf", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator
		db.Spec.Configuration = &olddbapi.ConfigurationSpec{Inline: map[string]string{kubedb.NATSAccountsConfKey: "TEAM_A { }"}}
		_, errs := w.ValidateCreateOrUpdate(ctx, db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "not allowed in Operator auth mode") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("operator block needs Operator mode", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Operator = &olddbapi.NATSOperatorSpec{}
		_, errs := w.ValidateCreateOrUpdate(ctx, db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "only allowed with spec.auth.mode: Operator") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("external operator secret must be named", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Mode = olddbapi.NATSAuthModeOperator
		db.Spec.Auth.Operator = &olddbapi.NATSOperatorSpec{ExternalSecret: &core.LocalObjectReference{}}
		_, errs := w.ValidateCreateOrUpdate(ctx, db, true)
		if !strings.Contains(errs.ToAggregate().Error(), "externalSecret.name can not be empty") {
			t.Fatalf("got %v", errs)
		}
	})
	t.Run("unknown mode", func(t *testing.T) {
		db := newNATS()
		db.Spec.Auth.Mode = "Kerberos"
		_, errs := w.ValidateCreateOrUpdate(ctx, db, true)
		if len(errs) == 0 {
			t.Fatalf("unknown mode must be rejected")
		}
	})
}

func TestNATSHaltedRule(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	db.Spec.Halted = true
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
	_, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true)
	if !strings.Contains(errs.ToAggregate().Error(), "DoNotTerminate") {
		t.Fatalf("halted with DoNotTerminate must be rejected, got %v", errs)
	}
	if err := w.Default(context.TODO(), db); err == nil {
		t.Fatalf("defaulting halted with DoNotTerminate must fail")
	}

	db = newNATS()
	db.Spec.Halted = true
	if err := w.Default(context.TODO(), db); err != nil {
		t.Fatal(err)
	}
	if db.Spec.DeletionPolicy != olddbapi.DeletionPolicyHalt {
		t.Fatalf("halted must set deletionPolicy Halt, got %q", db.Spec.DeletionPolicy)
	}
}

func TestNATSVersionRule(t *testing.T) {
	w := natsTestWebhook(t)
	ctx := context.TODO()

	db := newNATS()
	db.Spec.Version = "9.9.9"
	if _, errs := w.ValidateCreateOrUpdate(ctx, db, true); !strings.Contains(errs.ToAggregate().Error(), "not found in the catalog") {
		t.Fatalf("unknown version must be rejected, got %v", errs)
	}

	db.Spec.Version = "2.12.0"
	if _, errs := w.ValidateCreateOrUpdate(ctx, db, true); !strings.Contains(errs.ToAggregate().Error(), "deprecated") {
		t.Fatalf("deprecated version must be rejected, got %v", errs)
	}
	// an update that keeps a now-deprecated version is not blocked
	if _, errs := w.ValidateCreateOrUpdate(ctx, db, false); len(errs) != 0 {
		t.Fatalf("keeping a deprecated version must pass, got %v", errs)
	}

	db.Spec.Version = "broken"
	if _, errs := w.ValidateCreateOrUpdate(ctx, db, true); !strings.Contains(errs.ToAggregate().Error(), "invalid NATSVersion") {
		t.Fatalf("a version with missing images must be rejected, got %v", errs)
	}

	db.Spec.Version = ""
	if _, errs := w.ValidateCreateOrUpdate(ctx, db, true); !strings.Contains(errs.ToAggregate().Error(), "spec.version is missing") {
		t.Fatalf("got %v", errs)
	}
}

func TestNATSDeleteRule(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
	if _, err := w.ValidateDelete(context.TODO(), db); err == nil {
		t.Fatalf("DoNotTerminate must block delete")
	}
	db.Spec.DeletionPolicy = olddbapi.DeletionPolicyDelete
	if _, err := w.ValidateDelete(context.TODO(), db); err != nil {
		t.Fatal(err)
	}
}

func TestNATSReservedVolumesAndEnv(t *testing.T) {
	w := natsTestWebhook(t)
	db := newNATS()
	c := natsContainer(db)
	c.Env = append(c.Env, core.EnvVar{Name: kubedb.NATSEnvGoMemLimit, Value: "1"})
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) == 0 {
		t.Fatalf("overriding GOMEMLIMIT must be rejected")
	}

	db = newNATS()
	c = natsContainer(db)
	c.VolumeMounts = append(c.VolumeMounts, core.VolumeMount{Name: "x", MountPath: kubedb.NATSConfigDir})
	if _, errs := w.ValidateCreateOrUpdate(context.TODO(), db, true); len(errs) == 0 {
		t.Fatalf("mounting over the config dir must be rejected")
	}
}
