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

// SetupNeo4jAutoscalerWebhookWithManager registers the webhook for Neo4jAutoscaler in the manager.
func SetupNeo4jAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.Neo4jAutoscaler{}).
		WithValidator(&Neo4jAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&Neo4jAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type Neo4jAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var neoLog = logf.Log.WithName("neo4j-autoscaler")

var _ webhook.CustomDefaulter = &Neo4jAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *Neo4jAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.Neo4jAutoscaler)
	if !ok {
		return fmt.Errorf("expected an Neo4jAutoscaler object but got %T", obj)
	}
	neoLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *Neo4jAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.Neo4jAutoscaler) {
	var db olddbapi.Neo4j
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		klog.Errorf("can't get Neo4j %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Neo4j)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Neo4j)
	}
}

func (w *Neo4jAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.Neo4jAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &Neo4jAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *Neo4jAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.Neo4jAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jAutoscaler object but got %T", obj)
	}
	neoLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *Neo4jAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.Neo4jAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an Neo4jAutoscaler object but got %T", newObj)
	}
	neoLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *Neo4jAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *Neo4jAutoscalerCustomWebhook) validate(scaler *autoscalingapi.Neo4jAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.Neo4j
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		klog.Errorf("can't get Neo4j %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
