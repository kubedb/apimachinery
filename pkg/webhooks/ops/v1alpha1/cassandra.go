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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
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

	oldOps, ok := oldObj.(*opsapi.CassandraOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraOpsRequest object but got %T", oldObj)
	}

	if err := validateCassandraOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *CassandraOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*opsapi.CassandraOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an CassandraOpsRequest object but got %T", obj)
	}
	return nil, nil
}

func validateCassandraOpsRequest(req *opsapi.CassandraOpsRequest, oldReq *opsapi.CassandraOpsRequest) error {
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

func (w *CassandraOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.CassandraOpsRequest) error {
	cassandra, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	var allErr field.ErrorList
	opsType := req.GetRequestType().(opsapi.CassandraOpsRequestType)

	switch opsType {
	case opsapi.CassandraOpsRequestTypeRestart:
		if _, err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeUpdateVersion:
		if err := w.validateCassandraUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeHorizontalScaling:
		if err := w.validateCassandraHorizontalScalingOpsRequest(req, cassandra); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeVerticalScaling:
		if err := w.validateCassandraVerticalScalingOpsRequest(req, cassandra); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeVolumeExpansion:
		if err := w.validateCassandraVolumeExpansionOpsRequest(req, cassandra); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeReconfigure:
		if err := w.validateCassandraReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeReconfigureTLS:
		if err := w.validateCassandraReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeRotateAuth:
		if err := w.validateCassandraRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Cassandra are %s", req.Spec.Type, strings.Join(opsapi.CassandraOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Cassandraopsrequests.kubedb.com", Kind: "CassandraOpsRequest"}, req.Name, allErr)
}

func (w *CassandraOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.CassandraOpsRequest) (*dbapi.Cassandra, error) {
	cassandra := dbapi.Cassandra{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &cassandra); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return &cassandra, nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraUpdateVersionOpsRequest(req *opsapi.CassandraOpsRequest) error {
	// right now, kubeDB support the following Cassandra version: 25.0.0, 28.0.1, 30.0.0, 30.0.1
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	nextCassandraVersion := catalog.CassandraVersion{}
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.UpdateVersion.TargetVersion,
		Namespace: req.GetNamespace(),
	}, &nextCassandraVersion)
	if err != nil {
		return fmt.Errorf("spec.updateVersion.targetVersion - %s, is not supported", req.Spec.UpdateVersion.TargetVersion)
	}
	// check if nextCassandraVersion is deprecated.if deprecated, return error
	if nextCassandraVersion.Spec.Deprecated {
		return fmt.Errorf("spec.updateVersion.targetVersion - %s, is depricated", req.Spec.UpdateVersion.TargetVersion)
	}
	return nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraHorizontalScalingOpsRequest(req *opsapi.CassandraOpsRequest, cassandra *dbapi.Cassandra) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}

	if cassandra.Spec.Topology != nil && horizontalScalingSpec.Topology == nil {
		return errors.New("spec.horizontalScaling.topology can not be empty as reference database mode is topology")
	}

	if horizontalScalingSpec.Topology != nil {
		if horizontalScalingSpec.Topology.Coordinators != nil && *horizontalScalingSpec.Topology.Coordinators <= 0 {
			return errors.New("spec.horizontalScaling.topology.broker must be positive")
		}
		if horizontalScalingSpec.Topology.Overlords != nil && *horizontalScalingSpec.Topology.Overlords <= 0 {
			return errors.New("spec.horizontalScaling.topology.overlords must be positive")
		}
		if horizontalScalingSpec.Topology.MiddleManagers != nil && *horizontalScalingSpec.Topology.MiddleManagers <= 0 {
			return errors.New("spec.horizontalScaling.topology.middleManagers must be positive")
		}
		if horizontalScalingSpec.Topology.Historicals != nil && *horizontalScalingSpec.Topology.Historicals <= 0 {
			return errors.New("spec.horizontalScaling.topology.historicals must be positive")
		}
		if horizontalScalingSpec.Topology.Brokers != nil && *horizontalScalingSpec.Topology.Brokers <= 0 {
			return errors.New("spec.horizontalScaling.topology.brokers must be positive")
		}
		if horizontalScalingSpec.Topology.Routers != nil && *horizontalScalingSpec.Topology.Routers <= 0 {
			return errors.New("spec.horizontalScaling.topology.routers must be positive")
		}
	}

	return nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraVerticalScalingOpsRequest(req *opsapi.CassandraOpsRequest, cassandra *dbapi.Cassandra) error {
	verticalScalingSpec := req.Spec.VerticalScaling

	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Coordinators == nil && verticalScalingSpec.Overlords == nil && verticalScalingSpec.MiddleManagers == nil && verticalScalingSpec.Historicals == nil && verticalScalingSpec.Brokers == nil && verticalScalingSpec.Routers == nil {
		return errors.New("spec.verticalScaling.topology can not be empty")
	}

	topology := cassandra.Spec.Topology
	if verticalScalingSpec.Coordinators != nil && topology.Coordinators == nil {
		return errors.New("spec.verticalScaling.Coordinators can not be set as Coordinators does not exist in the database instance")
	}
	if verticalScalingSpec.Overlords != nil && topology.Overlords == nil {
		return errors.New("spec.verticalScaling.Overlords can not be set as Overlords does not exist in the database instance")
	}
	if verticalScalingSpec.MiddleManagers != nil && topology.MiddleManagers == nil {
		return errors.New("spec.verticalScaling.MiddleManagers can not be set as MiddleManagers does not exist in the database instance")
	}
	if verticalScalingSpec.Historicals != nil && topology.Historicals == nil {
		return errors.New("spec.verticalScaling.Historicals can not be set as Historicals does not exist in the database instance")
	}
	if verticalScalingSpec.Brokers != nil && topology.Brokers == nil {
		return errors.New("spec.verticalScaling.Brokers can not be set as Brokers does not exist in the database instance")
	}
	if verticalScalingSpec.Routers != nil && topology.Routers == nil {
		return errors.New("spec.verticalScaling.Routers can not be set as Routers does not exist in the database instance")
	}
	return nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraVolumeExpansionOpsRequest(req *opsapi.CassandraOpsRequest, cassandra *dbapi.Cassandra) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion

	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.MiddleManagers == nil && volumeExpansionSpec.Historicals == nil {
		return errors.New("spec.verticalScaling.topology can not be empty")
	}

	topology := cassandra.Spec.Topology
	if volumeExpansionSpec.MiddleManagers != nil && topology.MiddleManagers == nil {
		return errors.New("spec.verticalScaling.MiddleManagers can not be set as MiddleManagers does not exist in the database instance")
	}
	if volumeExpansionSpec.Historicals != nil && topology.Historicals == nil {
		return errors.New("spec.verticalScaling.Historicals can not be set as Historicals does not exist in the database instance")
	}

	if volumeExpansionSpec.Historicals != nil && cassandra.Spec.Topology.Historicals.StorageType != dbapi.StorageTypeDurable {
		return errors.New("volumeExpansionSpec.historicals can not be set when storageType of historicals of the database is not Durable ")
	}
	if volumeExpansionSpec.MiddleManagers != nil && cassandra.Spec.Topology.MiddleManagers.StorageType != dbapi.StorageTypeDurable {
		return errors.New("volumeExpansionSpec.middleManagers can not be set when storageType of middleManagers of the database is not Durable ")
	}
	return nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraReconfigurationOpsRequest(req *opsapi.CassandraOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	if _, err := w.hasDatabaseRef(req); err != nil {
		return err
	}

	if configurationSpec.RemoveCustomConfig && (configurationSpec.ConfigSecret != nil || len(configurationSpec.ApplyConfig) != 0) {
		return errors.New("at a time one configuration is allowed to run one operation(`RemoveCustomConfig` or `ConfigSecret with or without ApplyConfig`) to reconfigure")
	}
	if !configurationSpec.RemoveCustomConfig && configurationSpec.ConfigSecret == nil && len(configurationSpec.ApplyConfig) == 0 {
		return errors.New("`RemoveCustomConfig`, `ConfigSecret` and `ApplyConfig`, all can not be empty together")
	}
	return nil
}

func (w *CassandraOpsRequestCustomWebhook) validateCassandraReconfigurationTLSOpsRequest(req *opsapi.CassandraOpsRequest) error {
	TLSSpec := req.Spec.TLS
	if TLSSpec == nil {
		return errors.New("spec.TLS nil not supported in ReconfigureTLS type")
	}
	if _, err := w.hasDatabaseRef(req); err != nil {
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
			if kerr.IsNotFound(err) {
				return fmt.Errorf("referenced secret %s not found", authSpec.SecretRef.Name)
			}
			return err
		}
	}

	return nil
}
