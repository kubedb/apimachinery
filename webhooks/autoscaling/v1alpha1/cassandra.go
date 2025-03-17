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
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupCassandraAutoscalerWebhookWithManager registers the webhook for CassandraAutoscaler in the manager.
func SetupCassandraAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.CassandraAutoscaler{}).
		WithValidator(&CassandraAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&CassandraAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type CassandraAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var casLog = logf.Log.WithName("cassandra-autoscaler")

var _ webhook.CustomDefaulter = &CassandraAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *CassandraAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.CassandraAutoscaler)
	if !ok {
		return fmt.Errorf("expected an CassandraAutoscaler object but got %T", obj)
	}
	casLog.Info("defaulting", "name", scaler.Name)
	r.setDefaults(scaler)
	return nil
}

func (r *CassandraAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.CassandraAutoscaler) {
	var db olddbapi.Cassandra
	err := r.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get Cassandra %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	r.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Cassandra)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Cassandra)
	}
}

func (r *CassandraAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.CassandraAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.CassandraOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &CassandraAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CassandraAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.CassandraAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraAutoscaler object but got %T", obj)
	}
	casLog.Info("validate create", "name", scaler.Name)
	return nil, r.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CassandraAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.CassandraAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraAutoscaler object but got %T", newObj)
	}
	casLog.Info("validate create", "name", scaler.Name)
	return nil, r.validate(scaler)
}

func (r *CassandraAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *CassandraAutoscalerCustomWebhook) validate(scaler *autoscalingapi.CassandraAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.Cassandra
	err := r.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		_ = fmt.Errorf("can't get Cassandra %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
