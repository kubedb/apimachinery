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

// SetupZooKeeperAutoscalerWebhookWithManager registers the webhook for ZooKeeperAutoscaler in the manager.
func SetupZooKeeperAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.ZooKeeperAutoscaler{}).
		WithValidator(&ZooKeeperAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ZooKeeperAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ZooKeeperAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var zkLog = logf.Log.WithName("zookeeper-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-zookeeperautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=zookeeperautoscaler,verbs=create;update,versions=v1alpha1,name=mzookeeperautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &ZooKeeperAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ZooKeeperAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.ZooKeeperAutoscaler)
	if !ok {
		return fmt.Errorf("expected an ZooKeeperAutoscaler object but got %T", obj)
	}

	zkLog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *ZooKeeperAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.ZooKeeperAutoscaler) {
	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.ZooKeeper)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.ZooKeeper)
	}
}

func (in *ZooKeeperAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.ZooKeeperAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.ZooKeeperOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-zookeeperautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=zookeeperautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vzookeeperautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &ZooKeeperAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ZooKeeperAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.ZooKeeperAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ZooKeeperAutoscaler object but got %T", obj)
	}

	zkLog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ZooKeeperAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.ZooKeeperAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ZooKeeperAutoscaler object but got %T", newObj)
	}

	return nil, in.validate(scaler)
}

func (_ ZooKeeperAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *ZooKeeperAutoscalerCustomWebhook) validate(scaler *autoscalingapi.ZooKeeperAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
