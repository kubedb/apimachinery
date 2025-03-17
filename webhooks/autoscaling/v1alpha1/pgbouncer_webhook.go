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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var pbLog = logf.Log.WithName("pgbouncer-autoscaler")

func (in *autoscalingapi.PgBouncerAutoscaler) SetupWebhookWithManager(mgr manager.Manager) error {
	return builder.WebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-autoscaling-kubedb-com-v1alpha1-pgbouncerautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=autoscaling.kubedb.com,resources=pgbouncerautoscaler,verbs=create;update,versions=v1alpha1,name=mpgbouncerautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomDefaulter = &autoscalingapi.PgBouncerAutoscaler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *autoscalingapi.PgBouncerAutoscaler) Default(ctx context.Context, obj runtime.Object) error {
	pbLog.Info("defaulting", "name", r.Name)
	r.setDefaults()
	return nil
}

func (r *autoscalingapi.PgBouncerAutoscaler) setDefaults() {
	var db dbapi.PgBouncer
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      r.Spec.DatabaseRef.Name,
		Namespace: r.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get PgBouncer %s/%s \n", r.Namespace, r.Spec.DatabaseRef.Name)
		return
	}

	r.setOpsReqOptsDefaults()

	if r.Spec.Compute != nil {
		setDefaultComputeValues(r.Spec.Compute.PgBouncer)
	}
}

func (r *autoscalingapi.PgBouncerAutoscaler) setOpsReqOptsDefaults() {
	if r.Spec.OpsRequestOptions == nil {
		r.Spec.OpsRequestOptions = &autoscalingapi.PgBouncerOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if r.Spec.OpsRequestOptions.Apply == "" {
		r.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

// +kubebuilder:webhook:path=/validate-schema-kubedb-com-v1alpha1-pgbouncerautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=schema.kubedb.com,resources=pgbouncerautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vpgbouncerautoscaler.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.CustomValidator = &autoscalingapi.PgBouncerAutoscaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *autoscalingapi.PgBouncerAutoscaler) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	pbLog.Info("validate create", "name", r.Name)
	return nil, r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *autoscalingapi.PgBouncerAutoscaler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	pbLog.Info("validate update", "name", r.Name)
	return nil, r.validate()
}

func (r *autoscalingapi.PgBouncerAutoscaler) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *autoscalingapi.PgBouncerAutoscaler) validate() error {
	if r.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var bouncer dbapi.PgBouncer
	err := autoscalingapi.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      r.Spec.DatabaseRef.Name,
		Namespace: r.Namespace,
	}, &bouncer)
	if err != nil {
		_ = fmt.Errorf("can't get PgBouncer %s/%s \n", r.Namespace, r.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
