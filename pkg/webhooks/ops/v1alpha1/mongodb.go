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
	"strconv"
	"strings"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gomodules.xyz/x/arrays"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMongoDBOpsRequestWebhookWithManager registers the webhook for MongoDBOpsRequest in the manager.
func SetupMongoDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MongoDBOpsRequest{}).
		WithValidator(&MongoDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MongoDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mongodbLog = logf.Log.WithName("mongodb-opsrequest")

var _ webhook.CustomValidator = &MongoDBOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MongoDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MongoDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MongoDBOpsRequest object but got %T", obj)
	}
	mongodbLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MongoDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MongoDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MongoDBOpsRequest object but got %T", newObj)
	}
	mongodbLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MongoDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MongoDBOpsRequest object but got %T", oldObj)
	}

	if err := validateMongoDBOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *MongoDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *MongoDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MongoDBOpsRequest) error {
	var allErr field.ErrorList
	// validate MongoDBOpsRequest specs
	if !IsOpsTypeSupported(opsapi.MongoDBOpsRequestTypeNames(), string(req.Spec.Type)) {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MongoDB are %s", req.Spec.Type, strings.Join(opsapi.MongoDBOpsRequestTypeNames(), ", "))))
	}

	var db dbapi.MongoDB
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &db)
	if err != nil && !kerr.IsNotFound(err) {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name,
			fmt.Sprintf("referenced database %s/%s is not found", req.Namespace, req.Spec.DatabaseRef.Name)))
	}

	if req.Spec.Type == opsapi.MongoDBOpsRequestTypeHorizontalScaling {
		if err = w.validateMongoDBHorizontalScalingOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec"),
				req.Name,
				err.Error()))
		}
	}
	if req.Spec.Type == opsapi.MongoDBOpsRequestTypeHorizons {
		if err = w.validateMongoDBHorizons(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec"),
				req.Name,
				err.Error()))
		}
	}
	if validType, _ := arrays.Contains(opsapi.MongoDBOpsRequestTypeNames(), req.Spec.Type); !validType {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MongoDB are %s", req.Spec.Type, strings.Join(opsapi.MongoDBOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MongoDBopsrequests.kubedb.com", Kind: "MongoDBOpsRequest"}, req.Name, allErr)
}

func (w *MongoDBOpsRequestCustomWebhook) validateMongoDBHorizontalScalingOpsRequest(db *dbapi.MongoDB, req *opsapi.MongoDBOpsRequest) error {
	if req.Spec.HorizontalScaling == nil {
		return errors.New("`spec.horizontalScaling` field is nil")
	}

	if req.Spec.HorizontalScaling.Replicas != nil {
		if req.Spec.HorizontalScaling.Shard != nil || req.Spec.HorizontalScaling.ConfigServer != nil || req.Spec.HorizontalScaling.Mongos != nil {
			return errors.New("shard, configServer or mongos field can not be used when replicas field is already set")
		}
	}
	if req.Spec.HorizontalScaling.Replicas != nil && db.Spec.ReplicaSet == nil {
		return errors.New("replicas field can't be set on a non-replica mongoDB")
	}
	if (req.Spec.HorizontalScaling.Shard != nil || req.Spec.HorizontalScaling.ConfigServer != nil || req.Spec.HorizontalScaling.Mongos != nil) && db.Spec.ShardTopology == nil {
		return errors.New("shard, configServer or mongos field can't be set on a non-sharded mongoDB")
	}

	// count=0 related validation starts

	if req.Spec.HorizontalScaling.Replicas != nil && *req.Spec.HorizontalScaling.Replicas == 0 {
		return errors.New("replicas count can not be 0")
	}
	if req.Spec.HorizontalScaling.Shard != nil {
		// TODO : if user tries to scale only one of Replicas & Shard, the other value can be by default 0 (null value of int), add This to mutator
		if req.Spec.HorizontalScaling.Shard.Replicas == 0 || req.Spec.HorizontalScaling.Shard.Shards == 0 {
			return errors.New("replicas or shard count can not be 0 in Sharded MongoDB")
		}
	}
	if req.Spec.HorizontalScaling.ConfigServer != nil && req.Spec.HorizontalScaling.ConfigServer.Replicas == 0 {
		return errors.New("configServer replicas count can not be 0")
	}
	if req.Spec.HorizontalScaling.Mongos != nil && req.Spec.HorizontalScaling.Mongos.Replicas == 0 {
		return errors.New("mongos count can not be 0")
	}

	if db.Spec.StorageEngine == dbapi.StorageEngineInMemory {
		if db.Spec.ShardTopology != nil {
			if req.Spec.HorizontalScaling.Shard != nil && req.Spec.HorizontalScaling.Shard.Replicas > 0 && req.Spec.HorizontalScaling.Shard.Replicas < 3 {
				return errors.New("`spec.horizontalScaling.shard.replicas` field can not be less then 3 for inMemory storage engine")
			}
			if req.Spec.HorizontalScaling.ConfigServer != nil && req.Spec.HorizontalScaling.ConfigServer.Replicas < 3 {
				return errors.New("`spec.horizontalScaling.configServer.replicas` field can not be less then 3 for inMemory storage engine")
			}
		} else if db.Spec.ReplicaSet != nil {
			if req.Spec.HorizontalScaling.Replicas != nil && *req.Spec.HorizontalScaling.Replicas < 3 {
				return errors.New("`spec.horizontalScaling.replicas` field can not be less then 3 for inMemory storage engine")
			}
		}
	}
	return nil
}

func (w *MongoDBOpsRequestCustomWebhook) validateMongoDBHorizons(db *dbapi.MongoDB, req *opsapi.MongoDBOpsRequest) error {
	if db.Spec.ReplicaSet == nil {
		return errors.New("horizon opsRequest is only supported for ReplicaSet")
	}
	if req.Spec.Horizons != nil {
		if db.Spec.TLS == nil {
			return errors.New("horizon opsRequest is only supported for TLS")
		}
		if len(req.Spec.Horizons.Pods) != int(*db.Spec.Replicas) {
			return errors.New("the length of ops.spec.horizons.pods has to be " + strconv.Itoa(int(*db.Spec.Replicas)))
		}
	}
	return nil
}

func validateMongoDBOpsRequest(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func IsOpsTypeSupported(supportedTypes []string, curOpsType string) bool {
	for _, s := range supportedTypes {
		if s == curOpsType {
			return true
		}
	}
	return false
}
