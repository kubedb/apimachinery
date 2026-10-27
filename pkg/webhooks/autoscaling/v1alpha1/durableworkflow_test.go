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

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// dwStubClient serves DurableWorkflows and DurableWorkflowVersions from memory.
type dwStubClient struct {
	client.Client
	dbs      map[string]*olddbapi.DurableWorkflow
	versions map[string]*catalog.DurableWorkflowVersion
}

func (c dwStubClient) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	switch o := obj.(type) {
	case *olddbapi.DurableWorkflow:
		found, ok := c.dbs[key.Namespace+"/"+key.Name]
		if !ok {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflows"}, key.Name)
		}
		*o = *found
	case *catalog.DurableWorkflowVersion:
		found, ok := c.versions[key.Name]
		if !ok {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflowversions"}, key.Name)
		}
		*o = *found
	default:
		panic("dwStubClient serves only DurableWorkflow and DurableWorkflowVersion")
	}
	return nil
}

func dwScalerWebhook(withInternalFrontend bool) *DurableWorkflowAutoscalerCustomWebhook {
	node := &olddbapi.DurableWorkflowNode{Replicas: ptr.To[int32](2)}
	topology := &olddbapi.DurableWorkflowTopology{Frontend: node, History: node, Matching: node, Worker: node}
	if withInternalFrontend {
		topology.InternalFrontend = node
	}
	mk := func(name, version string) *olddbapi.DurableWorkflow {
		return &olddbapi.DurableWorkflow{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "demo"},
			Spec:       olddbapi.DurableWorkflowSpec{Version: version, Topology: topology},
		}
	}
	return &DurableWorkflowAutoscalerCustomWebhook{DefaultClient: dwStubClient{
		dbs: map[string]*olddbapi.DurableWorkflow{
			"demo/temporal": mk("temporal", "temporal-1.32.1"),
			"demo/cadence":  mk("cadence", "cadence-1.4.2"),
		},
		versions: map[string]*catalog.DurableWorkflowVersion{
			"temporal-1.32.1": {
				ObjectMeta: metav1.ObjectMeta{Name: "temporal-1.32.1"},
				Spec:       catalog.DurableWorkflowVersionSpec{Version: "1.32.1", Distribution: catalog.DurableWorkflowDistroTemporal},
			},
			"cadence-1.4.2": {
				ObjectMeta: metav1.ObjectMeta{Name: "cadence-1.4.2"},
				Spec:       catalog.DurableWorkflowVersionSpec{Version: "1.4.2", Distribution: catalog.DurableWorkflowDistroCadence},
			},
		},
	}}
}

func dwScaler(db string, compute *autoscalingapi.DurableWorkflowComputeAutoscalerSpec) *autoscalingapi.DurableWorkflowAutoscaler {
	return &autoscalingapi.DurableWorkflowAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: "scaler", Namespace: "demo"},
		Spec: autoscalingapi.DurableWorkflowAutoscalerSpec{
			DatabaseRef: &core.LocalObjectReference{Name: db},
			Compute:     compute,
		},
	}
}

func TestDurableWorkflowAutoscalerDefaults(t *testing.T) {
	w := dwScalerWebhook(true)
	scaler := dwScaler("temporal", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{
		Frontend: &autoscalingapi.ComputeAutoscalerSpec{Trigger: autoscalingapi.AutoscalerTriggerOn},
		History:  &autoscalingapi.ComputeAutoscalerSpec{},
	})
	if err := w.Default(context.TODO(), scaler); err != nil {
		t.Fatal(err)
	}

	if o := scaler.Spec.OpsRequestOptions; o == nil || o.Apply != opsapi.ApplyOptionIfReady || o.MaxRetries != 1 {
		t.Errorf("unexpected opsRequestOptions: %+v", o)
	}
	fe := scaler.Spec.Compute.Frontend
	if fe.Trigger != autoscalingapi.AutoscalerTriggerOn {
		t.Errorf("an explicit trigger must be kept, got %q", fe.Trigger)
	}
	if len(fe.ControlledResources) != 2 || fe.ResourceDiffPercentage != autoscalingapi.DefaultResourceDiffPercentage {
		t.Errorf("frontend defaults not set: %+v", fe)
	}
	if h := scaler.Spec.Compute.History; h.Trigger != autoscalingapi.AutoscalerTriggerOff || h.ContainerControlledValues == nil {
		t.Errorf("history defaults not set: %+v", h)
	}
	if scaler.Spec.Compute.Worker != nil {
		t.Error("a role that is not configured must stay nil")
	}
}

func TestDurableWorkflowAutoscalerDefaultsMissingDatabase(t *testing.T) {
	w := dwScalerWebhook(false)
	scaler := dwScaler("missing", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{
		Frontend: &autoscalingapi.ComputeAutoscalerSpec{},
	})
	if err := w.Default(context.TODO(), scaler); err != nil {
		t.Fatal(err)
	}
	if scaler.Spec.Compute.Frontend.Trigger != "" || scaler.Spec.OpsRequestOptions != nil {
		t.Error("nothing must be defaulted when the database can't be read")
	}
}

func TestDurableWorkflowAutoscalerValidate(t *testing.T) {
	cpu := func(v string) core.ResourceList { return core.ResourceList{core.ResourceCPU: resource.MustParse(v)} }
	on := func() *autoscalingapi.ComputeAutoscalerSpec {
		return &autoscalingapi.ComputeAutoscalerSpec{Trigger: autoscalingapi.AutoscalerTriggerOn}
	}
	tests := []struct {
		name    string
		webhook *DurableWorkflowAutoscalerCustomWebhook
		scaler  *autoscalingapi.DurableWorkflowAutoscaler
		wantErr string
	}{
		{
			name:    "all roles on Temporal",
			webhook: dwScalerWebhook(true),
			scaler: dwScaler("temporal", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{
				Frontend: on(), History: on(), Matching: on(), Worker: on(), InternalFrontend: on(),
			}),
		},
		{
			name:    "single role on Cadence",
			webhook: dwScalerWebhook(false),
			scaler:  dwScaler("cadence", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{History: on()}),
		},
		{
			name:    "no compute is accepted",
			webhook: dwScalerWebhook(false),
			scaler:  dwScaler("temporal", nil),
		},
		{
			name:    "missing databaseRef",
			webhook: dwScalerWebhook(false),
			scaler:  &autoscalingapi.DurableWorkflowAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "scaler", Namespace: "demo"}},
			wantErr: "databaseRef can't be empty",
		},
		{
			name:    "database not found",
			webhook: dwScalerWebhook(false),
			scaler:  dwScaler("missing", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{History: on()}),
			wantErr: "not found",
		},
		{
			name:    "empty compute",
			webhook: dwScalerWebhook(false),
			scaler:  dwScaler("temporal", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{}),
			wantErr: "at least one",
		},
		{
			name:    "internalFrontend on Cadence",
			webhook: dwScalerWebhook(true),
			scaler:  dwScaler("cadence", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{InternalFrontend: on()}),
			wantErr: "only available for Temporal",
		},
		{
			name:    "internalFrontend not in the topology",
			webhook: dwScalerWebhook(false),
			scaler:  dwScaler("temporal", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{InternalFrontend: on()}),
			wantErr: "no spec.topology.internalFrontend",
		},
		{
			name:    "minAllowed above maxAllowed",
			webhook: dwScalerWebhook(false),
			scaler: dwScaler("temporal", &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{
				Worker: &autoscalingapi.ComputeAutoscalerSpec{MinAllowed: cpu("2"), MaxAllowed: cpu("1")},
			}),
			wantErr: "spec.compute.worker: minAllowed cpu",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.webhook.ValidateCreate(context.TODO(), tt.scaler)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestDurableWorkflowComputeSpecForRole(t *testing.T) {
	spec := &autoscalingapi.DurableWorkflowComputeAutoscalerSpec{
		Frontend: &autoscalingapi.ComputeAutoscalerSpec{}, InternalFrontend: &autoscalingapi.ComputeAutoscalerSpec{},
	}
	for role, want := range map[string]*autoscalingapi.ComputeAutoscalerSpec{
		"frontend": spec.Frontend, "internal-frontend": spec.InternalFrontend, "history": nil, "worker": nil, "ui": nil,
	} {
		if got := spec.ComputeSpecForRole(role); got != want {
			t.Errorf("role %s: got %p want %p", role, got, want)
		}
	}
	var nilSpec *autoscalingapi.DurableWorkflowComputeAutoscalerSpec
	if nilSpec.ComputeSpecForRole("frontend") != nil || !nilSpec.IsEmpty() {
		t.Error("a nil spec must be empty")
	}
}
