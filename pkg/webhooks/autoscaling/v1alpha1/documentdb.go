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
	"context"
	"errors"
	"fmt"

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupDocumentDBAutoscalerWebhookWithManager registers the webhook for DocumentDBAutoscaler in the manager.
func SetupDocumentDBAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.DocumentDBAutoscaler{}).
		WithValidator(&DocumentDBAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&DocumentDBAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DocumentDBAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var documentDBLog = logf.Log.WithName("documentdb-autoscaler")

var _ webhook.CustomDefaulter = &DocumentDBAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *DocumentDBAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.DocumentDBAutoscaler)
	if !ok {
		return fmt.Errorf("expected a DocumentDBAutoscaler object but got %T", obj)
	}
	documentDBLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *DocumentDBAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.DocumentDBAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.DocumentDB)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.DocumentDB)
	}
}

func (w *DocumentDBAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.DocumentDBAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &DocumentDBAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.DocumentDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a DocumentDBAutoscaler object but got %T", obj)
	}
	documentDBLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.DocumentDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a DocumentDBAutoscaler object but got %T", newObj)
	}
	documentDBLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *DocumentDBAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *DocumentDBAutoscalerCustomWebhook) validate(scaler *autoscalingapi.DocumentDBAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
