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
	"errors"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mongoLog = logf.Log.WithName("mongodb-autoscaler")

func (in *MongoDBAutoscaler) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-mongodbautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=mongodbautoscaler,verbs=create;update,versions=v1alpha1,name=mmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &MongoDBAutoscaler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *MongoDBAutoscaler) Default() {
	mongoLog.Info("defaulting", "name", in.Name)
}

func (in *MongoDBAutoscaler) SetDefaults(db *dbapi.MongoDB) {
	if in.Spec.Storage != nil {
		setDefaultStorageValues(in.Spec.Storage.Standalone)
		setDefaultStorageValues(in.Spec.Storage.ReplicaSet)
		setDefaultStorageValues(in.Spec.Storage.Shard)
		setDefaultStorageValues(in.Spec.Storage.ConfigServer)
	}

	if in.Spec.Compute != nil {
		setDefaultComputeValues(in.Spec.Compute.Standalone)
		setInMemoryDefaults(in.Spec.Compute.Standalone, db.Spec.StorageEngine)
		setDefaultComputeValues(in.Spec.Compute.ReplicaSet)
		setInMemoryDefaults(in.Spec.Compute.ReplicaSet, db.Spec.StorageEngine)
		setDefaultComputeValues(in.Spec.Compute.Shard)
		setInMemoryDefaults(in.Spec.Compute.Shard, db.Spec.StorageEngine)
		setDefaultComputeValues(in.Spec.Compute.ConfigServer)
		setInMemoryDefaults(in.Spec.Compute.ConfigServer, db.Spec.StorageEngine)
		setDefaultComputeValues(in.Spec.Compute.Mongos)
		setInMemoryDefaults(in.Spec.Compute.Mongos, db.Spec.StorageEngine)
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-mongodbautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=mongodbautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vmongodbautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &MongoDBAutoscaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscaler) ValidateCreate() error {
	mongoLog.Info("validate create", "name", in.Name)
	return in.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MongoDBAutoscaler) ValidateUpdate(old runtime.Object) error {
	mongoLog.Info("validate create", "name", in.Name)
	return in.validate()
}

func (_ MongoDBAutoscaler) ValidateDelete() error {
	return nil
}

func (in *MongoDBAutoscaler) validate() error {
	if in.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	return nil
}

func (in *MongoDBAutoscaler) ValidateFields(mg *dbapi.MongoDB) error {
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
	}
	return nil
}
