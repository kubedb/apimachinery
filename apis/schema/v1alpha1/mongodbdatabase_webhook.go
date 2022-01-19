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
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mongodbdatabaselog = logf.Log.WithName("mongodbdatabase-resource")

func (db *MongoDBDatabase) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(db).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update,versions=v1alpha1,name=mmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MongoDBDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (db *MongoDBDatabase) Default() {
	mongodbdatabaselog.Info("default", "name", db.Name)

	if db.Spec.Restore != nil {
		if db.Spec.Restore.Snapshot == "" {
			db.Spec.Restore.Snapshot = "latest"
		}
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update;delete,versions=v1alpha1,name=vmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MongoDBDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (db *MongoDBDatabase) ValidateCreate() error {
	mongodbdatabaselog.Info("validate create", "name", db.Name)
	return db.ValidateMongoDBDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (db *MongoDBDatabase) ValidateUpdate(old runtime.Object) error {
	mongodbdatabaselog.Info("validate update", "name", db.Name)
	oldDb := old.(*MongoDBDatabase)
	if oldDb.Status.Phase == SchemaDatabasePhaseSuccessfull && oldDb.Spec.DatabaseConfig.Name != db.Spec.DatabaseConfig.Name {
		return errors.New("you can't change the Database Schema name now")
	}

	// making VaultRef & DatabaseRef fields immutable
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if oldDb.Spec.DatabaseRef != db.Spec.DatabaseRef {
		allErrs = append(allErrs, field.Invalid(path.Child("databaseRef"), db.Name, `Cannot change mongodb reference`))
	}
	if oldDb.Spec.VaultRef != db.Spec.VaultRef {
		allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), db.Name, `Cannot change vault reference`))
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(db.GroupVersionKind().GroupKind(), db.Name, allErrs)
	}
	return db.ValidateMongoDBDatabase()
}

const ValidateDeleteMessage = "MongoDBDatabase schema can't be deleted if the deletion policy is DoNotDelete"

//ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (db *MongoDBDatabase) ValidateDelete() error {
	mongodbdatabaselog.Info("validate delete", "name", db.Name)
	if db.Spec.DeletionPolicy == DeletionPolicyDoNotDelete {
		return errors.New(ValidateDeleteMessage)
	}
	return nil
}

func (db *MongoDBDatabase) ValidateMongoDBDatabase() error {
	var allErrs field.ErrorList
	if err := db.validateSchemaInitRestore(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := db.validateMongoDBDatabaseSchemaName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := db.validateCRDName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := db.CheckIfNameFieldsAreOkOrNot(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(db.GroupVersionKind().GroupKind(), db.Name, allErrs)
}

func (db *MongoDBDatabase) validateSchemaInitRestore() *field.Error {
	path := field.NewPath("spec")
	if db.Spec.Init != nil && db.Spec.Restore != nil {
		return field.Invalid(path, db.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (db *MongoDBDatabase) validateMongoDBDatabaseSchemaName() *field.Error {
	path := field.NewPath("spec").Child("databaseSchema").Child("name")
	name := db.Spec.DatabaseConfig.Name

	if name == MongoDatabaseNameForEntry || name == "admin" || name == "config" || name == "local" {
		str := fmt.Sprintf("cannot use \"%v\" as the database name", name)
		return field.Invalid(path, db.Name, str)
	}
	return nil
}

func (db *MongoDBDatabase) validateCRDName() *field.Error {
	// This is a workaround value. So that after appending some characters after the schema name, it doesn't exceed the standard size
	if len(db.ObjectMeta.Name) > 40 {
		return field.Invalid(field.NewPath("metadata").Child("name"), db.Name, "must be no more than 40 characters")
	}
	return nil
}

/*
Ensure that the name of database, vault & repository are not empty
*/

func (db *MongoDBDatabase) CheckIfNameFieldsAreOkOrNot() *field.Error {
	if db.Spec.DatabaseRef.Name == "" {
		str := fmt.Sprintf("Database Ref name cant be empty")
		return field.Invalid(field.NewPath("spec").Child("databaseRef").Child("name"), db.Name, str)
	}
	if db.Spec.VaultRef.Name == "" {
		str := fmt.Sprintf("Vault Ref name cant be empty")
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), db.Name, str)
	}
	if db.Spec.Restore != nil {
		if db.Spec.Restore.Repository.Name == "" {
			str := fmt.Sprintf("Repository name cant be empty")
			return field.Invalid(field.NewPath("spec").Child("restore").Child("repository").Child("name"), db.Name, str)
		}
	}
	return nil
}
