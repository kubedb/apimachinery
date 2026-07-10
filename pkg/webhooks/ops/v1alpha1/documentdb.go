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
	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

	"github.com/pkg/errors"
	"gomodules.xyz/x/arrays"
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

// SetupDocumentDBOpsRequestWebhookWithManager registers the webhook for DocumentDBOpsRequest in the manager.
func SetupDocumentDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.DocumentDBOpsRequest{}).
		WithValidator(&DocumentDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DocumentDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var documentdbLog = logf.Log.WithName("documentdb-opsrequest")

var _ webhook.CustomValidator = &DocumentDBOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.DocumentDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDBOpsRequest object but got %T", obj)
	}
	documentdbLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DocumentDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.DocumentDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDBOpsRequest object but got %T", newObj)
	}
	documentdbLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.DocumentDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DocumentDBOpsRequest object but got %T", oldObj)
	}

	if err := validateDocumentDBOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.DocumentDB
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *DocumentDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.DocumentDBOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.DocumentDBOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for DocumentDB are %s", req.Spec.Type, strings.Join(opsapi.DocumentDBOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList

	switch opsapi.DocumentDBOpsRequestType(req.GetRequestType()) {
	case opsapi.DocumentDBOpsRequestTypeRestart:

	case opsapi.DocumentDBOpsRequestTypeVerticalScaling:
		if err := w.validateDocumentDBVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeHorizontalScaling:
		if err := w.validateDocumentDBHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeReconfigure:
		if err := w.validateDocumentDBReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeUpdateVersion:
		if err := w.validateDocumentDBUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeReconfigureTLS:
		if err := w.validateDocumentDBReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeRotateAuth:
		if err := w.validateDocumentDBRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeStorageMigration:
		if err := w.validateDocumentDBStorageMigrationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeVolumeExpansion:
		if err := w.validateDocumentDBVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.DocumentDBOpsRequestTypeReconnectStandby:
	case opsapi.DocumentDBOpsRequestTypeForceFailOver:
	case opsapi.DocumentDBOpsRequestTypeSetRaftKeyPair:
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Documentdbopsrequests.kubedb.com", Kind: "DocumentDBOpsRequest"}, req.Name, allErr)
}

func (w *DocumentDBOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.DocumentDBOpsRequest) (*dbapi.DocumentDB, error) {
	documentdb := &dbapi.DocumentDB{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, documentdb); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return documentdb, nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBVerticalScalingOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.DocumentDB == nil && verticalScalingSpec.Coordinator == nil && verticalScalingSpec.Arbiter == nil && verticalScalingSpec.ReadReplicas == nil {
		return errors.New("`spec.verticalScaling.DocumentDB`, `spec.verticalScaling.Coordinator`, `spec.verticalScaling.Arbiter`, `spec.verticalScaling.ReadReplica` at least any of them should be present in vertical scaling ops request")
	}
	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBHorizontalScalingOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}
	if horizontalScalingSpec.Replicas == nil && horizontalScalingSpec.ReadReplicas == nil {
		return errors.New("`spec.horizontalScaling.Replicas or `spec.readReplica` has to be mentioned")
	}
	if horizontalScalingSpec.Replicas != nil && len(horizontalScalingSpec.ReadReplicas) > 0 {
		return errors.New("either `spec.horizontalScaling.Replicas` or `spec.horizontalScaling.readReplicas` can be provided at a time")
	}
	if horizontalScalingSpec.Replicas != nil && *horizontalScalingSpec.Replicas <= 0 {
		return errors.New("`spec.horizontalScaling.Replicas` can't be less than or equal 0")
	}
	if horizontalScalingSpec.ReadReplicas != nil {
		for _, rrSpec := range horizontalScalingSpec.ReadReplicas {
			if rrSpec.Replicas != nil && *rrSpec.Replicas <= 0 {
				return errors.New("`spec.horizontalScaling.readReplica.replicas` can't be less than or equal 0")
			}
		}
	}

	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBReconfigureOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	if !req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret == nil && !applyConfigExistsForDocumentDB(req.Spec.Configuration.ApplyConfig) && req.Spec.Configuration.Tuning == nil {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, `Tuning` or `ApplyConfig` must be specified")
	}

	if applyConfigExistsForDocumentDB(req.Spec.Configuration.ApplyConfig) {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.DocumentDBBackendConfigFile]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.DocumentDBBackendConfigFile)
		}
	}

	if req.Spec.Configuration.Restart == opsapi.ReconfigureRestartFalse && req.Spec.Configuration.RemoveCustomConfig {
		return errors.New("`spec.configuration.restart: false` is not allowed when `spec.configuration.removeCustomConfig: true` for documentdb")
	}

	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBUpdateVersionOpsRequest(db *dbapi.DocumentDB, req *opsapi.DocumentDBOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	if req.Status.Phase == opsapi.OpsRequestPhasePending || req.Status.Phase == "" {
		yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindDocumentDBVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
		if err != nil {
			return err
		}
		if !yes {
			return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
		}
	}

	documentdbTargetVersion := &catalog.DocumentDBVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, documentdbTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBReconfigureTLSOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}

	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBRotateAuthenticationOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	db := &dbapi.DocumentDB{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get documentdb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if err := validateRotateAuthSecretRef(context.TODO(), w.DefaultClient, req.Namespace, authSpec.SecretRef, db.GetAdminAuthSecretName(), dbapi.IsVirtualAuthSecretReferred(db.Spec.AuthSecret)); err != nil {
			return err
		}
	}

	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBStorageMigrationOpsRequest(req *opsapi.DocumentDBOpsRequest) error {
	db := &dbapi.DocumentDB{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get documentdb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.Migration == nil {
		return errors.New("spec.migration is required for StorageMigration type")
	}
	if req.Spec.Migration.StorageClassName == nil {
		return errors.New("spec.migration.storageClassName is required")
	}
	if req.Spec.Timeout == nil {
		// timeout is required for Storage Migration ops request because it's a long-running operation
		// default timeout is len(pods) * 5 minute
		return errors.New("spec.timeout is required for Storage Migration ops request,adjust timeout according to the size of your database")
	}
	if db.Spec.Storage == nil || db.Spec.Storage.StorageClassName == nil {
		return fmt.Errorf("db.Spec.Storage.StorageClassName can't be nil in the database yaml")
	}
	// check new storageClass
	var newstorage, oldstorage storagev1.StorageClass
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: *req.Spec.Migration.StorageClassName,
	}, &newstorage)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Wrap(err, fmt.Sprintf("storage class %s not found", *req.Spec.Migration.StorageClassName))
		}
		return err
	}

	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: db.GetStorageClassName(),
	}, &oldstorage)
	if err != nil {
		return err
	}

	if *oldstorage.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		if *newstorage.VolumeBindingMode != storagev1.VolumeBindingWaitForFirstConsumer {
			return errors.New(fmt.Sprintf("volume binding mode should be WaitForFirstConsumer for %s storageClass", newstorage.Name))
		}
	}

	return nil
}

func (w *DocumentDBOpsRequestCustomWebhook) validateDocumentDBVolumeExpansionOpsRequest(db *dbapi.DocumentDB, req *opsapi.DocumentDBOpsRequest) error {
	if req.Spec.VolumeExpansion == nil {
		return errors.New("`spec.volumeExpansion` field is required, can not be nil.")
	}

	if req.Spec.VolumeExpansion.DocumentDB == nil && req.Spec.VolumeExpansion.Arbiter == nil {
		return errors.New("at least one of `spec.volumeExpansion.documentdb` or `spec.volumeExpansion.arbiter` must be specified in volume expansion ops request")
	}

	if req.Spec.VolumeExpansion.DocumentDB != nil {
		if err := opsutil.ValidateStorageExpansion(db.Spec.Storage, req.Spec.VolumeExpansion.DocumentDB, req.Status.Phase, "DocumentDB"); err != nil {
			return err
		}
	}

	return nil
}

func applyConfigExistsForDocumentDB(applyConfig map[string]string) bool {
	if applyConfig == nil {
		return false
	}
	_, exists := applyConfig[kubedb.DocumentDBBackendConfigFile]
	return exists
}

func validateDocumentDBOpsRequest(req *opsapi.DocumentDBOpsRequest, oldReq *opsapi.DocumentDBOpsRequest) error {
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
