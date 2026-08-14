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
	secret_lib "kubedb.dev/apimachinery/pkg/secret"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

	"github.com/pkg/errors"
	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
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

func SetupEtcdOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.EtcdOpsRequest{}).
		WithValidator(&EtcdOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type EtcdOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var etcdOpsLog = logf.Log.WithName("etcd-opsrequest")

var _ webhook.CustomValidator = &EtcdOpsRequestCustomWebhook{}

func (w *EtcdOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.EtcdOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a EtcdOpsRequest object but got %T", obj)
	}
	etcdOpsLog.Info("validate create", "name", req.Name)
	return nil, w.validateCreateOrUpdate(req)
}

func (w *EtcdOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.EtcdOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a EtcdOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.EtcdOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a EtcdOpsRequest object but got %T", oldObj)
	}
	etcdOpsLog.Info("validate update", "name", req.Name)

	if err := validateEtcdOpsRequest(req, oldReq); err != nil {
		return nil, err
	}
	if isOpsReqCompleted(req.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db olddbapi.Etcd
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *EtcdOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateEtcdOpsRequest(req *opsapi.EtcdOpsRequest, oldReq *opsapi.EtcdOpsRequest) error {
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

func (w *EtcdOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.EtcdOpsRequest) error {
	if validType := req.Spec.Type.IsValid(); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Etcd are %s", req.Spec.Type, strings.Join(opsapi.EtcdOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch req.Spec.Type {
	case opsapi.EtcdOpsRequestTypeUpdateVersion:
		if err := w.validateEtcdUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeRestart:
	case opsapi.EtcdOpsRequestTypeHorizontalScaling:
		if err := w.validateEtcdHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeVerticalScaling:
		if err := w.validateEtcdVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeVolumeExpansion:
		if err := w.validateEtcdVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeReconfigure:
		if err := w.validateEtcdReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeReconfigureTLS:
		if err := w.validateEtcdReconfigureTLSOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeRotateAuth:
		if err := w.validateEtcdRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeStorageMigration:
		if err := w.validateEtcdStorageMigrationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"), req.Name, err.Error()))
		}
	case opsapi.EtcdOpsRequestTypeMoveLeader:
		// spec.moveLeader is optional; an empty NewLeader lets the operator pick a
		// healthy, up-to-date member, so there is nothing more to validate here.
	case opsapi.EtcdOpsRequestTypeDefragment:
		// spec.defragment carries no fields; defragmentation always walks every
		// member sequentially, so there is nothing to validate here.
	case opsapi.EtcdOpsRequestTypeCompact:
		if err := w.validateEtcdCompactOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("compact"), req.Name, err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "etcdopsrequests.ops.kubedb.com", Kind: opsapi.ResourceKindEtcdOpsRequest}, req.Name, allErr)
}

func (w *EtcdOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.EtcdOpsRequest) (*olddbapi.Etcd, error) {
	db := &olddbapi.Etcd{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.DatabaseRef.Name,
		Namespace: req.Namespace,
	}, db); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s is invalid or not found", req.Namespace, req.Spec.DatabaseRef.Name)
	}
	return db, nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdUpdateVersionOpsRequest(req *opsapi.EtcdOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion is nil, not supported in UpdateVersion type")
	}
	if updateVersionSpec.TargetVersion == "" {
		return errors.New("spec.updateVersion.targetVersion can not be empty")
	}
	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdHorizontalScalingOpsRequest(req *opsapi.EtcdOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling is nil, not supported in HorizontalScaling type")
	}
	if horizontalScalingSpec.Replicas == nil {
		return errors.New("spec.horizontalScaling.replicas can not be empty")
	}
	if *horizontalScalingSpec.Replicas < 1 {
		return errors.New("spec.horizontalScaling.replicas must be at least 1")
	}
	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdVerticalScalingOpsRequest(req *opsapi.EtcdOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling is nil, not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Etcd == nil && verticalScalingSpec.Exporter == nil {
		return errors.New("at least one of spec.verticalScaling.etcd or spec.verticalScaling.exporter must be specified")
	}

	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdVolumeExpansionOpsRequest(db *olddbapi.Etcd, req *opsapi.EtcdOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion is nil, not supported in VolumeExpansion type")
	}
	if volumeExpansionSpec.Etcd == nil {
		return errors.New("spec.volumeExpansion.etcd can not be empty")
	}

	if err := opsutil.ValidateStorageExpansion(db.Spec.Storage, volumeExpansionSpec.Etcd, req.Status.Phase, "Etcd"); err != nil {
		return err
	}

	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdStorageMigrationOpsRequest(db *olddbapi.Etcd, req *opsapi.EtcdOpsRequest) error {
	migrationSpec := req.Spec.Migration
	if migrationSpec == nil {
		return errors.New("spec.migration is required for StorageMigration type")
	}
	if migrationSpec.StorageClassName == nil {
		return errors.New("spec.migration.storageClassName is required")
	}
	if req.Spec.Timeout == nil {
		return errors.New("spec.timeout is required for Storage Migration ops request, adjust timeout according to the size of your database")
	}
	if db.Spec.Storage == nil || db.Spec.Storage.StorageClassName == nil {
		return nil
	}

	var newStorage, oldStorage storagev1.StorageClass
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: *migrationSpec.StorageClassName}, &newStorage); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("storage class %s not found: %w", *migrationSpec.StorageClassName, err)
		}
		return err
	}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: *db.Spec.Storage.StorageClassName}, &oldStorage); err != nil {
		return err
	}
	if oldStorage.VolumeBindingMode != nil && *oldStorage.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		if newStorage.VolumeBindingMode == nil || *newStorage.VolumeBindingMode != storagev1.VolumeBindingWaitForFirstConsumer {
			return fmt.Errorf("volume binding mode should be WaitForFirstConsumer for %s storageClass", newStorage.Name)
		}
	}

	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdRotateAuthenticationOpsRequest(req *opsapi.EtcdOpsRequest) error {
	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if authSpec.SecretRef.Name == "" {
			return errors.New("spec.authentication.secretRef.name can not be empty")
		}

		isVirtual := authSpec.SecretRef.APIGroup == vsecretapi.GroupName
		newData, err := secret_lib.GetData(context.TODO(), w.DefaultClient, req.Namespace, authSpec.SecretRef.Name, isVirtual)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return errors.Errorf("referenced secret %s/%s not found", req.Namespace, authSpec.SecretRef.Name)
			}
			return err
		}
		if password := newData[core.BasicAuthPasswordKey]; len(password) == 0 {
			return errors.Errorf("referenced secret %s/%s must contain a non-empty %q", req.Namespace, authSpec.SecretRef.Name, core.BasicAuthPasswordKey)
		}
		if username := newData[core.BasicAuthUsernameKey]; len(username) > 0 && string(username) != kubedb.EtcdRootUser {
			return errors.Errorf("username in referenced secret %s/%s must be %q", req.Namespace, authSpec.SecretRef.Name, kubedb.EtcdRootUser)
		}
	}

	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdReconfigureOpsRequest(req *opsapi.EtcdOpsRequest) error {
	cfg := req.Spec.Configuration
	if cfg == nil {
		return errors.New("spec.configuration is nil, not supported in Reconfigure type")
	}
	if cfg.Tuning == nil && !cfg.RemoveCustomConfig {
		return errors.New("spec.configuration.tuning must be specified, since Etcd does not support a mounted custom config file (--config-file is mutually exclusive with etcd command-line flags)")
	}
	if len(cfg.ApplyConfig) > 0 || (cfg.ConfigSecret != nil && cfg.ConfigSecret.Name != "") {
		return errors.New("spec.configuration.applyConfig and spec.configuration.configSecret are not supported for Etcd, use spec.configuration.tuning instead")
	}
	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdReconfigureTLSOpsRequest(db *olddbapi.Etcd, req *opsapi.EtcdOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("spec.tls is nil, not supported in ReconfigureTLS type")
	}

	certUpdateRequested := tls.IssuerRef != nil || len(tls.Certificates) > 0

	opCount := 0
	if tls.Remove {
		opCount++
	}
	if tls.RotateCertificates {
		opCount++
	}
	if certUpdateRequested {
		opCount++
	}
	if opCount == 0 {
		return errors.New("at least one of Remove, RotateCertificates, IssuerRef, or Certificates must be specified")
	}
	if opCount > 1 {
		return errors.New("only one TLS reconfiguration operation is allowed at a time")
	}

	if tls.Remove {
		return nil
	}

	if tls.RotateCertificates {
		if db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil {
			return errors.New("rotateCertificates requires TLS to already be enabled with issuerRef on Etcd")
		}
		return nil
	}

	if certUpdateRequested && tls.IssuerRef == nil && (db.Spec.TLS == nil || db.Spec.TLS.IssuerRef == nil) {
		return errors.New("tls.issuerRef is required for Etcd ReconfigureTLS")
	}

	return nil
}

func (w *EtcdOpsRequestCustomWebhook) validateEtcdCompactOpsRequest(req *opsapi.EtcdOpsRequest) error {
	compactSpec := req.Spec.Compact
	if compactSpec == nil {
		return errors.New("spec.compact is nil, not supported in Compact type")
	}
	if compactSpec.Revision != nil && *compactSpec.Revision < 0 {
		return errors.New("spec.compact.revision can not be negative")
	}
	return nil
}
