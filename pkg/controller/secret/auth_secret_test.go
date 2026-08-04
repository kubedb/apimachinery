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

package secret

import (
	"context"
	"testing"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// stubClient implements only the three methods CreateOrPatch reaches on the patch
// path for a non-official API group: Scheme, Get and Patch. The embedded interface
// is nil, so anything else panics loudly instead of silently misbehaving.
type stubClient struct {
	client.Client
	scheme *runtime.Scheme
	stored *dbapi.Solr
}

func (s *stubClient) Scheme() *runtime.Scheme { return s.scheme }

// Get mirrors the real client: the fetched object replaces the contents of the
// target it is handed.
func (s *stubClient) Get(_ context.Context, _ client.ObjectKey, out client.Object, _ ...client.GetOption) error {
	s.stored.DeepCopyInto(out.(*dbapi.Solr))
	return nil
}

func (s *stubClient) Patch(_ context.Context, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
	s.stored = obj.(*dbapi.Solr).DeepCopy()
	return nil
}

// TestPatchAuthSecretRefDoesNotMutateCaller pins the contract that ops requests
// depend on: patchAuthSecretRef records the auth secret reference without replacing
// the caller's object. CreateOrPatch writes post-patch server state back over
// whatever object it is given, so passing o.DB straight through would wipe the
// desired state an ops request has staged but not yet persisted -- which silently
// drops the backup configuration and credentials from the rendered workload.
func TestPatchAuthSecretRefDoesNotMutateCaller(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := dbapi.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to register scheme: %v", err)
	}

	// What the API server holds mid-reconfigure: no spec.configuration yet, because
	// the ops request only patches the CR once the workload has been updated.
	stored := &dbapi.Solr{
		ObjectMeta: metav1.ObjectMeta{Name: "solr-src", Namespace: "demo"},
	}

	// What the ops request reconciles from: desired state that is deliberately not
	// on the API server, mirroring what applyBackupCredentials produces.
	db := &dbapi.Solr{
		ObjectMeta: metav1.ObjectMeta{Name: "solr-src", Namespace: "demo"},
		Spec: dbapi.SolrSpec{
			Configuration: &dbapi.SolrConfiguration{
				ConfigurationSpec: dbapi.ConfigurationSpec{
					Inline: map[string]string{"solr.xml": "<solr/>"},
				},
				BackupCredentials: &dbapi.SolrBackupCredentials{
					S3: &dbapi.SolrS3Credential{SecretName: "aws-s3-secret"},
				},
			},
		},
	}

	kc := &stubClient{scheme: scheme, stored: stored}
	if err := (Options{KBClient: kc, DB: db}).patchAuthSecretRef(context.TODO(), "solr-src-auth", nil); err != nil {
		t.Fatalf("patchAuthSecretRef returned an error: %v", err)
	}

	// The staged desired state must survive the call.
	if db.Spec.Configuration == nil {
		t.Fatal("Spec.Configuration was wiped: the caller's object was replaced with server state")
	}
	if got, want := db.Spec.Configuration.Inline["solr.xml"], "<solr/>"; got != want {
		t.Errorf("inline config = %q, want %q", got, want)
	}
	if db.Spec.Configuration.BackupCredentials == nil ||
		db.Spec.Configuration.BackupCredentials.S3 == nil {
		t.Error("backup credentials were wiped from the caller's object")
	}

	// The auth reference must still be recorded -- on the caller's object...
	if db.Spec.AuthSecret == nil {
		t.Fatal("Spec.AuthSecret was not set on the caller's object")
	}
	if got, want := db.Spec.AuthSecret.Name, "solr-src-auth"; got != want {
		t.Errorf("caller Spec.AuthSecret.Name = %q, want %q", got, want)
	}

	// ...and on the server.
	if kc.stored.Spec.AuthSecret == nil {
		t.Fatal("Spec.AuthSecret was not patched onto the server object")
	}
	if got, want := kc.stored.Spec.AuthSecret.Name, "solr-src-auth"; got != want {
		t.Errorf("server Spec.AuthSecret.Name = %q, want %q", got, want)
	}
}
