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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// stubClient serves DurableWorkflowVersions by name; anything else is a programming error in the test.
type stubClient struct {
	client.Client
	versions map[string]*catalog.DurableWorkflowVersion
}

func (c stubClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	v, ok := obj.(*catalog.DurableWorkflowVersion)
	if !ok {
		panic("stubClient only serves DurableWorkflowVersion")
	}
	found, ok := c.versions[key.Name]
	if !ok {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflowversions"}, key.Name)
	}
	*v = *found
	return nil
}

func dwTestWebhook(t *testing.T) *DurableWorkflowCustomWebhook {
	t.Helper()
	versions := []*catalog.DurableWorkflowVersion{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "temporal-1.32.1"},
			Spec: catalog.DurableWorkflowVersionSpec{
				Version: "1.32.1", Distribution: catalog.DurableWorkflowDistroTemporal, DefaultHistoryShards: 16,
				Schema: catalog.DurableWorkflowVersionSchema{Stores: []catalog.DurableWorkflowVersionStore{
					{Type: "Postgres", Default: true, Visibility: true},
					{Type: "MySQL", Default: true, Visibility: true},
					{Type: "Cassandra", Default: true},
					{Type: "Elasticsearch", Visibility: true},
				}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "cadence-1.4.2-upstream"},
			Spec: catalog.DurableWorkflowVersionSpec{
				Version: "1.4.2", Distribution: catalog.DurableWorkflowDistroCadence, DefaultHistoryShards: 16,
				Schema: catalog.DurableWorkflowVersionSchema{Stores: []catalog.DurableWorkflowVersionStore{
					{Type: "Postgres", Default: true, Visibility: true},
				}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "cadence-1.4.2"},
			Spec: catalog.DurableWorkflowVersionSpec{
				Version: "1.4.2", Distribution: catalog.DurableWorkflowDistroCadence, DefaultHistoryShards: 16, PublicClientTLS: true,
				Schema: catalog.DurableWorkflowVersionSchema{Stores: []catalog.DurableWorkflowVersionStore{
					{Type: "Postgres", Default: true, Visibility: true},
					{Type: "Cassandra", Default: true, Visibility: true},
					{Type: "Elasticsearch", Visibility: true},
				}},
			},
		},
	}
	c := stubClient{versions: map[string]*catalog.DurableWorkflowVersion{}}
	for _, v := range versions {
		c.versions[v.Name] = v
	}
	return &DurableWorkflowCustomWebhook{DefaultClient: c}
}

func dwTestObj(version string) *olddbapi.DurableWorkflow {
	return &olddbapi.DurableWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "dw", Namespace: "demo"},
		Spec: olddbapi.DurableWorkflowSpec{
			Version: version,
			Persistence: olddbapi.DurableWorkflowPersistence{
				Default: olddbapi.DurableWorkflowStore{
					ObjectReference: &kmapi.ObjectReference{Name: "pg", Namespace: "demo"},
					Type:            olddbapi.DurableWorkflowStorePostgres,
				},
			},
		},
	}
}

func archCase(name, version string, a *olddbapi.DurableWorkflowArchival, wantErr, wantWarn string) struct {
	name     string
	mutate   func(*olddbapi.DurableWorkflow)
	old      func(*olddbapi.DurableWorkflow)
	version  string
	wantErr  string // substring, empty means valid
	wantWarn string
} {
	c := struct {
		name     string
		mutate   func(*olddbapi.DurableWorkflow)
		old      func(*olddbapi.DurableWorkflow)
		version  string
		wantErr  string // substring, empty means valid
		wantWarn string
	}{name: name, version: version, wantErr: wantErr, wantWarn: wantWarn}
	c.mutate = func(d *olddbapi.DurableWorkflow) { d.Spec.Archival = a }
	return c
}

func TestDurableWorkflowValidate(t *testing.T) {
	es := olddbapi.DurableWorkflowStore{ObjectReference: &kmapi.ObjectReference{Name: "es"}, Type: olddbapi.DurableWorkflowStoreElasticsearch}
	cassandra := olddbapi.DurableWorkflowStore{ObjectReference: &kmapi.ObjectReference{Name: "c"}, Type: olddbapi.DurableWorkflowStoreCassandra}
	tests := []struct {
		name     string
		mutate   func(*olddbapi.DurableWorkflow)
		old      func(*olddbapi.DurableWorkflow)
		version  string
		wantErr  string // substring, empty means valid
		wantWarn string
	}{
		{name: "valid temporal", version: "temporal-1.32.1"},
		{name: "valid cadence", version: "cadence-1.4.2"},
		{name: "unknown version", version: "temporal-9.9.9", wantErr: "version not supported"},
		{
			name: "internal frontend on cadence", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Topology = &olddbapi.DurableWorkflowTopology{InternalFrontend: &olddbapi.DurableWorkflowNode{}}
			},
			wantErr: "internalFrontend is only available for Temporal",
		},
		{
			name: "temporal rejects cassandra visibility", version: "temporal-1.32.1",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.Persistence.Visibility = &cassandra },
			wantErr: "not supported as the visibility store",
		},
		{
			name: "cadence elasticsearch visibility is not supported yet", version: "cadence-1.4.2",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.Persistence.Visibility = &es },
			wantErr: "not supported for Cadence yet",
		},
		{
			name: "cadence with elasticsearch and kafka is not supported yet", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Persistence.Visibility = &es
				d.Spec.Persistence.KafkaRef = &kmapi.ObjectReference{Name: "kafka"}
			},
			wantErr: "not supported for Cadence yet",
		},
		{
			name: "temporal rejects kafka", version: "temporal-1.32.1",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.Persistence.KafkaRef = &kmapi.ObjectReference{Name: "kafka"} },
			wantErr: "kafkaRef is not used by Temporal",
		},
		{
			name: "auth secret required when not creating the database", version: "temporal-1.32.1",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.Persistence.Default.CreateDatabase = ptr.To(false) },
			wantErr: "authSecret is required",
		},
		{
			name: "cadence ui with required client auth and tls", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.UI = &olddbapi.DurableWorkflowUI{Enabled: true}
				d.Spec.TLS = &kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "issuer"}}
				d.Spec.FrontendClientAuth = ptr.To(olddbapi.DurableWorkflowClientAuthRequired)
			},
			wantErr: "can't present a client certificate",
		},
		{
			name: "cadence ui with tls defaults to no client auth", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.UI = &olddbapi.DurableWorkflowUI{Enabled: true}
				d.Spec.TLS = &kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "issuer"}}
			},
		},
		{
			name: "cadence tls without publicClient.tls", version: "cadence-1.4.2-upstream",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.TLS = &kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "issuer"}}
			},
			wantErr: "doesn't support TLS",
		},
		{
			name: "cadence tls with required client auth", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				required := olddbapi.DurableWorkflowClientAuthRequired
				d.Spec.TLS = &kmapi.TLSConfig{IssuerRef: &core.TypedLocalObjectReference{Name: "issuer"}}
				d.Spec.FrontendClientAuth = &required
			},
			wantErr: "can't present a client certificate",
		},
		{
			name: "temporal authorization needs an internal frontend", version: "temporal-1.32.1",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Authorization = &olddbapi.DurableWorkflowAuthorization{JWKSURL: "https://idp/jwks"}
			},
			wantErr: "needs an internalFrontend",
		},
		{
			name: "temporal authorization with an internal frontend", version: "temporal-1.32.1",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Topology = &olddbapi.DurableWorkflowTopology{InternalFrontend: &olddbapi.DurableWorkflowNode{}}
				d.Spec.Authorization = &olddbapi.DurableWorkflowAuthorization{JWKSURL: "https://idp/jwks"}
			},
		},
		{
			name: "cadence authorization verifies external tokens by JWKS", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Authorization = &olddbapi.DurableWorkflowAuthorization{PublicKeySecret: &core.LocalObjectReference{Name: "keys"}}
			},
			wantErr: "Cadence verifies external tokens with a JWKS URL",
		},
		{
			name: "authorization needs a key source", version: "cadence-1.4.2",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Authorization = &olddbapi.DurableWorkflowAuthorization{}
			},
			wantErr: "one of jwksURL and publicKeySecret is required",
		},
		{
			name: "ui jwt auth on temporal", version: "temporal-1.32.1",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.UI = &olddbapi.DurableWorkflowUI{Enabled: true, Auth: olddbapi.DurableWorkflowUIAuthJWT}
			},
			wantErr: "only available for Cadence",
		},
		archCase("valid s3 on temporal", "temporal-1.32.1", &olddbapi.DurableWorkflowArchival{
			History:    &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/h", ConfigSecret: &core.LocalObjectReference{Name: "s3"}},
			Visibility: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/v", ConfigSecret: &core.LocalObjectReference{Name: "s3"}},
		}, "", ""),
		archCase("valid gcs on cadence", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "gcs", URI: "gs://b/h", ConfigSecret: &core.LocalObjectReference{Name: "gcs"}},
		}, "", ""),
		archCase("gcs on temporal warns about the build tag", "temporal-1.32.1", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "gcs", URI: "gs://b/h", ConfigSecret: &core.LocalObjectReference{Name: "gcs"}},
		}, "", "temporal_with_gstorage"),
		archCase("filestore warns it is not durable", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file:///var/archive/h"},
		}, "", "not shared between pods"),
		archCase("provider and scheme mismatch", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "gs://b/h", ConfigSecret: &core.LocalObjectReference{Name: "s3"}},
		}, "needs a s3:// URI", ""),
		archCase("s3 without a secret", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/h"},
		}, "needs a configSecret", ""),
		archCase("s3 without a bucket", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3:///h", ConfigSecret: &core.LocalObjectReference{Name: "s3"}},
		}, "bucket is missing", ""),
		archCase("filestore with a secret", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file:///a", ConfigSecret: &core.LocalObjectReference{Name: "x"}},
		}, "no credentials", ""),
		archCase("filestore relative path", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file://archive/h"},
		}, "file:///absolute/path", ""),
		archCase("filestore root", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file:///"},
		}, "below /", ""),
		archCase("nested filestore paths", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History:    &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file:///a"},
			Visibility: &olddbapi.DurableWorkflowArchivalTarget{Provider: "filestore", URI: "file:///a/v"},
		}, "can't be nested", ""),
		archCase("two s3 secrets", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History:    &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/h", ConfigSecret: &core.LocalObjectReference{Name: "a"}},
			Visibility: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/v", ConfigSecret: &core.LocalObjectReference{Name: "b"}},
		}, "same s3 secret", ""),
		archCase("unknown provider", "cadence-1.4.2", &olddbapi.DurableWorkflowArchival{
			History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "azure", URI: "az://b"},
		}, "Unsupported value", ""),
		{
			name: "tls without issuer", version: "temporal-1.32.1",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.TLS = &kmapi.TLSConfig{} },
			wantErr: "issuerRef is missing",
		},
		{
			name: "shards are immutable", version: "temporal-1.32.1",
			mutate:  func(d *olddbapi.DurableWorkflow) { d.Spec.NumHistoryShards = ptr.To(int32(32)) },
			old:     func(d *olddbapi.DurableWorkflow) { d.Spec.NumHistoryShards = ptr.To(int32(16)) },
			wantErr: "numHistoryShards is immutable",
		},
		{
			name: "history replicas above shards warn", version: "temporal-1.32.1",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.NumHistoryShards = ptr.To(int32(4))
				d.Spec.Topology = &olddbapi.DurableWorkflowTopology{History: &olddbapi.DurableWorkflowNode{Replicas: ptr.To(int32(8))}}
			},
			wantWarn: "exceed numHistoryShards",
		},
		{
			name: "zero replicas", version: "temporal-1.32.1",
			mutate: func(d *olddbapi.DurableWorkflow) {
				d.Spec.Topology = &olddbapi.DurableWorkflowTopology{Matching: &olddbapi.DurableWorkflowNode{Replicas: ptr.To(int32(0))}}
			},
			wantErr: "number of replicas can not be 0 or less",
		},
	}

	w := dwTestWebhook(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := dwTestObj(tc.version)
			if tc.mutate != nil {
				tc.mutate(db)
			}
			var old *olddbapi.DurableWorkflow
			if tc.old != nil {
				old = dwTestObj(tc.version)
				tc.old(old)
			}
			warnings, err := w.validate(db, old)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
			if tc.wantWarn != "" && !strings.Contains(strings.Join(warnings, ";"), tc.wantWarn) {
				t.Fatalf("want warning containing %q, got %v", tc.wantWarn, warnings)
			}
		})
	}
}

func TestDurableWorkflowSetDefaults(t *testing.T) {
	w := dwTestWebhook(t)
	db := dwTestObj("temporal-1.32.1")
	db.SetDefaults(w.DefaultClient)

	if got := ptr.Deref(db.Spec.NumHistoryShards, 0); got != 16 {
		t.Errorf("numHistoryShards default = %d, want 16", got)
	}
	if db.Spec.DeletionPolicy != olddbapi.DeletionPolicyDelete {
		t.Errorf("deletionPolicy default = %q", db.Spec.DeletionPolicy)
	}
	for _, role := range []string{"frontend", "history", "matching", "worker"} {
		if db.Replicas(role) != 1 {
			t.Errorf("%s replicas = %d, want 1", role, db.Replicas(role))
		}
	}
	if db.Spec.Persistence.Default.DatabaseName != "dw" || !db.Spec.Persistence.Default.ShouldCreateDatabase() {
		t.Errorf("default store defaults not applied: %+v", db.Spec.Persistence.Default)
	}
	if got := db.StoreDatabaseName(true); got != "dw_visibility" || got == db.StoreDatabaseName(false) {
		t.Errorf("the visibility store needs a database of its own, got %q next to %q", got, db.StoreDatabaseName(false))
	}
	if db.Spec.Topology.InternalFrontend != nil {
		t.Errorf("internal frontend must not be defaulted on")
	}

	// explicit values are kept
	db = dwTestObj("cadence-1.4.2")
	db.Spec.NumHistoryShards = ptr.To(int32(1024))
	db.SetDefaults(w.DefaultClient)
	if got := ptr.Deref(db.Spec.NumHistoryShards, 0); got != 1024 {
		t.Errorf("explicit numHistoryShards overwritten: %d", got)
	}
}
