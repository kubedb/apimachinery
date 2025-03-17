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
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"github.com/pkg/errors"
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

// SetupPerconaXtraDBOpsRequestWebhookWithManager registers the webhook for PerconaXtraDBOpsRequest in the manager.
func SetupPerconaXtraDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.PerconaXtraDBOpsRequest{}).
		WithValidator(&PerconaXtraDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PerconaXtraDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pxLog = logf.Log.WithName("percona-opsrequest")

var _ webhook.CustomValidator = &PerconaXtraDBOpsRequestCustomWebhook{}

// ValidateCreate implements webhooin.Validator so a webhook will be registered for the type
func (in *PerconaXtraDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.PerconaXtraDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PerconaXtraDBOpsRequest object but got %T", obj)
	}
	pxLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhooin.Validator so a webhook will be registered for the type
func (in *PerconaXtraDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.PerconaXtraDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PerconaXtraDBOpsRequest object but got %T", newObj)
	}
	pxLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.PerconaXtraDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PerconaXtraDBOpsRequest object but got %T", oldObj)
	}

	if err := in.validatePerconaXtraDBOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.PerconaXtraDBOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.PerconaXtraDBOpsRequestType) {
	case opsapi.PerconaXtraDBOpsRequestTypeRestart:
		if err := in.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeVerticalScaling:
		if err := in.validatePerconaXtraDBScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeHorizontalScaling:
		if err := in.validatePerconaXtraDBScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeReconfigure:
		if err := in.validatePerconaXtraDBReconfigurationOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeUpdateVersion:
		if err := in.validatePerconaXtraDBUpgradeOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeReconfigureTLS:
		if err := in.validatePerconaXtraDBReconfigurationTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	case opsapi.PerconaXtraDBOpsRequestTypeVolumeExpansion:
		if err := in.validatePerconaXtraDBVolumeExpansionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("volumeExpansion"),
				req.Name,
				err.Error()))
		}

	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for PerconaXtraDB are %s", req.Spec.Type, strings.Join(opsapi.PerconaXtraDBOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "PerconaXtraDBopsrequests.kubedb.com", Kind: "PerconaXtraDBOpsRequest"}, req.Name, allErr)
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.PerconaXtraDBOpsRequest) error {
	px := dbapi.PerconaXtraDB{}
	if err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &px); err != nil {
		return errors.New(fmt.Sprintf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName()))
	}
	return nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBOpsRequest(obj, oldObj runtime.Object) error {
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

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBUpgradeOpsRequest(req *opsapi.PerconaXtraDBOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.Upgrade & spec.UpdateVersion both nil not supported")
	}

	db := &dbapi.PerconaXtraDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: req.Spec.DatabaseRef.Name,
	}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get percona-xtradbdb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}
	pxNextVersion := &catalog.PerconaXtraDBVersion{}
	err = in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, pxNextVersion)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get percona-xtradbdbVersion: %s", updateVersionSpec.TargetVersion))
	}
	// check if pxNextVersion is deprecated.if deprecated, return error
	if pxNextVersion.Spec.Deprecated {
		return fmt.Errorf("percona-xtradbdb target version %s/%s is deprecated. Skipped processing", db.Namespace, pxNextVersion.Name)
	}
	return nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBScalingOpsRequest(req *opsapi.PerconaXtraDBOpsRequest) error {
	if req.Spec.Type == opsapi.PerconaXtraDBOpsRequestTypeHorizontalScaling {
		if req.Spec.HorizontalScaling == nil {
			return errors.New("`spec.Scale.HorizontalScaling` field is nil")
		}

		if *req.Spec.HorizontalScaling.Member <= int32(2) || *req.Spec.HorizontalScaling.Member > int32(9) {
			return errors.New("Group size can not be less than 3 or greater than 9, range: [3,9]")
		}
		return nil
	}

	if req.Spec.VerticalScaling == nil {
		return errors.New("`spec.Scale.Vertical` field is empty")
	}

	return nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBVolumeExpansionOpsRequest(req *opsapi.PerconaXtraDBOpsRequest) error {
	if req.Spec.VolumeExpansion == nil || req.Spec.VolumeExpansion.PerconaXtraDB == nil {
		return errors.New("`.Spec.VolumeExpansion` field is nil")
	}
	db := &dbapi.PerconaXtraDB{}
	err := in.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: req.Spec.DatabaseRef.Name,
	}, db)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get percona-xtradb: %s/%s", req.Namespace, req.Spec.DatabaseRef.Name))
	}

	cur, ok := db.Spec.Storage.Resources.Requests[core.ResourceStorage]
	if !ok {
		return errors.Wrap(err, "failed to parse current storage size")
	}

	if cur.Cmp(*req.Spec.VolumeExpansion.PerconaXtraDB) >= 0 {
		return fmt.Errorf("Desired storage size must be greater than current storage. Current storage: %v", cur.String())
	}
	return nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBReconfigurationOpsRequest(req *opsapi.PerconaXtraDBOpsRequest) error {
	if req.Spec.Configuration == nil || (!req.Spec.Configuration.RemoveCustomConfig && len(req.Spec.Configuration.ApplyConfig) == 0 && req.Spec.Configuration.ConfigSecret == nil) {
		return errors.New("`.Spec.Configuration` field is nil/not assigned properly")
	}

	assign := 0
	if req.Spec.Configuration.RemoveCustomConfig {
		assign++
	}
	if req.Spec.Configuration.ApplyConfig != nil {
		assign++
	}
	if req.Spec.Configuration.ConfigSecret != nil {
		assign++
	}
	if assign > 1 {
		return errors.New("more than 1 field have assigned to reconfigure your database but at a time you you are allowed to run one operation(`RemoveCustomConfig`, `ApplyConfig` or `ConfigSecret`) to reconfigure")
	}

	return nil
}

func (in *PerconaXtraDBOpsRequestCustomWebhook) validatePerconaXtraDBReconfigurationTLSOpsRequest(req *opsapi.PerconaXtraDBOpsRequest) error {
	if req.Spec.TLS == nil {
		return errors.New("TLS Spec is empty")
	}
	configCount := 0
	if req.Spec.TLS.Remove {
		configCount++
	}
	if req.Spec.TLS.RotateCertificates {
		configCount++
	}
	if req.Spec.TLS.TLSConfig.IssuerRef != nil || req.Spec.TLS.TLSConfig.Certificates != nil || req.Spec.TLS.RequireSSL != nil {
		configCount++
	}

	if configCount == 0 {
		return errors.New("No of incomplete reconfiguration is provided in TLS Spec.")
	}

	if configCount > 1 {
		return errors.New("more than 1 field have assigned to reconfigureTLS to your database but at a time you you are allowed to run one operation")
	}
	return nil
}
