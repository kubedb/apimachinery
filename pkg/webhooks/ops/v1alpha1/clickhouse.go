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

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"gomodules.xyz/x/arrays"
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

// SetupClickHouseOpsRequestWebhookWithManager registers the webhook for ClickHouseOpsRequest in the manager.
func SetupClickHouseOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.ClickHouseOpsRequest{}).
		WithValidator(&ClickHouseOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type ClickHouseOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var clickhouseLog = logf.Log.WithName("clickhosue-opsrequest")

var _ webhook.CustomValidator = &ClickHouseOpsRequestCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ClickHouseOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	ops, ok := obj.(*opsapi.ClickHouseOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ClickHouseOpsRequest object but got %T", obj)
	}
	clickhouseLog.Info("validate create", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ClickHouseOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	ops, ok := newObj.(*opsapi.ClickHouseOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected an ClickHouseOpsRequest object but got %T", newObj)
	}
	clickhouseLog.Info("validate update", "name", ops.Name)
	return nil, w.validateCreateOrUpdate(ops)
}

func (w *ClickHouseOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (rv *ClickHouseOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.ClickHouseOpsRequest) error {
	var allErr field.ErrorList
	switch req.GetRequestType().(opsapi.ClickHouseOpsRequestType) {
	case opsapi.ClickHouseOpsRequestTypeRestart:
		if err := rv.hasDatabaseRef(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("restart"),
				req.Name,
				err.Error()))
		}
	case opsapi.ClickHouseOpsRequestTypeVerticalScaling:
		if err := rv.validateClickHouseVerticalScalingOpsRequest(req); err != nil {
			allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("verticalScaling"),
				req.Name,
				err.Error()))
		}
	}
	if validType, _ := arrays.Contains(opsapi.ClickHouseOpsRequestTypeNames(), req.Spec.Type); !validType {
		allErr = append(allErr, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for ClickHouse are %s", req.Spec.Type, strings.Join(opsapi.ClickHouseOpsRequestTypeNames(), ", "))))
	}
	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "ClickHouseopsrequests.kubedb.com", Kind: "ClickHouseOpsRequest"}, req.Name, allErr)
}

func (rv *ClickHouseOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.ClickHouseOpsRequest) error {
	clickhouse := &dbapi.ClickHouse{}
	if err := rv.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, clickhouse); err != nil {
		return fmt.Errorf("spec.databaseRef %s/%s, is invalid or not found", req.GetNamespace(), req.GetDBRefName())
	}
	return nil
}

func (rv *ClickHouseOpsRequestCustomWebhook) validateClickHouseVerticalScalingOpsRequest(req *opsapi.ClickHouseOpsRequest) error {
	verticalScalingSpec := req.Spec.VerticalScaling
	if verticalScalingSpec == nil {
		return errors.New("spec.verticalScaling nil not supported in VerticalScaling type")
	}
	if verticalScalingSpec.Standalone != nil && verticalScalingSpec.Cluster != nil {
		return errors.New("spec.standalone and spec.cluster cannot be set at the same time")
	}
	err := rv.hasDatabaseRef(req)
	if err != nil {
		return err
	}
	if verticalScalingSpec.Cluster != nil {
		for _, cluster := range verticalScalingSpec.Cluster {
			if cluster.ClusterName == "" {
				return errors.New("spec.verticalScaling.Cluster.Name can't be empty")
			}
			if cluster.Node == nil {
				return errors.New("spec.verticalScaling.Cluster.Node can't be empty")
			}
		}
	}

	return nil
}
