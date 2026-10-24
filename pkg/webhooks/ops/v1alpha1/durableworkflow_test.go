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
	"strings"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// dwStubClient serves the DurableWorkflow and the versions of a test by name.
type dwStubClient struct {
	client.Client
	db       *dbapi.DurableWorkflow
	versions map[string]*catalog.DurableWorkflowVersion
}

func (c dwStubClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	switch o := obj.(type) {
	case *dbapi.DurableWorkflow:
		if c.db == nil || c.db.Name != key.Name {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflows"}, key.Name)
		}
		*o = *c.db
	case *catalog.DurableWorkflowVersion:
		v, ok := c.versions[key.Name]
		if !ok {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflowversions"}, key.Name)
		}
		*o = *v
	default:
		panic("unexpected type in dwStubClient")
	}
	return nil
}

func dwOpsWebhook() *DurableWorkflowOpsRequestCustomWebhook {
	version := func(name string, d catalog.DurableWorkflowDistro) *catalog.DurableWorkflowVersion {
		return &catalog.DurableWorkflowVersion{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       catalog.DurableWorkflowVersionSpec{Version: name, Distribution: d},
		}
	}
	db := &dbapi.DurableWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "dw", Namespace: "demo"},
		Spec:       dbapi.DurableWorkflowSpec{Version: "temporal-1.32.1"},
	}
	return &DurableWorkflowOpsRequestCustomWebhook{DefaultClient: dwStubClient{
		db: db,
		versions: map[string]*catalog.DurableWorkflowVersion{
			"temporal-1.32.1": version("temporal-1.32.1", catalog.DurableWorkflowDistroTemporal),
			"cadence-1.4.2":   version("cadence-1.4.2", catalog.DurableWorkflowDistroCadence),
		},
	}}
}

func TestDurableWorkflowOpsRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    opsapi.DurableWorkflowOpsRequestSpec
		wantErr string
	}{
		{name: "restart", spec: opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeRestart}},
		{name: "unknown type", spec: opsapi.DurableWorkflowOpsRequestSpec{Type: "VolumeExpansion"}, wantErr: "is not supported"},
		{
			name:    "horizontal scaling needs its spec",
			spec:    opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeHorizontalScaling},
			wantErr: "spec.horizontalScaling is required",
		},
		{
			name: "horizontal scaling",
			spec: opsapi.DurableWorkflowOpsRequestSpec{
				Type:              opsapi.DurableWorkflowOpsRequestTypeHorizontalScaling,
				HorizontalScaling: &opsapi.DurableWorkflowHorizontalScalingSpec{History: ptr.To(int32(3))},
			},
		},
		{
			name:    "vertical scaling needs its spec",
			spec:    opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeVerticalScaling},
			wantErr: "spec.verticalScaling is required",
		},
		{
			name: "the number of history shards can't be reconfigured",
			spec: opsapi.DurableWorkflowOpsRequestSpec{
				Type:          opsapi.DurableWorkflowOpsRequestTypeReconfigure,
				Configuration: &opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"numHistoryShards": "64"}},
			},
			wantErr: "numHistoryShards can't be changed",
		},
		{
			name: "reconfigure",
			spec: opsapi.DurableWorkflowOpsRequestSpec{
				Type:          opsapi.DurableWorkflowOpsRequestTypeReconfigure,
				Configuration: &opsapi.ReconfigurationSpec{ApplyConfig: map[string]string{"dynamic_config.yaml": "a: []"}},
			},
		},
		{
			name: "update to another engine is a migration",
			spec: opsapi.DurableWorkflowOpsRequestSpec{
				Type:          opsapi.DurableWorkflowOpsRequestTypeUpdateVersion,
				UpdateVersion: &opsapi.DurableWorkflowUpdateVersionSpec{TargetVersion: "cadence-1.4.2"},
			},
			wantErr: "not supported by UpdateVersion",
		},
		{
			name: "update to an unknown version",
			spec: opsapi.DurableWorkflowOpsRequestSpec{
				Type:          opsapi.DurableWorkflowOpsRequestTypeUpdateVersion,
				UpdateVersion: &opsapi.DurableWorkflowUpdateVersionSpec{TargetVersion: "temporal-9.9.9"},
			},
			wantErr: "target version temporal-9.9.9",
		},
		{
			name:    "update needs its spec",
			spec:    opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeUpdateVersion},
			wantErr: "spec.updateVersion nil",
		},
	}

	w := dwOpsWebhook()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.spec.DatabaseRef.Name = "dw"
			req := &opsapi.DurableWorkflowOpsRequest{
				ObjectMeta: metav1.ObjectMeta{Name: "req", Namespace: "demo"},
				Spec:       tc.spec,
			}
			_, err := w.validateCreateOrUpdate(req)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}

	t.Run("the cluster must exist", func(t *testing.T) {
		req := &opsapi.DurableWorkflowOpsRequest{
			ObjectMeta: metav1.ObjectMeta{Name: "req", Namespace: "demo"},
			Spec:       opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeRestart, DatabaseRef: corev1LocalRef("missing")},
		}
		if _, err := w.validateCreateOrUpdate(req); err == nil || !strings.Contains(err.Error(), "is invalid or not found") {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("the spec is immutable", func(t *testing.T) {
		a := &opsapi.DurableWorkflowOpsRequest{Spec: opsapi.DurableWorkflowOpsRequestSpec{Type: opsapi.DurableWorkflowOpsRequestTypeRestart}}
		b := a.DeepCopy()
		if err := validateImmutableSpec(a, b); err != nil {
			t.Errorf("an unchanged spec must pass: %v", err)
		}
		b.Spec.Type = opsapi.DurableWorkflowOpsRequestTypeRotateAuth
		if err := validateImmutableSpec(a, b); err == nil {
			t.Error("a changed spec must be refused")
		}
	})
}

func corev1LocalRef(name string) corev1.LocalObjectReference {
	return corev1.LocalObjectReference{Name: name}
}
