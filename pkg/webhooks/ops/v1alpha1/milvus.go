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
	meta_util "kmodules.xyz/client-go/meta"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupMilvusOpsRequestWebhookWithManager registers the webhook for MilvusOpsRequest in the manager.
func SetupMilvusOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.MilvusOpsRequest{}).
		WithValidator(&MilvusOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type MilvusOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomValidator = &MilvusOpsRequestCustomWebhook{}

var milvusOpsReqLog = logf.Log.WithName("milvus-opsrequest")

var _ webhook.CustomValidator = &MilvusOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *MilvusOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.MilvusOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a MilvusOpsRequest object but got %T", obj)
	}
	milvusOpsReqLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *MilvusOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.MilvusOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MilvusOpsRequest object but got %T", newObj)
	}
	milvusOpsReqLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.MilvusOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an MilvusOpsRequest object but got %T", oldObj)
	}

	if err := validateMilvusOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	if err := w.validateCreateOrUpdate(ops); err != nil {
		return nil, err
	}

	if isOpsReqCompleted(ops.Status.Phase) && !isOpsReqCompleted(oldOps.Status.Phase) { // just completed
		var db dbapi.Milvus
		err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: ops.Spec.DatabaseRef.Name, Namespace: ops.Namespace}, &db)
		if err != nil {
			return nil, err
		}
		return nil, resumeDatabase(w.DefaultClient, &db)
	}
	return nil, nil
}

func (w *MilvusOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateMilvusOpsRequest(req *opsapi.MilvusOpsRequest, oldReq *opsapi.MilvusOpsRequest) error {
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

func (w *MilvusOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.MilvusOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.MilvusOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Milvus are %s", req.Spec.Type, strings.Join(opsapi.MilvusOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	var allErr field.ErrorList

	opsType := opsapi.MilvusOpsRequestType(req.GetRequestType())

	switch opsType {
	case opsapi.MilvusOpsRequestTypeRestart:
		// no additional validation needed

	case opsapi.MilvusOpsRequestTypeUpdateVersion:
		if err := w.validateMilvusUpdateVersionOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeHorizontalScaling:
		if err := w.validateMilvusHorizontalScalingOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeVerticalScaling:
		if err := w.validateMilvusVerticalScalingOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeVolumeExpansion:
		if err := w.validateMilvusVolumeExpansionOpsRequest(req, db); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeStorageMigration:
		if err := w.validateMilvusStorageMigrationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("migration"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeReconfigure:
		if err := w.validateMilvusReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeReconfigureTLS:
		if err := w.validateMilvusReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}

	case opsapi.MilvusOpsRequestTypeRotateAuth:
		if err := w.validateMilvusRotateAuthOpsRequest(db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "milvusopsrequests.kubedb.com", Kind: "MilvusOpsRequest"}, req.Name, allErr)
}

func (w *MilvusOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.MilvusOpsRequest) (*dbapi.Milvus, error) {
	milvus := &dbapi.Milvus{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, milvus); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return milvus, nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusUpdateVersionOpsRequest(db *dbapi.Milvus, req *opsapi.MilvusOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, "MilvusVersion", db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, updateVersionSpec.TargetVersion)
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusHorizontalScalingOpsRequest(req *opsapi.MilvusOpsRequest, milvus *dbapi.Milvus) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}

	if !milvus.IsDistributed() {
		return errors.New("horizontal scaling is not applicable for Milvus Standalone mode")
	}

	if horizontalScalingSpec.Topology == nil {
		return errors.New("spec.horizontalScaling.topology can not be empty for Distributed mode")
	}

	t := horizontalScalingSpec.Topology
	if t.Proxy != nil && *t.Proxy <= 0 {
		return errors.New("spec.horizontalScaling.topology.proxy must be positive")
	}
	if t.MixCoord != nil && *t.MixCoord <= 0 {
		return errors.New("spec.horizontalScaling.topology.mixcoord must be positive")
	}
	if t.QueryNode != nil && *t.QueryNode <= 0 {
		return errors.New("spec.horizontalScaling.topology.querynode must be positive")
	}
	if t.StreamingNode != nil && *t.StreamingNode <= 0 {
		return errors.New("spec.horizontalScaling.topology.streamingnode must be positive")
	}
	if t.DataNode != nil && *t.DataNode <= 0 {
		return errors.New("spec.horizontalScaling.topology.dataNode must be positive")
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusVerticalScalingOpsRequest(req *opsapi.MilvusOpsRequest, milvus *dbapi.Milvus) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}

	isDistributed := milvus.IsDistributed()

	if isDistributed {
		if verticalScalingSpec.Node != nil {
			return errors.New("spec.verticalScaling.node can not be set when database mode is Distributed")
		}
		if verticalScalingSpec.Proxy == nil && verticalScalingSpec.MixCoord == nil &&
			verticalScalingSpec.DataNode == nil && verticalScalingSpec.QueryNode == nil &&
			verticalScalingSpec.StreamingNode == nil {
			return errors.New("at least one distributed node type must be specified in spec.verticalScaling")
		}

		dist := milvus.Spec.Topology.Distributed
		if verticalScalingSpec.Proxy != nil && (dist == nil || dist.Proxy == nil) {
			return errors.New("spec.verticalScaling.proxy can not be set as Proxy does not exist in the database instance")
		}
		if verticalScalingSpec.MixCoord != nil && (dist == nil || dist.MixCoord == nil) {
			return errors.New("spec.verticalScaling.mixcoord can not be set as MixCoord does not exist in the database instance")
		}
		if verticalScalingSpec.DataNode != nil && (dist == nil || dist.DataNode == nil) {
			return errors.New("spec.verticalScaling.datanode can not be set as DataNode does not exist in the database instance")
		}
		if verticalScalingSpec.QueryNode != nil && (dist == nil || dist.QueryNode == nil) {
			return errors.New("spec.verticalScaling.querynode can not be set as QueryNode does not exist in the database instance")
		}
		if verticalScalingSpec.StreamingNode != nil && (dist == nil || dist.StreamingNode == nil) {
			return errors.New("spec.verticalScaling.streamingnode can not be set as StreamingNode does not exist in the database instance")
		}
	} else {
		if verticalScalingSpec.Node == nil {
			return errors.New("spec.verticalScaling.node must be set when database mode is Standalone")
		}
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusVolumeExpansionOpsRequest(req *opsapi.MilvusOpsRequest, milvus *dbapi.Milvus) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	isDistributed := milvus.IsDistributed()

	if isDistributed {
		if volumeExpansionSpec.Node != nil {
			return errors.New("spec.volumeExpansion.node can not be set when database mode is Distributed")
		}
		if volumeExpansionSpec.StreamingNode == nil {
			return errors.New("spec.volumeExpansion.streamingnode must be specified in Distributed mode")
		}
		dist := milvus.Spec.Topology.Distributed
		if volumeExpansionSpec.StreamingNode != nil && (dist == nil || dist.StreamingNode == nil) {
			return errors.New("spec.volumeExpansion.streamingnode can not be set as StreamingNode does not exist in the database instance")
		}
		if dist != nil && dist.StreamingNode != nil && dist.StreamingNode.StorageType != dbapi.StorageTypeDurable {
			return errors.New("spec.volumeExpansion.streamingnode can not be set when storageType of StreamingNode is not Durable")
		}
	} else {
		if volumeExpansionSpec.Node == nil {
			return errors.New("spec.volumeExpansion.node must be set when database mode is Standalone")
		}
		if milvus.Spec.StorageType != dbapi.StorageTypeDurable {
			return errors.New("spec.volumeExpansion.node can not be set when storageType of the database is not Durable")
		}
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusReconfigurationOpsRequest(req *opsapi.MilvusOpsRequest) error {
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

		if _, ok := secret.Data[kubedb.MilvusConfigFileName]; !ok {
			return fmt.Errorf("config secret %s/%s does not have file named '%v'", req.Namespace, reconfigureSpec.ConfigSecret.Name, kubedb.MilvusConfigFileName)
		}
	}

	// Validate ApplyConfig has the required config file if provided
	if req.Spec.Configuration.ApplyConfig != nil {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.MilvusConfigFileName]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.MilvusConfigFileName)
		}
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusReconfigureTLSOpsRequest(req *opsapi.MilvusOpsRequest) error {
	tlsSpec := req.Spec.TLS
	if tlsSpec == nil {
		return errors.New("spec.tls nil not supported in ReconfigureTLS type")
	}

	configCount := 0
	if tlsSpec.Remove {
		configCount++
	}
	if tlsSpec.RotateCertificates {
		configCount++
	}
	if tlsSpec.IssuerRef != nil || tlsSpec.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no reconfiguration is provided in TLS spec")
	}
	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.tls but at a time only one operation is allowed")
	}

	return nil
}

func (w *MilvusOpsRequestCustomWebhook) validateMilvusRotateAuthOpsRequest(db *dbapi.Milvus, req *opsapi.MilvusOpsRequest) error {
	if db.Spec.DisableSecurity {
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

func (w *MilvusOpsRequestCustomWebhook) validateMilvusStorageMigrationOpsRequest(req *opsapi.MilvusOpsRequest) error {
	db := &dbapi.Milvus{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.GetDBRefName(), Namespace: req.GetNamespace()}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get Milvus: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
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
