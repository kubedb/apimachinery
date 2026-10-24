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
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupDurableWorkflowOpsRequestWebhookWithManager registers the webhook for DurableWorkflowOpsRequest in the manager.
func SetupDurableWorkflowOpsRequestWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&opsapi.DurableWorkflowOpsRequest{}).
		WithValidator(&DurableWorkflowOpsRequestCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DurableWorkflowOpsRequestCustomWebhook struct {
	DefaultClient client.Client
}

var durableWorkflowOpsLog = logf.Log.WithName("durableworkflow-opsrequest")

var _ webhook.CustomValidator = &DurableWorkflowOpsRequestCustomWebhook{}

func (w *DurableWorkflowOpsRequestCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, ok := obj.(*opsapi.DurableWorkflowOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowOpsRequest object but got %T", obj)
	}
	durableWorkflowOpsLog.Info("validate create", "name", req.Name)
	return w.validateCreateOrUpdate(req)
}

func (w *DurableWorkflowOpsRequestCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, ok := newObj.(*opsapi.DurableWorkflowOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowOpsRequest object but got %T", newObj)
	}
	oldReq, ok := oldObj.(*opsapi.DurableWorkflowOpsRequest)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowOpsRequest object but got %T", oldObj)
	}
	durableWorkflowOpsLog.Info("validate update", "name", req.Name)

	if err := validateImmutableSpec(oldReq, req); err != nil {
		return nil, err
	}
	warnings, err := w.validateCreateOrUpdate(req)
	if err != nil {
		return warnings, err
	}
	if isOpsReqCompleted(req.Status.Phase) && !isOpsReqCompleted(oldReq.Status.Phase) { // just completed
		var db dbapi.DurableWorkflow
		if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: req.Spec.DatabaseRef.Name, Namespace: req.Namespace}, &db); err != nil {
			return warnings, err
		}
		return warnings, resumeDatabase(w.DefaultClient, &db)
	}
	return warnings, nil
}

func (w *DurableWorkflowOpsRequestCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if _, ok := obj.(*opsapi.DurableWorkflowOpsRequest); !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowOpsRequest object but got %T", obj)
	}
	return nil, nil
}

func validateImmutableSpec(oldReq, req *opsapi.DurableWorkflowOpsRequest) error {
	if fmt.Sprintf("%+v", oldReq.Spec) != fmt.Sprintf("%+v", req.Spec) {
		return errors.New("spec is immutable")
	}
	return nil
}

func (w *DurableWorkflowOpsRequestCustomWebhook) validateCreateOrUpdate(req *opsapi.DurableWorkflowOpsRequest) (admission.Warnings, error) {
	if validType, _ := arrays.Contains(opsapi.DurableWorkflowOpsRequestTypeNames(), string(req.Spec.Type)); !validType {
		return nil, field.Invalid(field.NewPath("spec").Child("type"), req.Name,
			fmt.Sprintf("defined OpsRequestType %s is not supported, supported types for DurableWorkflow are %s", req.Spec.Type, strings.Join(opsapi.DurableWorkflowOpsRequestTypeNames(), ", ")))
	}

	db, err := w.hasDatabaseRef(req)
	if err != nil {
		return nil, err
	}

	var allErr field.ErrorList
	var warnings admission.Warnings
	spec := field.NewPath("spec")
	switch opsapi.DurableWorkflowOpsRequestType(req.GetRequestType()) {
	case opsapi.DurableWorkflowOpsRequestTypeUpdateVersion:
		warns, err := w.validateUpdateVersion(db, req)
		warnings = append(warnings, warns...)
		if err != nil {
			allErr = append(allErr, field.Invalid(spec.Child("updateVersion"), req.Name, err.Error()))
		}
	case opsapi.DurableWorkflowOpsRequestTypeHorizontalScaling:
		if req.Spec.HorizontalScaling == nil {
			allErr = append(allErr, field.Required(spec.Child("horizontalScaling"), "spec.horizontalScaling is required for HorizontalScaling"))
		} else if req.Spec.HorizontalScaling.InternalFrontend != nil && !w.isTemporal(db) {
			allErr = append(allErr, field.Forbidden(spec.Child("horizontalScaling").Child("internalFrontend"), "internalFrontend is only available for Temporal"))
		}
	case opsapi.DurableWorkflowOpsRequestTypeVerticalScaling:
		if req.Spec.VerticalScaling == nil {
			allErr = append(allErr, field.Required(spec.Child("verticalScaling"), "spec.verticalScaling is required for VerticalScaling"))
		}
	case opsapi.DurableWorkflowOpsRequestTypeReconfigure:
		if req.Spec.Configuration == nil {
			allErr = append(allErr, field.Required(spec.Child("configuration"), "spec.configuration is required for Reconfigure"))
		} else if _, ok := req.Spec.Configuration.ApplyConfig["numHistoryShards"]; ok {
			allErr = append(allErr, field.Forbidden(spec.Child("configuration").Child("applyConfig"), "numHistoryShards can't be changed"))
		}
	case opsapi.DurableWorkflowOpsRequestTypeReconfigureTLS:
		if req.Spec.TLS == nil {
			allErr = append(allErr, field.Required(spec.Child("tls"), "spec.tls is required for ReconfigureTLS"))
		}
	case opsapi.DurableWorkflowOpsRequestTypeRestart, opsapi.DurableWorkflowOpsRequestTypeRotateAuth:
	}

	if len(allErr) == 0 {
		return warnings, nil
	}
	return warnings, apierrors.NewInvalid(schema.GroupKind{Group: "durableworkflowopsrequests.kubedb.com", Kind: "DurableWorkflowOpsRequest"}, req.Name, allErr)
}

func (w *DurableWorkflowOpsRequestCustomWebhook) isTemporal(db *dbapi.DurableWorkflow) bool {
	var v catalog.DurableWorkflowVersion
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &v); err != nil {
		return false
	}
	return v.IsTemporal()
}

func (w *DurableWorkflowOpsRequestCustomWebhook) validateUpdateVersion(db *dbapi.DurableWorkflow, req *opsapi.DurableWorkflowOpsRequest) (admission.Warnings, error) {
	spec := req.Spec.UpdateVersion
	if spec == nil {
		return nil, errors.New("spec.updateVersion nil not supported in UpdateVersion type")
	}

	var current, target catalog.DurableWorkflowVersion
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: db.Spec.Version}, &current); err != nil {
		return nil, fmt.Errorf("current version %s: %w", db.Spec.Version, err)
	}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{Name: spec.TargetVersion}, &target); err != nil {
		return nil, fmt.Errorf("target version %s: %w", spec.TargetVersion, err)
	}
	if current.Spec.Distribution != target.Spec.Distribution {
		return nil, fmt.Errorf("changing the distribution from %s to %s is a data migration and not supported by UpdateVersion",
			current.Spec.Distribution, target.Spec.Distribution)
	}

	yes, err := IsUpgradable(w.DefaultClient, catalog.ResourceKindDurableWorkflowVersion, db.Spec.Version, spec.TargetVersion)
	if err != nil {
		return nil, err
	}
	if !yes {
		return nil, fmt.Errorf("upgrade from version %v to %v is not supported", db.Spec.Version, spec.TargetVersion)
	}

	var warnings admission.Warnings
	if !db.IsHighlyAvailable(&current) && !spec.AllowDowntime {
		warnings = append(warnings, "the cluster is not highly available (a role has fewer than 2 replicas), so a role will be unavailable while its pod restarts")
	}
	if !target.OnlineSchemaUpdate() && !spec.AllowDowntime {
		return warnings, fmt.Errorf("the schema update of %s can't run online; set updateVersion.allowDowntime to proceed", target.Name)
	}
	return warnings, nil
}

func (w *DurableWorkflowOpsRequestCustomWebhook) hasDatabaseRef(req *opsapi.DurableWorkflowOpsRequest) (*dbapi.DurableWorkflow, error) {
	db := dbapi.DurableWorkflow{}
	if err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      req.GetDBRefName(),
		Namespace: req.GetNamespace(),
	}, &db); err != nil {
		allErr := field.ErrorList{field.Invalid(field.NewPath("spec").Child("databaseRef"), req.GetDBRefName(), fmt.Sprintf("%s/%s is invalid or not found", req.GetNamespace(), req.GetDBRefName()))}
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "durableworkflowopsrequests.kubedb.com", Kind: "DurableWorkflowOpsRequest"}, req.Name, allErr)
	}
	return &db, nil
}
