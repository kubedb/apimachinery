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

// SetupDruidAutoscalerWebhookWithManager registers the webhook for DruidAutoscaler in the manager.
func SetupDruidAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.DruidAutoscaler{}).
		WithValidator(&DruidAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&DruidAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DruidAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var druidLog = logf.Log.WithName("druid-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-druidautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=druidautoscaler,verbs=create;update,versions=v1alpha1,name=mdruidautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &DruidAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *DruidAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.DruidAutoscaler)
	if !ok {
		return fmt.Errorf("expected an DruidAutoscaler object but got %T", obj)
	}

	druidLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *DruidAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.DruidAutoscaler) {
	var db olddbapi.Druid
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get Druid %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		if db.Spec.Topology.MiddleManagers != nil && scaler.Spec.Storage.MiddleManagers != nil {
			setDefaultStorageValues(scaler.Spec.Storage.MiddleManagers)
		}
		if db.Spec.Topology.Historicals != nil && scaler.Spec.Storage.Historicals != nil {
			setDefaultStorageValues(scaler.Spec.Storage.Historicals)
		}
	}

	if scaler.Spec.Compute != nil {
		if db.Spec.Topology != nil {
			if db.Spec.Topology.Coordinators != nil && scaler.Spec.Compute.Coordinators != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Coordinators)
			}
			if db.Spec.Topology.Overlords != nil && scaler.Spec.Compute.Overlords != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Overlords)
			}
			if db.Spec.Topology.MiddleManagers != nil && scaler.Spec.Compute.MiddleManagers != nil {
				setDefaultComputeValues(scaler.Spec.Compute.MiddleManagers)
			}
			if db.Spec.Topology.Historicals != nil && scaler.Spec.Compute.Historicals != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Historicals)
			}
			if db.Spec.Topology.Brokers != nil && scaler.Spec.Compute.Brokers != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Brokers)
			}
			if db.Spec.Topology.Routers != nil && scaler.Spec.Compute.Routers != nil {
				setDefaultComputeValues(scaler.Spec.Compute.Routers)
			}

		}
	}
}

func (w *DruidAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.DruidAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.DruidOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-druidautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=druidautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vdruidautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &DruidAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DruidAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.DruidAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an DruidAutoscaler object but got %T", obj)
	}

	druidLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DruidAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.DruidAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an DruidAutoscaler object but got %T", newObj)
	}

	return nil, w.validate(scaler)
}

func (w DruidAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *DruidAutoscalerCustomWebhook) validate(scaler *autoscalingapi.DruidAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var dr olddbapi.Druid
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &dr)
	if err != nil {
		_ = fmt.Errorf("can't get Druid %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if dr.Spec.Topology != nil {
			if cm.Coordinators == nil && cm.Overlords == nil && cm.MiddleManagers == nil && cm.Historicals == nil && cm.Brokers == nil && cm.Routers == nil {
				return errors.New("Spec.Compute.Coordinators, Spec.Compute.Overlords, Spec.Compute.MiddleManagers, Spec.Compute.Historicals, Spec.Compute.Brokers, Spec.Compute.Routers all are empty")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		if dr.Spec.Topology != nil {
			if scaler.Spec.Storage.MiddleManagers == nil && scaler.Spec.Storage.Historicals == nil {
				return errors.New("Spec.Storage.MiddleManagers and Spec.Storage.Historicals both are empty")
			}
		}
	}
	return nil
}
