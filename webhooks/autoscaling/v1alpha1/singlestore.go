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

// SetupSinglestoreAutoscalerWebhookWithManager registers the webhook for SinglestoreAutoscaler in the manager.
func SetupSinglestoreAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.SinglestoreAutoscaler{}).
		WithValidator(&SinglestoreAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&SinglestoreAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type SinglestoreAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

var sdbLog = logf.Log.WithName("singlestore-autoscaler")

// log is for logging in this package.
var singlestoreLog = logf.Log.WithName("singlestore-autoscaler")

var _ webhook.CustomDefaulter = &SinglestoreAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (s *SinglestoreAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.SinglestoreAutoscaler)
	if !ok {
		return fmt.Errorf("expected an SinglestoreAutoscaler object but got %T", obj)
	}
	singlestoreLog.Info("defaulting", "name", scaler.GetName())
	s.setDefaults(scaler)
	return nil
}

func (s *SinglestoreAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.SinglestoreAutoscaler) {
	var db olddbapi.Singlestore
	err := s.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get Singlestore %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	s.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		if db.Spec.Topology != nil {
			setDefaultStorageValues(scaler.Spec.Storage.Aggregator)
			setDefaultStorageValues(scaler.Spec.Storage.Leaf)
		} else {
			setDefaultStorageValues(scaler.Spec.Storage.Node)
		}
	}

	if scaler.Spec.Compute != nil {
		if db.Spec.Topology != nil {
			setDefaultComputeValues(scaler.Spec.Compute.Aggregator)
			setDefaultComputeValues(scaler.Spec.Compute.Leaf)
		} else {
			setDefaultComputeValues(scaler.Spec.Compute.Node)
		}
	}
}

func (s *SinglestoreAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.SinglestoreAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.SinglestoreOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &SinglestoreAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (s *SinglestoreAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.SinglestoreAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an SinglestoreAutoscaler object but got %T", obj)
	}
	sdbLog.Info("validate create", "name", scaler.Name)
	return nil, s.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (s *SinglestoreAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.SinglestoreAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an SinglestoreAutoscaler object but got %T", newObj)
	}
	sdbLog.Info("validate update", "name", scaler.Name)
	return nil, s.validate(scaler)
}

func (_ *SinglestoreAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (s *SinglestoreAutoscalerCustomWebhook) validate(scaler *autoscalingapi.SinglestoreAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var sdb olddbapi.Singlestore
	err := s.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &sdb)
	if err != nil {
		_ = fmt.Errorf("can't get Singlestore %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if sdb.Spec.Topology != nil {
			if cm.Node != nil {
				return errors.New("Spec.Compute.Node is invalid for singlestore with cluster")
			}
		} else {
			if cm.Aggregator != nil {
				return errors.New("Spec.Compute.Aggregator is invalid for standalone")
			}
			if cm.Leaf != nil {
				return errors.New("Spec.Compute.Leaf is invalid for combined standalone")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		st := scaler.Spec.Storage
		if sdb.Spec.Topology != nil {
			if st.Node != nil {
				return errors.New("Spec.Storage.Node is invalid for Singlestore with cluster")
			}
		} else {
			if st.Aggregator != nil {
				return errors.New("Spec.Storage.Aggregator is invalid for standalone")
			}
			if st.Leaf != nil {
				return errors.New("Spec.Storage.Leaf is invalid for standalone")
			}
		}
	}

	return nil
}
