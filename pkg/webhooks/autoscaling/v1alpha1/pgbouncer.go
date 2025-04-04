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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupPgBouncerAutoscalerWebhookWithManager registers the webhook for PgBouncerAutoscaler in the manager.
func SetupPgBouncerAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.PgBouncerAutoscaler{}).
		WithValidator(&PgBouncerAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&PgBouncerAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PgBouncerAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pbLog = logf.Log.WithName("pgbouncer-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-pgbouncerautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=pgbouncerautoscaler,verbs=create;update,versions=v1alpha1,name=mpgbouncerautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &PgBouncerAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *PgBouncerAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.PgBouncerAutoscaler)
	if !ok {
		return fmt.Errorf("expected an PgBouncerAutoscaler object but got %T", obj)
	}
	pbLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *PgBouncerAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.PgBouncerAutoscaler) {
	var db dbapi.PgBouncer
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get PgBouncer %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.PgBouncer)
	}
}

func (w *PgBouncerAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.PgBouncerAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.PgBouncerOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-pgbouncerautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=pgbouncerautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vpgbouncerautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &PgBouncerAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *PgBouncerAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.PgBouncerAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PgBouncerAutoscaler object but got %T", obj)
	}
	pbLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *PgBouncerAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.PgBouncerAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PgBouncerAutoscaler object but got %T", newObj)
	}
	pbLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *PgBouncerAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *PgBouncerAutoscalerCustomWebhook) validate(scaler *autoscalingapi.PgBouncerAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var bouncer dbapi.PgBouncer
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &bouncer)
	if err != nil {
		_ = fmt.Errorf("can't get PgBouncer %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}
	return nil
}
