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
	"fmt"

	smapi "kubedb.dev/apimachinery/apis/schema/v1alpha1"

	gocmp "github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mysqldatabaselog = logf.Log.WithName("mysqldatabase-resource")

// SetupMySQLSchemaWebhookWithManager registers the webhook for Cassandra in the manager.
func SetupMySQLSchemaWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&smapi.MySQLDatabase{}).
		WithValidator(&MySQLDatabaseCustomWebhook{}).
		WithDefaulter(&MySQLDatabaseCustomWebhook{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update,versions=v1alpha1,name=mmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false

type MySQLDatabaseCustomWebhook struct{}

var _ webhook.CustomDefaulter = &MySQLDatabaseCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MySQLDatabaseCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db, ok := obj.(*smapi.MySQLDatabase)
	if !ok {
		return fmt.Errorf("expected mysqldatabase but got %T", obj)
	}
	mysqldatabaselog.Info("default", "name", db.Name)

	if db.Spec.Init != nil {
		if db.Spec.Init.Snapshot != nil {
			if db.Spec.Init.Snapshot.SnapshotID == "" {
				db.Spec.Init.Snapshot.SnapshotID = "latest"
			}
		}
	}
	val := db.Spec.Database.Config.Encryption
	if val == "enable" || val == smapi.MySQLEncryptionEnabled {
		db.Spec.Database.Config.Encryption = smapi.MySQLEncryptionEnabled
	} else {
		db.Spec.Database.Config.Encryption = smapi.MySQLEncryptionDisabled
	}
	if db.Spec.Database.Config.ReadOnly != 1 {
		db.Spec.Database.Config.ReadOnly = 0
	}
	if db.Spec.Database.Config.CharacterSet == "" {
		db.Spec.Database.Config.CharacterSet = "utf8"
	}
	return nil
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update;delete,versions=v1alpha1,name=vmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MySQLDatabaseCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MySQLDatabaseCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.MySQLDatabase)
	if !ok {
		return nil, fmt.Errorf("expected mysqldatabase but got %T", obj)
	}
	mysqldatabaselog.Info("validate create", "name", db.Name)
	var allErrs field.ErrorList
	//if db.Spec.Database.Config.ReadOnly == 1 { //todo handle this case if possible
	//	allErrs = append(allErrs, field.Invalid(field.NewPath("spec.database.config"), db.Name, "Cannot create readOnly database"))
	//}
	if err := in.ValidateMySQLDatabase(db); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath(""), db.Name, err.Error()))
	}
	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, db.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MySQLDatabaseCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	db, ok := newObj.(*smapi.MySQLDatabase)
	if !ok {
		return nil, fmt.Errorf("expected mysqldatabase but got %T", newObj)
	}
	mysqldatabaselog.Info("validate update", "name", db.Name)
	oldobj := old.(*smapi.MySQLDatabase)
	return nil, in.ValidateMySQLDatabaseUpdate(db, oldobj)
}

func (in *MySQLDatabaseCustomWebhook) ValidateMySQLDatabaseUpdate(newobj *smapi.MySQLDatabase, oldobj *smapi.MySQLDatabase) error {
	if newobj.Finalizers == nil {
		return nil
	}
	var allErrs field.ErrorList
	if !gocmp.Equal(oldobj.Spec.Database.ServerRef, newobj.Spec.Database.ServerRef) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.database.serverRef"), newobj.Name, "cannot change database serverRef"))
	}
	if !gocmp.Equal(oldobj.Spec.Database.Config.Name, newobj.Spec.Database.Config.Name) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.database.config.name"), newobj.Name, "cannot change database name configuration"))
	}
	if !gocmp.Equal(oldobj.Spec.VaultRef, newobj.Spec.VaultRef) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.vaultRef"), newobj.Name, "cannot change vaultRef"))
	}
	if !gocmp.Equal(oldobj.Spec.AccessPolicy, newobj.Spec.AccessPolicy) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.accessPolicy"), newobj.Name, "cannot change accessPolicy"))
	}
	if newobj.Spec.Init != nil {
		if oldobj.Spec.Init == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.init"), newobj.Name, "cannot change init"))
		}
	}
	if oldobj.Spec.Init != nil {
		if newobj.Spec.Init == nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.init"), newobj.Name, "cannot change init"))
		}
	}
	if newobj.Spec.Init != nil && oldobj.Spec.Init != nil {
		if !gocmp.Equal(newobj.Spec.Init.Script, oldobj.Spec.Init.Script) || !gocmp.Equal(newobj.Spec.Init.Snapshot, oldobj.Spec.Init.Snapshot) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.init"), newobj.Name, "cannot change init"))
		}
	}
	er := in.ValidateMySQLDatabase(newobj)
	if er != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), newobj.Name, er.Error()))
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, newobj.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *MySQLDatabaseCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.MySQLDatabase)
	if !ok {
		return nil, fmt.Errorf("expected mysqldatabase but got %T", obj)
	}
	mysqldatabaselog.Info("validate delete", "name", db.Name)
	if db.Spec.DeletionPolicy == smapi.DeletionPolicyDoNotDelete {
		return nil, field.Invalid(field.NewPath("spec").Child("terminationPolicy"), db.Name, `cannot delete object when terminationPolicy is set to "DoNotDelete"`)
	}
	if db.Spec.Database.Config.ReadOnly == 1 {
		return nil, field.Invalid(field.NewPath("spec").Child("databaseConfig.readOnly"), db.Name, `schema manger cannot be deleted : database is read only enabled`)
	}
	return nil, nil
}

func (in *MySQLDatabaseCustomWebhook) ValidateMySQLDatabase(db *smapi.MySQLDatabase) error {
	var allErrs field.ErrorList
	if err := in.validateInitailizationSchema(db); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.validateMySQLDatabaseConfig(db); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, db.Name, allErrs)
}

func (in *MySQLDatabaseCustomWebhook) validateInitailizationSchema(db *smapi.MySQLDatabase) *field.Error {
	path := field.NewPath("spec.init")
	if db.Spec.Init != nil {
		if db.Spec.Init.Script != nil && db.Spec.Init.Snapshot != nil {
			return field.Invalid(path, db.Name, `cannot initialize database using both restore and initSpec`)
		}
	}
	return nil
}

func (in *MySQLDatabaseCustomWebhook) validateMySQLDatabaseConfig(db *smapi.MySQLDatabase) *field.Error {
	path := field.NewPath("spec").Child("database.config").Child("name")
	name := db.Spec.Database.Config.Name
	if name == "sys" {
		return field.Invalid(path, db.Name, `cannot use "sys" as the database name`)
	}
	if name == "performance_schema" {
		return field.Invalid(path, db.Name, `cannot use "performance_schema" as the database name`)
	}
	if name == "mysql" {
		return field.Invalid(path, db.Name, `cannot use "mysql" as the database name`)
	}
	if name == smapi.DatabaseForEntry {
		return field.Invalid(path, db.Name, `cannot use "kubedb_system" as the database name`)
	}
	if name == "information_schema" {
		return field.Invalid(path, db.Name, `cannot use "information_schema" as the database name`)
	}
	if name == smapi.DatabaseNameAdmin {
		return field.Invalid(path, db.Name, `cannot use "admin" as the database name`)
	}
	if name == smapi.DatabaseNameConfig {
		return field.Invalid(path, db.Name, `cannot use "config" as the database name`)
	}
	path = field.NewPath("spec").Child("database.config")
	val := db.Spec.Database.Config.ReadOnly
	if val == 1 {
		if db.Spec.Init != nil {
			if (db.Spec.Init.Script != nil || db.Spec.Init.Snapshot != nil) && db.Status.Phase != smapi.DatabaseSchemaPhaseCurrent {
				return field.Invalid(path.Child("readOnly"), db.Name, `cannot make the database readonly , init/restore yet to be applied`)
			}
		}
	} else if db.Spec.Database.Config.Encryption == smapi.MySQLEncryptionEnabled {
		if db.Spec.Init != nil {
			if (db.Spec.Init.Script != nil || db.Spec.Init.Snapshot != nil) && db.Status.Phase != smapi.DatabaseSchemaPhaseCurrent {
				return field.Invalid(path.Child("encryption"), db.Name, `cannot make the database encryption enables , init/restore yet to be applied`)
			}
		}
	}
	return nil
}
