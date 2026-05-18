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

	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

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

func SetupHanaDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.HanaDBOpsRequest{}).
		WithValidator(&HanaDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type HanaDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var hanadbOpsLog = logf.Log.WithName("hanadb-opsrequest")

var _ webhook.CustomValidator = &HanaDBOpsRequestCustomWebhook{}

func (w *HanaDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.HanaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a HanaDBOpsRequest object but got %T", obj)
	}
	hanadbOpsLog.Info("validate create", "name", req.Name)
	return nil, w.validateCreateOrUpdate(req)
}

func (w *HanaDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.HanaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a HanaDBOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.HanaDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a HanaDBOpsRequest object but got %T", oldObj)
	}
	hanadbOpsLog.Info("validate update", "name", req.Name)

	if err := validateHanaDBOpsRequest(req, oldReq); err != nil {
		return nil, err
	}
	if isOpsReqCompleted(req.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db olddbapi.HanaDB
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *HanaDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateHanaDBOpsRequest(req *opsapi.HanaDBOpsRequest, oldReq *opsapi.HanaDBOpsRequest) error {
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

func (w *HanaDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.HanaDBOpsRequest) error {
	if validType := req.Spec.Type.IsValid(); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for HanaDB are %s", req.Spec.Type, strings.Join(opsapi.HanaDBOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch req.Spec.Type {
	case opsapi.HanaDBOpsRequestTypeRestart:
	case opsapi.HanaDBOpsRequestTypeVerticalScaling:
		if err := w.validateHanaDBVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"), req.Name, err.Error()))
		}
	case opsapi.HanaDBOpsRequestTypeVolumeExpansion:
		if err := w.validateHanaDBVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"), req.Name, err.Error()))
		}
	case opsapi.HanaDBOpsRequestTypeReconfigure:
		if err := w.validateHanaDBReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	case opsapi.HanaDBOpsRequestTypeReconfigureTLS:
		if err := w.validateHanaDBReconfigureTLSOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"), req.Name, err.Error()))
		}
	case opsapi.HanaDBOpsRequestTypeRotateAuth:
		if err := w.validateHanaDBRotateAuthenticationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"), req.Name, err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "hanadbopsrequests.ops.kubedb.com", Kind: opsapi.ResourceKindHanaDBOpsRequest}, req.Name, allErr)
}

func (w *HanaDBOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.HanaDBOpsRequest) (*olddbapi.HanaDB, error) {
	db := &olddbapi.HanaDB{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.DatabaseRef.Name,
		Namespace: req.Namespace,
	}, db); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s is invalid or not found", req.Namespace, req.Spec.DatabaseRef.Name)
	}
	return db, nil
}

func (w *HanaDBOpsRequestCustomWebhook) validateHanaDBVerticalScalingOpsRequest(req *opsapi.HanaDBOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling is nil, not supported in VerticalScaling type")
	}

	if verticalScalingSpec.HanaDB == nil && verticalScalingSpec.Coordinator == nil && verticalScalingSpec.Exporter == nil {
		return errors.New("at least one of spec.verticalScaling.hanadb, spec.verticalScaling.coordinator, or spec.verticalScaling.exporter must be specified")
	}

	return nil
}

func (w *HanaDBOpsRequestCustomWebhook) validateHanaDBVolumeExpansionOpsRequest(db *olddbapi.HanaDB, req *opsapi.HanaDBOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion is nil, not supported in VolumeExpansion type")
	}
	if volumeExpansionSpec.HanaDB == nil {
		return errors.New("spec.volumeExpansion.hanadb can not be empty")
	}

	if err := opsutil.ValidateStorageExpansion(db.Spec.Storage, volumeExpansionSpec.HanaDB, req.Status.Phase, "HanaDB"); err != nil {
		return err
	}

	return nil
}

func (w *HanaDBOpsRequestCustomWebhook) validateHanaDBRotateAuthenticationOpsRequest(db *olddbapi.HanaDB, req *opsapi.HanaDBOpsRequest) error {
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
				return errors.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
			}
			return err
		}
		if password, ok := newAuthSecret.Data[core.BasicAuthPasswordKey]; !ok || len(password) == 0 {
			return errors.Errorf("referenced secret %s/%s must contain non-empty %q key", req.Namespace, authSpec.SecretRef.Name, core.BasicAuthPasswordKey)
		}

		err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      db.GetAuthSecretName(),
			Namespace: db.GetNamespace(),
		}, &oldAuthSecret)
		if err != nil {
			return err
		}

		newUsername := newAuthSecret.Data[core.BasicAuthUsernameKey]
		if len(newUsername) > 0 && string(oldAuthSecret.Data[core.BasicAuthUsernameKey]) != string(newUsername) {
			return errors.New("database username cannot be changed")
		}
	}

	return nil
}

func (w *HanaDBOpsRequestCustomWebhook) validateHanaDBReconfigureOpsRequest(req *opsapi.HanaDBOpsRequest) error {
	cfg := req.Spec.Configuration
	if cfg == nil {
		return errors.New("spec.configuration is nil, not supported in Reconfigure type")
	}
	if !cfg.RemoveCustomConfig &&
		(cfg.ConfigSecret == nil || cfg.ConfigSecret.Name == "") &&
		len(cfg.ApplyConfig) == 0 {
		return errors.New("at least one of removeCustomConfig, configSecret, or applyConfig must be specified")
	}
	if cfg.ConfigSecret != nil && cfg.ConfigSecret.Name != "" {
		var secret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      cfg.ConfigSecret.Name,
			Namespace: req.Namespace,
		}, &secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Errorf("referenced config secret %s/%s not found", req.Namespace, cfg.ConfigSecret.Name)
			}
			return err
		}
		if _, ok := secret.Data[kubedb.HanaDBConfigFileName]; !ok {
			return errors.Errorf("config secret %s/%s does not have file named %q", req.Namespace, cfg.ConfigSecret.Name, kubedb.HanaDBConfigFileName)
		}
	}
	for fileName := range cfg.ApplyConfig {
		if fileName != kubedb.HanaDBConfigFileName {
			return errors.Errorf("unsupported HanaDB config file %q, supported file is %q", fileName, kubedb.HanaDBConfigFileName)
		}
	}
	return nil
}

func (w *HanaDBOpsRequestCustomWebhook) validateHanaDBReconfigureTLSOpsRequest(db *olddbapi.HanaDB, req *opsapi.HanaDBOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("spec.tls is nil, not supported in ReconfigureTLS type")
	}

	certUpdateRequested := tls.IssuerRef != nil || len(tls.Certificates) > 0
	clientTLSUpdateRequested := tls.ClientTLS != nil || tls.ServerName != "" || tls.InsecureSkipVerify

	opCount := 0
	if tls.Remove {
		opCount++
	}
	if tls.RotateCertificates {
		opCount++
	}
	if certUpdateRequested || clientTLSUpdateRequested {
		opCount++
	}
	if opCount == 0 {
		return errors.New("at least one of Remove, RotateCertificates, IssuerRef, Certificates, or client TLS settings must be specified")
	}
	if opCount > 1 {
		return errors.New("only one TLS reconfiguration operation is allowed at a time")
	}

	if tls.Remove {
		return nil
	}

	if tls.RotateCertificates {
		if db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil {
			return errors.New("rotateCertificates requires TLS to already be enabled with issuerRef on HanaDB")
		}
		return nil
	}

	if clientTLSUpdateRequested && !certUpdateRequested && db.Spec.TLS == nil {
		return errors.New("client TLS settings require TLS to already be configured on HanaDB")
	}

	if certUpdateRequested && tls.IssuerRef == nil && (db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil) {
		return errors.New("tls.issuerRef is required for HanaDB ReconfigureTLS")
	}

	return nil
}
