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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupYugabyteDBAutoscalerWebhookWithManager registers the webhook for YugabyteDBAutoscaler in the manager.
func SetupYugabyteDBAutoscalerWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&autoscalingapi.YugabyteDBAutoscaler{}).
		WithValidator(&YugabyteDBAutoscalerCustomWebhook{mgr.GetClient()}).
		WithDefaulter(&YugabyteDBAutoscalerCustomWebhook{mgr.GetClient()}).
		Complete()
}

type YugabyteDBAutoscalerCustomWebhook struct {
	DefaultClient client.Client
}

// log is for logging in this package.
var ybLog = logf.Log.WithName("yugabytedb-autoscaler")

var _ webhook.CustomDefaulter = &YugabyteDBAutoscalerCustomWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (w *YugabyteDBAutoscalerCustomWebhook) Default(ctx context.Context, obj runtime.Object) error {
	scaler, ok := obj.(*autoscalingapi.YugabyteDBAutoscaler)
	if !ok {
		return fmt.Errorf("expected an YugabyteDBAutoscaler object but got %T", obj)
	}
	ybLog.Info("defaulting", "name", scaler.Name)
	w.setDefaults(scaler)
	return nil
}

func (w *YugabyteDBAutoscalerCustomWebhook) setDefaults(scaler *autoscalingapi.YugabyteDBAutoscaler) {
	w.setOpsReqOptsDefaults(scaler)

	if scaler.Spec.Storage != nil {
		setDefaultStorageValues(scaler.Spec.Storage.Master)
		setDefaultStorageValues(scaler.Spec.Storage.TServer)
		setDefaultStorageValues(scaler.Spec.Storage.ReadReplica)
	}

	if scaler.Spec.Compute != nil {
		setDefaultComputeValues(scaler.Spec.Compute.Master)
		setDefaultComputeValues(scaler.Spec.Compute.TServer)
		setDefaultComputeValues(scaler.Spec.Compute.ReadReplica)
	}
}

func (w *YugabyteDBAutoscalerCustomWebhook) setOpsReqOptsDefaults(scaler *autoscalingapi.YugabyteDBAutoscaler) {
	if scaler.Spec.OpsRequestOptions == nil {
		scaler.Spec.OpsRequestOptions = &autoscalingapi.OpsRequestOptions{
			Apply:      opsapi.ApplyOptionIfReady,
			MaxRetries: 1,
		}
	}
}

var _ webhook.CustomValidator = &YugabyteDBAutoscalerCustomWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *YugabyteDBAutoscalerCustomWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	scaler, ok := obj.(*autoscalingapi.YugabyteDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an YugabyteDBAutoscaler object but got %T", obj)
	}
	ybLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *YugabyteDBAutoscalerCustomWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	scaler, ok := newObj.(*autoscalingapi.YugabyteDBAutoscaler)
	if !ok {
		return nil, fmt.Errorf("expected an YugabyteDBAutoscaler object but got %T", newObj)
	}
	ybLog.Info("validate create", "name", scaler.Name)
	return nil, w.validate(scaler)
}

func (w *YugabyteDBAutoscalerCustomWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *YugabyteDBAutoscalerCustomWebhook) validate(scaler *autoscalingapi.YugabyteDBAutoscaler) error {
	if scaler.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var db olddbapi.YugabyteDB
	err := w.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      scaler.Spec.DatabaseRef.Name,
		Namespace: scaler.Namespace,
	}, &db)
	if err != nil {
		klog.Errorf("can't get YugabyteDB %s/%s", scaler.Namespace, scaler.Spec.DatabaseRef.Name)
		return err
	}
	return validateYugabyteDBAutoscalerTopology(scaler, &db)
}

// validateYugabyteDBAutoscalerTopology rejects a role section the database can not act on, as the MongoDB
// webhook does for its topology: a read replica section without read replicas, and a storage section for a
// role on ephemeral storage. Such a section would be accepted and silently do nothing.
func validateYugabyteDBAutoscalerTopology(scaler *autoscalingapi.YugabyteDBAutoscaler, db *olddbapi.YugabyteDB) error {
	rr := db.IsReadReplicaEnabled()
	if c := scaler.Spec.Compute; c != nil && c.ReadReplica != nil && !rr {
		return errors.New("spec.compute.readReplica: the database has no read replicas")
	}
	s := scaler.Spec.Storage
	if s == nil {
		return nil
	}
	if s.ReadReplica != nil && !rr {
		return errors.New("spec.storage.readReplica: the database has no read replicas")
	}
	for _, r := range []struct {
		name string
		spec *autoscalingapi.StorageAutoscalerSpec
		role olddbapi.YugabyteDBRole
	}{
		{"master", s.Master, olddbapi.YugabyteDBRoleMaster},
		{"tserver", s.TServer, olddbapi.YugabyteDBRoleTServer},
		{"readReplica", s.ReadReplica, olddbapi.YugabyteDBRoleReadReplica},
	} {
		if r.spec == nil {
			continue
		}
		if n := db.Node(r.role); n != nil && n.StorageType == olddbapi.StorageTypeEphemeral {
			return fmt.Errorf("spec.storage.%s: the %s pods use ephemeral storage, there is no volume to expand", r.name, r.name)
		}
	}
	return nil
}
