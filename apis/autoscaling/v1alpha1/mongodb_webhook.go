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

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var mongoLog = logf.Log.WithName("mongodb-autoscaler")

func (in *MongoDBAutoscaler) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-mongodbautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=mongodbautoscaler,verbs=create;update,versions=v1alpha1,name=mmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &MongoDBAutoscaler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDBAutoscaler) Default(ctx context.Context, obj runtime.Object) error {
	mongoLog.Info("defaulting", "name", in.Name)
	in.setDefaults()
	return nil
}

func (in *MongoDBAutoscaler) setDefaults() {
	var db dbapi.MongoDB
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      in.Spec.DatabaseRef.Name,
		Namespace: in.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get MongoDB %s/%s \n", in.Namespace, in.Spec.DatabaseRef.Name)
		return
	}

	in.setOpsReqOptsDefaults()

	if in.Spec.Storage != nil {
		setDefaultStorageValues(in.Spec.Storage.Standalone)
		setDefaultStorageValues(in.Spec.Storage.ReplicaSet)
		setDefaultStorageValues(in.Spec.Storage.Shard)
		setDefaultStorageValues(in.Spec.Storage.ConfigServer)
		setDefaultStorageValues(in.Spec.Storage.Hidden)
	}

	if in.Spec.Compute != nil {
		setDefaultComputeValues(in.Spec.Compute.Standalone)
		setDefaultComputeValues(in.Spec.Compute.ReplicaSet)
		setDefaultComputeValues(in.Spec.Compute.Shard)
		setDefaultComputeValues(in.Spec.Compute.ConfigServer)
		setDefaultComputeValues(in.Spec.Compute.Mongos)
		setDefaultComputeValues(in.Spec.Compute.Arbiter)
		setDefaultComputeValues(in.Spec.Compute.Hidden)

		setInMemoryDefaults(in.Spec.Compute.Standalone, db.Spec.StorageEngine)
		setInMemoryDefaults(in.Spec.Compute.ReplicaSet, db.Spec.StorageEngine)
		setInMemoryDefaults(in.Spec.Compute.Shard, db.Spec.StorageEngine)
		setInMemoryDefaults(in.Spec.Compute.ConfigServer, db.Spec.StorageEngine)
		setInMemoryDefaults(in.Spec.Compute.Mongos, db.Spec.StorageEngine)
		// no need for Defaulting the Arbiter & Hidden PodResources.
		// As arbiter is not a data-node.  And hidden doesn't have the impact of storageEngine (it can't be InMemory).
	}
}

func (in *MongoDBAutoscaler) setOpsReqOptsDefaults() {
	if in.Spec.OpsRequestOptions == nil {
		in.Spec.OpsRequestOptions = &MongoDBOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if in.Spec.OpsRequestOptions.Apply == "" {
		in.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &MongoDBAutoscaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscaler) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	mongoLog.Info("validate create", "name", in.Name)
	return nil, in.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscaler) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	mongoLog.Info("validate update", "name", in.Name)
	return nil, in.validate()
}

func (_ MongoDBAutoscaler) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *MongoDBAutoscaler) validate() error {
	if in.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var mg dbapi.MongoDB
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      in.Spec.DatabaseRef.Name,
		Namespace: in.Namespace,
	}, &mg)
	if err != nil {
		_ = fmt.Errorf("can't get MongoDB %s/%s \n", in.Namespace, in.Spec.DatabaseRef.Name)
		return err
	}

	if in.Spec.Compute != nil {
		cm := in.Spec.Compute
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

	if in.Spec.Storage != nil {
		st := in.Spec.Storage
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
