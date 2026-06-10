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

// SetupOracleAutoscalerWebhookWithManager registers the webhook for OracleAutoscaler in the manager.
func SetupOracleAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.OracleAutoscaler{}).
		WithValidator(&OracleAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&OracleAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type OracleAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var oraLog = logf.Log.WithName("oracle-autoscaler")

var _ webhook.CustomDefaulter = &OracleAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *OracleAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.OracleAutoscaler)
	if !ok {
		return fmt.Errorf("expected an OracleAutoscaler object but got %T", obj)
	}
	oraLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *OracleAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.OracleAutoscaler) {
	var db olddbapi.Oracle
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		klog.Errorf("can't get Oracle %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Node)
		setDefaultStorageValues(scaler.Spec.Storage.Observer)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Node)
		setDefaultComputeValues(scaler.Spec.Compute.Observer)
	}
}

func (w *OracleAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.OracleAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &OracleAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.OracleAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an OracleAutoscaler object but got %T", obj)
	}
	oraLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.OracleAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an OracleAutoscaler object but got %T", newObj)
	}
	oraLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *OracleAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *OracleAutoscalerCustomWebhook) validate(scaler *autoscalingapi.OracleAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.Oracle
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		klog.Errorf("can't get Oracle %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
