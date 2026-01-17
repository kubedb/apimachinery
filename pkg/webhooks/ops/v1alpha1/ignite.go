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

// SetupIgniteOpsRequestWebhookWithManager registers the webhook for IgniteOpsRequest in the manager.
func SetupIgniteOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.IgniteOpsRequest{}).
		WithValidator(&IgniteOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type IgniteOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var igniteLog = logf.Log.WithName("ignite-opsrequest")

var _ webhook.CustomValidator = &IgniteOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *IgniteOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.IgniteOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an IgniteOpsRequest object but got %T", obj)
	}
	igniteLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *IgniteOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.IgniteOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an IgniteOpsRequest object but got %T", newObj)
	}
	igniteLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.IgniteOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an IgniteOpsRequest object but got %T", oldObj)
	}

	if err := validateIgniteOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *IgniteOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateIgniteOpsRequest(req *opsapi.IgniteOpsRequest, oldReq *opsapi.IgniteOpsRequest) error {
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

func (rv *IgniteOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.IgniteOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.IgniteOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Ignite are %s", req.Spec.Type, strings.Join(opsapi.IgniteOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	var db olddbapi.Ignite
	switch opsapi.IgniteOpsRequestType(req.GetRequestType()) {
	case opsapi.IgniteOpsRequestTypeRestart:
		if err := rv.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeVerticalScaling:
		if err := rv.validateIgniteVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeVolumeExpansion:
		if err := rv.validateIgniteVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeHorizontalScaling:
		if err := rv.validateIgniteHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeReconfigure:
		if err := rv.validateIgniteReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeReconfigureTLS:
		if err := rv.validateIgniteReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeUpdateVersion:
		if err := rv.validateIgniteUpdateVersionOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.IgniteOpsRequestTypeRotateAuth:
		if err := rv.validateIgniteRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Igniteopsrequests.kubedb.com", Kind: "IgniteOpsRequest"}, req.Name, allErr)
}

func (rv *IgniteOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.IgniteOpsRequest) error {
	ignite := &olddbapi.Ignite{}
	if err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, ignite); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (w *IgniteOpsRequestCustomWebhook) validateIgniteRotateAuthenticationOpsRequest(req *opsapi.IgniteOpsRequest) error {
	ignite := &olddbapi.Ignite{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.GetDBRefName(),
	}, ignite)
	if err != nil {
		return err
	}
	if ignite.Spec.DisableSecurity {
		return fmt.Errorf("disableSecurity is on, RotateAuth is not applicable")
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

func (rv *IgniteOpsRequestCustomWebhook) validateIgniteVerticalScalingOpsRequest(req *opsapi.IgniteOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Ignite == nil {
		return errors.New("spec.verticalScaling.Node can't be empty")
	}

	return nil
}

func (rv *IgniteOpsRequestCustomWebhook) validateIgniteVolumeExpansionOpsRequest(req *opsapi.IgniteOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if volumeExpansionSpec.Ignite == nil {
		return errors.New("spec.volumeExpansion.Node can't be empty")
	}

	return nil
}

func (rv *IgniteOpsRequestCustomWebhook) validateIgniteUpdateVersionOpsRequest(db *olddbapi.Ignite, req *opsapi.IgniteOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(rv.DefaultClient, catalog.ResourceKindIgniteVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (rv *IgniteOpsRequestCustomWebhook) validateIgniteHorizontalScalingOpsRequest(req *opsapi.IgniteOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if horizontalScalingSpec.Ignite == nil {
		return errors.New("spec.horizontalScaling.node can not be empty")
	}

	if *horizontalScalingSpec.Ignite <= 0 {
		return errors.New("spec.horizontalScaling.node must be positive")
	}

	return nil
}

func (w *IgniteOpsRequestCustomWebhook) validateIgniteReconfigurationOpsRequest(req *opsapi.IgniteOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}

	if err := w.hasDatabaseRef(req); err != nil {
		return err
	}

	if !reconfigureSpec.RemoveCustomConfig && reconfigureSpec.ConfigSecret == nil && len(reconfigureSpec.ApplyConfig) == 0 {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	if reconfigureSpec.ConfigSecret != nil && reconfigureSpec.ConfigSecret.Name != "" {
		var secret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      reconfigureSpec.ConfigSecret.Name,
			Namespace: req.Namespace,
		}, &secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("referenced config secret %s/%s not found", req.Namespace, reconfigureSpec.ConfigSecret.Name)
			}
			return err
		}

		if _, ok := secret.Data[kubedb.IgniteConfigFileName]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.IgniteConfigFileName)
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.IgniteConfigFileName]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.IgniteConfigFileName)
		}
	}

	return nil
}

func (rv *IgniteOpsRequestCustomWebhook) validateIgniteReconfigurationTLSOpsRequest(req *opsapi.IgniteOpsRequest) error {
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
	if req.Spec.TLS.IssuerRef != nil || req.Spec.TLS.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no reconfiguration is provided in TLS spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}
	return nil
}
