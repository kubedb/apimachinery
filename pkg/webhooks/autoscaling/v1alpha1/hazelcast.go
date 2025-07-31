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

// SetupHazelcastAutoscalerWebhookWithManager registers the webhook for HazelcastAutoscaler in the manager.
func SetupHazelcastAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.HazelcastAutoscaler{}).
		WithValidator(&HazelcastAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&HazelcastAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type HazelcastAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var HazelcastLog = logf.Log.WithName("Hazelcast-autoscaler")

var _ webhook.CustomDefaulter = &HazelcastAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *HazelcastAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.HazelcastAutoscaler)
	if !ok {
		return fmt.Errorf("expected an HazelcastAutoscaler object but got %T", obj)
	}

	HazelcastLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *HazelcastAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.HazelcastAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Hazelcast)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Hazelcast)
	}
}

func (w *HazelcastAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.HazelcastAutoscaler) {
	HazelcastLog.Info("deafualting Opsrequest", "name", scaler.Name)
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.HazelcastOpsrequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &HazelcastAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *HazelcastAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.HazelcastAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an HazelcastAutoscaler object but got %T", obj)
	}

	HazelcastLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *HazelcastAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.HazelcastAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an HazelcastAutoscaler object but got %T", newObj)
	}

	return nil, w.validate(scaler)
}

func (w HazelcastAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *HazelcastAutoscalerCustomWebhook) validate(scaler *autoscalingapi.HazelcastAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
