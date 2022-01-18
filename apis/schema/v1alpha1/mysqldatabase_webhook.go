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
	gocmp "github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	kmodulesv1 "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mysqldatabaselog = logf.Log.WithName("mysqldatabase-resource")

func (r *MySQLDatabase) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update,versions=v1alpha1,name=mmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MySQLDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MySQLDatabase) Default() {
	mysqldatabaselog.Info("default", "name", r.Name)

	if r.Spec.Restore != nil {
		if r.Spec.Restore.Snapshot == "" {
			r.Spec.Restore.Snapshot = "latest"
		}
	}
	val := r.Spec.DatabaseConfig.Encryption
	if val == "enable" || val == ENCRYPTIONENABLE {
		r.Spec.DatabaseConfig.Encryption = ENCRYPTIONENABLE
	} else {
		r.Spec.DatabaseConfig.Encryption = ENCRYPTIONDISABLE
	}
	if r.Spec.DatabaseConfig.ReadOnly != 1 {
		r.Spec.DatabaseConfig.ReadOnly = 0
	}
	if r.Spec.DatabaseConfig.CharacterSet == "" {
		r.Spec.DatabaseConfig.CharacterSet = "utf8"
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update;delete,versions=v1alpha1,name=vmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MySQLDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MySQLDatabase) ValidateCreate() error {
	mysqldatabaselog.Info("validate create", "name", r.Name)

	return r.ValidateMySQLDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MySQLDatabase) ValidateUpdate(old runtime.Object) error {
	mysqldatabaselog.Info("validate update", "name", r.Name)
	oldobj := old.(*MySQLDatabase)
	return validateMySQLDatabaseUpdate(oldobj, r)
}

func validateMySQLDatabaseUpdate(oldobj *MySQLDatabase, newobj *MySQLDatabase) error {
	if newobj.Finalizers == nil {
		return nil
	}
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if !kmodulesv1.IsConditionTrue(oldobj.Status.Conditions, string(SchemaIgnored)) {
		if oldobj.Spec.DatabaseConfig.Name != newobj.Spec.DatabaseConfig.Name {
			allErrs = append(allErrs, field.Invalid(path.Child("databaseConfig"), newobj.Name, `Cannot change target database name`))
		}
		if oldobj.Spec.MySQLRef != newobj.Spec.MySQLRef {
			allErrs = append(allErrs, field.Invalid(path.Child("mysqlRef"), newobj.Name, `Cannot change mysql reference`))
		}
		if oldobj.Spec.VaultRef != newobj.Spec.VaultRef {
			allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), newobj.Name, `Cannot change vault reference`))
		}
	}
	if err := newobj.ValidateMySQLDatabase(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath(""), newobj.Name, err.Error()))
	}
	if kmodulesv1.IsConditionTrue(oldobj.Status.Conditions, string(ScriptApplied)) {

		if !gocmp.Equal(oldobj.Spec.InitSpec.Script, newobj.Spec.InitSpec.Script) {
			allErrs = append(allErrs, field.Invalid(path.Child("initSpec.script"), newobj.Name, "Cannot change initSpec script, former script already applied"))
		}
		//if oldobj.Spec.InitSpec.Script !=  newobj.Spec.InitSpec.Script {
		//	klog.Info("\nupdated2\n")
		//	klog.Infof("printing old object %+v\n", &oldobj.Spec.InitSpec)
		//	klog.Infof("printing new object %+v\n", &newobj.Spec.InitSpec)
		//}
		if oldobj.Spec.InitSpec.PodTemplate != newobj.Spec.InitSpec.PodTemplate {
			if newobj.Spec.InitSpec.PodTemplate != nil {
				klog.Infof("script already applied : Changes in the pod template won't be applied")
			}
		}
	}
	if kmodulesv1.IsConditionTrue(oldobj.Status.Conditions, string(RestoredFromRepository)) {
		if oldobj.Spec.Restore != newobj.Spec.Restore {
			allErrs = append(allErrs, field.Invalid(path.Child("restore"), newobj.Name, "Cannot change restore, former restore session already applied"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, newobj.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MySQLDatabase) ValidateDelete() error {
	mysqldatabaselog.Info("validate delete", "name", r.Name)
	if r.Spec.TerminationPolicy == TerminationPolicyDoNotDelete {
		return field.Invalid(field.NewPath("spec").Child("terminationPolicy"), r.Name, `cannot delete object when terminationPolicy is set to "DoNotDelete"`)
	}
	if r.Spec.DatabaseConfig.ReadOnly == 1 {
		return field.Invalid(field.NewPath("spec").Child("databaseConfig.readOnly"), r.Name, `schema manger cannot be deleted : database is read only enabled`)
	}
	return nil
}

func (r *MySQLDatabase) ValidateMySQLDatabase() error {
	var allErrs field.ErrorList
	if err := r.validateTerminationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateInitailizationSchema(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateMySQLDatabaseConfig(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateMySQLDatabaseNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateMySQLDatabaseName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, r.Name, allErrs)
}

func (r *MySQLDatabase) validateTerminationPolicy() *field.Error {
	val := r.Spec.TerminationPolicy
	if val != TerminationPolicyDelete && val != TerminationPolicyDoNotDelete {
		return field.Invalid(field.NewPath("spec").Child("terminationPolicy"), r.Name, `terminationPolicy must be either "Delete" or "DoNotDelete"`)
	}
	return nil
}

func (r *MySQLDatabase) validateInitailizationSchema() *field.Error {
	path := field.NewPath("spec")
	if r.Spec.InitSpec != nil && r.Spec.Restore != nil {
		return field.Invalid(path, r.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (r *MySQLDatabase) validateMySQLDatabaseConfig() *field.Error {
	path := field.NewPath("spec").Child("databaseConfig").Child("name")
	name := r.Spec.DatabaseConfig.Name
	if name == "sys" {
		return field.Invalid(path, r.Name, `cannot use "sys" as the database name`)
	}
	if name == "performance_schema" {
		return field.Invalid(path, r.Name, `cannot use "performance_schema" as the database name`)
	}
	if name == "mysql" {
		return field.Invalid(path, r.Name, `cannot use "mysql" as the database name`)
	}
	if name == "kubedb_system" {
		return field.Invalid(path, r.Name, `cannot use "kubedb_system" as the database name`)
	}
	if name == "information_schema" {
		return field.Invalid(path, r.Name, `cannot use "information_schema" as the database name`)
	}
	if name == "admin" {
		return field.Invalid(path, r.Name, `cannot use "admin" as the database name`)
	}
	if name == "config" {
		return field.Invalid(path, r.Name, `cannot use "config" as the database name`)
	}
	path = field.NewPath("spec").Child("databaseConfig").Child("readOnly")
	val := r.Spec.DatabaseConfig.ReadOnly
	if val == 1 {
		if (r.Spec.InitSpec != nil || r.Spec.Restore != nil) && r.Status.Phase != Success {
			return field.Invalid(path, r.Name, `cannot make the database readonly , init/restore yet to be applied`)
		}
	}
	return nil
}

func (r *MySQLDatabase) validateMySQLDatabaseNamespace() *field.Error {
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

func (r *MySQLDatabase) validateMySQLDatabaseName() *field.Error {
	if len(r.ObjectMeta.Name) > 30 {
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 30 characters")
	}
	return nil
}
