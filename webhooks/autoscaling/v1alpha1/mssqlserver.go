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

// SetupMSSQLServerAutoscalerWebhookWithManager registers the webhook for MSSQLServerAutoscaler in the manager.
func SetupMSSQLServerAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.MSSQLServerAutoscaler{}).
		WithValidator(&MSSQLServerAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MSSQLServerAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MSSQLServerAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mssqlLog = logf.Log.WithName("mssqlserver-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-mssqlserverautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=mssqlserverautoscaler,verbs=create;update,versions=v1alpha1,name=mmssqlserverautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &MSSQLServerAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *MSSQLServerAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.MSSQLServerAutoscaler)
	if !ok {
		return fmt.Errorf("expected an MSSQLServerAutoscaler object but got %T", obj)
	}

	mssqlLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *MSSQLServerAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.MSSQLServerAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.MSSQLServer)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.MSSQLServer)
	}
}

func (w *MSSQLServerAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.MSSQLServerAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.MSSQLServerOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mssqlserverautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mssqlserverautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmssqlserverautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MSSQLServerAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MSSQLServerAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.MSSQLServerAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerAutoscaler object but got %T", obj)
	}

	mssqlLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MSSQLServerAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.MSSQLServerAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerAutoscaler object but got %T", newObj)
	}

	return nil, w.validate(scaler)
}

func (w MSSQLServerAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *MSSQLServerAutoscalerCustomWebhook) validate(scaler *autoscalingapi.MSSQLServerAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
