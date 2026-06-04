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

// SetupMilvusAutoscalerWebhookWithManager registers the webhook for MilvusAutoscaler in the manager.
func SetupMilvusAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.MilvusAutoscaler{}).
		WithValidator(&MilvusAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MilvusAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MilvusAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var milvusLog = logf.Log.WithName("milvus-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-milvusautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=milvusautoscaler,verbs=create;update,versions=v1alpha1,name=mmilvusautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &MilvusAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *MilvusAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.MilvusAutoscaler)
	if !ok {
		return fmt.Errorf("expected an MilvusAutoscaler object but got %T", obj)
	}

	milvusLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *MilvusAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.MilvusAutoscaler) {
	var db olddbapi.Milvus
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		klog.Errorf("can't get Milvus %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		if db.Spec.Topology.Distributed.StreamingNode != nil && scaler.Spec.Storage.StreamingNode != nil {
			setDefaultStorageValues(scaler.Spec.Storage.StreamingNode)
		}
	}

	if scaler.Spec.Compute != nil {
		if db.Spec.Topology != nil {
			if db.Spec.Topology.Distributed.Proxy != nil && scaler.Spec.Compute.Proxy != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Proxy)
			}
			if db.Spec.Topology.Distributed.DataNode != nil && scaler.Spec.Compute.DataNode != nil {
				setDefaultComputeValues(scaler.Spec.Compute.DataNode)
			}
			if db.Spec.Topology.Distributed.QueryNode != nil && scaler.Spec.Compute.QueryNode != nil {
				setDefaultComputeValues(scaler.Spec.Compute.QueryNode)
			}
			if db.Spec.Topology.Distributed.MixCoord != nil && scaler.Spec.Compute.MixCoord != nil {
				setDefaultComputeValues(scaler.Spec.Compute.MixCoord)
			}
			if db.Spec.Topology.Distributed.StreamingNode != nil && scaler.Spec.Compute.StreamingNode != nil {
				setDefaultComputeValues(scaler.Spec.Compute.StreamingNode)
			}
		}
	}
}

func (w *MilvusAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.MilvusAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-milvusautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=milvusautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmilvusautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MilvusAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MilvusAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.MilvusAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MilvusAutoscaler object but got %T", obj)
	}

	milvusLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MilvusAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.MilvusAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MilvusAutoscaler object but got %T", newObj)
	}

	return nil, w.validate(scaler)
}

func (w MilvusAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *MilvusAutoscalerCustomWebhook) validate(scaler *autoscalingapi.MilvusAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var dr olddbapi.Milvus
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &dr)
	if err != nil {
		klog.Errorf("can't get Milvus %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if dr.Spec.Topology != nil {
			if cm.Proxy == nil && cm.DataNode == nil && cm.QueryNode == nil && cm.MixCoord == nil && cm.StreamingNode == nil {
				return errors.New("Spec.Compute.Proxy, Spec.Compute.Datanode, Spec.Compute.Querynode, Spec.Compute.Mixcoord, Spec.Compute.Streamingnode all are empty")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		if dr.Spec.Topology != nil {
			if scaler.Spec.Storage.StreamingNode == nil {
				return errors.New("Spec.Storage.Streamingnode is empty")
			}
		}
	}
	return nil
}
