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

// log is for logging in this package.
var kafkaLog = logf.Log.WithName("kafka-autoscaler")

// SetupKafkaAutoscalerWebhookWithManager registers the webhook for KafkaAutoscaler in the manager.
func SetupKafkaAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.KafkaAutoscaler{}).
		WithValidator(&KafkaAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&KafkaAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type KafkaAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &KafkaAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (k *KafkaAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.KafkaAutoscaler)
	if !ok {
		return fmt.Errorf("expected an KafkaAutoscaler object but got %T", obj)
	}
	kafkaLog.Info("defaulting", "name", scaler.Name)
	k.setDefaults(scaler)
	return nil
}

func (k *KafkaAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.KafkaAutoscaler) {
	var db dbapi.Kafka
	err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get Kafka %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	k.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		if db.Spec.Topology != nil {
			setDefaultStorageValues(scaler.Spec.Storage.Broker)
			setDefaultStorageValues(scaler.Spec.Storage.Controller)
		} else {
			setDefaultStorageValues(scaler.Spec.Storage.Node)
		}
	}

	if scaler.Spec.Compute != nil {
		if db.Spec.Topology != nil {
			setDefaultComputeValues(scaler.Spec.Compute.Broker)
			setDefaultComputeValues(scaler.Spec.Compute.Controller)
		} else {
			setDefaultComputeValues(scaler.Spec.Compute.Node)
		}
	}
}

func (k *KafkaAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.KafkaAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.KafkaOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &KafkaAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KafkaAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.KafkaAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaAutoscaler object but got %T", obj)
	}
	kafkaLog.Info("validate create", "name", scaler.Name)
	return nil, k.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KafkaAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.KafkaAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaAutoscaler object but got %T", newObj)
	}
	kafkaLog.Info("validate create", "name", scaler.Name)
	return nil, k.validate(scaler)
}

func (_ *KafkaAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*autoscalingapi.KafkaAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an KafkaAutoscaler object but got %T", obj)
	}
	return nil, nil
}

func (k *KafkaAutoscalerCustomWebhook) validate(scaler *autoscalingapi.KafkaAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf dbapi.Kafka
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &kf)
	if err != nil {
		_ = fmt.Errorf("can't get Kafka %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if kf.Spec.Topology != nil {
			if cm.Node != nil {
				return errors.New("Spec.Compute.Node is invalid for kafka with topology")
			}
		} else {
			if cm.Broker != nil {
				return errors.New("Spec.Compute.Broker is invalid for combined kafka")
			}
			if cm.Controller != nil {
				return errors.New("Spec.Compute.Controller is invalid for combined kafka")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		st := scaler.Spec.Storage
		if kf.Spec.Topology != nil {
			if st.Node != nil {
				return errors.New("Spec.Storage.Node is invalid for kafka with topology")
			}
		} else {
			if st.Broker != nil {
				return errors.New("Spec.Storage.Broker is invalid for combined kafka")
			}
			if st.Controller != nil {
				return errors.New("Spec.Storage.Controller is invalid for combined kafka")
			}
		}
	}

	return nil
}
