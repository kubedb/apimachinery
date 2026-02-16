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

// SetupIgniteAutoscalerWebhookWithManager registers the webhook for IgniteAutoscaler in the manager.
func SetupIgniteAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.IgniteAutoscaler{}).
		WithValidator(&IgniteAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&IgniteAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type IgniteAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var igniteLog = logf.Log.WithName("ignite-autoscaler")

var _ webhook.CustomDefaulter = &IgniteAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *IgniteAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.IgniteAutoscaler)
	if !ok {
		return fmt.Errorf("expected an IgniteAutoscaler object but got %T", obj)
	}
	igniteLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *IgniteAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.IgniteAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Ignite)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Ignite)
	}
}

func (w *IgniteAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.IgniteAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.IgniteOpsrequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &IgniteAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *IgniteAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.IgniteAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an IgniteAutoscaler object but got %T", obj)
	}

	igniteLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *IgniteAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.IgniteAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an IgniteAutoscaler object but got %T", newObj)
	}
	igniteLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w IgniteAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *IgniteAutoscalerCustomWebhook) validate(scaler *autoscalingapi.IgniteAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
