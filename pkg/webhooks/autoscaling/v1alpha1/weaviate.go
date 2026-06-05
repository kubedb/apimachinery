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

// SetupWeaviateAutoscalerWebhookWithManager registers the webhook for WeaviateAutoscaler in the manager.
func SetupWeaviateAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.WeaviateAutoscaler{}).
		WithValidator(&WeaviateAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&WeaviateAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type WeaviateAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var WeaviateLog = logf.Log.WithName("weaviate-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-weaviateautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=weaviateautoscaler,verbs=create;update,versions=v1alpha1,name=mweaviateautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &WeaviateAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *WeaviateAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.WeaviateAutoscaler)
	if !ok {
		return fmt.Errorf("expected an WeaviateAutoscaler object but got %T", obj)
	}

	WeaviateLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *WeaviateAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.WeaviateAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Weaviate)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Weaviate)
	}
}

func (w *WeaviateAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.WeaviateAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-weaviateautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=weaviateautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vweaviateautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &WeaviateAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *WeaviateAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.WeaviateAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an WeaviateAutoscaler object but got %T", obj)
	}

	WeaviateLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *WeaviateAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.WeaviateAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an WeaviateAutoscaler object but got %T", newObj)
	}

	return nil, w.validate(scaler)
}

func (w WeaviateAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *WeaviateAutoscalerCustomWebhook) validate(scaler *autoscalingapi.WeaviateAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
