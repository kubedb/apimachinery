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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// versionClient serves one OracleVersion; the validation reads nothing else
// for an Ephemeral Oracle.
type versionClient struct {
	client.Client
	version catalog.OracleVersion
}

func (c versionClient) Get(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	*obj.(*catalog.OracleVersion) = c.version
	return nil
}

func oracleWithInitArchiver(mutate func(*olddbapi.ArchiverRecovery)) *olddbapi.Oracle {
	a := &olddbapi.ArchiverRecovery{
		RecoveryTimestamp: metav1.NewTime(time.Date(2026, 11, 11, 8, 35, 30, 0, time.UTC)),
		FullDBRepository:  &kmapi.ObjectReference{Name: "sa-oracle-full"},
		EncryptionSecret:  &kmapi.ObjectReference{Name: "encrypt-secret"},
	}
	if mutate != nil {
		mutate(a)
	}
	return &olddbapi.Oracle{Spec: olddbapi.OracleSpec{
		Mode: olddbapi.OracleModeStandalone,
		Init: &olddbapi.InitSpec{Archiver: a},
	}}
}

func TestOracleValidateArchiver(t *testing.T) {
	if errs := oracleValidateArchiver(&olddbapi.Oracle{}); len(errs) != 0 {
		t.Errorf("no init.archiver: %v", errs)
	}
	if errs := oracleValidateArchiver(oracleWithInitArchiver(nil)); len(errs) != 0 {
		t.Errorf("complete init.archiver: %v", errs)
	}

	for name, tc := range map[string]struct {
		db   *olddbapi.Oracle
		path string
	}{
		"no timestamp":  {oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) { a.RecoveryTimestamp = metav1.Time{} }), "spec.init.archiver.recoveryTimestamp"},
		"no repository": {oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) { a.FullDBRepository = nil }), "spec.init.archiver.fullDBRepository"},
		"no encryption": {oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) { a.EncryptionSecret = nil }), "spec.init.archiver.encryptionSecret"},
		"into DataGuard": {func() *olddbapi.Oracle {
			db := oracleWithInitArchiver(nil)
			db.Spec.Mode = olddbapi.OracleModeDataGuard
			return db
		}(), "spec.init.archiver"},
	} {
		errs := oracleValidateArchiver(tc.db)
		if len(errs) != 1 || errs[0].Field != tc.path {
			t.Errorf("%s: want one error on %s, got %v", name, tc.path, errs)
		}
	}
}

// Once restored, the recovery request is history: changing it would claim a
// restore that never happened.
func TestOracleValidateArchiverImmutability(t *testing.T) {
	old := oracleWithInitArchiver(nil)
	if errs := oracleValidateArchiverImmutability(old, oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) {
		a.RecoveryTimestamp = metav1.NewTime(time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC))
	})); len(errs) != 0 {
		t.Errorf("before initialization it may change: %v", errs)
	}

	old.Spec.Init.Initialized = true
	changed := oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) {
		a.RecoveryTimestamp = metav1.NewTime(time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC))
	})
	changed.Spec.Init.Initialized = true
	errs := oracleValidateArchiverImmutability(old, changed)
	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "cannot be changed") {
		t.Errorf("after initialization it must not change: %v", errs)
	}

	same := oracleWithInitArchiver(nil)
	same.Spec.Init.Initialized = true
	if errs := oracleValidateArchiverImmutability(old, same); len(errs) != 0 {
		t.Errorf("unchanged: %v", errs)
	}
}

// The webhook runs the archiver checks on an Oracle that is otherwise valid,
// not only next to other errors.
func TestOracleValidateCreateOrUpdateRunsArchiverChecks(t *testing.T) {
	w := &OracleCustomWebhook{DefaultClient: versionClient{version: catalog.OracleVersion{
		ObjectMeta: metav1.ObjectMeta{Name: "21.3.0"},
		Spec: catalog.OracleVersionSpec{
			Version:       "21.3.0",
			DB:            catalog.OracleVersionDatabase{Image: "oracle"},
			Exporter:      catalog.OracleVersionExporter{Image: "exporter"},
			InitContainer: catalog.OracleVersionInitContainer{Image: "init"},
			Coordinator:   catalog.OracleVersionCoordinator{Image: "coordinator"},
		},
	}}}

	db := oracleWithInitArchiver(func(a *olddbapi.ArchiverRecovery) { a.EncryptionSecret = nil })
	db.Name = "restored"
	db.Spec.Version = "21.3.0"
	db.Spec.Edition = "enterprise"
	db.Spec.Replicas = ptr.To[int32](1)
	db.Spec.StorageType = olddbapi.StorageTypeEphemeral

	errs := w.ValidateCreateOrUpdate(db)
	if len(errs) != 1 || errs[0].Field != "spec.init.archiver.encryptionSecret" {
		t.Errorf("want only the missing encryptionSecret, got %v", errs)
	}

	db.Spec.Init.Archiver.EncryptionSecret = &kmapi.ObjectReference{Name: "encrypt-secret"}
	if errs := w.ValidateCreateOrUpdate(db); len(errs) != 0 {
		t.Errorf("complete Oracle rejected: %v", errs)
	}
}
