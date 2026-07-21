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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	opsutil "kubedb.dev/apimachinery/pkg/webhooks/ops"

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
	metautil "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func SetupWeaviateOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.WeaviateOpsRequest{}).
		WithValidator(&WeaviateOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type WeaviateOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var weaviateOpsReqLog = logf.Log.WithName("weaviate-opsrequest")

var _ webhook.CustomValidator = &WeaviateOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *WeaviateOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.WeaviateOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an WeaviateOpsRequest object but got %T", obj)
	}
	weaviateOpsReqLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *WeaviateOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.WeaviateOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an WeaviateOpsRequest object but got %T", newObj)
	}
	weaviateOpsReqLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.WeaviateOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an WeaviateOpsRequest object but got %T", oldObj)
	}

	if err := validateWeaviateOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Weaviate
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *WeaviateOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateWeaviateOpsRequest(req *opsapi.WeaviateOpsRequest, oldReq *opsapi.WeaviateOpsRequest) error {
	preconditions := metautil.PreConditionSet{Set: sets.New[string]("spec")}
	_, err := metautil.CreateStrategicPatch(oldReq, req, preconditions.PreconditionFunc()...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditions.Error())
		}
		return err
	}
	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.WeaviateOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.WeaviateOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Weaviate are %s", req.Spec.Type, strings.Join(opsapi.WeaviateOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch opsapi.WeaviateOpsRequestType(req.GetRequestType()) {
	case opsapi.WeaviateOpsRequestTypeHorizontalScaling:
		if err := w.validateWeaviateHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.WeaviateOpsRequestTypeReconfigure:
		if err := w.validateWeaviateReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.WeaviateOpsRequestTypeReconfigureTLS:
		if err := w.validateWeaviateReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.WeaviateOpsRequestTypeRestart:

	case opsapi.WeaviateOpsRequestTypeRotateAuth:
		if err := w.validateWeaviateRotateAuthOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}

	case opsapi.WeaviateOpsRequestTypeStorageMigration:
		if err := w.validateWeaviateStorageMigrationOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"),
				req.Name,
				err.Error()))
		}

	case opsapi.WeaviateOpsRequestTypeVerticalScaling:
		if err := w.validateWeaviateVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.WeaviateOpsRequestTypeVolumeExpansion:
		if err := w.validateWeaviateVolumeExpansionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "weaviateopsrequests.kubedb.com", Kind: "WeaviateOpsRequest"}, req.Name, allErr)
}

func (w *WeaviateOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.WeaviateOpsRequest) (*dbapi.Weaviate, error) {
	weaviate := &dbapi.Weaviate{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, weaviate); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}

	return weaviate, nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateHorizontalScalingOpsRequest(req *opsapi.WeaviateOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}

	if horizontalScalingSpec.Node == nil {
		return errors.New("spec.horizontalScaling.node can not be empty")
	}

	if *horizontalScalingSpec.Node <= 0 {
		return errors.New("spec.horizontalScaling.node must be positive")
	}

	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateReconfigurationOpsRequest(req *opsapi.WeaviateOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("spec configuration nil not supported in Reconfigure type")
	}

	if !reconfigureSpec.RemoveCustomConfig && reconfigureSpec.ConfigSecret == nil && len(reconfigureSpec.ApplyConfig) == 0 && reconfigureSpec.BackupConfigSecret == nil {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig`, or `BackupConfigSecret` must be specified")
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

		if _, ok := secret.Data[kubedb.WeaviateConfigFileName]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.WeaviateConfigFileName)
		}
	}

	// Validate BackupConfigSecret if provided
	if reconfigureSpec.BackupConfigSecret != nil && reconfigureSpec.BackupConfigSecret.Name != "" {
		var secret core.Secret
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
			Name:      reconfigureSpec.BackupConfigSecret.Name,
			Namespace: req.Namespace,
		}, &secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf(
					"referenced backup config secret %s/%s not found",
					req.Namespace,
					reconfigureSpec.BackupConfigSecret.Name,
				)
			}
			return err
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.WeaviateConfigFileName]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.WeaviateConfigFileName)
		}
	}

	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateReconfigureTLSOpsRequest(req *opsapi.WeaviateOpsRequest) error {
	tlsSpec := req.Spec.TLS
	if tlsSpec == nil {
		return errors.New("spec.tls nil not supported in ReconfigureTLS type")
	}

	if tlsSpec.Remove {
		if tlsSpec.RotateCertificates || tlsSpec.IssuerRef != nil || tlsSpec.Certificates != nil || tlsSpec.ClientAuth != nil {
			return errors.New("remove can not be combined with other TLS reconfiguration fields")
		}
		return nil
	}

	if !tlsSpec.RotateCertificates && tlsSpec.IssuerRef == nil && tlsSpec.Certificates == nil && tlsSpec.ClientAuth == nil {
		return errors.New("no reconfiguration is provided in TLS spec")
	}

	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateRotateAuthOpsRequest(db *dbapi.Weaviate, req *opsapi.WeaviateOpsRequest) error {
	if db.Spec.DisableSecurity {
		return fmt.Errorf("disableSecurity is on, RotateAuth is not applicable")
	}

	authSpec := req.Spec.Authentication
	if authSpec != nil && authSpec.SecretRef != nil {
		if err := validateAuthSecretRef(context.TODO(), w.DefaultClient, req.Namespace, authSpec.SecretRef); err != nil {
			return err
		}
	}
	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateStorageMigrationOpsRequest(db *dbapi.Weaviate, req *opsapi.WeaviateOpsRequest) error {
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get Weaviate: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}
	if req.Spec.Migration == nil {
		return errors.New("spec.migration is required")
	}
	if req.Spec.Migration.StorageClassName == nil {
		return errors.New("spec.migration.storageClassName is required")
	}
	if req.Spec.Timeout == nil {
		// timeout is required for Storage Migration ops request because it's a long-running operation
		// default timeout is len(pods) * 5 minute
		return errors.New("spec.timeout is required for Storage Migration ops request,adjust timeout according to the size of your database")
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

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateVerticalScalingOpsRequest(req *opsapi.WeaviateOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}

	if verticalScalingSpec.Node == nil {
		return errors.New("spec.verticalScaling.Node can't be empty")
	}

	return nil
}

func (w *WeaviateOpsRequestCustomWebhook) validateWeaviateVolumeExpansionOpsRequest(db *dbapi.Weaviate, req *opsapi.WeaviateOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.Node == nil {
		return errors.New("spec.volumeExpansion.Node can't be empty")
	}

	if err := opsutil.ValidateStorageExpansion(db.Spec.Storage, volumeExpansionSpec.Node, req.Status.Phase, "Weaviate"); err != nil {
		return err
	}

	return nil
}
