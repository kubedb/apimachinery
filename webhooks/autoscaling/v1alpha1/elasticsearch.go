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

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var esLog = logf.Log.WithName("elasticsearch-autoscaler")

// SetupElasticsearchAutoscalerWebhookWithManager registers the webhook for ElasticsearchAutoscaler in the manager.
func SetupElasticsearchAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.ElasticsearchAutoscaler{}).
		WithValidator(&ElasticsearchAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&ElasticsearchAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ElasticsearchAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-elasticsearchautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=elasticsearchautoscaler,verbs=create;update,versions=v1alpha1,name=melasticsearchautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &ElasticsearchAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ElasticsearchAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.ElasticsearchAutoscaler)
	if !ok {
		return fmt.Errorf("expected an MariaDBAutoscaler object but got %T", obj)
	}

	esLog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *ElasticsearchAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.ElasticsearchAutoscaler) {
	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Node)
		setDefaultStorageValues(scaler.Spec.Storage.Master)
		setDefaultStorageValues(scaler.Spec.Storage.Data)
		setDefaultStorageValues(scaler.Spec.Storage.Ingest)
		setDefaultStorageValues(scaler.Spec.Storage.DataContent)
		setDefaultStorageValues(scaler.Spec.Storage.DataCold)
		setDefaultStorageValues(scaler.Spec.Storage.DataWarm)
		setDefaultStorageValues(scaler.Spec.Storage.DataFrozen)
		setDefaultStorageValues(scaler.Spec.Storage.DataHot)
		setDefaultStorageValues(scaler.Spec.Storage.ML)
		setDefaultStorageValues(scaler.Spec.Storage.Transform)
		setDefaultStorageValues(scaler.Spec.Storage.Coordinating)
	}
	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Node)
		setDefaultComputeValues(scaler.Spec.Compute.Master)
		setDefaultComputeValues(scaler.Spec.Compute.Data)
		setDefaultComputeValues(scaler.Spec.Compute.Ingest)
		setDefaultComputeValues(scaler.Spec.Compute.DataContent)
		setDefaultComputeValues(scaler.Spec.Compute.DataCold)
		setDefaultComputeValues(scaler.Spec.Compute.DataWarm)
		setDefaultComputeValues(scaler.Spec.Compute.DataFrozen)
		setDefaultComputeValues(scaler.Spec.Compute.DataHot)
		setDefaultComputeValues(scaler.Spec.Compute.ML)
		setDefaultComputeValues(scaler.Spec.Compute.Transform)
		setDefaultComputeValues(scaler.Spec.Compute.Coordinating)
	}
}

func (in *ElasticsearchAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.ElasticsearchAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.ElasticsearchOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-elasticsearchautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=elasticsearchautoscalers,verbs=create;update;delete,versions=v1alpha1,name=velasticsearchautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &ElasticsearchAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ElasticsearchAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.ElasticsearchAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ElasticsearchAutoscaler object but got %T", obj)
	}

	esLog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ElasticsearchAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.ElasticsearchAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an ElasticsearchAutoscaler object but got %T", newObj)
	}

	return nil, in.validate(scaler)
}

func (_ ElasticsearchAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *ElasticsearchAutoscalerCustomWebhook) validate(scaler *autoscalingapi.ElasticsearchAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}

	var es dbapi.Elasticsearch
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &es)
	if err != nil {
		_ = fmt.Errorf("can't get Elasticsearch %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if es.Spec.Topology != nil {
			if cm.Node != nil {
				return errors.New("Spec.Compute.PodResources is invalid for elastic-search topology")
			}
		} else {
			if cm.Master != nil || cm.Data != nil || cm.Ingest != nil || cm.DataContent != nil || cm.DataCold != nil || cm.DataFrozen != nil ||
				cm.DataWarm != nil || cm.DataHot != nil || cm.ML != nil || cm.Transform != nil || cm.Coordinating != nil {
				return errors.New("only Spec.Compute.Node is valid for basic elastic search structure")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		st := scaler.Spec.Storage
		if es.Spec.Topology != nil {
			if st.Node != nil {
				return errors.New("Spec.Storage.PodResources is invalid for elastic-search topology")
			}
		} else {
			if st.Master != nil || st.Data != nil || st.Ingest != nil || st.DataContent != nil || st.DataCold != nil || st.DataFrozen != nil ||
				st.DataWarm != nil || st.DataHot != nil || st.ML != nil || st.Transform != nil || st.Coordinating != nil {
				return errors.New("only Spec.Storage.Node is valid for basic elastic search structure")
			}
		}
	}
	return nil
}
