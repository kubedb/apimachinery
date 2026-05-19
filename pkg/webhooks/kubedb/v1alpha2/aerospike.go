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
	"fmt"

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupAerospikeWebhookWithManager registers the webhook for Aerospike in the manager.
func SetupAerospikeWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.Aerospike{}).
		WithValidator(&AerospikeCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&AerospikeCustomWebhook{mgr.GetClient()}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-aerospike-kubedb-com-v1alpha1-aerospike,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=aerospikes,verbs=create;update,versions=v1alpha1,name=maerospike.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type AerospikeCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &AerospikeCustomWebhook{}

// log is for logging in this package.
var aerospikelog = logf.Log.WithName("aerospike-resource")

var _ webhook.CustomDefaulter = &AerospikeCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *AerospikeCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	ar, ok := obj.(*olddbapi.Aerospike)
	if !ok {
		return fmt.Errorf("expected an aerospike object but got %T", obj)
	}
	aerospikelog.Info("default", "name", ar.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-aerospike,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=aerospikes,verbs=create;update;delete,versions=v1alpha2,name=vaerospike.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &AerospikeCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *AerospikeCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ar, ok := obj.(*olddbapi.Aerospike)
	if !ok {
		return nil, fmt.Errorf("expected an aerospike object but got %T", obj)
	}
	aerospikelog.Info("validate create", "name", ar.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *AerospikeCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	ar, ok := newObj.(*olddbapi.Aerospike)
	if !ok {
		return nil, fmt.Errorf("expected an aerospike object but got %T", ar)
	}
	aerospikelog.Info("validate update", "name", ar.Name)
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *AerospikeCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ar, ok := obj.(*olddbapi.Aerospike)
	if !ok {
		return nil, fmt.Errorf("expected an aerospike object but got %T", ar)
	}
	aerospikelog.Info("validate delete", "name", ar.Name)

	var errorList field.ErrorList
	if ar.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		errorList = append(errorList, field.Invalid(field.NewPath("spec").Child("terminationPolicy"),
			ar.Name,
			"Can not delete as terminationPolicy is set to \"DoNotTerminate\""))
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "Aerospike"}, ar.Name, errorList)
	}
	return nil, nil
}
