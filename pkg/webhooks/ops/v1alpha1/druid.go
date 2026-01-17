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

	"gomodules.xyz/x/arrays"
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

// SetupDruidOpsRequestWebhookWithManager registers the webhook for DruidOpsRequest in the manager.
func SetupDruidOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.DruidOpsRequest{}).
		WithValidator(&DruidOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DruidOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var druidLog = logf.Log.WithName("druid-opsrequest")

var _ webhook.CustomValidator = &DruidOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DruidOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.DruidOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DruidOpsRequest object but got %T", obj)
	}
	druidLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DruidOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.DruidOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DruidOpsRequest object but got %T", newObj)
	}
	druidLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.DruidOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DruidOpsRequest object but got %T", oldObj)
	}

	if err := validateDruidOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *DruidOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*opsapi.DruidOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an DruidOpsRequest object but got %T", obj)
	}
	return nil, nil
}

func validateDruidOpsRequest(req *opsapi.DruidOpsRequest, oldReq *opsapi.DruidOpsRequest) error {
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

func (w *DruidOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.DruidOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.DruidOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Druid are %s", req.Spec.Type, strings.Join(opsapi.DruidOpsRequestTypeNames(), ", ")))
	}
	druid, err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	var allErr field.ErrorList
	var db dbapi.Druid
	opsType := opsapi.DruidOpsRequestType(req.GetRequestType())

	switch opsType {
	case opsapi.DruidOpsRequestTypeRestart:
		if _, err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeUpdateVersion:
		if err := w.validateDruidUpdateVersionOpsRequest(&db, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeHorizontalScaling:
		if err := w.validateDruidHorizontalScalingOpsRequest(req, druid); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeVerticalScaling:
		if err := w.validateDruidVerticalScalingOpsRequest(req, druid); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeVolumeExpansion:
		if err := w.validateDruidVolumeExpansionOpsRequest(req, druid); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeReconfigure:
		if err := w.validateDruidReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeReconfigureTLS:
		if err := w.validateDruidReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.DruidOpsRequestTypeRotateAuth:
		if err := w.validateDruidRotateAuthenticationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("authentication"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Druidopsrequests.kubedb.com", Kind: "DruidOpsRequest"}, req.Name, allErr)
}

func (w *DruidOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.DruidOpsRequest) (*dbapi.Druid, error) {
	druid := dbapi.Druid{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &druid); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return &druid, nil
}

func (w *DruidOpsRequestCustomWebhook) validateDruidUpdateVersionOpsRequest(db *dbapi.Druid, req *opsapi.DruidOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindDruidVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}

func (w *DruidOpsRequestCustomWebhook) validateDruidHorizontalScalingOpsRequest(req *opsapi.DruidOpsRequest, druid *dbapi.Druid) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("spec.horizontalScaling nil not supported in HorizontalScaling type")
	}

	if druid.Spec.Topology != nil && horizontalScalingSpec.Topology == nil {
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

func (w *DruidOpsRequestCustomWebhook) validateDruidVerticalScalingOpsRequest(req *opsapi.DruidOpsRequest, druid *dbapi.Druid) error {
	verticalScalingSpec := req.Spec.VerticalScaling

	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Coordinators == nil && verticalScalingSpec.Overlords == nil && verticalScalingSpec.MiddleManagers == nil && verticalScalingSpec.Historicals == nil && verticalScalingSpec.Brokers == nil && verticalScalingSpec.Routers == nil {
		return errors.New("spec.verticalScaling.topology can not be empty")
	}

	topology := druid.Spec.Topology
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

func (w *DruidOpsRequestCustomWebhook) validateDruidVolumeExpansionOpsRequest(req *opsapi.DruidOpsRequest, druid *dbapi.Druid) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion

	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}

	if volumeExpansionSpec.MiddleManagers == nil && volumeExpansionSpec.Historicals == nil {
		return errors.New("spec.verticalScaling.topology can not be empty")
	}

	topology := druid.Spec.Topology
	if volumeExpansionSpec.MiddleManagers != nil && topology.MiddleManagers == nil {
		return errors.New("spec.verticalScaling.MiddleManagers can not be set as MiddleManagers does not exist in the database instance")
	}
	if volumeExpansionSpec.Historicals != nil && topology.Historicals == nil {
		return errors.New("spec.verticalScaling.Historicals can not be set as Historicals does not exist in the database instance")
	}

	if volumeExpansionSpec.Historicals != nil && druid.Spec.Topology.Historicals.StorageType != dbapi.StorageTypeDurable {
		return errors.New("volumeExpansionSpec.historicals can not be set when storageType of historicals of the database is not Durable ")
	}
	if volumeExpansionSpec.MiddleManagers != nil && druid.Spec.Topology.MiddleManagers.StorageType != dbapi.StorageTypeDurable {
		return errors.New("volumeExpansionSpec.middleManagers can not be set when storageType of middleManagers of the database is not Durable ")
	}
	return nil
}

func (w *DruidOpsRequestCustomWebhook) validateDruidReconfigurationOpsRequest(req *opsapi.DruidOpsRequest) error {
	configurationSpec := req.Spec.Configuration
	if configurationSpec == nil {
		return errors.New("spec.configuration nil not supported in Reconfigure type")
	}
	if _, err := w.hasDatabaseRef(req); err != nil {
		return err
	}

	if !configurationSpec.RemoveCustomConfig && configurationSpec.ConfigSecret == nil && len(configurationSpec.ApplyConfig) == 0 {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}
	return nil
}

func (w *DruidOpsRequestCustomWebhook) validateDruidReconfigurationTLSOpsRequest(req *opsapi.DruidOpsRequest) error {
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
	if req.Spec.TLS.IssuerRef != nil || req.Spec.TLS.Certificates != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("no reconfiguration is provided in TLS spec")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to spec.reconfigureTLS but at a time one is allowed to run one operation")
	}

	return nil
}

func (w *DruidOpsRequestCustomWebhook) validateDruidRotateAuthenticationOpsRequest(req *opsapi.DruidOpsRequest) error {
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
