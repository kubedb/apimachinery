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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupRedisAutoscalerWebhookWithManager registers the webhook for RedisAutoscaler in the manager.
func SetupRedisAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.RedisAutoscaler{}).
		WithValidator(&RedisAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&RedisAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RedisAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var redisLog = logf.Log.WithName("redis-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-redisautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=redisautoscaler,verbs=create;update,versions=v1alpha1,name=mredisautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &RedisAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *RedisAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.RedisAutoscaler)
	if !ok {
		return fmt.Errorf("expected an RedisAutoscaler object but got %T", obj)
	}
	redisLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *RedisAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.RedisAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Standalone)
		setDefaultStorageValues(scaler.Spec.Storage.Cluster)
		setDefaultStorageValues(scaler.Spec.Storage.Sentinel)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Standalone)
		setDefaultComputeValues(scaler.Spec.Compute.Cluster)
		setDefaultComputeValues(scaler.Spec.Compute.Sentinel)
	}
}

func (w *RedisAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.RedisAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.RedisOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s w ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-redisautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=redisautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vredisautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &RedisAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *RedisAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.RedisAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an RedisAutoscaler object but got %T", obj)
	}
	redisLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *RedisAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.RedisAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an RedisAutoscaler object but got %T", newObj)
	}
	redisLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w RedisAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *RedisAutoscalerCustomWebhook) validate(scaler *autoscalingapi.RedisAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}

	var rd dbapi.Redis
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &rd)
	if err != nil {
		_ = fmt.Errorf("can't get Redis %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if rd.Spec.Mode == dbapi.RedisModeCluster {
			if cm.Standalone != nil {
				return errors.New("Spec.Compute.Standalone is invalid for clustered redis")
			}
			if cm.Sentinel != nil {
				return errors.New("Spec.Compute.Sentinel is invalid for clustered redis")
			}
		} else if rd.Spec.Mode == dbapi.RedisModeSentinel {
			if cm.Standalone != nil {
				return errors.New("Spec.Compute.Standalone is invalid for redis sentinel")
			}
			if cm.Cluster != nil {
				return errors.New("Spec.Compute.Cluster is invalid for redis sentinel")
			}
		} else if rd.Spec.Mode == dbapi.RedisModeStandalone {
			if cm.Cluster != nil {
				return errors.New("Spec.Compute.Cluster is invalid for standalone redis")
			}
			if cm.Cluster != nil {
				return errors.New("Spec.Compute.Sentinel is invalid for standalone redis")
			}

		}
	}

	if scaler.Spec.Storage != nil {
		st := scaler.Spec.Storage
		if rd.Spec.Mode == dbapi.RedisModeCluster {
			if st.Standalone != nil {
				return errors.New("Spec.Storage.Standalone is invalid for clustered redis")
			}
			if st.Sentinel != nil {
				return errors.New("Spec.Storage.Sentinel is invalid for clustered redis")
			}
		} else if rd.Spec.Mode == dbapi.RedisModeSentinel {
			if st.Standalone != nil {
				return errors.New("Spec.Storage.Standalone is invalid for redis sentinel")
			}
			if st.Cluster != nil {
				return errors.New("Spec.Storage.Cluster is invalid for redis sentinel")
			}
		} else if rd.Spec.Mode == dbapi.RedisModeStandalone {
			if st.Cluster != nil {
				return errors.New("Spec.Storage.Cluster is invalid for standalone redis")
			}
			if st.Sentinel != nil {
				return errors.New("Spec.Storage.Sentinel is invalid for standalone redis")
			}
		}
	}
	return nil
}
