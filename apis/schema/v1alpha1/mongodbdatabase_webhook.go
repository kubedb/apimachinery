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

func (in *MongoDBDatabase) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update,versions=v1alpha1,name=mmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MongoDBDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDBDatabase) Default() {
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbdatabases,verbs=create;update;delete,versions=v1alpha1,name=vmongodbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MongoDBDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBDatabase) ValidateCreate() error {
	mongodbdatabaselog.Info("validate create", "name", in.Name)
	return in.ValidateMongoDBDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBDatabase) ValidateUpdate(old runtime.Object) error {
	mongodbdatabaselog.Info("validate update", "name", in.Name)
	oldDb := old.(*MongoDBDatabase)
	if oldDb.Status.Phase == Success && oldDb.Spec.DatabaseConfig.Name != in.Spec.DatabaseConfig.Name {
		return errors.New("you can't change the Database Schema name now")
	}

	// making VaultRef & DatabaseRef fields immutable
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if oldDb.Spec.DatabaseRef != in.Spec.DatabaseRef {
		allErrs = append(allErrs, field.Invalid(path.Child("databaseRef"), in.Name, `Cannot change mongodb reference`))
	}
	if oldDb.Spec.VaultRef != in.Spec.VaultRef {
		allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), in.Name, `Cannot change vault reference`))
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
	}
	return in.ValidateMongoDBDatabase()
}

const ValidateDeleteMessage = "MongoDBDatabase schema can't be deleted if the deletion policy is DoNotDelete"

//ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBDatabase) ValidateDelete() error {
	mongodbdatabaselog.Info("validate delete", "name", in.Name)
	if in.Spec.DeletionPolicy == DeletionPolicyDoNotDelete {
		return errors.New(ValidateDeleteMessage)
	}
	return nil
}

func (in *MongoDBDatabase) ValidateMongoDBDatabase() error {
	var allErrs field.ErrorList
	if err := in.validateSchemaInitRestore(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.validateMongoDBDatabaseSchemaName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.validateCRDName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := in.CheckIfNameFieldsAreOkOrNot(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
}

func (in *MongoDBDatabase) validateSchemaInitRestore() *field.Error {
	path := field.NewPath("spec")
	if in.Spec.Init != nil && in.Spec.Init.Snapshot != nil {
		return field.Invalid(path, in.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (in *MongoDBDatabase) validateMongoDBDatabaseSchemaName() *field.Error {
	path := field.NewPath("spec").Child("databaseSchema").Child("name")
	name := in.Spec.DatabaseConfig.Name

	if name == MongoDatabaseNameForEntry || name == "admin" || name == "config" || name == "local" {
		str := fmt.Sprintf("cannot use \"%v\" as the database name", name)
		return field.Invalid(path, in.Name, str)
	}
	return nil
}

func (in *MongoDBDatabase) validateCRDName() *field.Error {
	// This is a workaround value. So that after appending some characters after the schema name, it doesn't exceed the standard size
	if len(in.ObjectMeta.Name) > 40 {
		return field.Invalid(field.NewPath("metadata").Child("name"), in.Name, "must be no more than 40 characters")
	}
	return nil
}

/*
Ensure that the name of database, vault & repository are not empty
*/

func (in *MongoDBDatabase) CheckIfNameFieldsAreOkOrNot() *field.Error {
	if in.Spec.DatabaseRef.Name == "" {
		str := fmt.Sprintf("Database Ref name cant be empty")
		return field.Invalid(field.NewPath("spec").Child("databaseRef").Child("name"), in.Name, str)
	}
	if in.Spec.VaultRef.Name == "" {
		str := fmt.Sprintf("Vault Ref name cant be empty")
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), in.Name, str)
	}
	if in.Spec.Init.Snapshot != nil && in.Spec.Init.Snapshot.Repository.Name == "" {
		str := fmt.Sprintf("Repository name cant be empty")
		return field.Invalid(field.NewPath("spec").Child("restore").Child("repository").Child("name"), in.Name, str)
	}
	return nil
}
