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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupFerretDBAutoscalerWebhookWithManager registers the webhook for FerretDBAutoscaler in the manager.
func SetupFerretDBAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.FerretDBAutoscaler{}).
		WithValidator(&FerretDBAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&FerretDBAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type FerretDBAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var ferretdbLog = logf.Log.WithName("ferretdb-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-ferretdbautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=ferretdbautoscaler,verbs=create;update,versions=v1alpha1,name=mferretdbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &FerretDBAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *FerretDBAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.FerretDBAutoscaler)
	if !ok {
		return fmt.Errorf("expected an FerretDBAutoscaler object but got %T", obj)
	}

	ferretdbLog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *FerretDBAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.FerretDBAutoscaler) {
	var db olddbapi.FerretDB
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get FerretDB %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.FerretDB)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Primary)
		setDefaultComputeValues(scaler.Spec.Compute.Secondary)
	}
}

func (in *FerretDBAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.FerretDBAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.FerretDBOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-ferretdbautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=ferretdbautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vferretdbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &FerretDBAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *FerretDBAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.FerretDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an FerretDBAutoscaler object but got %T", obj)
	}

	ferretdbLog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *FerretDBAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.FerretDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an FerretDBAutoscaler object but got %T", newObj)
	}

	return nil, in.validate(scaler)
}

func (_ FerretDBAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *FerretDBAutoscalerCustomWebhook) validate(scaler *autoscalingapi.FerretDBAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.FerretDB
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		_ = fmt.Errorf("can't get FerretDB %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
