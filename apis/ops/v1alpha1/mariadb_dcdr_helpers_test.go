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

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func TestMariaDBDCDROpsRefusal(t *testing.T) {
	issuer := func(name string) *core.TypedLocalObjectReference {
		g := "cert-manager.io"
		return &core.TypedLocalObjectReference{APIGroup: &g, Kind: "Issuer", Name: name}
	}
	dcdr := &dbapi.MariaDB{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{dbapi.DCDREnabledAnnotation: "true"}},
		Spec:       dbapi.MariaDBSpec{Distributed: true},
	}
	dcdrTLS := dcdr.DeepCopy()
	dcdrTLS.Spec.TLS = &kmapi.TLSConfig{IssuerRef: issuer("ca")}
	plain := &dbapi.MariaDB{}
	q := resource.MustParse("2Gi")
	req := func(t MariaDBOpsRequestType, mut ...func(*MariaDBOpsRequestSpec)) *MariaDBOpsRequest {
		r := &MariaDBOpsRequest{Spec: MariaDBOpsRequestSpec{Type: t}}
		for _, m := range mut {
			m(&r.Spec)
		}
		return r
	}
	tests := []struct {
		name    string
		db      *dbapi.MariaDB
		req     *MariaDBOpsRequest
		refused bool
		substr  string
	}{
		{"restart allowed", dcdr, req(MariaDBOpsRequestTypeRestart), false, ""},
		{"horizontal scaling allowed", dcdr, req(MariaDBOpsRequestTypeHorizontalScaling), false, ""},
		{"horizontal scaling of maxscale", dcdr, req(MariaDBOpsRequestTypeHorizontalScaling, func(s *MariaDBOpsRequestSpec) {
			s.HorizontalScaling = &MariaDBHorizontalScalingSpec{MaxScale: true}
		}), true, "without MaxScale"},
		{"not dc-dr: anything goes", plain, req(MariaDBOpsRequestTypeStorageMigration), false, ""},
		{"distributed without the annotation is not dc-dr", &dbapi.MariaDB{Spec: dbapi.MariaDBSpec{Distributed: true}}, req(MariaDBOpsRequestTypeArchiverRestore), false, ""},
		{"storage migration never", dcdr, req(MariaDBOpsRequestTypeStorageMigration), true, "spoke clusters"},
		{"archiver restore never", dcdr, req(MariaDBOpsRequestTypeArchiverRestore), true, "new DC-DR MariaDB"},
		{"vertical scaling in place", dcdr, req(MariaDBOpsRequestTypeVerticalScaling, func(s *MariaDBOpsRequestSpec) {
			s.VerticalScaling = &MariaDBVerticalScalingSpec{Mode: VerticalScalingModeInPlace}
		}), true, "mode Restart"},
		{"vertical scaling of maxscale", dcdr, req(MariaDBOpsRequestTypeVerticalScaling, func(s *MariaDBOpsRequestSpec) {
			s.VerticalScaling = &MariaDBVerticalScalingSpec{MaxScale: &PodResources{}}
		}), true, "without MaxScale"},
		{"vertical scaling, mode Restart", dcdr, req(MariaDBOpsRequestTypeVerticalScaling), false, ""},
		{"volume expansion offline", dcdr, req(MariaDBOpsRequestTypeVolumeExpansion, func(s *MariaDBOpsRequestSpec) {
			s.VolumeExpansion = &MariaDBVolumeExpansionSpec{Mode: VolumeExpansionModeOffline}
		}), true, "mode Online"},
		{"volume expansion of maxscale", dcdr, req(MariaDBOpsRequestTypeVolumeExpansion, func(s *MariaDBOpsRequestSpec) {
			s.VolumeExpansion = &MariaDBVolumeExpansionSpec{Mode: VolumeExpansionModeOnline, MaxScale: &q}
		}), true, "without MaxScale"},
		{"volume expansion online", dcdr, req(MariaDBOpsRequestTypeVolumeExpansion, func(s *MariaDBOpsRequestSpec) {
			s.VolumeExpansion = &MariaDBVolumeExpansionSpec{Mode: VolumeExpansionModeOnline, MariaDB: &q}
		}), false, ""},
		{"rotate auth", dcdr, req(MariaDBOpsRequestTypeRotateAuth), false, ""},
		{"reconfigure", dcdr, req(MariaDBOpsRequestTypeReconfigure), false, ""},
		{"update version", dcdr, req(MariaDBOpsRequestTypeUpdateVersion), false, ""},
		{"tls on", dcdr, req(MariaDBOpsRequestTypeReconfigureTLS, func(s *MariaDBOpsRequestSpec) {
			s.TLS = &MariaDBTLSSpec{TLSSpec: TLSSpec{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer("ca")}}}
		}), true, "turning TLS on"},
		{"tls off", dcdrTLS, req(MariaDBOpsRequestTypeReconfigureTLS, func(s *MariaDBOpsRequestSpec) {
			s.TLS = &MariaDBTLSSpec{TLSSpec: TLSSpec{Remove: true}}
		}), true, "removing TLS"},
		{"new issuer", dcdrTLS, req(MariaDBOpsRequestTypeReconfigureTLS, func(s *MariaDBOpsRequestSpec) {
			s.TLS = &MariaDBTLSSpec{TLSSpec: TLSSpec{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer("other")}}}
		}), true, "changing the issuer"},
		{"same issuer, rotate", dcdrTLS, req(MariaDBOpsRequestTypeReconfigureTLS, func(s *MariaDBOpsRequestSpec) {
			s.TLS = &MariaDBTLSSpec{TLSSpec: TLSSpec{TLSConfig: kmapi.TLSConfig{IssuerRef: issuer("ca")}, RotateCertificates: true}}
		}), false, ""},
		{"requireSSL on", dcdrTLS, req(MariaDBOpsRequestTypeReconfigureTLS, func(s *MariaDBOpsRequestSpec) {
			on := true
			s.TLS = &MariaDBTLSSpec{RequireSSL: &on}
		}), false, ""},
		{"an unknown type", dcdr, req("Foo"), true, "not supported for a DC-DR MariaDB yet (supported: HorizontalScaling, Restart, VerticalScaling"},
		{"nil request", dcdr, nil, false, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MariaDBDCDROpsRefusal(tc.db, tc.req)
			if (got != "") != tc.refused {
				t.Fatalf("refused=%v, want %v (%q)", got != "", tc.refused, got)
			}
			if tc.substr != "" && !strings.Contains(got, tc.substr) {
				t.Fatalf("%q does not mention %q", got, tc.substr)
			}
		})
	}
}
