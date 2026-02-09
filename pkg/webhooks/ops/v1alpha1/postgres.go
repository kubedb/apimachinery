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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/x/arrays"
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

// SetupPostgresOpsRequestWebhookWithManager registers the webhook for PostgresOpsRequest in the manager.
func SetupPostgresOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.PostgresOpsRequest{}).
		WithValidator(&PostgresOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PostgresOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var postgresLog = logf.Log.WithName("postgres-opsrequest")

var _ webhook.CustomValidator = &PostgresOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *PostgresOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.PostgresOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PostgresOpsRequest object but got %T", obj)
	}
	postgresLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *PostgresOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.PostgresOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PostgresOpsRequest object but got %T", newObj)
	}
	postgresLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.PostgresOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PostgresOpsRequest object but got %T", oldObj)
	}

	if err := validatePostgresOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Postgres
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *PostgresOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *PostgresOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.PostgresOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.PostgresOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Postgres are %s", req.Spec.Type, strings.Join(opsapi.PostgresOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList

	switch opsapi.PostgresOpsRequestType(req.GetRequestType()) {
	case opsapi.PostgresOpsRequestTypeRestart:

	case opsapi.PostgresOpsRequestTypeVerticalScaling:
		if err := w.validatePostgresVerticalScalingOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeHorizontalScaling:
		if err := w.validatePostgresHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeReconfigure:
		if err := w.validatePostgresReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeUpdateVersion:
		if err := w.validatePostgresUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeReconfigureTLS:
		if err := w.validatePostgresReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeRotateAuth:
		if err := w.validatePostgresRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeStorageMigration:
		if err := w.validatePostgresStorageMigrationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeVolumeExpansion:
		if err := w.validatePostgresVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeReconnectStandby:
	case opsapi.PostgresOpsRequestTypeForceFailOver:
	case opsapi.PostgresOpsRequestTypeSetRaftKeyPair:
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Postgresopsrequests.kubedb.com", Kind: "PostgresOpsRequest"}, req.Name, allErr)
}

func (w *PostgresOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.PostgresOpsRequest) (*dbapi.Postgres, error) {
	postgres := &dbapi.Postgres{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, postgres); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return postgres, nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresVerticalScalingOpsRequest(db *dbapi.Postgres, req *opsapi.PostgresOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Postgres == nil && verticalScalingSpec.Coordinator == nil && verticalScalingSpec.Arbiter == nil && verticalScalingSpec.ReadReplicas == nil {
		return errors.New("`spec.verticalScaling.Postgres`, `spec.verticalScaling.Coordinator`, `spec.verticalScaling.Arbiter`, `spec.verticalScaling.ReadReplica` at least any of them should be present in vertical scaling ops request")
	}
	if verticalScalingSpec.ReadReplicas != nil {
		for _, rrSpec := range verticalScalingSpec.ReadReplicas {
			if !hasReadReplica(db, rrSpec.Name) {
				return errors.New("referenced read replica " + rrSpec.Name + " is not found in database spec")
			}
		}
	}
	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresHorizontalScalingOpsRequest(req *opsapi.PostgresOpsRequest) error {
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

func (w *PostgresOpsRequestCustomWebhook) validatePostgresReconfigureOpsRequest(req *opsapi.PostgresOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	if !req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret == nil && !applyConfigExistsForPostgres(req.Spec.Configuration.ApplyConfig) && req.Spec.Configuration.Tuning == nil {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, `Tuning` or `ApplyConfig` must be specified")
	}

	if applyConfigExistsForPostgres(req.Spec.Configuration.ApplyConfig) {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.PostgresCustomConfigFile]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.PostgresCustomConfigFile)
		}
	}

	if req.Spec.Configuration.Restart == opsapi.ReconfigureRestartFalse && req.Spec.Configuration.RemoveCustomConfig {
		return errors.New("`spec.configuration.restart: false` is not allowed when `spec.configuration.removeCustomConfig: true` for postgres")
	}

	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresUpdateVersionOpsRequest(db *dbapi.Postgres, req *opsapi.PostgresOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	if req.Status.Phase == opsapi.OpsRequestPhasePending || req.Status.Phase == "" {
		yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindPostgresVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
		if err != nil {
			return err
		}
		if !yes {
			return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
		}
	}

	postgresTargetVersion := &catalog.PostgresVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, postgresTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresReconfigureTLSOpsRequest(req *opsapi.PostgresOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}

	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresRotateAuthenticationOpsRequest(req *opsapi.PostgresOpsRequest) error {
	db := &dbapi.Postgres{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get postgres: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
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

func (w *PostgresOpsRequestCustomWebhook) validatePostgresStorageMigrationOpsRequest(req *opsapi.PostgresOpsRequest) error {
	db := &dbapi.Postgres{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get postgres: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.Migration.StorageClassName == nil {
		return errors.New("spec.migration.storageClassName is required")
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

func (w *PostgresOpsRequestCustomWebhook) validatePostgresVolumeExpansionOpsRequest(req *opsapi.PostgresOpsRequest) error {
	if req.Spec.VolumeExpansion == nil {
		return errors.New("`spec.volumeExpansion` field is required, can not be nil.")
	}

	if req.Spec.VolumeExpansion.Postgres == nil && req.Spec.VolumeExpansion.Arbiter == nil {
		return errors.New("at least one of `spec.volumeExpansion.postgres` or `spec.volumeExpansion.arbiter` must be specified in volume expansion ops request")
	}

	db := &dbapi.Postgres{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get postgres: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.New("failed to parse current storage size")
	}

	if req.Spec.VolumeExpansion.Postgres != nil && (req.Status.Phase == opsapi.OpsRequestPhasePending ||
		req.Status.Phase == "") {
		if cur.Cmp(*req.Spec.VolumeExpansion.Postgres) >= 0 {
			return fmt.Errorf("desired storage size must be greater than current storage. Current storage: %v", cur.String())
		}
	}
	if req.Spec.VolumeExpansion.Arbiter != nil && (req.Status.Phase == opsapi.OpsRequestPhasePending ||
		req.Status.Phase == "") {
		if db.Spec.Arbiter != nil && db.Spec.Arbiter.Resources.Requests.Storage() != nil {
			curArbiter, ok := db.Spec.Arbiter.Resources.Requests[core.ResourceStorage]
			if !ok {
				return errors.New("failed to parse current arbiter storage size")
			}
			if curArbiter.Cmp(*req.Spec.VolumeExpansion.Arbiter) >= 0 {
				return fmt.Errorf("desired arbiter storage size must be greater than current arbiter storage. Current arbiter storage: %v", curArbiter.String())
			}
		} else {
			return errors.New("arbiter storage is not configured for this Postgres")
		}
	}

	return nil
}

func applyConfigExistsForPostgres(applyConfig map[string]string) bool {
	if applyConfig == nil {
		return false
	}
	_, exists := applyConfig[kubedb.PostgresCustomConfigFile]
	return exists
}

func validatePostgresOpsRequest(req *opsapi.PostgresOpsRequest, oldReq *opsapi.PostgresOpsRequest) error {
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

func hasReadReplica(db *dbapi.Postgres, replica string) bool {
	for _, r := range db.Spec.ReadReplicas {
		if r.Name == replica {
			return true
		}
	}
	return false
}
