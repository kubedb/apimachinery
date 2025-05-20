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
	"fmt"
	core "k8s.io/api/core/v1"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// SetupRedisOpsRequestWebhookWithManager registers the webhook for RedisOpsRequest in the manager.
func SetupRedisOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.RedisOpsRequest{}).
		WithValidator(&RedisOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RedisOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var redisLog = logf.Log.WithName("redis-opsrequest")

var _ webhook.CustomValidator = &RedisOpsRequestCustomWebhook{}

func (w *RedisOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", obj)
	}

	redisLog.Info("validate create", "name", req.Name)
	err := w.isDatabaseRefValid(req)
	if err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(req)
}

func (w *RedisOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newReq, ok := newObj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.RedisOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisOpsRequest object but got %T", oldObj)
	}

	if err := validateRedisOpsRequest(newReq, oldReq); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	err := w.isDatabaseRefValid(newReq)
	if err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(newReq)
}

func (w *RedisOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateRedisOpsRequest(obj, oldObj runtime.Object) error {
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

func (w *RedisOpsRequestCustomWebhook) isDatabaseRefValid(req *opsapi.RedisOpsRequest) error {
	redis := &v1.Redis{ObjectMeta: metav1.ObjectMeta{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}}
	return w.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(redis), redis)
}

func (w *RedisOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.RedisOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.RedisOpsRequestType) {
	case opsapi.RedisOpsRequestTypeHorizontalScaling:
		if err := w.validateRedisHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeUpdateVersion:
		if err := w.validateRedisUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisOpsRequestTypeRotateAuth:
		if err := w.validateRedisRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("rotateauth"),
				req.Name,
				err.Error()))
		}
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Redisopsrequests.kubedb.com", Kind: "RedisOpsRequest"}, req.Name, allErr)
}

func (w *RedisOpsRequestCustomWebhook) validateRedisUpdateVersionOpsRequest(req *opsapi.RedisOpsRequest) error {
	updatedVersionName := req.Spec.UpdateVersion.TargetVersion
	redis := v1.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}
	currentVersionName := redis.Spec.Version

	updatedVersion := &v1alpha1.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updatedVersionName}, updatedVersion); err != nil {
		return err
	}

	currentVersion := &v1alpha1.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: currentVersionName}, currentVersion); err != nil {
		return err
	}

	if redis.Spec.Mode == v1.RedisModeSentinel {
		currentSentinel := &v1.RedisSentinel{}
		if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: redis.Spec.SentinelRef.Name, Namespace: redis.Spec.SentinelRef.Namespace}, currentSentinel); err != nil {
			return err
		}
		if currentSentinel.Spec.Version != updatedVersion.Spec.Version {
			return fmt.Errorf("RedisOpsRequest %s/%s: can't upgrade to a different version of sentinel", req.Namespace, req.Name)
		}
	}

	if updatedVersion.Spec.Distribution != currentVersion.Spec.Distribution {
		updatedSemver, err := semver.NewVersion(updatedVersion.Spec.Version)
		if err != nil {
			return err
		}
		currentSemver, err := semver.NewVersion(currentVersion.Spec.Version)
		if err != nil {
			return err
		}
		if updatedSemver.Major() != currentSemver.Major() {
			return fmt.Errorf("RedisOpsRequest %s/%s: can't upgrade Official to Valkey with a different major version", req.Namespace, req.Name)
		}
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisHorizontalScalingOpsRequest(req *opsapi.RedisOpsRequest) error {
	redis := v1.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}
	if redis.Spec.Mode == v1.RedisModeStandalone {
		return fmt.Errorf("horizontal scaling is not allowed for standalone redis")
	} else if redis.Spec.Mode == v1.RedisModeCluster {
		return w.checkHorizontalOpsReqForClusterMode(req)
	} else {
		return w.checkHorizontalOpsReqForSentinelMode(req)
	}
}

func (w *RedisOpsRequestCustomWebhook) checkHorizontalOpsReqForClusterMode(req *opsapi.RedisOpsRequest) error {
	if req.Spec.HorizontalScaling.Shards != nil {
		if *req.Spec.HorizontalScaling.Shards < 3 {
			return fmt.Errorf("shards should be >= 3 in cluster mode")
		}
	}
	if req.Spec.HorizontalScaling.Replicas != nil {
		if *req.Spec.HorizontalScaling.Replicas < 2 {
			return fmt.Errorf("replica should be >= 2 in cluster mode")
		}
	}
	if req.Spec.HorizontalScaling.Shards == nil && req.Spec.HorizontalScaling.Replicas == nil {
		return fmt.Errorf("please specify shards or replica for scaling")
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) checkHorizontalOpsReqForSentinelMode(req *opsapi.RedisOpsRequest) error {
	if req.Spec.HorizontalScaling.Shards != nil {
		return fmt.Errorf("master count should not be specified for Sentinel mode")
	}
	if req.Spec.HorizontalScaling.Replicas != nil {
		if *req.Spec.HorizontalScaling.Replicas < 1 {
			return fmt.Errorf("replica should be >= 1")
		}
	} else {
		return fmt.Errorf("pleaase specify replica for scaling")
	}
	return nil
}

func (w *RedisOpsRequestCustomWebhook) validateRedisRotateAuthenticationOpsRequest(req *opsapi.RedisOpsRequest) error {
	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return fmt.Errorf("spec.authentication.secretRef.name can not be empty")
		}
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
			}
			return err
		}
	}

	redis := v1.Redis{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Namespace: req.Namespace, Name: req.GetDBRefName()}, &redis)
	if err != nil {
		return err
	}

	if redis.Spec.DisableAuth {
		return fmt.Errorf("%s is running in disable auth mode. RotateAuth is not applicable", req.GetDBRefName())
	}

	return nil
}
