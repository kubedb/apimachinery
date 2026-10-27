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
	"testing"

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
)

func TestNATSAutoscalerWebhook(t *testing.T) {
	w := &NATSAutoscalerCustomWebhook{}
	ctx := context.TODO()

	scaler := &autoscalingapi.NATSAutoscaler{}
	if _, err := w.ValidateCreate(ctx, scaler); err == nil {
		t.Fatalf("a NATSAutoscaler without databaseRef must be rejected")
	}

	scaler.Spec.DatabaseRef = &core.LocalObjectReference{Name: "nats"}
	if _, err := w.ValidateCreate(ctx, scaler); err != nil {
		t.Fatal(err)
	}
	if _, err := w.ValidateUpdate(ctx, scaler, scaler); err != nil {
		t.Fatal(err)
	}

	if err := w.Default(ctx, scaler); err != nil {
		t.Fatal(err)
	}
	if scaler.Spec.OpsRequestOptions == nil || scaler.Spec.OpsRequestOptions.Apply != opsapi.ApplyOptionIfReady {
		t.Fatalf("opsRequestOptions must default to IfReady, got %+v", scaler.Spec.OpsRequestOptions)
	}
}
