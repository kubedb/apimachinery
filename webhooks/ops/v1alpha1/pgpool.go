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
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// SetupPgpoolOpsRequestWebhookWithManager registers the webhook for PgpoolOpsRequest in the manager.
func SetupPgpoolOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.PgpoolOpsRequest{}).
		WithValidator(&PgpoolOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type PgpoolOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var pgpoolLog = logf.Log.WithName("pgpool-opsrequest")

var _ webhook.CustomValidator = &PgpoolOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PgpoolOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.PgpoolOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgpoolOpsRequest object but got %T", obj)
	}
	pgpoolLog.Info("validate create", "name", ops.Name)
	return nil, in.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PgpoolOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.PgpoolOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgpoolOpsRequest object but got %T", newObj)
	}
	pgpoolLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.PgpoolOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an PgpoolOpsRequest object but got %T", oldObj)
	}

	if err := validatePgpoolOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, in.validateCreateOrUpdate(ops)
}

func (in *PgpoolOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validatePgpoolOpsRequest(req *opsapi.PgpoolOpsRequest, oldReq *opsapi.PgpoolOpsRequest) error {
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

func (k *PgpoolOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.PgpoolOpsRequest) error {
	pp := &olddbapi.Pgpool{ObjectMeta: metav1.ObjectMeta{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}}
	err := k.DefaultClient.Get(context.TODO(), client.ObjectKeyFromObject(pp), pp)
	if err != nil && apierrors.IsNotFound(err) {
		return fmt.Errorf("referenced database %s/%s is not found", req.Namespace, req.Spec.DatabaseRef.Name)
	}

	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.PgpoolOpsRequestType) {
	case opsapi.PgpoolOpsRequestTypeRestart:
		if err := k.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.PgpoolOpsRequestTypeVerticalScaling:
		if err := k.validatePgpoolVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PgpoolOpsRequestTypeHorizontalScaling:
		if err := k.validatePgpoolHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.PgpoolOpsRequestTypeReconfigure:
		if err := k.validatePgpoolReconfigureOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("configuration"),
				req.Name,
				err.Error()))
		}
	case opsapi.PgpoolOpsRequestTypeUpdateVersion:
		if err := k.validatePgpoolUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
				req.Name,
				err.Error()))
		}
	case opsapi.PgpoolOpsRequestTypeReconfigureTLS:
		if err := k.validatePgpoolReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Pgpool are %s", req.Spec.Type, strings.Join(opsapi.PgpoolOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Pgpoolopsrequests.kubedb.com", Kind: "PgpoolOpsRequest"}, req.Name, allErr)
}

func (k *PgpoolOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.PgpoolOpsRequest) error {
	pgpool := olddbapi.Pgpool{}
	if err := k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &pgpool); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (k *PgpoolOpsRequestCustomWebhook) validatePgpoolVerticalScalingOpsRequest(req *opsapi.PgpoolOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Node == nil {
		return errors.New("`spec.verticalScaling.Node` can't be non-empty at vertical scaling ops request")
	}

	return nil
}

func (k *PgpoolOpsRequestCustomWebhook) validatePgpoolHorizontalScalingOpsRequest(req *opsapi.PgpoolOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if horizontalScalingSpec.Node == nil {
		return errors.New("`spec.horizontalScaling.Node` can't be non-empty at HorizontalScaling ops request")
	}
	if *horizontalScalingSpec.Node <= 0 {
		return errors.New("`spec.horizontalScaling.Node` can't be less than or equal 0")
	}
	return nil
}

func (k *PgpoolOpsRequestCustomWebhook) validatePgpoolReconfigureOpsRequest(req *opsapi.PgpoolOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if applyConfigExists(req.Spec.Configuration.ApplyConfig) {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.PgpoolCustomConfigFile]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.PgpoolCustomConfigFile)
		}
	}

	if req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret != nil {
		return errors.New("can not use `spec.configuration.removeCustomConfig` and `spec.configuration.configSecret` is not supported together")
	}

	return nil
}

func (k *PgpoolOpsRequestCustomWebhook) validatePgpoolUpdateVersionOpsRequest(req *opsapi.PgpoolOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	pgpoolTargetVersion := &catalog.PgpoolVersion{}
	err = k.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: updateVersionSpec.TargetVersion,
	}, pgpoolTargetVersion)
	if err != nil {
		return err
	}
	return nil
}

func (k *PgpoolOpsRequestCustomWebhook) validatePgpoolReconfigureTLSOpsRequest(req *opsapi.PgpoolOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}
	err := k.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	return nil
}

func applyConfigExists(applyConfig map[string]string) bool {
	if applyConfig == nil {
		return false
	}
	_, exists := applyConfig[kubedb.PgpoolCustomConfigFile]
	return exists
}
