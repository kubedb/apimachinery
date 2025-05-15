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

// SetupRedisSentinelOpsRequestWebhookWithManager registers the webhook for RedisSentinelOpsRequest in the manager.
func SetupRedisSentinelOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.RedisSentinelOpsRequest{}).
		WithValidator(&RedisSentinelOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RedisSentinelOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var redissentinelLog = logf.Log.WithName("redissentinel-opsrequest")

var _ webhook.CustomValidator = &RedisSentinelOpsRequestCustomWebhook{}

func (w *RedisSentinelOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", obj)
	}

	redissentinelLog.Info("validate create", "name", req.Name)
	err := w.isDatabaseRefValid(req)
	if err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(req)
}

func (w *RedisSentinelOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newReq, ok := newObj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.RedisSentinelOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RedisSentinelOpsRequest object but got %T", oldObj)
	}

	if err := validateRedisSentinelOpsRequest(newReq, oldReq); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	err := w.isDatabaseRefValid(newReq)
	if err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(newReq)
}

func (w *RedisSentinelOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateRedisSentinelOpsRequest(obj, oldObj runtime.Object) error {
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

func (w *RedisSentinelOpsRequestCustomWebhook) isDatabaseRefValid(obj *opsapi.RedisSentinelOpsRequest) error {
	rs := &v1.RedisSentinel{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.DatabaseRef.Name, Namespace: obj.Namespace}}
	return w.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(rs), rs)
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.RedisSentinelOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.RedisSentinelOpsRequestType) {
	case opsapi.RedisSentinelOpsRequestTypeUpdateVersion:
		if err := w.validateRedisSentinelUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "redissentinelopsrequest.kubedb.com", Kind: "RedisSentinelOpsRequest"}, req.Name, allErr)
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelUpdateVersionOpsRequest(req *opsapi.RedisSentinelOpsRequest) error {
	updatedVersionName := req.Spec.UpdateVersion.TargetVersion
	redisSentinel := v1.RedisSentinel{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.Namespace}, &redisSentinel)
	if err != nil {
		return err
	}
	currentVersionName := redisSentinel.Spec.Version

	updatedVersion := &v1alpha1.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updatedVersionName}, updatedVersion); err != nil {
		return err
	}

	currentVersion := &v1alpha1.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: currentVersionName}, currentVersion); err != nil {
		return err
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
			return fmt.Errorf("redisSentinelOpsRequest %s/%s: can't upgrade Official to Valkey with a different major version", req.Namespace, req.Name)
		}
	}
	return nil
}
