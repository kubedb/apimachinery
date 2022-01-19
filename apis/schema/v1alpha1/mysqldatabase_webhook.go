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
	"reflect"

	gocmp "github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mysqldatabaselog = logf.Log.WithName("mysqldatabase-resource")

func (obj *MySQLDatabase) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(obj).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=true,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update,versions=v1alpha1,name=mmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MySQLDatabase{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (obj *MySQLDatabase) Default() {
	mysqldatabaselog.Info("default", "name", obj.Name)

	if obj.Spec.Init.Snapshot != nil {
		if obj.Spec.Init.Snapshot.SnapshotID == "" {
			obj.Spec.Init.Snapshot.SnapshotID = "latest"
		}
	}
	val := obj.Spec.DatabaseConfig.Encryption
	if val == "enable" || val == ENCRYPTIONENABLE {
		obj.Spec.DatabaseConfig.Encryption = ENCRYPTIONENABLE
	} else {
		obj.Spec.DatabaseConfig.Encryption = ENCRYPTIONDISABLE
	}
	if obj.Spec.DatabaseConfig.ReadOnly != 1 {
		obj.Spec.DatabaseConfig.ReadOnly = 0
	}
	if obj.Spec.DatabaseConfig.CharacterSet == "" {
		obj.Spec.DatabaseConfig.CharacterSet = "utf8"
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mysqldatabase,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqldatabases,verbs=create;update;delete,versions=v1alpha1,name=vmysqldatabase.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MySQLDatabase{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (obj *MySQLDatabase) ValidateCreate() error {
	mysqldatabaselog.Info("validate create", "name", obj.Name)

	return obj.ValidateMySQLDatabase()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (obj *MySQLDatabase) ValidateUpdate(old runtime.Object) error {
	mysqldatabaselog.Info("validate update", "name", obj.Name)
	oldobj := old.(*MySQLDatabase)
	return validateMySQLDatabaseUpdate(oldobj, obj)
}

func validateMySQLDatabaseUpdate(oldobj *MySQLDatabase, newobj *MySQLDatabase) error {
	if newobj.Finalizers == nil {
		return nil
	}
	var allErrs field.ErrorList
	path := field.NewPath("spec")
	if !kmapi.IsConditionTrue(oldobj.Status.Conditions, string(SchemaIgnored)) {
		if oldobj.Spec.DatabaseConfig.Name != newobj.Spec.DatabaseConfig.Name {
			allErrs = append(allErrs, field.Invalid(path.Child("databaseConfig"), newobj.Name, `Cannot change target database name`))
		}
		if oldobj.Spec.DatabaseRef != newobj.Spec.DatabaseRef {
			allErrs = append(allErrs, field.Invalid(path.Child("mysqlRef"), newobj.Name, `Cannot change mysql reference`))
		}
		if oldobj.Spec.VaultRef != newobj.Spec.VaultRef {
			allErrs = append(allErrs, field.Invalid(path.Child("vaultRef"), newobj.Name, `Cannot change vault reference`))
		}
	}
	if err := newobj.ValidateMySQLDatabase(); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath(""), newobj.Name, err.Error()))
	}
	if kmapi.IsConditionTrue(oldobj.Status.Conditions, string(ScriptApplied)) {

		if !gocmp.Equal(oldobj.Spec.Init.Script, newobj.Spec.Init.Script) {
			allErrs = append(allErrs, field.Invalid(path.Child("initSpec.script"), newobj.Name, "Cannot change initSpec script, former script already applied"))
		}
		//if oldobj.Spec.Init.Script !=  newobj.Spec.Init.Script {
		//	klog.Info("\nupdated2\n")
		//	klog.Infof("printing old object %+v\n", &oldobj.Spec.Init)
		//	klog.Infof("printing new object %+v\n", &newobj.Spec.Init)
		//}
		if oldobj.Spec.Init.PodTemplate != newobj.Spec.Init.PodTemplate {
			if newobj.Spec.Init.PodTemplate != nil {
				klog.Infof("script already applied : Changes in the pod template won't be applied")
			}
		}
	}
	if kmapi.IsConditionTrue(oldobj.Status.Conditions, string(RestoredFromRepository)) {
		if !reflect.DeepEqual(oldobj.Spec.Init.Snapshot, newobj.Spec.Init.Snapshot) {
			allErrs = append(allErrs, field.Invalid(path.Child("restore"), newobj.Name, "Cannot change restore, former restore session already applied"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, newobj.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (obj *MySQLDatabase) ValidateDelete() error {
	mysqldatabaselog.Info("validate delete", "name", obj.Name)
	if obj.Spec.TerminationPolicy == TerminationPolicyDoNotDelete {
		return field.Invalid(field.NewPath("spec").Child("terminationPolicy"), obj.Name, `cannot delete object when terminationPolicy is set to "DoNotDelete"`)
	}
	if obj.Spec.DatabaseConfig.ReadOnly == 1 {
		return field.Invalid(field.NewPath("spec").Child("databaseConfig.readOnly"), obj.Name, `schema manger cannot be deleted : database is read only enabled`)
	}
	return nil
}

func (obj *MySQLDatabase) ValidateMySQLDatabase() error {
	var allErrs field.ErrorList
	if err := obj.validateTerminationPolicy(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := obj.validateInitailizationSchema(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := obj.validateMySQLDatabaseConfig(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := obj.validateMySQLDatabaseNamespace(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := obj.validateMySQLDatabaseName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "schema.kubedb.com", Kind: "MySQLDatabase"}, obj.Name, allErrs)
}

func (obj *MySQLDatabase) validateTerminationPolicy() *field.Error {
	val := obj.Spec.TerminationPolicy
	if val != TerminationPolicyDelete && val != TerminationPolicyDoNotDelete {
		return field.Invalid(field.NewPath("spec").Child("terminationPolicy"), obj.Name, `terminationPolicy must be either "Delete" or "DoNotDelete"`)
	}
	return nil
}

func (obj *MySQLDatabase) validateInitailizationSchema() *field.Error {
	path := field.NewPath("spec")
	if obj.Spec.Init != nil && obj.Spec.Init.Snapshot != nil {
		return field.Invalid(path, obj.Name, `cannot initialize database using both restore and initSpec`)
	}
	return nil
}

func (obj *MySQLDatabase) validateMySQLDatabaseConfig() *field.Error {
	path := field.NewPath("spec").Child("databaseConfig").Child("name")
	name := obj.Spec.DatabaseConfig.Name
	if name == "sys" {
		return field.Invalid(path, obj.Name, `cannot use "sys" as the database name`)
	}
	if name == "performance_schema" {
		return field.Invalid(path, obj.Name, `cannot use "performance_schema" as the database name`)
	}
	if name == "mysql" {
		return field.Invalid(path, obj.Name, `cannot use "mysql" as the database name`)
	}
	if name == "kubedb_system" {
		return field.Invalid(path, obj.Name, `cannot use "kubedb_system" as the database name`)
	}
	if name == "information_schema" {
		return field.Invalid(path, obj.Name, `cannot use "information_schema" as the database name`)
	}
	if name == "admin" {
		return field.Invalid(path, obj.Name, `cannot use "admin" as the database name`)
	}
	if name == "config" {
		return field.Invalid(path, obj.Name, `cannot use "config" as the database name`)
	}
	path = field.NewPath("spec").Child("databaseConfig").Child("readOnly")
	val := obj.Spec.DatabaseConfig.ReadOnly
	if val == 1 {
		if (obj.Spec.Init != nil || obj.Spec.Init.Snapshot != nil) && obj.Status.Phase != Success {
			return field.Invalid(path, obj.Name, `cannot make the database readonly , init/restore yet to be applied`)
		}
	}
	return nil
}

func (obj *MySQLDatabase) validateMySQLDatabaseNamespace() *field.Error {
	path := field.NewPath("metadata").Child("namespace")
	ns := obj.ObjectMeta.Namespace
	if ns == "cert-manager" {
		return field.Invalid(path, obj.Name, `cannot use namespace "cert-manager" to create schema manager`)
	}
	if ns == "kube-system" {
		return field.Invalid(path, obj.Name, `cannot use namespace "kube-system" to create schema manager`)
	}
	if ns == "kubedb-system" {
		return field.Invalid(path, obj.Name, `cannot use namespace "kubedb-system" to create schema manager`)
	}
	if ns == "kubedb" {
		return field.Invalid(path, obj.Name, `cannot use namespace "kubedb" to create schema manager`)
	}
	if ns == "kubevault" {
		return field.Invalid(path, obj.Name, `cannot use namespace "kubevault" to create schema manager`)
	}
	if ns == "local-path-storage" {
		return field.Invalid(path, obj.Name, `cannot use namespace "local-path-storage" to create schema manager`)
	}
	return nil
}

func (obj *MySQLDatabase) validateMySQLDatabaseName() *field.Error {
	if len(obj.ObjectMeta.Name) > 30 {
		return field.Invalid(field.NewPath("metadata").Child("name"), obj.Name, "must be no more than 30 characters")
	}
	return nil
}
