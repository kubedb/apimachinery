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

// SetupPostgresAutoscalerWebhookWithManager registers the webhook for MariaDBAutoscaler in the manager.
func SetupPostgresAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.PostgresAutoscaler{}).
		WithValidator(&PostgresAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&PostgresAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PostgresAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pgLog = logf.Log.WithName("postgres-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-postgresautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=postgresautoscaler,verbs=create;update,versions=v1alpha1,name=mpostgresautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &PostgresAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *PostgresAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.PostgresAutoscaler)
	if !ok {
		return fmt.Errorf("expected an PostgresAutoscaler object but got %T", obj)
	}
	pgLog.Info("defaulting", "name", scaler.GetName())
	w.setDefaults(scaler)
	return nil
}

func (w *PostgresAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.PostgresAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Postgres)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Postgres)
	}
}

func (w *PostgresAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.PostgresAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.PostgresOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-postgresautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=postgresautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vpostgresautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &PostgresAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *PostgresAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.PostgresAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PostgresAutoscaler object but got %T", obj)
	}
	pgLog.Info("validate create", "name", scaler.GetName())
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *PostgresAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.PostgresAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PostgresAutoscaler object but got %T", newObj)
	}
	pgLog.Info("validate update", "name", scaler.GetName())
	return nil, w.validate(scaler)
}

func (w PostgresAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *PostgresAutoscalerCustomWebhook) validate(scaler *autoscalingapi.PostgresAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
