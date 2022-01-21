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
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if in.Spec.Init != nil && in.Spec.Init.Initialized {
		allErrs = append(allErrs, field.Invalid(path.Child("init").Child("initialized"), in.Name, `cannot set the initialized field to true directly`))
		return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
	}
	return in.ValidateMongoDBDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBDatabase) ValidateUpdate(old runtime.Object) error {
	mongodbdatabaselog.Info("validate update", "name", in.Name)
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	oldDb := old.(*MongoDBDatabase)

	// if phase is 'Successful', do not give permission to change the DatabaseConfig.Name
	if oldDb.Status.Phase == DatabaseSchemaPhaseSuccessful && oldDb.Spec.Database.Config.Name != in.Spec.Database.Config.Name {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("config"), in.Name, `you can't change the Database Config name now`))
		return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
	}

	// If Initialized==false, Do not give permission to set it to true directly
	if oldDb.Spec.Init != nil && !oldDb.Spec.Init.Initialized && in.Spec.Init != nil && in.Spec.Init.Initialized {
		allErrs = append(allErrs, field.Invalid(path.Child("init").Child("initialized"), in.Name, `cannot set the initialized field to true directly`))
		return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
	}

	// making VaultRef & DatabaseRef fields immutable
	if oldDb.Spec.Database.ServerRef != in.Spec.Database.ServerRef {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("serverRef"), in.Name, `Cannot change mongodb reference`))
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
		var allErrs field.ErrorList
		path := field.NewPath("spec").Child("deletionPolicy")
		allErrs = append(allErrs, field.Invalid(path, in.Name, ValidateDeleteMessage))
		return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
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
	if err := in.CheckIfNameFieldsAreOkOrNot(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(in.GroupVersionKind().GroupKind(), in.Name, allErrs)
}

func (in *MongoDBDatabase) validateSchemaInitRestore() *field.Error {
	path := field.NewPath("spec").Child("init")
	if in.Spec.Init != nil && in.Spec.Init.Script != nil && in.Spec.Init.Snapshot != nil {
		return field.Invalid(path, in.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (in *MongoDBDatabase) validateMongoDBDatabaseSchemaName() *field.Error {
	path := field.NewPath("spec").Child("database").Child("config").Child("name")
	name := in.Spec.Database.Config.Name

	if name == MongoDatabaseNameForEntry || name == "admin" || name == "config" || name == "local" {
		str := fmt.Sprintf("cannot use \"%v\" as the database name", name)
		return field.Invalid(path, in.Name, str)
	}
	return nil
}

/*
Ensure that the name of database, vault & repository are not empty
*/

func (in *MongoDBDatabase) CheckIfNameFieldsAreOkOrNot() *field.Error {
	if in.Spec.Database.ServerRef.Name == "" {
		str := "Database Ref name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("database").Child("serverRef").Child("name"), in.Name, str)
	}
	if in.Spec.VaultRef.Name == "" {
		str := "Vault Ref name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), in.Name, str)
	}
	if in.Spec.Init != nil && in.Spec.Init.Snapshot != nil && in.Spec.Init.Snapshot.Repository.Name == "" {
		str := "Repository name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("init").Child("snapshot").Child("repository").Child("name"), in.Name, str)
	}
	return nil
}
