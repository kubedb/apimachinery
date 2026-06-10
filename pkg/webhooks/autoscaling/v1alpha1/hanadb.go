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

// SetupHanaDBAutoscalerWebhookWithManager registers the webhook for HanaDBAutoscaler in the manager.
func SetupHanaDBAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.HanaDBAutoscaler{}).
		WithValidator(&HanaDBAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&HanaDBAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type HanaDBAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var hanaDBLog = logf.Log.WithName("hanadb-autoscaler")

var _ webhook.CustomDefaulter = &HanaDBAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *HanaDBAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.HanaDBAutoscaler)
	if !ok {
		return fmt.Errorf("expected a HanaDBAutoscaler object but got %T", obj)
	}
	hanaDBLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *HanaDBAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.HanaDBAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.HanaDB)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.HanaDB)
	}
}

func (w *HanaDBAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.HanaDBAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &HanaDBAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *HanaDBAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.HanaDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a HanaDBAutoscaler object but got %T", obj)
	}
	hanaDBLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *HanaDBAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.HanaDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a HanaDBAutoscaler object but got %T", newObj)
	}
	hanaDBLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *HanaDBAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *HanaDBAutoscalerCustomWebhook) validate(scaler *autoscalingapi.HanaDBAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
