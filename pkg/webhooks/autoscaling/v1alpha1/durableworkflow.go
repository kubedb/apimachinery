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

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupDurableWorkflowAutoscalerWebhookWithManager registers the webhook for DurableWorkflowAutoscaler in the manager.
func SetupDurableWorkflowAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.DurableWorkflowAutoscaler{}).
		WithValidator(&DurableWorkflowAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&DurableWorkflowAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type DurableWorkflowAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var durableWorkflowLog = logf.Log.WithName("durableworkflow-autoscaler")

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-durableworkflowautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=durableworkflowautoscalers,verbs=create;update,versions=v1alpha1,name=mdurableworkflowautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &DurableWorkflowAutoscalerCustomWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *DurableWorkflowAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.DurableWorkflowAutoscaler)
	if !ok {
		return fmt.Errorf("expected a DurableWorkflowAutoscaler object but got %T", obj)
	}

	durableWorkflowLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(ctx, scaler)
	return nil
}

func (w *DurableWorkflowAutoscalerCustomWebhook) setDefaults(ctx context.Context, scaler *autoscalingapi.DurableWorkflowAutoscaler) {
	if scaler.Spec.DatabaseRef == nil {
		return
	}
	db, version, err := w.getDatabase(ctx, scaler)
	if err != nil {
		klog.Errorf("can't get DurableWorkflow %s/%s: %v", scaler.Namespace, scaler.Spec.DatabaseRef.Name, err)
		return
	}

	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Compute == nil {
		return
	}
	for _, role := range db.Roles(version) {
		setDefaultComputeValues(scaler.Spec.Compute.ComputeSpecForRole(role))
	}
}

func (w *DurableWorkflowAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.DurableWorkflowAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

// getDatabase returns the referenced DurableWorkflow and its DurableWorkflowVersion.
func (w *DurableWorkflowAutoscalerCustomWebhook) getDatabase(ctx context.Context, scaler *autoscalingapi.DurableWorkflowAutoscaler) (*olddbapi.DurableWorkflow, *catalog.DurableWorkflowVersion, error) {
	var db olddbapi.DurableWorkflow
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db); err != nil {
		return nil, nil, err
	}
	var version catalog.DurableWorkflowVersion
	if err := w.DefaultClient.Get(ctx, types.NamespacedName{Name: db.Spec.Version}, &version); err != nil {
		return nil, nil, err
	}
	return &db, &version, nil
}

// +kubebuilder:webhook:path=/validate-autoscaling-kubedb-com-v1alpha1-durableworkflowautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=durableworkflowautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vdurableworkflowautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &DurableWorkflowAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *DurableWorkflowAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.DurableWorkflowAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowAutoscaler object but got %T", obj)
	}

	durableWorkflowLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(ctx, scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *DurableWorkflowAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.DurableWorkflowAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected a DurableWorkflowAutoscaler object but got %T", newObj)
	}

	durableWorkflowLog.Info("validate update", "name", scaler.Name)
	return nil, w.validate(ctx, scaler)
}

func (w *DurableWorkflowAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *DurableWorkflowAutoscalerCustomWebhook) validate(ctx context.Context, scaler *autoscalingapi.DurableWorkflowAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	db, version, err := w.getDatabase(ctx, scaler)
	if err != nil {
		klog.Errorf("can't get DurableWorkflow %s/%s: %v", scaler.Namespace, scaler.Spec.DatabaseRef.Name, err)
		return err
	}

	cm := scaler.Spec.Compute
	if cm == nil {
		return nil
	}
	if cm.IsEmpty() {
		return errors.New("spec.compute must configure at least one of frontend, history, matching, worker, internalFrontend")
	}
	if cm.InternalFrontend != nil {
		if !version.IsTemporal() {
			return errors.New("spec.compute.internalFrontend is only available for Temporal")
		}
		if db.Spec.Topology == nil || db.Spec.Topology.InternalFrontend == nil {
			return fmt.Errorf("spec.compute.internalFrontend is set, but DurableWorkflow %s/%s has no spec.topology.internalFrontend", db.Namespace, db.Name)
		}
	}
	for name, spec := range map[string]*autoscalingapi.ComputeAutoscalerSpec{
		"frontend":         cm.Frontend,
		"history":          cm.History,
		"matching":         cm.Matching,
		"worker":           cm.Worker,
		"internalFrontend": cm.InternalFrontend,
	} {
		if err := validateComputeSpec(spec); err != nil {
			return fmt.Errorf("spec.compute.%s: %w", name, err)
		}
	}
	return nil
}

// validateComputeSpec checks that minAllowed does not exceed maxAllowed for CPU and memory.
func validateComputeSpec(spec *autoscalingapi.ComputeAutoscalerSpec) error {
	if spec == nil {
		return nil
	}
	for _, res := range []core.ResourceName{core.ResourceCPU, core.ResourceMemory} {
		min, hasMin := spec.MinAllowed[res]
		max, hasMax := spec.MaxAllowed[res]
		if hasMin && hasMax && min.Cmp(max) > 0 {
			return fmt.Errorf("minAllowed %s (%s) is greater than maxAllowed (%s)", res, min.String(), max.String())
		}
	}
	return nil
}
