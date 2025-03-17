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
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
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

// SetupMSSQLServerOpsRequestWebhookWithManager registers the webhook for MSSQLServerOpsRequest in the manager.
func SetupMSSQLServerOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MSSQLServerOpsRequest{}).
		WithValidator(&MSSQLServerOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MSSQLServerOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var mssqlserverLog = logf.Log.WithName("mssqlserver-opsrequest")

var _ webhook.CustomValidator = &MSSQLServerOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *MSSQLServerOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MSSQLServerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerOpsRequest object but got %T", obj)
	}
	mssqlserverLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *MSSQLServerOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MSSQLServerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerOpsRequest object but got %T", newObj)
	}
	mssqlserverLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MSSQLServerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerOpsRequest object but got %T", oldObj)
	}

	if err := validateMSSQLServerOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *MSSQLServerOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateMSSQLServerOpsRequest(req *opsapi.MSSQLServerOpsRequest, oldReq *opsapi.MSSQLServerOpsRequest) error {
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

func (k *MSSQLServerOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MSSQLServerOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.MSSQLServerOpsRequestType) {
	case opsapi.MSSQLServerOpsRequestTypeRestart:
		if err := k.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeVerticalScaling:
		if err := k.validateMSSQLServerVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeVolumeExpansion:
		if err := k.validateMSSQLServerVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeHorizontalScaling:
		if err := k.validateMSSQLServerHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeReconfigure:
		if err := k.validateMSSQLServerReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeUpdateVersion:
		if err := k.validateMSSQLServerUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeReconfigureTLS:
		if err := k.validateMSSQLServerReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeRotateAuth:
		if err := k.validateMSSQLServerRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MSSQLServer are %s", req.Spec.Type, strings.Join(opsapi.MSSQLServerOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MSSQLServeropsrequests.kubedb.com", Kind: "MSSQLServerOpsRequest"}, req.Name, allErr)
}

func (k *MSSQLServerOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MSSQLServerOpsRequest) error {
	mssqlserver := olddbapi.MSSQLServer{}
	if err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &mssqlserver); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerVerticalScalingOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` is nil. Not supported in VerticalScaling type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.MSSQLServer == nil {
		return errors.New("`spec.verticalScaling.mssqlserver` can't be nil in vertical scaling ops request")
	}

	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerVolumeExpansionOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion is nil, not supported in VolumeExpansion type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if volumeExpansionSpec.MSSQLServer == nil {
		return errors.New("spec.volumeExpansion.mssqlserver is nil, not supported in VolumeExpansion type")
	}

	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerHorizontalScalingOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` is nil. Not supported in HorizontalScaling type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if horizontalScalingSpec.Replicas == nil {
		return errors.New("`spec.horizontalScaling.replicas` can't be nil in HorizontalScaling ops request")
	}
	if *horizontalScalingSpec.Replicas <= 0 {
		return errors.New("`spec.horizontalScaling.replicas` can't be less than or equal 0")
	}
	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerReconfigureOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if mssqlApplyConfigExists(req.Spec.Configuration.ApplyConfig) {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.MSSQLConfigKey]
		if !ok {
			return errors.New(fmt.Sprintf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.MSSQLConfigKey))
		}
	}

	if req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret != nil {
		return errors.New("`spec.configuration.removeCustomConfig` and `spec.configuration.configSecret` is not supported together")
	}

	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerUpdateVersionOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	mssqlserverTargetVersion := &catalog.MSSQLServerVersion{}
	err = k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, mssqlserverTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerReconfigureTLSOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
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
		return errors.New("no reconfiguration is provided in TLS Spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}

	return nil
}

func (k *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerRotateAuthenticationOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	authSpec := req.Spec.Authentication
	if err := k.hasDatabaseRef(req); err != nil {
		return err
	}
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}
		err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &core.Secret{})
		if err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}

	return nil
}

func mssqlApplyConfigExists(applyConfig map[string]string) bool {
	if applyConfig == nil {
		return false
	}
	_, exists := applyConfig[kubedb.MSSQLConfigKey]
	return exists
}
