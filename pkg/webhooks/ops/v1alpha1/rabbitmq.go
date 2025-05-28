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

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

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

// SetupRabbitMQOpsRequestWebhookWithManager registers the webhook for RabbitMQOpsRequest in the manager.
func SetupRabbitMQOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.RabbitMQOpsRequest{}).
		WithValidator(&RabbitMQOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type RabbitMQOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var rabbitmqLog = logf.Log.WithName("rabbitmq-opsrequest")

var _ webhook.CustomValidator = &RabbitMQOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *RabbitMQOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.RabbitMQOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RabbitMQOpsRequest object but got %T", obj)
	}
	rabbitmqLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *RabbitMQOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.RabbitMQOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RabbitMQOpsRequest object but got %T", newObj)
	}
	rabbitmqLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.RabbitMQOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an RabbitMQOpsRequest object but got %T", oldObj)
	}

	if err := validateRabbitMQOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *RabbitMQOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateRabbitMQOpsRequest(req *opsapi.RabbitMQOpsRequest, oldReq *opsapi.RabbitMQOpsRequest) error {
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

func (rv *RabbitMQOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.RabbitMQOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.RabbitMQOpsRequestType) {
	case opsapi.RabbitMQOpsRequestTypeRestart:
		if err := rv.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeVerticalScaling:
		if err := rv.validateRabbitMQVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeVolumeExpansion:
		if err := rv.validateRabbitMQVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeHorizontalScaling:
		if err := rv.validateRabbitMQHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeReconfigure:
		if err := rv.validateRabbitMQReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeReconfigureTLS:
		if err := rv.validateRabbitMQReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeUpdateVersion:
		if err := rv.validateRabbitMQUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.RabbitMQOpsRequestTypeRotateAuth:
		if err := rv.validateRabbitMQRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for RabbitMQ are %s", req.Spec.Type, strings.Join(opsapi.RabbitMQOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "RabbitMQopsrequests.kubedb.com", Kind: "RabbitMQOpsRequest"}, req.Name, allErr)
}

func (rv *RabbitMQOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.RabbitMQOpsRequest) error {
	rabbitmq := &olddbapi.RabbitMQ{}
	if err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, rabbitmq); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (w *RabbitMQOpsRequestCustomWebhook) validateRabbitMQRotateAuthenticationOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	rabbitmq := &olddbapi.RabbitMQ{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.GetDBRefName(),
	}, rabbitmq)
	if err != nil {
		return err
	}
	if rabbitmq.Spec.DisableSecurity {
		return fmt.Errorf("DisableSecurity is on, RotateAuth is not applicable")
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
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}
	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQVerticalScalingOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Node == nil {
		return errors.New("spec.verticalScaling.Node can't be empty")
	}

	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQVolumeExpansionOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if volumeExpansionSpec.Node == nil {
		return errors.New("spec.volumeExpansion.Node can't be empty")
	}

	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQUpdateVersionOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	if err := rv.hasDatabaseRef(req); err != nil {
		return err
	}

	nextRabbitMQVersion := catalog.RabbitMQVersion{}
	err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.UpdateVersion.TargetVersion,
		Namespace: req.GetNamespace(),
	}, &nextRabbitMQVersion)
	if err != nil {
		return fmt.Errorf("spec.updateVersion.targetVersion - %s, is not supported", req.Spec.UpdateVersion.TargetVersion)
	}
	// check if nextRabbitMQVersion is deprecated.if deprecated, return error
	if nextRabbitMQVersion.Spec.Deprecated {
		return fmt.Errorf("spec.updateVersion.targetVersion - %s, is depricated", req.Spec.UpdateVersion.TargetVersion)
	}
	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQHorizontalScalingOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if horizontalScalingSpec.Node == nil {
		return errors.New("spec.horizontalScaling.node can not be empty")
	}

	if *horizontalScalingSpec.Node <= 0 {
		return errors.New("spec.horizontalScaling.node must be positive")
	}

	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQReconfigurationOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	if err := rv.hasDatabaseRef(req); err != nil {
		return err
	}
	if configurationSpec.RemoveCustomConfig && (configurationSpec.ConfigSecret != nil || len(configurationSpec.ApplyConfig) != 0) {
		return errors.New("at a time one configuration is allowed to run one operation(`RemoveCustomConfig` or `ConfigSecret with or without ApplyConfig`) to reconfigure")
	}
	return nil
}

func (rv *RabbitMQOpsRequestCustomWebhook) validateRabbitMQReconfigurationTLSOpsRequest(req *opsapi.RabbitMQOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}
	if err := rv.hasDatabaseRef(req); err != nil {
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
