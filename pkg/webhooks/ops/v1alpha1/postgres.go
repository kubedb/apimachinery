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
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *PostgresOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
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

func (w *PostgresOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.PostgresOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.PostgresOpsRequestType) {
	case opsapi.PostgresOpsRequestTypeRestart:
		if err := w.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.PostgresOpsRequestTypeVerticalScaling:
		if err := w.validatePostgresVerticalScalingOpsRequest(req); err != nil {
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
		if err := w.validatePostgresUpdateVersionOpsRequest(req); err != nil {
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
	default:
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for Postgres are %s", req.Spec.Type, strings.Join(opsapi.PostgresOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "Postgresopsrequests.kubedb.com", Kind: "PostgresOpsRequest"}, req.Name, allErr)
}

func (w *PostgresOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.PostgresOpsRequest) error {
	postgres := olddbapi.Postgres{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &postgres); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresVerticalScalingOpsRequest(req *opsapi.PostgresOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Postgres == nil && verticalScalingSpec.Coordinator == nil && verticalScalingSpec.Arbiter == nil {
		return errors.New("`spec.verticalScaling.Postgres`, `spec.verticalScaling.Coordinator`, `spec.verticalScaling.Arbiter` at least any of them should be present in vertical scaling ops request")
	}

	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresHorizontalScalingOpsRequest(req *opsapi.PostgresOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if horizontalScalingSpec.Replicas == nil {
		return errors.New("`spec.horizontalScaling.Replicas has to be mentioned")
	}
	if *horizontalScalingSpec.Replicas <= 0 {
		return errors.New("`spec.horizontalScaling.Replicas` can't be less than or equal 0")
	}
	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresReconfigureOpsRequest(req *opsapi.PostgresOpsRequest) error {
	reconfigureSpec := req.Spec.Configuration
	if reconfigureSpec == nil {
		return errors.New("`spec.configuration` nil not supported in Reconfigure type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}

	if applyConfigExistsForPostgres(req.Spec.Configuration.ApplyConfig) {
		_, ok := req.Spec.Configuration.ApplyConfig[kubedb.PostgresCustomConfigFile]
		if !ok {
			return fmt.Errorf("`spec.configuration.applyConfig` does not have file named '%v'", kubedb.PostgresCustomConfigFile)
		}
	}

	if req.Spec.Configuration.RemoveCustomConfig && req.Spec.Configuration.ConfigSecret != nil {
		return errors.New("can not use `spec.configuration.removeCustomConfig` and `spec.configuration.configSecret` is not supported together")
	}

	return nil
}

func (w *PostgresOpsRequestCustomWebhook) validatePostgresUpdateVersionOpsRequest(req *opsapi.PostgresOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("`spec.updateVersion` nil not supported in UpdateVersion type")
	}
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	postgresTargetVersion := &catalog.PostgresVersion{}
	err = w.DefaultClient.Get(context.TODO(), types.NamespacedName{
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
	err := w.hasDatabaseRef(req)
	if err != nil {
		return err
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
