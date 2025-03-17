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

// SetupMongoDBAutoscalerWebhookWithManager registers the webhook for MongoDBAutoscaler in the manager.
func SetupMongoDBAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.MongoDBAutoscaler{}).
		WithValidator(&MongoDBAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&MongoDBAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MongoDBAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mongoLog = logf.Log.WithName("mongodb-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-mongodbautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=mongodbautoscaler,verbs=create;update,versions=v1alpha1,name=mmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &MongoDBAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDBAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.MongoDBAutoscaler)
	if !ok {
		return fmt.Errorf("expected an MongoDBAutoscaler object but got %T", obj)
	}

	mongoLog.Info("defaulting", "name", scaler.Name)
	in.setDefaults(scaler)
	return nil
}

func (in *MongoDBAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.MongoDBAutoscaler) {
	var db dbapi.MongoDB
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get MongoDB %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return
	}

	in.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Standalone)
		setDefaultStorageValues(scaler.Spec.Storage.ReplicaSet)
		setDefaultStorageValues(scaler.Spec.Storage.Shard)
		setDefaultStorageValues(scaler.Spec.Storage.ConfigServer)
		setDefaultStorageValues(scaler.Spec.Storage.Hidden)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Standalone)
		setDefaultComputeValues(scaler.Spec.Compute.ReplicaSet)
		setDefaultComputeValues(scaler.Spec.Compute.Shard)
		setDefaultComputeValues(scaler.Spec.Compute.ConfigServer)
		setDefaultComputeValues(scaler.Spec.Compute.Mongos)
		setDefaultComputeValues(scaler.Spec.Compute.Arbiter)
		setDefaultComputeValues(scaler.Spec.Compute.Hidden)

		setInMemoryDefaults(scaler.Spec.Compute.Standalone, db.Spec.StorageEngine)
		setInMemoryDefaults(scaler.Spec.Compute.ReplicaSet, db.Spec.StorageEngine)
		setInMemoryDefaults(scaler.Spec.Compute.Shard, db.Spec.StorageEngine)
		setInMemoryDefaults(scaler.Spec.Compute.ConfigServer, db.Spec.StorageEngine)
		setInMemoryDefaults(scaler.Spec.Compute.Mongos, db.Spec.StorageEngine)
		// no need for Defaulting the Arbiter & Hidden PodResources.
		// As arbiter is not a data-node.  And hidden doesn't have the impact of storageEngine (it can't be InMemory).
	}
}

func (in *MongoDBAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.MongoDBAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.MongoDBOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if scaler.Spec.OpsRequestOptions.Apply == "" {
		scaler.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MongoDBAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.MongoDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MongoDBAutoscaler object but got %T", obj)
	}

	mongoLog.Info("validate create", "name", scaler.Name)
	return nil, in.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.MongoDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an MongoDBAutoscaler object but got %T", newObj)
	}

	return nil, in.validate(scaler)
}

func (_ MongoDBAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *MongoDBAutoscalerCustomWebhook) validate(scaler *autoscalingapi.MongoDBAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var mg dbapi.MongoDB
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &mg)
	if err != nil {
		_ = fmt.Errorf("can't get MongoDB %s/%s \n", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}

	if scaler.Spec.Compute != nil {
		cm := scaler.Spec.Compute
		if mg.Spec.ShardTopology != nil {
			if cm.ReplicaSet != nil {
				return errors.New("Spec.Compute.ReplicaSet is invalid for sharded mongoDB")
			}
			if cm.Standalone != nil {
				return errors.New("Spec.Compute.Standalone is invalid for sharded mongoDB")
			}
		} else if mg.Spec.ReplicaSet != nil {
			if cm.Standalone != nil {
				return errors.New("Spec.Compute.Standalone is invalid for replicaSet mongoDB")
			}
			if cm.Shard != nil {
				return errors.New("Spec.Compute.Shard is invalid for replicaSet mongoDB")
			}
			if cm.ConfigServer != nil {
				return errors.New("Spec.Compute.ConfigServer is invalid for replicaSet mongoDB")
			}
			if cm.Mongos != nil {
				return errors.New("Spec.Compute.Mongos is invalid for replicaSet mongoDB")
			}
		} else {
			if cm.ReplicaSet != nil {
				return errors.New("Spec.Compute.Replicaset is invalid for Standalone mongoDB")
			}
			if cm.Shard != nil {
				return errors.New("Spec.Compute.Shard is invalid for Standalone mongoDB")
			}
			if cm.ConfigServer != nil {
				return errors.New("Spec.Compute.ConfigServer is invalid for Standalone mongoDB")
			}
			if cm.Mongos != nil {
				return errors.New("Spec.Compute.Mongos is invalid for Standalone mongoDB")
			}
			if cm.Arbiter != nil {
				return errors.New("Spec.Compute.Arbiter is invalid for Standalone mongoDB")
			}
			if cm.Hidden != nil {
				return errors.New("Spec.Compute.Hidden is invalid for Standalone mongoDB")
			}
		}
	}

	if scaler.Spec.Storage != nil {
		st := scaler.Spec.Storage
		if mg.Spec.ShardTopology != nil {
			if st.ReplicaSet != nil {
				return errors.New("Spec.Storage.ReplicaSet is invalid for sharded mongoDB")
			}
			if st.Standalone != nil {
				return errors.New("Spec.Storage.Standalone is invalid for sharded mongoDB")
			}

			if err = validateScalingRules(st.ConfigServer); err != nil {
				return err
			}
			if err = validateScalingRules(st.Shard); err != nil {
				return err
			}

		} else if mg.Spec.ReplicaSet != nil {
			if st.Standalone != nil {
				return errors.New("Spec.Storage.Standalone is invalid for replicaSet mongoDB")
			}
			if st.Shard != nil {
				return errors.New("Spec.Storage.Shard is invalid for replicaSet mongoDB")
			}
			if st.ConfigServer != nil {
				return errors.New("Spec.Storage.ConfigServer is invalid for replicaSet mongoDB")
			}

			if err = validateScalingRules(st.ReplicaSet); err != nil {
				return err
			}
		} else {
			if st.ReplicaSet != nil {
				return errors.New("Spec.Storage.Replicaset is invalid for Standalone mongoDB")
			}
			if st.Shard != nil {
				return errors.New("Spec.Storage.Shard is invalid for Standalone mongoDB")
			}
			if st.ConfigServer != nil {
				return errors.New("Spec.Storage.ConfigServer is invalid for Standalone mongoDB")
			}
			if st.Hidden != nil {
				return errors.New("Spec.Storage.Hidden is invalid for Standalone mongoDB")
			}

			if err = validateScalingRules(st.Standalone); err != nil {
				return err
			}
		}

		if mg.Spec.Hidden != nil {
			if err = validateScalingRules(st.Hidden); err != nil {
				return err
			}
		}
	}

	return nil
}
