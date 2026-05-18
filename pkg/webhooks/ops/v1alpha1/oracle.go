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

func SetupOracleOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.OracleOpsRequest{}).
		WithValidator(&OracleOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type OracleOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var oracleOpsReqLog = logf.Log.WithName("oracle-opsrequest")

var _ webhook.CustomValidator = &OracleOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", obj)
	}
	oracleOpsReqLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *OracleOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", newObj)
	}
	oracleOpsReqLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.OracleOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an OracleOpsRequest object but got %T", oldObj)
	}

	if err := validateOracleOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Oracle
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *OracleOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateOracleOpsRequest(req *opsapi.OracleOpsRequest, oldReq *opsapi.OracleOpsRequest) error {
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

func (w *OracleOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.OracleOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.OracleOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Oracle are %s", req.Spec.Type, strings.Join(opsapi.OracleOpsRequestTypeNames(), ", ")))
	}

	_, err := w.hasDatabaseRef(req)
	if err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch opsapi.OracleOpsRequestType(req.GetRequestType()) {
	case opsapi.OracleOpsRequestTypeReconfigure:
		if err := w.validateOracleReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.OracleOpsRequestTypeRestart:

	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "oracleopsrequests.kubedb.com", Kind: "OracleOpsRequest"}, req.Name, allErr)
}

func (w *OracleOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.OracleOpsRequest) (*dbapi.Oracle, error) {
	oracle := &dbapi.Oracle{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, oracle); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}

	return oracle, nil
}

func (w *OracleOpsRequestCustomWebhook) validateOracleStorageMigrationOpsRequest(req *opsapi.OracleOpsRequest) error {
	db := &dbapi.Oracle{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get postgres: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
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

func (w *OracleOpsRequestCustomWebhook) validateOracleReconfigurationOpsRequest(req *opsapi.OracleOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("spec configuration nil not supported in Reconfigure type")
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

		if _, ok := secret.Data[kubedb.OracleCustomConfigFileName]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.OracleCustomConfigFileName)
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.OracleCustomConfigFileName]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.OracleCustomConfigFileName)
		}
	}

	return nil
}
