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

// SetupMySQLAutoscalerWebhookWithManager registers the webhook for MySQLAutoscaler in the manager.
func SetupMySQLAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.MySQLAutoscaler{}).
		WithValidator(&MySQLAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MySQLAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MySQLAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mylog = logf.Log.WithName("mysql-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-mysqlautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=mysqlautoscaler,verbs=create;update,versions=v1alpha1,name=mmysqlautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &MySQLAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MySQLAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.MySQLAutoscaler)
	if !ok {
		return fmt.Errorf("expected an MySQLAutoscaler object but got %T", obj)
	}

	mylog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *MySQLAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.MySQLAutoscaler) {
	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.MySQL)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.MySQL)
	}
}

func (in *MySQLAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.MySQLAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.MySQLOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mysqlautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mysqlautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmysqlautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MySQLAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MySQLAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.MySQLAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLAutoscaler object but got %T", obj)
	}

	mylog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MySQLAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.MySQLAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLAutoscaler object but got %T", newObj)
	}

	return nil, in.validate(scaler)
}

func (_ MySQLAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *MySQLAutoscalerCustomWebhook) validate(scaler *autoscalingapi.MySQLAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}
