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

	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var msLog = logf.Log.WithName("mssqlserver-autoscaler")

var _ webhook.CustomDefaulter = &MSSQLServerAutoscaler{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *MSSQLServerAutoscaler) Default(ctx context.Context, obj runtime.Object) error {
	msLog.Info("defaulting", "name", r.Name)
	r.setDefaults()
	return nil
}

func (r *MSSQLServerAutoscaler) setDefaults() {
	var db olddbapi.MSSQLServer
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      r.Spec.DatabaseRef.Name,
		Namespace: r.Namespace,
	}, &db)
	if err != nil {
		_ = fmt.Errorf("can't get MSSQLServer %s/%s \n", r.Namespace, r.Spec.DatabaseRef.Name)
		return
	}

	r.setOpsReqOptsDefaults()

	if r.Spec.Storage != nil {
		setDefaultStorageValues(r.Spec.Storage.MSSQLServer)
	}

	if r.Spec.Compute != nil {
		setDefaultComputeValues(r.Spec.Compute.MSSQLServer)
	}
}

func (r *MSSQLServerAutoscaler) setOpsReqOptsDefaults() {
	if r.Spec.OpsRequestOptions == nil {
		r.Spec.OpsRequestOptions = &MSSQLServerOpsRequestOptions{}
	}
	// Timeout is defaulted to 600s in ops-manager retries.go (to retry 120 times with 5sec pause between each)
	// OplogMaxLagSeconds & ObjectsCountDiffPercentage are defaults to 0
	if r.Spec.OpsRequestOptions.Apply == "" {
		r.Spec.OpsRequestOptions.Apply = opsapi.ApplyOptionIfReady
	}
}

var _ webhook.CustomValidator = &MSSQLServerAutoscaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MSSQLServerAutoscaler) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	msLog.Info("validate create", "name", r.Name)
	return nil, r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MSSQLServerAutoscaler) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	msLog.Info("validate create", "name", r.Name)
	return nil, r.validate()
}

func (r *MSSQLServerAutoscaler) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (r *MSSQLServerAutoscaler) validate() error {
	if r.Spec.DatabaseRef == nil {
		return errors.New("databaseRef can't be empty")
	}
	var kf olddbapi.MSSQLServer
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name:      r.Spec.DatabaseRef.Name,
		Namespace: r.Namespace,
	}, &kf)
	if err != nil {
		_ = fmt.Errorf("can't get MSSQLServer %s/%s \n", r.Namespace, r.Spec.DatabaseRef.Name)
		return err
	}

	return nil
}
