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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mongodbdatabaselog = logf.Log.WithName("mongodbdatabase-resource")

// SetupMongoDBDatabaseWebhookWithManager registers the webhook for Cassandra in the manager.
func SetupMongoDBDatabaseWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&smapi.MongoDBDatabase{}).
		WithValidator(&MongoDatabaseCustomWebhook{}).
		WithDefaulter(&MongoDatabaseCustomWebhook{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update,versions=v1alpha1,name=mmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type MongoDatabaseCustomWebhook struct{}

var _ webhook.CustomDefaulter = &MongoDatabaseCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDatabaseCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update;delete,versions=v1alpha1,name=vmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MongoDatabaseCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDatabaseCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.MongoDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBDatabase, but got a %T", obj)
	}
	mongodbdatabaselog.Info("validate create", "name", db.Name)
	return nil, in.ValidateMongoDBDatabase(db)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MongoDatabaseCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	newDB, ok := newObj.(*smapi.MongoDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBDatabase, but got a %T", newObj)
	}
	mongodbdatabaselog.Info("validate update", "name", newDB.Name)
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	oldDb := old.(*smapi.MongoDBDatabase)

	// if phase is 'Current', do not give permission to change the DatabaseConfig.Name
	if oldDb.Status.Phase == smapi.DatabaseSchemaPhaseCurrent && oldDb.Spec.Database.Config.Name != newDB.Spec.Database.Config.Name {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("config"), newDB.Name, MongoDBValidateDatabaseNameChangeError))
		return nil, apierrors.NewInvalid(newDB.GroupVersionKind().GroupKind(), newDB.Name, allErrs)
	}

	// If Initialized==true, Do not give permission to unset it
	if oldDb.Spec.Init != nil && oldDb.Spec.Init.Initialized { // initialized is already set in old object
		// If user updated the Schema-yaml with no Spec.Init
		// Or
		// user updated the Schema-yaml with Spec.Init.Initialized = false
		if newDB.Spec.Init == nil || (newDB.Spec.Init != nil && !newDB.Spec.Init.Initialized) {
			allErrs = append(allErrs, field.Invalid(path.Child("init").Child("initialized"), newDB.Name, MongoDBValidateInitializedUnsetError))
			return nil, apierrors.NewInvalid(newDB.GroupVersionKind().GroupKind(), newDB.Name, allErrs)
		}
	}

	// making VaultRef & DatabaseRef fields immutable
	if oldDb.Spec.Database.ServerRef != newDB.Spec.Database.ServerRef {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("serverRef"), newDB.Name, MongoDBValidateDBServerRefChangeError))
	}
	if oldDb.Spec.VaultRef != newDB.Spec.VaultRef {
		allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), newDB.Name, MongoDBValidateVaultRefChangeError))
	}
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(newDB.GroupVersionKind().GroupKind(), newDB.Name, allErrs)
	}
	return nil, w.ValidateMongoDBDatabase(newDB)
}

const (
	// these constants are here for purpose of Testing of MongoDB Operator code

	MongoDBValidateDeletionPolicyError     = "schema can't be deleted if the deletion policy is DoNotDelete"
	MongoDBValidateInitTypeBothError       = "cannot initialize database using both restore and initSpec"
	MongoDBValidateInitializedUnsetError   = "cannot unset the initialized field directly"
	MongoDBValidateDatabaseNameChangeError = "you can't change the Database Config name now"
	MongoDBValidateDBServerRefChangeError  = "cannot change mongodb reference"
	MongoDBValidateVaultRefChangeError     = "cannot change vault reference"
)

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDatabaseCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.MongoDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MongoDBDatabase, but got a %T", obj)
	}
	mongodbdatabaselog.Info("validate delete", "name", db.Name)
	if db.Spec.DeletionPolicy == smapi.DeletionPolicyDoNotDelete {
		var allErrs field.ErrorList
		path := field.NewPath("spec").Child("deletionPolicy")
		allErrs = append(allErrs, field.Invalid(path, db.Name, MongoDBValidateDeletionPolicyError))
		return nil, apierrors.NewInvalid(db.GroupVersionKind().GroupKind(), db.Name, allErrs)
	}
	return nil, nil
}

func (in *MongoDatabaseCustomWebhook) ValidateMongoDBDatabase(db *smapi.MongoDBDatabase) error {
	var allErrs field.ErrorList
	if err := in.validateSchemaInitRestore(db); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.validateMongoDBDatabaseSchemaName(db); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.CheckIfNameFieldsAreOkOrNot(db); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(db.GroupVersionKind().GroupKind(), db.Name, allErrs)
}

func (in *MongoDatabaseCustomWebhook) validateSchemaInitRestore(db *smapi.MongoDBDatabase) *field.Error {
	path := field.NewPath("spec").Child("init")
	if db.Spec.Init != nil && db.Spec.Init.Script != nil && db.Spec.Init.Snapshot != nil {
		return field.Invalid(path, db.Name, MongoDBValidateInitTypeBothError)
	}
	return nil
}

func (in *MongoDatabaseCustomWebhook) validateMongoDBDatabaseSchemaName(db *smapi.MongoDBDatabase) *field.Error {
	path := field.NewPath("spec").Child("database").Child("config").Child("name")
	name := db.Spec.Database.Config.Name

	if name == smapi.MongoDatabaseNameForEntry || name == smapi.DatabaseNameAdmin || name == smapi.DatabaseNameConfig || name == smapi.DatabaseNameLocal {
		str := fmt.Sprintf("cannot use \"%v\" as the database name", name)
		return field.Invalid(path, db.Name, str)
	}
	return nil
}

/*
Ensure that the name of database, vault & repository are not empty
*/

func (in *MongoDatabaseCustomWebhook) CheckIfNameFieldsAreOkOrNot(db *smapi.MongoDBDatabase) *field.Error {
	if db.Spec.Database.ServerRef.Name == "" {
		str := "Database Ref name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("database").Child("serverRef").Child("name"), db.Name, str)
	}
	if db.Spec.VaultRef.Name == "" {
		str := "Vault Ref name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), db.Name, str)
	}
	if db.Spec.Init != nil && db.Spec.Init.Snapshot != nil && db.Spec.Init.Snapshot.Repository.Name == "" {
		str := "Repository name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("init").Child("snapshot").Child("repository").Child("name"), db.Name, str)
	}
	return nil
}
