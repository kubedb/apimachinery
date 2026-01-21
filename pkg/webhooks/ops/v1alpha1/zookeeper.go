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

// SetupZooKeeperOpsRequestWebhookWithManager registers the webhook for ZooKeeperOpsRequest in the manager.
func SetupZooKeeperOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.ZooKeeperOpsRequest{}).
		WithValidator(&ZooKeeperOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ZooKeeperOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var zookeeperLog = logf.Log.WithName("zookeeper-opsrequest")

var _ webhook.CustomValidator = &ZooKeeperOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ZooKeeperOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.ZooKeeperOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ZooKeeperOpsRequest object but got %T", obj)
	}
	zookeeperLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ZooKeeperOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.ZooKeeperOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ZooKeeperOpsRequest object but got %T", newObj)
	}
	zookeeperLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.ZooKeeperOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ZooKeeperOpsRequest object but got %T", oldObj)
	}

	if err := validateZooKeeperOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}

	return nil, w.validateCreateOrUpdate(ops)
}

func (w *ZooKeeperOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateZooKeeperOpsRequest(req *opsapi.ZooKeeperOpsRequest, oldReq *opsapi.ZooKeeperOpsRequest) error {
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

func (z *ZooKeeperOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.ZooKeeperOpsRequest) error {
	if validType, _ := arrays.Contains(opsapi.ZooKeeperOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for ZooKeeper are %s", req.Spec.Type, strings.Join(opsapi.ZooKeeperOpsRequestTypeNames(), ", ")))
	}
	var (
		zookeeper *dbapi.ZooKeeper
		err       error
	)
	if zookeeper, err = z.hasDatabaseRef(req); err != nil {
		return field.Invalid(field.NewPath("spec").Child("databaseRef"), req.Name, err.Error())
	}

	var allErr field.ErrorList
	switch opsapi.ZooKeeperOpsRequestType(req.GetRequestType()) {
	case opsapi.ZooKeeperOpsRequestTypeUpdateVersion:
		if err := z.validateZooKeeperUpdateVersionOpsRequest(zookeeper, req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.ZooKeeperOpsRequestTypeVerticalScaling:
		if err := z.validateZooKeeperVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("VerticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.ZooKeeperOpsRequestTypeHorizontalScaling:
		if err := z.validateZooKeeperHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.ZooKeeperOpsRequestTypeVolumeExpansion:
		if err := z.validateZooKeeperVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("VolumeExpansion"),
				req.Name,
				err.Error()))
		}
	case opsapi.ZooKeeperOpsRequestTypeReconfigure:
		if err := z.validateZooKeeperReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type").Child("reconfigure"),
				req.Name,
				err.Error()))
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "ZooKeeperopsrequests.kubedb.com", Kind: "ZooKeeperOpsRequest"}, req.Name, allErr)
}

func (z *ZooKeeperOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.ZooKeeperOpsRequest) (*dbapi.ZooKeeper, error) {
	zk := dbapi.ZooKeeper{}
	if err := z.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &zk); err != nil {
		return nil, fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return &zk, nil
}

func (z *ZooKeeperOpsRequestCustomWebhook) validateZooKeeperVerticalScalingOpsRequest(req *opsapi.ZooKeeperOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Node != nil {
		return errors.New("spec.verticalScaling.Node && spec.verticalScaling.Topology both can't be non-empty at the same ops request")
	}

	return nil
}

func (z *ZooKeeperOpsRequestCustomWebhook) validateZooKeeperUpdateVersionOpsRequest(db *dbapi.ZooKeeper, req *opsapi.ZooKeeperOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}

	yes, err := IsUpgradable(z.DefaultClient, catalog.ResourceKindZooKeeperVersion, db.Spec.Version, updateVersionSpec.TargetVersion)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, req.Spec.UpdateVersion.TargetVersion)
	}

	zookeeperTargetVersion := &catalog.ZooKeeperVersion{}
	err = z.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, zookeeperTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (z *ZooKeeperOpsRequestCustomWebhook) validateZooKeeperVolumeExpansionOpsRequest(req *opsapi.ZooKeeperOpsRequest) error {
	volumeExpansionSpec := req.Spec.VolumeExpansion
	if volumeExpansionSpec == nil {
		return errors.New("spec.volumeExpansion nil not supported in VolumeExpansion type")
	}
	if volumeExpansionSpec.Node != nil {
		return errors.New("spec.volumeExpansion.Node && spec.volumeExpansion.Topology both can't be non-empty at the same ops request")
	}

	return nil
}

func (z *ZooKeeperOpsRequestCustomWebhook) validateZooKeeperReconfigureOpsRequest(req *opsapi.ZooKeeperOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}

	if !req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret == nil && req.Spec.Configuration.ApplyConfig == nil {
		return errors.New("at least one of `RemoveCustomConfig`, `ConfigSecret`, or `ApplyConfig` must be specified")
	}

	return nil
}

func (z *ZooKeeperOpsRequestCustomWebhook) validateZooKeeperHorizontalScalingOpsRequest(req *opsapi.ZooKeeperOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}

	if horizontalScalingSpec.Replicas == nil {
		return errors.New("`spec.horizontalScaling.Node` can't be non-empty at HorizontalScaling ops request")
	}
	if *horizontalScalingSpec.Replicas < 3 {
		return errors.New("`spec.horizontalScaling.Node` can't be less than 3")
	}
	return nil
}
