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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var postgresdatabaselog = logf.Log.WithName("postgresdatabase-resource")

// SetupPostgresSchemaWebhookWithManager registers the webhook for SchemaManager in the managedb.
func SetupPostgresSchemaWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&smapi.PostgresDatabase{}).
		WithValidator(&PostgresDatabaseCustomWebhook{}).
		WithDefaulter(&PostgresDatabaseCustomWebhook{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-postgresdatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=postgresdatabases,verbs=create;update,versions=v1alpha1,name=mpostgresdatabase.kb.io,admissionReviewVersions={v1,v1beta1}
// +kubebuilder:object:generate=false
type PostgresDatabaseCustomWebhook struct{}

var _ webhook.CustomDefaulter = &PostgresDatabaseCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PostgresDatabaseCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db, ok := obj.(*smapi.PostgresDatabase)
	if !ok {
		return fmt.Errorf("expected a PostgresDatabase, got: %T", obj)
	}
	postgresdatabaselog.Info("default", "name", db.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-postgresdatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=postgresdatabases,verbs=create;update;delete,versions=v1alpha1,name=vpostgresdatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &PostgresDatabaseCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabaseCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.PostgresDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a PostgresDatabase, got: %T", obj)
	}
	postgresdatabaselog.Info("validate create", "name", db.Name)
	if db.Spec.Init != nil && db.Spec.Init.Initialized {
		return nil, field.Invalid(field.NewPath("spec").Child("init").Child("initialized"), db.Spec.Init.Initialized, fmt.Sprintf(`can't set spec.init.initialized true while creating postgresSchema %s/%s`, db.Namespace, db.Name))
	}
	return nil, r.ValidatePostgresDatabase(db)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabaseCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	newDB, ok := newObj.(*smapi.PostgresDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a PostgresDatabase, got: %T", newObj)
	}
	postgresdatabaselog.Info("validate update", "name", newDB.Name)
	oldobj := old.(*smapi.PostgresDatabase)
	return nil, r.ValidatePostgresDatabaseUpdate(oldobj, newDB)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PostgresDatabaseCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*smapi.PostgresDatabase)
	if !ok {
		return nil, fmt.Errorf("expected a PostgresDatabase, got: %T", obj)
	}
	postgresdatabaselog.Info("validate delete", "name", db.Name)
	if db.Spec.DeletionPolicy == smapi.DeletionPolicyDoNotDelete {
		return nil, field.Invalid(field.NewPath("spec").Child("deletionPolicy"), db.Spec.DeletionPolicy, fmt.Sprintf(`can't delete postgresSchema %s/%s when deletionPolicy is "DoNotDelete"`, db.Namespace, db.Name))
	}
	return nil, nil
}

func (r *PostgresDatabaseCustomWebhook) ValidateReadOnly(db *smapi.PostgresDatabase) *field.Error {
	if db.Spec.Database.Config.Params == nil {
		return nil
	}
	for _, param := range db.Spec.Database.Config.Params {
		if param.ConfigParameter == "default_transaction_read_only" && *param.Value == "on" && db.Spec.Init != nil && !db.Spec.Init.Initialized {
			return field.Invalid(field.NewPath("spec").Child("database").Child("config").Child("params"), db.Spec.Init, fmt.Sprintf("can't initialize a read-only database in postgresSchema %s/%s", db.Namespace, db.Name))
		}
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidatePostgresDatabaseUpdate(oldobj *smapi.PostgresDatabase, newobj *smapi.PostgresDatabase) error {
	if newobj.Finalizers == nil {
		return nil
	}
	path := field.NewPath("spec")
	if !gocmp.Equal(oldobj.Spec.Database.Config.Name, newobj.Spec.Database.Config.Name) {
		return field.Invalid(path.Child("database").Child("config").Child("name"), newobj.Spec.Database.Config.Name, fmt.Sprintf("can't change the database name in postgresSchema %s/%s", newobj.Namespace, newobj.Name))
	}
	if !gocmp.Equal(oldobj.Spec.Database.ServerRef, newobj.Spec.Database.ServerRef) {
		return field.Invalid(path.Child("database").Child("serverRef"), newobj.Spec.Database.ServerRef, fmt.Sprintf("can't change the kubedb server reference in postgresSchema %s/%s", newobj.Namespace, newobj.Name))
	}
	if !gocmp.Equal(oldobj.Spec.VaultRef, newobj.Spec.VaultRef) {
		return field.Invalid(path.Child("vaultRef"), newobj.Spec.VaultRef, fmt.Sprintf("can't change the vault reference in postgresSchema %s/%s", newobj.Namespace, newobj.Name))
	}
	if err := r.ValidatePostgresDatabase(newobj); err != nil {
		return err
	}
	if oldobj.Spec.Init != nil && oldobj.Spec.Init.Initialized && !gocmp.Equal(oldobj.Spec.Init, newobj.Spec.Init) {
		return field.Invalid(path.Child("init"), newobj.Spec.Init, fmt.Sprintf("can't change spec.init in postgresSchema %s/%s, is already initialized", newobj.Namespace, newobj.Name))
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidatePostgresDBName(db *smapi.PostgresDatabase) *field.Error {
	path := field.NewPath("spec").Child("database").Child("config").Child("name")
	name := db.Spec.Database.Config.Name
	if name == smapi.PostgresSchemaKubeSystem || name == smapi.DatabaseNameAdmin || name == smapi.DatabaseNameConfig || name == smapi.DatabaseNameLocal || name == "postgres" || name == "sys" || name == "template0" || name == "template1" {
		str := fmt.Sprintf("can't set spec.database.config.name \"%v\" in postgresSchema %s/%s", name, db.Namespace, db.Name)
		return field.Invalid(path, name, str)
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidateSchemaInitRestore(db *smapi.PostgresDatabase) *field.Error {
	path := field.NewPath("spec").Child("init")
	if db.Spec.Init != nil && db.Spec.Init.Snapshot != nil && db.Spec.Init.Script != nil {
		return field.Invalid(path, db.Name, fmt.Sprintf("can't set both spec.init.snapshot and spec.init.script in postgresSchema %s/%s", db.Namespace, db.Name))
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidateParams(db *smapi.PostgresDatabase) *field.Error {
	if db.Spec.Database.Config.Params == nil {
		return nil
	}
	for _, param := range db.Spec.Database.Config.Params {
		if param.ConfigParameter == "" || param.Value == nil {
			msg := fmt.Sprintf("can't set empty spec.database.config.params.configParameter or spec.database.config.params.value in postgresSchema %s/%s", db.Namespace, db.Name)
			return field.Invalid(field.NewPath("spec").Child("database").Child("config").Child("params"), db.Spec.Database.Config.Params, msg)
		}
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidateFields(db *smapi.PostgresDatabase) *field.Error {
	if db.Spec.Database.ServerRef.Name == "" {
		str := fmt.Sprintf("spec.database.serverRef.Name can't set empty in postgresSchema %s/%s", db.Namespace, db.Name)
		return field.Invalid(field.NewPath("spec").Child("database").Child("serverRef").Child("name"), db.Spec.Database.ServerRef, str)
	}
	if db.Spec.VaultRef.Name == "" {
		str := fmt.Sprintf("spec.database.vaultRef.Name can't set empty in postgresSchema %s/%s", db.Namespace, db.Name)
		return field.Invalid(field.NewPath("spec").Child("vaultRef").Child("name"), db.Spec.VaultRef, str)
	}
	if db.Spec.Init != nil && db.Spec.Init.Snapshot != nil {
		if db.Spec.Init.Snapshot.Repository.Name == "" {
			str := fmt.Sprintf("spec.init.snapshot.repository.name can't set empty in postgresSchema %s/%s", db.Namespace, db.Name)
			return field.Invalid(field.NewPath("spec").Child("init").Child("snapshot").Child("repository").Child("name"), db.Spec.Init.Snapshot.Repository.Name, str)
		}
	}
	if db.Spec.AccessPolicy.Subjects == nil {
		str := fmt.Sprintf("spec.accessPolicy.subjects can't set empty in postgresSchema %s/%s", db.Namespace, db.Name)
		return field.Invalid(field.NewPath("spec").Child("accessPolicy").Child("subjects"), db.Spec.AccessPolicy.Subjects, str)
	}
	return nil
}

func (r *PostgresDatabaseCustomWebhook) ValidatePostgresDatabase(db *smapi.PostgresDatabase) error {
	// check if Init and Restore both are present
	if err := r.ValidateSchemaInitRestore(db); err != nil {
		return err
	}
	// check the database name is conflicted with some constant name
	if err := r.ValidatePostgresDBName(db); err != nil {
		return err
	}
	// check the spec fields
	if err := r.ValidateFields(db); err != nil {
		return err
	}
	// check configuration params
	if err := r.ValidateParams(db); err != nil {
		return err
	}
	// check read-only
	if err := r.ValidateReadOnly(db); err != nil {
		return err
	}
	return nil
}
