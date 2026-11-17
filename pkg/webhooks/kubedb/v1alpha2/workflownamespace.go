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

package v1alpha2

import (
	"context"
	"fmt"
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

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

// SetupWorkflowNamespaceWebhookWithManager registers the webhook for WorkflowNamespace in the manager.
func SetupWorkflowNamespaceWebhookWithManager(mgr ctrl.Manager) error {
	w := &WorkflowNamespaceCustomWebhook{DefaultClient: mgr.GetClient()}
	return ctrl.NewWebhookManagedBy(mgr).For(&olddbapi.WorkflowNamespace{}).
		WithValidator(w).
		WithDefaulter(w).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kubedb-com-v1alpha2-workflownamespace,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=workflownamespaces,verbs=create;update,versions=v1alpha2,name=mworkflownamespace.kb.io,admissionReviewVersions={v1,v1beta1}

// +kubebuilder:object:generate=false
type WorkflowNamespaceCustomWebhook struct {
	DefaultClient client.Client
}

var _ webhook.CustomDefaulter = &WorkflowNamespaceCustomWebhook{}

var workflowNamespaceLog = logf.Log.WithName("workflownamespace-resource")

// Default implements webhook.CustomDefaulter.
func (w *WorkflowNamespaceCustomWebhook) Default(_ context.Context, obj runtime.Object) error {
	if isDeletionInProgress(obj) {
		return nil
	}
	ns, ok := obj.(*olddbapi.WorkflowNamespace)
	if !ok {
		return fmt.Errorf("expected a WorkflowNamespace object but got %T", obj)
	}
	workflowNamespaceLog.Info("default", "name", ns.GetName())
	ns.SetDefaults()
	return nil
}

//+kubebuilder:webhook:path=/validate-kubedb-com-v1alpha2-workflownamespace,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubedb.com,resources=workflownamespaces,verbs=create;update,versions=v1alpha2,name=vworkflownamespace.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &WorkflowNamespaceCustomWebhook{}

// ValidateCreate implements webhook.CustomValidator.
func (w *WorkflowNamespaceCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(obj) {
		return nil, nil
	}
	ns, ok := obj.(*olddbapi.WorkflowNamespace)
	if !ok {
		return nil, fmt.Errorf("expected a WorkflowNamespace object but got %T", obj)
	}
	workflowNamespaceLog.Info("validate create", "name", ns.GetName())
	return nil, w.validate(ctx, ns, nil)
}

// ValidateUpdate implements webhook.CustomValidator.
func (w *WorkflowNamespaceCustomWebhook) ValidateUpdate(ctx context.Context, old, newObj runtime.Object) (admission.Warnings, error) {
	if isDeletionInProgress(newObj) {
		return nil, nil
	}
	ns, ok := newObj.(*olddbapi.WorkflowNamespace)
	if !ok {
		return nil, fmt.Errorf("expected a WorkflowNamespace object but got %T", newObj)
	}
	oldNS, ok := old.(*olddbapi.WorkflowNamespace)
	if !ok {
		return nil, fmt.Errorf("expected a WorkflowNamespace object but got %T", old)
	}
	workflowNamespaceLog.Info("validate update", "name", ns.GetName())
	return nil, w.validate(ctx, ns, oldNS)
}

// ValidateDelete implements webhook.CustomValidator. DoNotTerminate refuses the deletion.
func (w *WorkflowNamespaceCustomWebhook) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	ns, ok := obj.(*olddbapi.WorkflowNamespace)
	if !ok {
		return nil, fmt.Errorf("expected a WorkflowNamespace object but got %T", obj)
	}
	if ns.Spec.DeletionPolicy == olddbapi.DeletionPolicyDoNotTerminate {
		return nil, apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "WorkflowNamespace"}, ns.GetName(), field.ErrorList{
			field.Invalid(field.NewPath("spec").Child("deletionPolicy"), ns.GetName(),
				"Can not delete as deletionPolicy is set to \"DoNotTerminate\""),
		})
	}
	return nil, nil
}

func (w *WorkflowNamespaceCustomWebhook) validate(ctx context.Context, ns, old *olddbapi.WorkflowNamespace) error {
	var allErr field.ErrorList
	spec := field.NewPath("spec")

	if old != nil {
		if old.EngineName() != ns.EngineName() {
			allErr = append(allErr, field.Forbidden(spec.Child("name"), "name is immutable"))
		}
		if old.Spec.DatabaseRef.Name != ns.Spec.DatabaseRef.Name {
			allErr = append(allErr, field.Forbidden(spec.Child("databaseRef"), "databaseRef is immutable"))
		}
	}

	if ns.Spec.TokenSecretRef != nil && ns.Spec.TokenSecretRef.Name == "" {
		allErr = append(allErr, field.Required(spec.Child("tokenSecretRef").Child("name"), "name is missing"))
	}
	if ns.Spec.Retention != nil && ns.Spec.Retention.Duration <= 0 {
		allErr = append(allErr, field.Invalid(spec.Child("retention"), ns.Spec.Retention.Duration.String(), "must be greater than 0"))
	}

	if ns.Spec.DatabaseRef.Name == "" {
		allErr = append(allErr, field.Required(spec.Child("databaseRef").Child("name"), "databaseRef.name is missing"))
	} else {
		dw := &olddbapi.DurableWorkflow{}
		err := w.DefaultClient.Get(ctx, types.NamespacedName{Namespace: ns.Namespace, Name: ns.Spec.DatabaseRef.Name}, dw)
		switch {
		case apierrors.IsNotFound(err):
			allErr = append(allErr, field.NotFound(spec.Child("databaseRef").Child("name"), ns.Spec.DatabaseRef.Name))
		case err != nil:
			return err
		default:
			allErr = append(allErr, w.validateAgainstDatabase(ctx, ns, dw)...)
		}
	}

	if len(allErr) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{Group: "kubedb.com", Kind: "WorkflowNamespace"}, ns.GetName(), allErr)
}

func (w *WorkflowNamespaceCustomWebhook) validateAgainstDatabase(ctx context.Context, ns *olddbapi.WorkflowNamespace, dw *olddbapi.DurableWorkflow) field.ErrorList {
	var allErr field.ErrorList
	spec := field.NewPath("spec")

	if ns.IsArchivalEnabled() && (dw.Spec.Archival == nil || (dw.Spec.Archival.History == nil && dw.Spec.Archival.Visibility == nil)) {
		allErr = append(allErr, field.Forbidden(spec.Child("archival").Child("enabled"),
			fmt.Sprintf("DurableWorkflow %s has no archival configured", dw.Name)))
	}

	// Without the version the engine is unknown, and the DurableWorkflow webhook reports that. Only the checks
	// that need the engine are skipped.
	version := &catalog.DurableWorkflowVersion{}
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: dw.Spec.Version}, version); err != nil {
		return allErr
	}
	if version.IsCadence() && ns.Spec.Retention != nil && ns.Spec.Retention.Duration < 24*time.Hour {
		allErr = append(allErr, field.Invalid(spec.Child("retention"), ns.Spec.Retention.Duration.String(),
			"Cadence keeps the closed workflows for whole days, the retention must be at least 24h"))
	}
	return allErr
}
