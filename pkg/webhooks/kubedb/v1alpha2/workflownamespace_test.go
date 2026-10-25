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
	"strings"
	"testing"
	"time"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// wfnsStub serves the DurableWorkflows and the versions the webhook reads.
type wfnsStub struct {
	client.Client
	dws      map[types.NamespacedName]*olddbapi.DurableWorkflow
	versions map[string]*catalog.DurableWorkflowVersion
}

func (c wfnsStub) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	switch o := obj.(type) {
	case *olddbapi.DurableWorkflow:
		if dw, ok := c.dws[key]; ok {
			*o = *dw
			return nil
		}
		return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflows"}, key.Name)
	case *catalog.DurableWorkflowVersion:
		if v, ok := c.versions[key.Name]; ok {
			*o = *v
			return nil
		}
		return apierrors.NewNotFound(schema.GroupResource{Resource: "durableworkflowversions"}, key.Name)
	}
	panic("wfnsStub serves DurableWorkflow and DurableWorkflowVersion only")
}

func wfnsTestWebhook(t *testing.T) *WorkflowNamespaceCustomWebhook {
	t.Helper()
	stub := wfnsStub{
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
		dws: map[types.NamespacedName]*olddbapi.DurableWorkflow{
			{Namespace: "demo", Name: "tmp"}: {
				ObjectMeta: metav1.ObjectMeta{Name: "tmp", Namespace: "demo"},
				Spec:       olddbapi.DurableWorkflowSpec{Version: "temporal-1.32.1"},
			},
			{Namespace: "demo", Name: "cad"}: {
				ObjectMeta: metav1.ObjectMeta{Name: "cad", Namespace: "demo"},
				Spec: olddbapi.DurableWorkflowSpec{Version: "cadence-1.4.2", Archival: &olddbapi.DurableWorkflowArchival{
					History: &olddbapi.DurableWorkflowArchivalTarget{Provider: "s3", URI: "s3://b/p"},
				}},
			},
		},
	}
	return &WorkflowNamespaceCustomWebhook{DefaultClient: stub}
}

func wfns(db string, mutate func(*olddbapi.WorkflowNamespace)) *olddbapi.WorkflowNamespace {
	ns := &olddbapi.WorkflowNamespace{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "demo"},
		Spec:       olddbapi.WorkflowNamespaceSpec{DatabaseRef: core.LocalObjectReference{Name: db}},
	}
	if mutate != nil {
		mutate(ns)
	}
	return ns
}

func TestWorkflowNamespaceDefault(t *testing.T) {
	w := wfnsTestWebhook(t)
	ns := wfns("tmp", nil)
	if err := w.Default(context.TODO(), ns); err != nil {
		t.Fatal(err)
	}
	if ns.Spec.Name != "orders" || ns.Spec.Retention.Duration != 72*time.Hour || ns.Spec.DeletionPolicy != olddbapi.DeletionPolicyDelete {
		t.Fatalf("unexpected defaults: %+v", ns.Spec)
	}
	// explicit values are kept
	ns = wfns("tmp", func(n *olddbapi.WorkflowNamespace) {
		n.Spec.Name = "other"
		n.Spec.Retention = &metav1.Duration{Duration: time.Hour}
		n.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate
	})
	_ = w.Default(context.TODO(), ns)
	if ns.Spec.Name != "other" || ns.Spec.Retention.Duration != time.Hour || ns.Spec.DeletionPolicy != olddbapi.DeletionPolicyDoNotTerminate {
		t.Fatalf("defaults overwrote values: %+v", ns.Spec)
	}
}

func TestWorkflowNamespaceValidateCreate(t *testing.T) {
	w := wfnsTestWebhook(t)
	cases := []struct {
		name    string
		ns      *olddbapi.WorkflowNamespace
		wantErr string
	}{
		{"ok temporal", wfns("tmp", nil), ""},
		{"ok temporal short retention", wfns("tmp", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Retention = &metav1.Duration{Duration: time.Hour}
		}), ""},
		{"missing database", wfns("", nil), "databaseRef.name is missing"},
		{"unknown database", wfns("nope", nil), "databaseRef.name"},
		{"cadence retention under a day", wfns("cad", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Retention = &metav1.Duration{Duration: 12 * time.Hour}
		}), "at least 24h"},
		{"cadence retention of a day", wfns("cad", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Retention = &metav1.Duration{Duration: 24 * time.Hour}
		}), ""},
		{"zero retention", wfns("tmp", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Retention = &metav1.Duration{}
		}), "greater than 0"},
		{"archival without archival on the database", wfns("tmp", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Archival = &olddbapi.WorkflowNamespaceArchival{Enabled: true}
		}), "no archival configured"},
		{"archival with archival on the database", wfns("cad", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.Archival = &olddbapi.WorkflowNamespaceArchival{Enabled: true}
		}), ""},
		{"empty token secret name", wfns("tmp", func(n *olddbapi.WorkflowNamespace) {
			n.Spec.TokenSecretRef = &core.LocalObjectReference{}
		}), "tokenSecretRef.name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := w.ValidateCreate(context.TODO(), tc.ns)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestWorkflowNamespaceValidateUpdate(t *testing.T) {
	w := wfnsTestWebhook(t)
	old := wfns("tmp", func(n *olddbapi.WorkflowNamespace) { n.Spec.Name = "orders" })

	rename := wfns("tmp", func(n *olddbapi.WorkflowNamespace) { n.Spec.Name = "payments" })
	if _, err := w.ValidateUpdate(context.TODO(), old, rename); err == nil || !strings.Contains(err.Error(), "name is immutable") {
		t.Fatalf("rename must be refused, got %v", err)
	}
	// an unset name equals the default, so it is not a rename
	unset := wfns("tmp", nil)
	if _, err := w.ValidateUpdate(context.TODO(), old, unset); err != nil {
		t.Fatalf("unset name is the default: %v", err)
	}
	move := wfns("cad", func(n *olddbapi.WorkflowNamespace) {
		n.Spec.Retention = &metav1.Duration{Duration: 48 * time.Hour}
	})
	if _, err := w.ValidateUpdate(context.TODO(), old, move); err == nil || !strings.Contains(err.Error(), "databaseRef is immutable") {
		t.Fatalf("databaseRef change must be refused, got %v", err)
	}
	longer := wfns("tmp", func(n *olddbapi.WorkflowNamespace) { n.Spec.Retention = &metav1.Duration{Duration: 240 * time.Hour} })
	if _, err := w.ValidateUpdate(context.TODO(), old, longer); err != nil {
		t.Fatalf("retention change must pass: %v", err)
	}
}

func TestWorkflowNamespaceValidateDelete(t *testing.T) {
	w := wfnsTestWebhook(t)
	if _, err := w.ValidateDelete(context.TODO(), wfns("tmp", nil)); err != nil {
		t.Fatal(err)
	}
	guarded := wfns("tmp", func(n *olddbapi.WorkflowNamespace) { n.Spec.DeletionPolicy = olddbapi.DeletionPolicyDoNotTerminate })
	if _, err := w.ValidateDelete(context.TODO(), guarded); err == nil {
		t.Fatal("DoNotTerminate must refuse the deletion")
	}
}
