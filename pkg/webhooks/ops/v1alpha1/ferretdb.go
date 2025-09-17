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

	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gomodules.xyz/x/arrays"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// SetupFerretDBOpsRequestWebhookWithManager registers the webhook for FerretDBOpsRequest in the manager.
func SetupFerretDBOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.FerretDBOpsRequest{}).
		WithValidator(&FerretDBOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type FerretDBOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var ferretdbLog = logf.Log.WithName("ferretdb-opsrequest")

var _ webhook.CustomValidator = &FerretDBOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *FerretDBOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.FerretDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an FerretDBOpsRequest object but got %T", obj)
	}
	ferretdbLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *FerretDBOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.FerretDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an FerretDBOpsRequest object but got %T", newObj)
	}
	ferretdbLog.Info("validate update", "name", ops.Name)

	oldOps, ok := oldObj.(*opsapi.FerretDBOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an FerretDBOpsRequest object but got %T", oldObj)
	}

	if err := validateFerretDBOpsRequest(ops, oldOps); err != nil {
		return nil, err
	}
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *FerretDBOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func validateFerretDBOpsRequest(req *opsapi.FerretDBOpsRequest, oldReq *opsapi.FerretDBOpsRequest) error {
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

func (w *FerretDBOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.FerretDBOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.FerretDBOpsRequestType) {
	case opsapi.FerretDBOpsRequestTypeRestart:
	case opsapi.FerretDBOpsRequestTypeVerticalScaling:
		if err := w.validateFerretDBVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.FerretDBOpsRequestTypeHorizontalScaling:
		if err := w.validateFerretDBHorizontalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("horizontalScaling"),
				req.Name,
				err.Error()))
		}
	case opsapi.FerretDBOpsRequestTypeReconfigureTLS:
		if err := w.validateFerretDBReconfigureTLSOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("tls"),
				req.Name,
				err.Error()))
		}
	}

	if validType, _ := arrays.Contains(opsapi.FerretDBOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for FerretDB are %s", req.Spec.Type, strings.Join(opsapi.FerretDBOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "FerretDBopsrequests.kubedb.com", Kind: "FerretDBOpsRequest"}, req.Name, allErr)
}

func (w *FerretDBOpsRequestCustomWebhook) validateFerretDBVerticalScalingOpsRequest(req *opsapi.FerretDBOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("`spec.verticalScaling` nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Primary == nil && verticalScalingSpec.Secondary == nil {
		return errors.New("both `spec.verticalScaling.primary` and `spec.verticalScaling.secondary` can't be non-empty at vertical scaling ops request")
	}

	return nil
}

func (w *FerretDBOpsRequestCustomWebhook) validateFerretDBHorizontalScalingOpsRequest(req *opsapi.FerretDBOpsRequest) error {
	horizontalScalingSpec := req.Spec.HorizontalScaling
	if horizontalScalingSpec == nil {
		return errors.New("`spec.horizontalScaling` nil not supported in HorizontalScaling type")
	}
	if horizontalScalingSpec.Primary == nil && horizontalScalingSpec.Secondary == nil {
		return errors.New("both `spec.horizontalScaling.primary.replicas` and `spec.horizontalScaling.secondary.replicas` can't be empty")
	}
	if horizontalScalingSpec.Primary != nil && *horizontalScalingSpec.Primary.Replicas <= 0 {
		return errors.New("`spec.horizontalScaling.primary.replicas` can't be less than or equal 0")
	}
	if horizontalScalingSpec.Secondary != nil && *horizontalScalingSpec.Secondary.Replicas <= 0 {
		return errors.New("`spec.horizontalScaling.secondary.replicas` can't be less than or equal 0")
	}
	return nil
}

func (w *FerretDBOpsRequestCustomWebhook) validateFerretDBReconfigureTLSOpsRequest(req *opsapi.FerretDBOpsRequest) error {
	tls := req.Spec.TLS
	if tls == nil {
		return errors.New("`spec.tls` nil not supported in ReconfigureTLS type")
	}

	return nil
}
