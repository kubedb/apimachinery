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

package v1alpha2

import (
	"context"
	"errors"
	"fmt"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupDocumentDBWebhookWithManager registers the webhook for DocumentDB in the manager.
func SetupDocumentDBWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.DocumentDB{}).
		WithValidator(&DocumentDBCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&DocumentDBCustomWebhook{mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-documentdb-kubedb-com-v1alpha1-documentdb,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=documentdbs,verbs=create;update,versions=v1alpha1,name=mdocumentdb.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type DocumentDBCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var documentdblog = logf.Log.WithName("documentdb-resource")

var _ webhook.CustomDefaulter = &DocumentDBCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *DocumentDBCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	db, ok := obj.(*olddbapi.DocumentDB)
	if !ok {
		return fmt.Errorf("expected an DocumentDB object but got %T", obj)
	}

	documentdblog.Info("default", "name", db.Name)

	db.SetDefaults(w.DefaultClient)
	return nil
}

var _ webhook.CustomValidator = &DocumentDBCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.DocumentDB)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDB object but got %T", obj)
	}

	documentdblog.Info("validate create", "name", db.Name)
	allErr := w.ValidateCreateOrUpdate(db)
	if len(allErr) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "DocumentDB"}, db.Name, allErr)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	db, ok := newObj.(*olddbapi.DocumentDB)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDB object but got %T", newObj)
	}
	documentdblog.Info("validate update", "name", db.Name)

	allErr := w.ValidateCreateOrUpdate(db)
	if len(allErr) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "DocumentDB"}, db.Name, allErr)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	db, ok := obj.(*olddbapi.DocumentDB)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDB object but got %T", obj)
	}
	documentdblog.Info("validate delete", "name", db.Name)

	var allErr field.ErrorList
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("terminationPolicy"),
			db.Name,
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "DocumentDB"}, db.Name, allErr)
	}
	return nil, nil
}

func (w *DocumentDBCustomWebhook) ValidateCreateOrUpdate(db *olddbapi.DocumentDB) field.ErrorList {
	var allErr field.ErrorList

	err := w.validateDocumentDBVersion(db)
	if err != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("version"),
			db.Name,
			err.Error()))
	}
	if db.Spec.Server.Primary == nil || *db.Spec.Server.Primary.Replicas < 1 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("server.primary.replicas"),
			db.Name,
			fmt.Sprintf(`spec.server.primary.replicas "%v" invalid. Must be greater than zero`, *db.Spec.Server.Primary.Replicas)))
	}

	if db.Spec.Server.Secondary != nil && *db.Spec.Server.Secondary.Replicas < 1 {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("server.secondary.replicas"),
			db.Name,
			fmt.Sprintf(`spec.server.secondary.replicas "%v" invalid. Must be greater than zero`, *db.Spec.Server.Secondary.Replicas)))
	}

	if db.Spec.Server.Primary != nil && db.Spec.Server.Primary.PodTemplate != nil {
		if err := DocumentDBValidateEnvVar(getMainContainerEnvs(db.Spec.Server.Primary.PodTemplate), forbiddenEnvVars, db.ResourceKind()); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("server.primary.podTemplate"),
				db.Name,
				err.Error()))
		}
	}

	if db.Spec.Server.Secondary != nil && db.Spec.Server.Secondary.PodTemplate != nil {
		if err := DocumentDBValidateEnvVar(getMainContainerEnvs(db.Spec.Server.Secondary.PodTemplate), forbiddenEnvVars, db.ResourceKind()); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("server.secondary.podTemplate"),
				db.Name,
				err.Error()))
		}
	}

	if db.Spec.Backend == nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("backend"),
			db.Name,
			`'spec.backend' is missing`))
	}

	// Storage related
	if db.Spec.Backend.StorageType == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("backend", "storageType"),
			db.Name,
			`'spec.backend.storageType' is missing`))
	}
	if db.Spec.Backend.StorageType != olddbapi.StorageTypeDurable && db.Spec.Backend.StorageType != olddbapi.StorageTypeEphemeral {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("backend", "storageType"),
			db.Name,
			fmt.Sprintf(`'spec.backend.storageType' %s is invalid`, db.Spec.Backend.StorageType)))
	}
	if db.Spec.Backend.StorageType == olddbapi.StorageTypeEphemeral && db.Spec.Backend.Storage != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("backend", "storageType"),
			db.Name,
			`'spec.backend.storageType' is set to Ephemeral, so 'spec.backend.storage' needs to be empty`))
	}

	// Auth secret related
	if db.Spec.AuthSecret != nil && db.Spec.AuthSecret.ExternallyManaged && db.Spec.AuthSecret.Name == "" {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authSecret"),
			db.Name,
			`'spec.authSecret.name' need to specify when auth secret is externally managed`))
	}

	// Termination policy related
	if db.Spec.DeletionPolicy == olddbapi.DeletionPolicyHalt {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("terminationPolicy"),
			db.Name,
			`'spec.terminationPolicy' value 'Halt' is not supported yet for DocumentDB`))
	}

	// TLS related
	if db.Spec.SSLMode == olddbapi.SSLModeAllowSSL || db.Spec.SSLMode == olddbapi.SSLModePreferSSL {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("sslMode"),
			db.Name,
			`'spec.sslMode' value 'allowSSL' or 'preferSSL' is not supported yet for DocumentDB`))
	}
	if db.Spec.SSLMode == olddbapi.SSLModeRequireSSL && db.Spec.TLS == nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("sslMode"),
			db.Name,
			`'spec.sslMode' is requireSSL but 'spec.tls' is not set`))
	}
	if db.Spec.SSLMode == olddbapi.SSLModeDisabled && db.Spec.TLS != nil {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("sslMode"),
			db.Name,
			`'spec.tls' is can't set when 'spec.sslMode' is disabled`))
	}

	return allErr
}

func DocumentDBValidateEnvVar(envs []core.EnvVar, forbiddenEnvs []string, resourceType string) error {
	for _, env := range envs {
		present, _ := arrays.Contains(forbiddenEnvs, env.Name)
		if present {
			return fmt.Errorf("environment variable %s is forbidden to use in %s spec", env.Name, resourceType)
		}
	}
	return nil
}

func (w *DocumentDBCustomWebhook) validateDocumentDBVersion(db *olddbapi.DocumentDB) error {
	dcVersion := v1alpha1.DocumentDBVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &dcVersion)
	if err != nil {
		return errors.New("version not supported")
	}

	pgVersion := v1alpha1.PostgresVersion{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: dcVersion.Spec.Postgres.Version}, &pgVersion)
	if err != nil {
		return errors.New("postgres version not supported in KubeDB")
	}
	return nil
}
