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
	"strings"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// SetupMemcachedOpsRequestWebhookWithManager registers the webhook for MemcachedOpsRequest in the manager.
func SetupMemcachedOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MemcachedOpsRequest{}).
		WithValidator(&MemcachedOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MemcachedOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var memcachedLog = logf.Log.WithName("memcached-opsrequest")

var _ webhook.CustomValidator = &MemcachedOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MemcachedOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MemcachedOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MemcachedOpsRequest object but got %T", obj)
	}
	memcachedLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MemcachedOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MemcachedOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MemcachedOpsRequest object but got %T", newObj)
	}
	memcachedLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MemcachedOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MemcachedOpsRequest object but got %T", oldObj)
	}

	if err := validateMemcachedOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *MemcachedOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateMemcachedOpsRequest(req *opsapi.MemcachedOpsRequest, oldReq *opsapi.MemcachedOpsRequest) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldReq, req, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MemcachedOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.MemcachedOpsRequestType) {
	case opsapi.MemcachedOpsRequestTypeRestart:
		if err := c.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MemcachedOpsRequestTypeHorizontalScaling:
		if err := c.validateMemcachedHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MemcachedOpsRequestTypeVerticalScaling:
		if err := c.validateMemcachedVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MemcachedOpsRequestTypeReconfigure:
		if err := c.validateMemcachedReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MemcachedOpsRequestTypeUpdateVersion:
		if err := c.validateMemcachedUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}

	case opsapi.MemcachedOpsRequestTypeReconfigureTLS:
		if err := c.validateMemcachedReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MemcachedOpsRequestTypeRotateAuth:
		if err := c.validateMemcachedRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Memcached are %s", req.Spec.Type, strings.Join(opsapi.MemcachedOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Memcachedopsrequests.kubedb.com", Kind: "MemcachedOpsRequest"}, req.Name, allErr)
}

func (c *MemcachedOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MemcachedOpsRequest) error {
	db := olddbapi.Memcached{}
	if err := c.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &db); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateMemcachedVerticalScalingOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateMemcachedReconfigurationOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateMemcachedHorizontalScalingOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateMemcachedUpdateVersionOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	err := c.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	memcachedTargetVersion := &catalog.MemcachedVersion{}
	err = c.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, memcachedTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (c *MemcachedOpsRequestCustomWebhook) validateMemcachedReconfigureTLSOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}
	if err := c.hasDatabaseRef(req); err != nil {
		return err
	}
	configCount := 0
	if req.Spec.TLS.Remove {
		configCount++
	}
	if req.Spec.TLS.RotateCertificates {
		configCount++
	}
	if req.Spec.TLS.TLSConfig.IssuerRef != nil || req.Spec.TLS.TLSConfig.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("No reconfiguration is provided in TLS Spec.")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}
	return nil
}

func (w *MemcachedOpsRequestCustomWebhook) validateMemcachedRotateAuthenticationOpsRequest(req *opsapi.MemcachedOpsRequest) error {
	db := &dbapi.Memcached{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get memcached: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		var newAuthSecret, oldAuthSecret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &newAuthSecret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Wrap(err, fmt.Sprintf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name))
			}
			return err
		}

		err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      db.GetMemcachedAuthSecretName(),
			Namespace: db.GetNamespace(),
		}, &oldAuthSecret)
		if err != nil {
			return err
		}
	}

	return nil
}
