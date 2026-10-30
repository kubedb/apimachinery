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

// SetupEtcdAutoscalerWebhookWithManager registers the webhook for EtcdAutoscaler in the manager.
func SetupEtcdAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.EtcdAutoscaler{}).
		WithValidator(&EtcdAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&EtcdAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type EtcdAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var etcdAutoscalerLog = logf.Log.WithName("etcd-autoscaler")

var _ webhook.CustomDefaulter = &EtcdAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *EtcdAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.EtcdAutoscaler)
	if !ok {
		return fmt.Errorf("expected an EtcdAutoscaler object but got %T", obj)
	}
	etcdAutoscalerLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *EtcdAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.EtcdAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Etcd)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Etcd)
	}
}

func (w *EtcdAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.EtcdAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &EtcdAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *EtcdAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.EtcdAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an EtcdAutoscaler object but got %T", obj)
	}
	etcdAutoscalerLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *EtcdAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.EtcdAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an EtcdAutoscaler object but got %T", newObj)
	}
	etcdAutoscalerLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *EtcdAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *EtcdAutoscalerCustomWebhook) validate(scaler *autoscalingapi.EtcdAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
