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

// SetupPgpoolAutoscalerWebhookWithManager registers the webhook for PgpoolAutoscaler in the manager.
func SetupPgpoolAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.PgpoolAutoscaler{}).
		WithValidator(&PgpoolAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&PgpoolAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PgpoolAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pgpoolLog = logf.Log.WithName("pgpool-autoscaler")

var _ webhook.CustomDefaulter = &PgpoolAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (in *PgpoolAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.PgpoolAutoscaler)
	if !ok {
		return fmt.Errorf("expected an PgpoolAutoscaler object but got %T", obj)
	}
	pgpoolLog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *PgpoolAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.PgpoolAutoscaler) {
	var db olddbapi.Pgpool
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get Pgpool %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Pgpool)
	}
}

func (in *PgpoolAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.PgpoolAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.PgpoolOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &PgpoolAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PgpoolAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.PgpoolAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PgpoolAutoscaler object but got %T", obj)
	}
	pgpoolLog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PgpoolAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.PgpoolAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an PgpoolAutoscaler object but got %T", newObj)
	}
	pgpoolLog.Info("validate update", "name", scaler.Name)
	return nil, in.validate(scaler)
}

func (in *PgpoolAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *PgpoolAutoscalerCustomWebhook) validate(scaler *autoscalingapi.PgpoolAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var pp olddbapi.Pgpool
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &pp)
	if err != nil {
		_ = fmt.Errorf("can't get Pgpool %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
