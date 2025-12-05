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

	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupCassandraOpsRequestWebhookWithManager registers the webhook for CassandraOpsRequest in the manager.
func SetupCassandraOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.CassandraOpsRequest{}).
		WithValidator(&CassandraOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type CassandraOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var cassandraLog = logf.Log.WithName("cassandra-opsrequest")

var _ webhook.CustomValidator = &CassandraOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *CassandraOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.CassandraOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraOpsRequest object but got %T", obj)
	}
	cassandraLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *CassandraOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.CassandraOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraOpsRequest object but got %T", newObj)
	}
	cassandraLog.Info("validate update", "name", ops.Name)

	return nil, w.validateCreateOrUpdate(ops)
}

func (w *CassandraOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.CassandraOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.CassandraOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Cassandra are %s", req.Spec.Type, strings.Join(opsapi.CassandraOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	var db olddbapi.Cassandra
	switch req.GetRequestType().(opsapi.CassandraOpsRequestType) {
	case opsapi.CassandraOpsRequestTypeRestart:
		if err := rv.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeHorizontalScaling:
		if err := rv.validateCassandraHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeVerticalScaling:
		if err := rv.validateCassandraVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.CassandraOpsRequestTypeVolumeExpansion:
		if err := rv.validateCassandraVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}

	case opsapi.CassandraOpsRequestTypeUpdateVersion:
		if err := rv.validateCassandraUpdateVersionOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeReconfigure:
		if err := rv.validateCassandraReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeReconfigureTLS:
		if err := rv.validateCassandraReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeRotateAuth:
		if err := rv.validateCassandraRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Cassandraopsrequests.kubedb.com", Kind: "CassandraOpsRequest"}, req.Name, allErr)
}

func (rv *CassandraOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.CassandraOpsRequest) error {
	cassandra := &olddbapi.Cassandra{}
	if err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, cassandra); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraVerticalScalingOpsRequest(req *opsapi.CassandraOpsRequest) error {
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

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraHorizontalScalingOpsRequest(req *opsapi.CassandraOpsRequest) error {
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

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraVolumeExpansionOpsRequest(req *opsapi.CassandraOpsRequest) error {
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

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraUpdateVersionOpsRequest(db *olddbapi.Cassandra, req *opsapi.CassandraOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(rv.DefaultClient, catalog.ResourceKindCassandraVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraReconfigurationOpsRequest(req *opsapi.CassandraOpsRequest) error {
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

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraReconfigurationTLSOpsRequest(req *opsapi.CassandraOpsRequest) error {
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

func (w *CassandraOpsRequestCustomWebhook) validateCassandraRotateAuthenticationOpsRequest(req *opsapi.CassandraOpsRequest) error {
	cassandra := &olddbapi.Cassandra{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.GetDBRefName(),
	}, cassandra)
	if err != nil {
		return err
	}
	if cassandra.Spec.DisableSecurity {
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
