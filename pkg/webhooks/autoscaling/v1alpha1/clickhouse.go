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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupClickHouseAutoscalerWebhookWithManager registers the webhook for ClickHouseAutoscaler in the manager.
func SetupClickHouseAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.ClickHouseAutoscaler{}).
		WithValidator(&ClickHouseAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ClickHouseAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ClickHouseAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var chLog = logf.Log.WithName("clickhouse-autoscaler")

var _ webhook.CustomDefaulter = &ClickHouseAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *ClickHouseAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.ClickHouseAutoscaler)
	if !ok {
		return fmt.Errorf("expected an ClickHouseAutoscaler object but got %T", obj)
	}
	chLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *ClickHouseAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.ClickHouseAutoscaler) {
	var db olddbapi.ClickHouse
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		klog.Errorf("can't get ClickHouse %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.ClickHouse)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.ClickHouse)
	}
}

func (w *ClickHouseAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.ClickHouseAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.ClickHouseOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &ClickHouseAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ClickHouseAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.ClickHouseAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ClickHouseAutoscaler object but got %T", obj)
	}
	chLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ClickHouseAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.ClickHouseAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ClickHouseAutoscaler object but got %T", newObj)
	}
	chLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *ClickHouseAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ClickHouseAutoscalerCustomWebhook) validate(scaler *autoscalingapi.ClickHouseAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.ClickHouse
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		klog.Errorf("can't get ClickHouse %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
