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

	"github.com/Masterminds/semver/v3"
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

// SetupMySQLOpsRequestWebhookWithManager registers the webhook for MySQLOpsRequest in the manager.
func SetupMySQLOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MySQLOpsRequest{}).
		WithValidator(&MySQLOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MySQLOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var myLog = logf.Log.WithName("mysql-opsrequest")

var _ webhook.CustomValidator = &MySQLOpsRequestCustomWebhook{}

// ValidateCreate implements webhooin.Validator so a webhook will be registered for the type
func (w *MySQLOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", obj)
	}
	myLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (w *MySQLOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", newObj)
	}
	myLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MySQLOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MySQLOpsRequest object but got %T", oldObj)
	}

	if err := w.validateMySQLOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *MySQLOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *MySQLOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MySQLOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.MySQLOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for MySQL are %s", req.Spec.Type, strings.Join(opsapi.MySQLOpsRequestTypeNames(), ", ")))
	}

	var allErr field.ErrorList
	var db olddbapi.MySQL
	switch req.GetRequestType().(opsapi.MySQLOpsRequestType) {
	case opsapi.MySQLOpsRequestTypeRestart:
		if err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeVerticalScaling:
		if err := w.validateMySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeHorizontalScaling:
		if err := w.validateMySQLScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReconfigure:
		if err := w.validateMySQLReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeUpdateVersion:
		if err := w.validateMySQLUpdateVersionOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReconfigureTLS:
		if err := w.validateMySQLReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeVolumeExpansion:
		if err := w.validateMySQLVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeReplicationModeTransformation:
		if err := w.validateMySQLReplicationModeTransformation(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("replicationModeTransformation"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeRotateAuth:
		if err := w.validateMySQLRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	case opsapi.MySQLOpsRequestTypeStorageMigration:
		if err := w.validateMySQLStorageMigrationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"),
				req.Name,
				err.Error()))
		}

	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "MySQLopsrequests.kubedb.com", Kind: "MySQLOpsRequest"}, req.Name, allErr)
}

func (w *MySQLOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MySQLOpsRequest) error {
	md := dbapi.MySQL{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &md); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLUpdateVersionOpsRequest(db *olddbapi.MySQL, req *opsapi.MySQLOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindMySQLVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLScalingOpsRequest(req *opsapi.MySQLOpsRequest) error {
	if req.Spec.Type == opsapi.MySQLOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if err := w.ensureMySQLGroupReplication(req); err != nil {
			return err
		}

		if int32(2) >= *req.Spec.HorizontalScaling.Member || int32(9) <= *req.Spec.HorizontalScaling.Member {
			return errors.New("Group size can not be less than 3 or greater than 9, range: [3,9]")
		}
		return nil
	}

	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLVolumeExpansionOpsRequest(req *opsapi.MySQLOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || req.Spec.VolumeExpansion.MySQL == nil {
		return errors.New("`.Spec.VolumeExpansion` field is nil")
	}
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.Wrap(err, "failed to parse current storage size")
	}

	if cur.Cmp(*req.Spec.VolumeExpansion.MySQL) >= 0 {
		return fmt.Errorf("desired storage size must be greater than current storage. Current storage: %v", cur.String())
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLReconfigurationOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.Configuration == nil || (!req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ApplyConfig == nil && req.Spec.Configuration.ConfigSecret == nil) {
		return errors.New("`.Spec.Configuration` field is nil/not assigned properly")
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLReconfigurationTLSOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if req.Spec.TLS == nil || (req.Spec.TLS.Remove && req.Spec.TLS.RotateCertificates) {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation(`Remove` or `RotateCertificates`)")
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLReplicationModeTransformation(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	curVersion := semver.MustParse(db.Spec.Version)
	refVersion := semver.MustParse("8.4.2")

	if curVersion.LessThan(refVersion) {
		return errors.Wrap(err, fmt.Sprintf("MySQL Replication Mode Transformation support only support for %s or upper.", refVersion))
	}

	if req.Spec.ReplicationModeTransformation != nil {
		if req.Spec.ReplicationModeTransformation.RequireSSL != nil && (req.Spec.ReplicationModeTransformation.IssuerRef == nil &&
			req.Spec.ReplicationModeTransformation.Certificates == nil) {
			return errors.Wrap(err, "MySQL Replication Mode Transformation requires TLS configuration to be enabled.")
		}
	}

	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLRotateAuthenticationOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
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

func (w *MySQLOpsRequestCustomWebhook) validateMySQLStorageMigrationOpsRequest(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
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

func (w *MySQLOpsRequestCustomWebhook) ensureMySQLGroupReplication(req *opsapi.MySQLOpsRequest) error {
	db := &dbapi.MySQL{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get mysql: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	if db.Spec.Topology == nil || db.Spec.Topology.Mode == nil {
		return errors.New("OpsRequest haven't pointed to a Group Replication, Horizontal scaling applicable only for group Replication")
	}
	return nil
}

func (w *MySQLOpsRequestCustomWebhook) validateMySQLOpsRequest(obj, oldObj runtime.Object) error {
	preconditions := meta_util.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}
