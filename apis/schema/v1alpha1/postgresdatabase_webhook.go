/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	gocmp "github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	config = "config"
	admin  = "admin"
	local  = "local"
)

// log is for logging in this package.
var postgresdatabaselog = logf.Log.WithName("postgresdatabase-resource")

func (r *PostgresDatabase) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-postgresdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=postgresdatabases,verbs=create;update,versions=v1alpha1,name=mpostgresdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &PostgresDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PostgresDatabase) Default() {
	postgresdatabaselog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-postgresdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=postgresdatabases,verbs=create;update;delete,versions=v1alpha1,name=vpostgresdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &PostgresDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabase) ValidateCreate() error {
	postgresdatabaselog.Info("validate create", "name", r.Name)
	if r.Spec.Init != nil && r.Spec.Init.Initialized {
		return field.Invalid(field.NewPath("spec").Child("init").Child("initialized"), r.Name, `cannot initialized true while creating the object`)
	}
	return r.ValidatePostgresDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabase) ValidateUpdate(old runtime.Object) error {
	postgresdatabaselog.Info("validate update", "name", r.Name)
	oldobj := old.(*PostgresDatabase)
	return r.ValidatePostgresDatabaseUpdate(oldobj, r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabase) ValidateDelete() error {
	postgresdatabaselog.Info("validate delete", "name", r.Name)
	if r.Spec.DeletionPolicy == DeletionPolicyDoNotDelete {
		return field.Invalid(field.NewPath("spec").Child("deletionPolicy"), r.Name, `cannot delete object when deletionPolicy is "DoNotDelete"`)
	}
	return nil
}

func (r *PostgresDatabase) ValidateReadOnly() *field.Error {
	if r.Spec.Database.Config.Params == nil {
		return nil
	}
	for _, param := range r.Spec.Database.Config.Params {
		if param.ConfigParameter == "default_transaction_read_only" && *param.Value == "on" && r.Spec.Init != nil {
			msg := "cannot initilize a read only database"
			return field.Invalid(field.NewPath("spec").Child("database").Child("config").Child("params"), r.Name, msg)
		}
	}
	return nil
}

func (r *PostgresDatabase) ValidateUpdateReadOnly() *field.Error {
	if r.Spec.Database.Config.Params == nil {
		return nil
	}
	for _, param := range r.Spec.Database.Config.Params {
		if param.ConfigParameter == "default_transaction_read_only" && *param.Value == "on" && r.Spec.Init != nil && !r.Spec.Init.Initialized {
			msg := "cannot initilize a read only database"
			return field.Invalid(field.NewPath("spec").Child("database").Child("config").Child("params"), r.Name, msg)
		}
	}
	return nil
}

func (r *PostgresDatabase) ValidatePostgresDatabaseUpdate(oldobj *PostgresDatabase, newobj *PostgresDatabase) error {
	if newobj.Finalizers == nil {
		return nil
	}
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if !gocmp.Equal(oldobj.Spec.Database.Config.Name, newobj.Spec.Database.Config.Name) {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("config").Child("name"), newobj.Name, `Cannot change target database name`))
	}
	if !gocmp.Equal(oldobj.Spec.Database.ServerRef, newobj.Spec.Database.ServerRef) {
		allErrs = append(allErrs, field.Invalid(path.Child("database").Child("serverRef"), newobj.Name, `Cannot change postgres reference`))
	}
	if !gocmp.Equal(oldobj.Spec.VaultRef, newobj.Spec.VaultRef) {
		allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), newobj.Name, `Cannot change vault reference`))
	}
	if err := newobj.ValidateUpdatePostgresDatabase(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath(""), newobj.Name, err.Error()))
	}
	if oldobj.Spec.Init != nil && oldobj.Spec.Init.Initialized && !gocmp.Equal(oldobj.Spec.Init, newobj.Spec.Init) {
		allErrs = append(allErrs, field.Invalid(path.Child("init"), newobj.Name, "cannot change init, former init already applied"))
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(r.GroupVersionKind().GroupKind(), newobj.Name, allErrs)
	}
	return nil
}

func (r *PostgresDatabase) ValidatePostgresDBName() *field.Error {
	path := field.NewPath("spec").Child("database").Child("config").Child("name")
	name := r.Spec.Database.Config.Name
	if name == PostgresSchemaKubeSystem || name == admin || name == config || name == local || name == "postgres" || name == "sys" || name == "template0" || name == "template1" {
		str := fmt.Sprintf("cannot use \"%v\" as the database name", name)
		return field.Invalid(path, r.Name, str)
	}
	return nil
}

func (r *PostgresDatabase) ValidateSchemaInitRestore() *field.Error {
	path := field.NewPath("spec").Child("init")
	if r.Spec.Init != nil && r.Spec.Init.Snapshot != nil && r.Spec.Init.Script != nil {
		return field.Invalid(path, r.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (r *PostgresDatabase) ValidateCRDName() *field.Error {
	if len(r.ObjectMeta.Name) > 40 {
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 40 characters")
	}
	return nil
}

func (r *PostgresDatabase) ValidateParams() *field.Error {
	if r.Spec.Database.Config.Params == nil {
		return nil
	}
	for _, param := range r.Spec.Database.Config.Params {
		if param.ConfigParameter == "" || param.Value == nil {
			msg := "cannot use empty parameter and value"
			return field.Invalid(field.NewPath("spec").Child("database").Child("config").Child("params"), r.Name, msg)
		}
	}
	return nil
}

func (r *PostgresDatabase) ValidateFields() *field.Error {
	if r.Spec.Database.ServerRef.Name == "" {
		str := "Database ServerRef name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("database").Child("serverRef").Child("name"), r.Name, str)
	}
	if r.Spec.VaultRef.Name == "" {
		str := "VaultRef name cant be empty"
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), r.Name, str)
	}
	if r.Spec.Init != nil && r.Spec.Init.Snapshot != nil {
		if r.Spec.Init.Snapshot.Repository.Name == "" {
			str := "Repository name cant be empty"
			return field.Invalid(field.NewPath("spec").Child("init").Child("snapshot").Child("repository").Child("name"), r.Name, str)
		}
	}
	if r.Spec.AccessPolicy.Subjects == nil {
		str := "AccessPolicy subjects can't be empty"
		return field.Invalid(field.NewPath("spec").Child("accessPolicy").Child("subjects"), r.Name, str)
	}
	return nil
}

func (r *PostgresDatabase) ValidatePostgresDatabaseNamespace() *field.Error {
	path := field.NewPath("metadata").Child("namespace")
	ns := r.ObjectMeta.Namespace
	if ns == "cert-manager" {
		return field.Invalid(path, r.Name, `cannot use namespace "cert-manager" to create schema manager`)
	}
	if ns == "kube-system" {
		return field.Invalid(path, r.Name, `cannot use namespace "kube-system" to create schema manager`)
	}
	if ns == "kubedb-system" {
		return field.Invalid(path, r.Name, `cannot use namespace "kubedb-system" to create schema manager`)
	}
	if ns == "kubedb" {
		return field.Invalid(path, r.Name, `cannot use namespace "kubedb" to create schema manager`)
	}
	if ns == "kubevault" {
		return field.Invalid(path, r.Name, `cannot use namespace "kubevault" to create schema manager`)
	}
	if ns == "local-path-storage" {
		return field.Invalid(path, r.Name, `cannot use namespace "local-path-storage" to create schema manager`)
	}
	return nil
}

func (r *PostgresDatabase) ValidatePostgresDatabase() error {
	var allErrs field.ErrorList
	// check if Init and Restore both are present
	if err := r.ValidateSchemaInitRestore(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the database name is conflicted with some constant name
	if err := r.ValidatePostgresDBName(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check if crd name length
	if err := r.ValidateCRDName(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the spec fields
	if err := r.ValidateFields(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the schema namespace
	if err := r.ValidatePostgresDatabaseNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.ValidateReadOnly(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.ValidateParams(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
	}
	return nil
}

func (r *PostgresDatabase) ValidateUpdatePostgresDatabase() error {
	var allErrs field.ErrorList
	// check if Init and Restore both are present
	if err := r.ValidateSchemaInitRestore(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the database name is conflicted with some constant name
	if err := r.ValidatePostgresDBName(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check if crd name length
	if err := r.ValidateCRDName(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the spec fields
	if err := r.ValidateFields(); err != nil {
		allErrs = append(allErrs, err)
	}
	// check the schema namespace
	if err := r.ValidatePostgresDatabaseNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.ValidateUpdateReadOnly(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.ValidateParams(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
	}
	return nil
}
