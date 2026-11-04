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
	"testing"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func TestMilvusCertificateSources(t *testing.T) {
	issuer := &core.TypedLocalObjectReference{Kind: "Issuer", Name: "ca"}
	ref := func(n string) *core.LocalObjectReference { return &core.LocalObjectReference{Name: n} }

	tests := []struct {
		name         string
		issuer       *core.TypedLocalObjectReference
		existing     *MilvusExistingTLSSecrets
		mode         TLSMode
		issuedServer bool
		issuedClient bool
		serverSecret string
		clientSecret string
	}{
		{
			name: "issuer only", issuer: issuer, mode: TLSModeTLS,
			issuedServer: true, issuedClient: true,
			serverSecret: "mv-server-cert", clientSecret: "mv-client-cert",
		},
		{
			name: "their server cert, no mTLS: KubeDB reads the CA from it", mode: TLSModeTLS,
			existing:     &MilvusExistingTLSSecrets{Server: ref("theirs")},
			serverSecret: "theirs", clientSecret: "theirs",
		},
		{
			name: "their server cert, no mTLS, an issuer changes nothing", issuer: issuer, mode: TLSModeTLS,
			existing:     &MilvusExistingTLSSecrets{Server: ref("theirs")},
			serverSecret: "theirs", clientSecret: "theirs",
		},
		{
			name: "their server cert with mTLS: the client cert is issued", issuer: issuer, mode: TLSModeMTLS,
			existing:     &MilvusExistingTLSSecrets{Server: ref("theirs")},
			issuedClient: true,
			serverSecret: "theirs", clientSecret: "mv-client-cert",
		},
		{
			name: "both of theirs", mode: TLSModeMTLS,
			existing:     &MilvusExistingTLSSecrets{Server: ref("theirs"), Client: ref("their-client")},
			serverSecret: "theirs", clientSecret: "their-client",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Milvus{ObjectMeta: metav1.ObjectMeta{Name: "mv"}}
			m.Spec.TLS = &MilvusTLSConfig{
				TLSConfig:       kmapi.TLSConfig{IssuerRef: tt.issuer},
				External:        &ProtocolTLSConfig{Mode: tt.mode},
				ExistingSecrets: tt.existing,
			}
			m.SetTLSDefaults()
			if got := m.CertificateIssued(MilvusCertificateTypeServer); got != tt.issuedServer {
				t.Errorf("server issued = %v, want %v", got, tt.issuedServer)
			}
			if got := m.CertificateIssued(MilvusCertificateTypeClient); got != tt.issuedClient {
				t.Errorf("client issued = %v, want %v", got, tt.issuedClient)
			}
			if got := m.GetCertSecretName(MilvusCertificateTypeServer); got != tt.serverSecret {
				t.Errorf("server secret = %q, want %q", got, tt.serverSecret)
			}
			if got := m.GetCertSecretName(MilvusCertificateTypeClient); got != tt.clientSecret {
				t.Errorf("client secret = %q, want %q", got, tt.clientSecret)
			}
		})
	}
}

func TestMilvusTLSConfigured(t *testing.T) {
	m := &Milvus{}
	if m.TLSConfigured() {
		t.Error("no spec.tls: want false")
	}
	m.Spec.TLS = &MilvusTLSConfig{}
	if m.TLSConfigured() {
		t.Error("empty spec.tls: want false")
	}
	m.Spec.TLS.ExistingSecrets = &MilvusExistingTLSSecrets{Server: &core.LocalObjectReference{Name: "theirs"}}
	if !m.TLSConfigured() {
		t.Error("existing server Secret: want true")
	}
}
