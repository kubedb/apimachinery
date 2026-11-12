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
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func ybAutoscalerWebhook(t *testing.T, objs ...*olddbapi.YugabyteDB) *YugabyteDBAutoscalerCustomWebhook {
	t.Helper()
	s := runtime.NewScheme()
	if err := olddbapi.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	b := fake.NewClientBuilder().WithScheme(s)
	for _, o := range objs {
		b = b.WithObjects(o)
	}
	return &YugabyteDBAutoscalerCustomWebhook{DefaultClient: b.Build()}
}

func TestYugabyteDBAutoscalerDefaultsAndValidation(t *testing.T) {
	db := &olddbapi.YugabyteDB{ObjectMeta: metav1.ObjectMeta{Name: "yb", Namespace: "demo"}}
	w := ybAutoscalerWebhook(t, db)

	scaler := &autoscalingapi.YugabyteDBAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: "as", Namespace: "demo"},
		Spec: autoscalingapi.YugabyteDBAutoscalerSpec{
			DatabaseRef: &core.LocalObjectReference{Name: "yb"},
			Compute: &autoscalingapi.YugabyteDBComputeAutoscalerSpec{
				TServer: &autoscalingapi.ComputeAutoscalerSpec{},
			},
			Storage: &autoscalingapi.YugabyteDBStorageAutoscalerSpec{
				Master: &autoscalingapi.StorageAutoscalerSpec{},
			},
		},
	}
	if err := w.Default(context.TODO(), scaler); err != nil {
		t.Fatal(err)
	}
	if o := scaler.Spec.OpsRequestOptions; o == nil || o.Apply != opsapi.ApplyOptionIfReady || o.MaxRetries != 1 {
		t.Fatalf("ops request options not defaulted: %+v", o)
	}
	if scaler.Spec.Compute.TServer.Trigger != autoscalingapi.AutoscalerTriggerOff {
		t.Fatalf("compute trigger = %q", scaler.Spec.Compute.TServer.Trigger)
	}
	if st := scaler.Spec.Storage.Master; st.Trigger != autoscalingapi.AutoscalerTriggerOff || st.UsageThreshold == nil {
		t.Fatalf("storage not defaulted: %+v", st)
	}
	if scaler.Spec.Compute.Master != nil || scaler.Spec.Storage.TServer != nil {
		t.Fatal("unset roles must stay nil")
	}

	if _, err := w.ValidateCreate(context.TODO(), scaler); err != nil {
		t.Fatalf("valid autoscaler rejected: %v", err)
	}

	missingRef := scaler.DeepCopy()
	missingRef.Spec.DatabaseRef = nil
	if _, err := w.ValidateCreate(context.TODO(), missingRef); err == nil {
		t.Fatal("nil databaseRef must be rejected")
	}
	unknownDB := scaler.DeepCopy()
	unknownDB.Spec.DatabaseRef.Name = "nope"
	if _, err := w.ValidateUpdate(context.TODO(), scaler, unknownDB); err == nil {
		t.Fatal("unknown database must be rejected")
	}
}
