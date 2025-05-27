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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

	return nil, w.validateCreateOrUpdate(ops)
}

func (w *CassandraOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.CassandraOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.CassandraOpsRequestType) {
	case opsapi.CassandraOpsRequestTypeRestart:
		if err := rv.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeVerticalScaling:
		if err := rv.validateCassandraVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.CassandraOpsRequestTypeUpdateVersion:
		if err := rv.validateCassandraUpdateVersionOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("updateVersion"),
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

func (rv *CassandraOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.CassandraOpsRequest) error {
	cassandra := &olddbapi.Cassandra{}
	if err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, cassandra); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraVerticalScalingOpsRequest(req *opsapi.CassandraOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Node == nil {
		return errors.New("spec.verticalScaling.Node can't be empty")
	}

	return nil
}

func (rv *CassandraOpsRequestCustomWebhook) validateCassandraUpdateVersionOpsRequest(req *opsapi.CassandraOpsRequest) error {
	updateVersionSpec := req.Spec.UpdateVersion
	if updateVersionSpec == nil {
		return errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	if err := rv.hasDatabaseRef(req); err != nil {
		return err
	}

	nextCassandraVersion := catalog.CassandraVersion{}
	err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.Spec.UpdateVersion.TargetVersion,
		Namespace: req.GetNamespace(),
	}, &nextCassandraVersion)
	if err != nil {
		return fmt.Errorf("spec.updateVersion.targetVersion - %s, is not found", req.Spec.UpdateVersion.TargetVersion)
	}

	return nil
}
