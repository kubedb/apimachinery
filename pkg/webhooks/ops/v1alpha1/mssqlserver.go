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

	"gomodules.xyz/x/arrays"
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
func (w *MSSQLServerOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MSSQLServerOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MSSQLServerOpsRequest object but got %T", obj)
	}
	mssqlserverLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MSSQLServerOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
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
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *MSSQLServerOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
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

func (w *MSSQLServerOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MSSQLServerOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.MSSQLServerOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MSSQLServer are %s", req.Spec.Type, strings.Join(opsapi.MSSQLServerOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.MSSQLServerOpsRequestType) {
	case opsapi.MSSQLServerOpsRequestTypeRestart:
		if err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeVerticalScaling:
		if err := w.validateMSSQLServerVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeVolumeExpansion:
		if err := w.validateMSSQLServerVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeHorizontalScaling:
		if err := w.validateMSSQLServerHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeReconfigure:
		if err := w.validateMSSQLServerReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeUpdateVersion:
		if err := w.validateMSSQLServerUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeReconfigureTLS:
		if err := w.validateMSSQLServerReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MSSQLServerOpsRequestTypeRotateAuth:
		if err := w.validateMSSQLServerRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MSSQLServeropsrequests.kubedb.com", Kind: "MSSQLServerOpsRequest"}, req.Name, allErr)
}

func (w *MSSQLServerOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MSSQLServerOpsRequest) error {
	mssqlserver := olddbapi.MSSQLServer{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &mssqlserver); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerVerticalScalingOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` is nil. Not supported in VerticalScaling type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.MSSQLServer == nil {
		return fmt.Errorf("`spec.verticalScaling.mssqlserver` can't be nil in vertical scaling ops request")
	}

	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerVolumeExpansionOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return fmt.Errorf("spec.volumeExpansion is nil, not supported in VolumeExpansion type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if volumeExpansionSpec.MSSQLServer == nil {
		return fmt.Errorf("spec.volumeExpansion.mssqlserver is nil, not supported in VolumeExpansion type")
	}

	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerHorizontalScalingOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return fmt.Errorf("`spec.horizontalScaling` is nil. Not supported in HorizontalScaling type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if horizontalScalingSpec.Replicas == nil {
		return fmt.Errorf("`spec.horizontalScaling.replicas` can't be nil in HorizontalScaling ops request")
	}
	if *horizontalScalingSpec.Replicas <= 0 {
		return fmt.Errorf("`spec.horizontalScaling.replicas` can't be less than or equal 0")
	}
	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerReconfigureOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return fmt.Errorf("`spec.configuration` nil not supported in Reconfigure type")
	}

	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	// Validate ConfigSecret exits if provided and has the required config file
	if reconfigureSpec.ConfigSecret != nil {
		var secret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      reconfigureSpec.ConfigSecret.Name,
			Namespace: req.Namespace,
		}, &secret)
		if err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf("referenced config secret %s/%s not found", req.Namespace, reconfigureSpec.ConfigSecret.Name)
			}
			return err
		}

		if _, ok := secret.Data[kubedb.MSSQLConfigKey]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.MSSQLConfigKey)
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.MSSQLConfigKey]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.MSSQLConfigKey)
		}
	}

	// Add validation to not allow both RemoveCustomConfig and ConfigSecret together
	if req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret != nil {
		return fmt.Errorf("`spec.configuration.removeCustomConfig` and `spec.configuration.configSecret` is not supported together")
	}

	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerUpdateVersionOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return fmt.Errorf("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	mssqlserverTargetVersion := &catalog.MSSQLServerVersion{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, mssqlserverTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerReconfigureTLSOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return fmt.Errorf("`spec.tls` nil not supported in ReconfigureTLS type")
	}
	err := w.hasDatabaseRef(req)
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
		return fmt.Errorf("no reconfiguration is provided in TLS Spec")
	}

	if configCount > 1 {
		return fmt.Errorf("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}

	return nil
}

func (w *MSSQLServerOpsRequestCustomWebhook) validateMSSQLServerRotateAuthenticationOpsRequest(req *opsapi.MSSQLServerOpsRequest) error {
	db := &olddbapi.MSSQLServer{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, db); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return fmt.Errorf("spec.authentication.secretRef.name can not be empty")
		}

		var newAuthSecret, oldAuthSecret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      authSpec.SecretRef.Name,
			Namespace: req.Namespace,
		}, &newAuthSecret)
		if err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
			}
			return err
		}

		err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      db.GetAuthSecretName(),
			Namespace: db.GetNamespace(),
		}, &oldAuthSecret)
		if err != nil {
			return err
		}

		if string(oldAuthSecret.Data[core.BasicAuthUsernameKey]) != string(newAuthSecret.Data[core.BasicAuthUsernameKey]) {
			return errors.New("database username cannot be changed")
		}
	}

	return nil
}
