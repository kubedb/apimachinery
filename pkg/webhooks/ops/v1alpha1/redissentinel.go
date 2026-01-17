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
	"strings"

	core "k8s.io/api/core/v1"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"gomodules.xyz/x/arrays"
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
	rs := &dbapi.RedisSentinel{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.DatabaseRef.Name, Namespace: obj.Namespace}}
	return w.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(rs), rs)
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.RedisSentinelOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.RedisSentinelOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for RedisSentinel are %s", req.Spec.Type, strings.Join(opsapi.RedisSentinelOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	var db *dbapi.RedisSentinel
	var err error

	if db, err = w.hasDatabaseRef(req); err != nil {
		return err
	}

	switch opsapi.RedisSentinelOpsRequestType(req.GetRequestType()) {
	case opsapi.RedisSentinelOpsRequestTypeRestart:
		// Restart just needs database ref validation which is already done above
	case opsapi.RedisSentinelOpsRequestTypeUpdateVersion:
		if err := w.validateRedisSentinelUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisSentinelOpsRequestTypeHorizontalScaling:
		if err := w.validateRedisSentinelHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisSentinelOpsRequestTypeVerticalScaling:
		if err := w.validateRedisSentinelVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisSentinelOpsRequestTypeReconfigure:
		if err := w.validateRedisSentinelReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisSentinelOpsRequestTypeReconfigureTLS:
		if err := w.validateRedisSentinelReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.RedisSentinelOpsRequestTypeRotateAuth:
		if err := w.validateRedisSentinelRotateAuthenticationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "redissentinelopsrequest.kubedb.com", Kind: "RedisSentinelOpsRequest"}, req.Name, allErr)
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelUpdateVersionOpsRequest(db *dbapi.RedisSentinel, req *opsapi.RedisSentinelOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindRedisVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	updatedVersionName := req.Spec.UpdateVersion.TargetVersion
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.Namespace}, db)
	if err != nil {
		return err
	}
	currentVersionName := db.Spec.Version

	updatedVersion := &catalog.RedisVersion{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: updatedVersionName}, updatedVersion); err != nil {
		return err
	}

	currentVersion := &catalog.RedisVersion{}
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

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelReconfigureOpsRequest(req *opsapi.RedisSentinelOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	if !reconfigureSpec.RemoveCustomConfig && reconfigureSpec.ConfigSecret == nil && req.Spec.Configuration.ApplyConfig == nil {
		return fmt.Errorf("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	return nil
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelHorizontalScalingOpsRequest(req *opsapi.RedisSentinelOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}

	if horizontalScalingSpec.Replicas == nil {
		return errors.New("`spec.horizontalScaling.replicas` must be specified")
	}

	if *horizontalScalingSpec.Replicas <= 0 {
		return fmt.Errorf("`spec.horizontalScaling.replicas` can not be less than or equal 0")
	}

	return nil
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelVerticalScalingOpsRequest(req *opsapi.RedisSentinelOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.RedisSentinel == nil && verticalScalingSpec.Exporter == nil {
		return errors.New("at least one of `spec.verticalScaling.redissentinel` or `spec.verticalScaling.exporter` must be specified")
	}

	return nil
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelReconfigureTLSOpsRequest(req *opsapi.RedisSentinelOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if tls.Remove {
		configCount++
	}
	if tls.RotateCertificates {
		configCount++
	}
	if tls.IssuerRef != nil || tls.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("at least one of `Remove`, `RotateCertificates`, `IssuerRef`, or `Certificates` must be specified in TLS spec")
	}

	if configCount > 1 {
		return errors.New("only one TLS reconfiguration operation (`Remove`, `RotateCertificates`, or certificate update) is allowed at a time")
	}

	return nil
}

func (w *RedisSentinelOpsRequestCustomWebhook) validateRedisSentinelRotateAuthenticationOpsRequest(db *dbapi.RedisSentinel, req *opsapi.RedisSentinelOpsRequest) error {
	if db.Spec.DisableAuth {
		return fmt.Errorf("%s is running in disable auth mode. RotateAuth is not applicable", req.GetDBRefName())
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
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

	return nil
}

func (w *RedisSentinelOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.RedisSentinelOpsRequest) (*dbapi.RedisSentinel, error) {
	db := dbapi.RedisSentinel{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &db); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return &db, nil
}
