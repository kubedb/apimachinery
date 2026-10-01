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
	"strings"
	"testing"
	"time"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

func milvusFor(mode olddbapi.MilvusMode) *olddbapi.Milvus {
	m := &olddbapi.Milvus{Spec: olddbapi.MilvusSpec{Version: "2.6.11", Topology: &olddbapi.MilvusTopology{Mode: &mode}}}
	return m
}

func joined(m *olddbapi.Milvus) string {
	var parts []string
	for _, e := range milvusValidateArchiver(m) {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, "; ")
}

func TestMilvusArchiverValidation(t *testing.T) {
	// a Distributed Milvus cannot pick RocksMQ
	d := milvusFor("Distributed")
	d.Spec.WAL = &olddbapi.MilvusWALSpec{Type: olddbapi.MilvusWALRocksMQ}
	if got := joined(d); !strings.Contains(got, "always uses the Woodpecker WAL") {
		t.Errorf("distributed+RocksMQ not rejected: %q", got)
	}

	// archiver with an externally managed etcd
	s := milvusFor("Standalone")
	s.Spec.Archiver = &olddbapi.Archiver{Ref: kmapi.ObjectReference{Name: "a"}}
	s.Spec.MetaStorage = &olddbapi.MetaStorageSpec{ExternallyManaged: true}
	if got := joined(s); !strings.Contains(got, "dedicated to this Milvus") {
		t.Errorf("external etcd with archiver not rejected: %q", got)
	}

	// Milvus 3.x is not validated
	s3 := milvusFor("Standalone")
	s3.Spec.Version = "3.0.0"
	s3.Spec.Archiver = &olddbapi.Archiver{Ref: kmapi.ObjectReference{Name: "a"}}
	if got := joined(s3); !strings.Contains(got, "3.x") {
		t.Errorf("3.x not rejected: %q", got)
	}

	// a restore needs time, repository and encryption secret
	r := milvusFor("Standalone")
	r.Spec.Init = &olddbapi.InitSpec{Archiver: &olddbapi.ArchiverRecovery{}}
	got := joined(r)
	for _, want := range []string{"recoveryTimestamp", "fullDBRepository", "encryptionSecret"} {
		if !strings.Contains(got, want) {
			t.Errorf("restore without %s not rejected: %q", want, got)
		}
	}
	r.Spec.Init.Archiver = &olddbapi.ArchiverRecovery{
		RecoveryTimestamp: metav1.NewTime(time.Now()),
		FullDBRepository:  &kmapi.ObjectReference{Name: "repo"},
		EncryptionSecret:  &kmapi.ObjectReference{Name: "enc"},
	}
	if got := joined(r); got != "" {
		t.Errorf("valid restore rejected: %q", got)
	}
}

func TestMilvusArchiverImmutability(t *testing.T) {
	oldDB := milvusFor("Standalone")
	oldDB.Spec.WAL = &olddbapi.MilvusWALSpec{Type: olddbapi.MilvusWALWoodpecker}
	newDB := oldDB.DeepCopy()
	if errs := milvusValidateArchiverImmutability(oldDB, newDB); len(errs) != 0 {
		t.Errorf("unchanged object rejected: %v", errs)
	}
	newDB.Spec.WAL = &olddbapi.MilvusWALSpec{Type: olddbapi.MilvusWALRocksMQ}
	if errs := milvusValidateArchiverImmutability(oldDB, newDB); len(errs) == 0 {
		t.Error("changing the WAL type must be rejected")
	}

	initDB := milvusFor("Standalone")
	initDB.Spec.Init = &olddbapi.InitSpec{Initialized: true, Archiver: &olddbapi.ArchiverRecovery{RecoveryTimestamp: metav1.NewTime(time.Unix(1, 0))}}
	changed := initDB.DeepCopy()
	changed.Spec.Init.Archiver.RecoveryTimestamp = metav1.NewTime(time.Unix(2, 0))
	if errs := milvusValidateArchiverImmutability(initDB, changed); len(errs) == 0 {
		t.Error("changing init.archiver after initialization must be rejected")
	}
}
