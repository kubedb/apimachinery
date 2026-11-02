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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestMetaStorageRootPath(t *testing.T) {
	m := &Milvus{}
	m.Name = "mv"
	if got := m.MetaStorageRootPath(); got != "mv-etcd" {
		t.Fatalf("default root path = %q, want mv-etcd", got)
	}
	m.Spec.MetaStorage = &MetaStorageSpec{RootPath: "by-dev"}
	if got := m.MetaStorageRootPath(); got != "by-dev" {
		t.Fatalf("root path = %q, want by-dev", got)
	}
}

func TestMetaStorageExternalSecrets(t *testing.T) {
	m := &Milvus{}
	m.Spec.MetaStorage = &MetaStorageSpec{
		ClientTLSSecret: &core.LocalObjectReference{Name: "tls"},
	}
	if m.MetaStorageExternalTLSSecretName() != "" {
		t.Fatal("managed etcd must not report an external TLS secret")
	}
	m.Spec.MetaStorage.ExternallyManaged = true
	if got := m.MetaStorageExternalTLSSecretName(); got != "tls" {
		t.Fatalf("tls secret = %q", got)
	}
	if m.MetaStorageExternalAuthSecretName() != "" {
		t.Fatal("no auth secret set")
	}
}

func TestRestrictedSCCSkipsUserDefaults(t *testing.T) {
	version := &catalog.MilvusVersion{}
	version.Spec.SecurityContext.RunAsUser = ptr.To(int64(1000))

	m := &Milvus{}
	rc := &core.SecurityContext{}
	m.AssignDefaultContainerSecurityContext(version, rc)
	if rc.RunAsUser == nil || *rc.RunAsUser != 1000 || rc.RunAsGroup == nil {
		t.Fatalf("default must keep the version's uid, got %+v", rc)
	}

	m.Spec.RestrictedSCC = true
	rc = &core.SecurityContext{}
	m.AssignDefaultContainerSecurityContext(version, rc)
	if rc.RunAsUser != nil || rc.RunAsGroup != nil {
		t.Fatalf("restrictedSCC must leave uid/gid unset, got %+v", rc)
	}
	if rc.RunAsNonRoot == nil || !*rc.RunAsNonRoot || rc.AllowPrivilegeEscalation == nil {
		t.Fatalf("hardening must stay on, got %+v", rc)
	}

	explicit := &core.SecurityContext{RunAsUser: ptr.To(int64(1234))}
	m.AssignDefaultContainerSecurityContext(version, explicit)
	if *explicit.RunAsUser != 1234 {
		t.Fatal("an explicit uid must win")
	}
}

func TestStreamingNodeStorageDefault(t *testing.T) {
	mv := &catalog.MilvusVersion{}
	mv.Spec.SecurityContext.RunAsUser = ptr.To[int64](1000)

	// The role is set (replicas, podTemplate) but storage is not: it must still get the default.
	m := &Milvus{}
	sn := &MilvusDataNode{MilvusNode: MilvusNode{Replicas: ptr.To[int32](2)}}
	m.setComponentDefaults(mv, &sn)
	if sn.StorageType != StorageTypeDurable || sn.Storage == nil {
		t.Fatalf("Durable StreamingNode without storage: got storageType=%q storage=%v", sn.StorageType, sn.Storage)
	}
	if got := sn.Storage.Resources.Requests[core.ResourceStorage]; got.String() != "1Gi" {
		t.Fatalf("default storage = %s, want 1Gi", got.String())
	}

	// A user's storage is kept, and Ephemeral gets none.
	own := &MilvusDataNode{Storage: &core.PersistentVolumeClaimSpec{StorageClassName: ptr.To("fast")}}
	m.setComponentDefaults(mv, &own)
	if own.Storage.StorageClassName == nil || *own.Storage.StorageClassName != "fast" {
		t.Fatal("user storage was replaced")
	}
	eph := &MilvusDataNode{StorageType: StorageTypeEphemeral}
	m.setComponentDefaults(mv, &eph)
	if eph.Storage != nil {
		t.Fatal("Ephemeral StreamingNode got a PVC")
	}
}
