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
var mariadbdatabaselog = logf.Log.WithName("mariadbdatabase-resource")

// SetupMariaDBDatabaseWebhookWithManager registers the webhook for SchemaManager in the manager.
func SetupMariaDBDatabaseWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&smapi.MariaDBDatabase{}).
		WithValidator(&MariaDBDatabaseCustomWebhook{}).
		WithDefaulter(&MariaDBDatabaseCustomWebhook{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mariadbdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mariadbdatabases,verbs=create;update,versions=v1alpha1,name=mmariadbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}
// +kubebuilder:object:generate=false

type MariaDBDatabaseCustomWebhook struct{}

var _ webhook.CustomDefaulter = &MariaDBDatabaseCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *MariaDBDatabaseCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	md, ok := obj.(*smapi.MariaDBDatabase)
	if !ok {
		return fmt.Errorf("expected a MariaDBDatabase, but got a %T", obj)
	}
	mariadbdatabaselog.Info("default", "name", md.GetName())

	if md.Spec.Init != nil {
		if md.Spec.Init.Snapshot != nil {
			if md.Spec.Init.Snapshot.SnapshotID == "" {
				md.Spec.Init.Snapshot.SnapshotID = "latest"
			}
		}
	}
	if md.Spec.Database.Config.CharacterSet == "" {
		md.Spec.Database.Config.CharacterSet = "utf8mb4"
	}
	return nil
}

//+kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mariadbdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mariadbdatabases,verbs=create;update;delete,versions=v1alpha1,name=vmariadbdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MariaDBDatabaseCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MariaDBDatabaseCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	md, ok := obj.(*smapi.MariaDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDBDatabase, but got a %T", obj)
	}
	mariadbdatabaselog.Info("validate create", "name", md.GetName())
	var allErrs field.ErrorList
	if err := w.ValidateMariaDBDatabase(md); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath(""), md.GetName(), err.Error()))
	}
	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MariaDBDatabase"}, md.GetName(), allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MariaDBDatabaseCustomWebhook) ValidateUpdate(_ context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	md, ok := newObj.(*smapi.MariaDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDBDatabase, but got a %T", newObj)
	}
	mariadbdatabaselog.Info("validate update", "name", md.GetName())
	oldobj := old.(*smapi.MariaDBDatabase)
	return nil, w.ValidateMariaDBDatabaseUpdate(md, oldobj)
}

func (w *MariaDBDatabaseCustomWebhook) ValidateMariaDBDatabaseUpdate(newobj *smapi.MariaDBDatabase, oldobj *smapi.MariaDBDatabase) error {
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
	er := w.ValidateMariaDBDatabase(newobj)
	if er != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), newobj.Name, er.Error()))
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MariaDBDatabase"}, newobj.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *MariaDBDatabaseCustomWebhook) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	md, ok := obj.(*smapi.MariaDBDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a MariaDBDatabase, but got a %T", obj)
	}
	mariadbdatabaselog.Info("validate delete", "name", md.GetName())
	if md.Spec.DeletionPolicy == smapi.DeletionPolicyDoNotDelete {
		return nil, field.Invalid(field.NewPath("spec").Child("terminationPolicy"), md.GetName(), `cannot delete object when terminationPolicy is set to "DoNotDelete"`)
	}
	return nil, nil
}

func (w *MariaDBDatabaseCustomWebhook) ValidateMariaDBDatabase(md *smapi.MariaDBDatabase) error {
	var allErrs field.ErrorList
	if err := w.validateInitailizationSchema(md); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := w.validateMariaDBDatabaseConfig(md); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MariaDBDatabase"}, md.GetName(), allErrs)
}

func (w *MariaDBDatabaseCustomWebhook) validateInitailizationSchema(md *smapi.MariaDBDatabase) *field.Error {
	path := field.NewPath("spec.init")
	if md.Spec.Init != nil {
		if md.Spec.Init.Script != nil && md.Spec.Init.Snapshot != nil {
			return field.Invalid(path, md.Name, `cannot initialize database using both restore and initSpec`)
		}
	}
	return nil
}

func (w *MariaDBDatabaseCustomWebhook) validateMariaDBDatabaseConfig(md *smapi.MariaDBDatabase) *field.Error {
	path := field.NewPath("spec").Child("database.config").Child("name")
	name := md.Spec.Database.Config.Name
	if name == smapi.SYSDatabase {
		return field.Invalid(path, md.GetName(), `cannot use "sys" as the database name`)
	}
	if name == "performance_schema" {
		return field.Invalid(path, md.GetName(), `cannot use "performance_schema" as the database name`)
	}
	if name == "mysql" {
		return field.Invalid(path, md.GetName(), `cannot use "mysql" as the database name`)
	}
	if name == smapi.DatabaseForEntry {
		return field.Invalid(path, md.GetName(), `cannot use "kubedb_system" as the database name`)
	}
	if name == "information_schema" {
		return field.Invalid(path, md.GetName(), `cannot use "information_schema" as the database name`)
	}
	if name == smapi.DatabaseNameAdmin {
		return field.Invalid(path, md.GetName(), `cannot use "admin" as the database name`)
	}
	if name == smapi.DatabaseNameConfig {
		return field.Invalid(path, md.GetName(), `cannot use "config" as the database name`)
	}
	return nil
}
