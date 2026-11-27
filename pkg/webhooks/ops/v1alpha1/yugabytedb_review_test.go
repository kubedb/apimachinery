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

	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestYugabyteDBOpsRollbackRefusedOnceFinalizing(t *testing.T) {
	for _, state := range []string{kubedb.YugabyteDBUpgradeStateFinalizing, kubedb.YugabyteDBUpgradeStateFinalized} {
		db := func() *dbapi.YugabyteDB {
			d := ybPendingFinalizeDB()
			d.Annotations = map[string]string{kubedb.YugabyteDBUpgradeStateAnnotation: state}
			return d
		}
		runYBOpsCases(t, []ybOpsCase{
			{name: "rollback " + state, db: db(), req: func() *opsapi.YugabyteDBOpsRequest {
				return ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeRollbackUpgrade, "")
			}, want: "can not be rolled back"},
			{name: "finalize " + state, db: db(), req: func() *opsapi.YugabyteDBOpsRequest {
				return ybUpgradeReq("op", opsapi.YugabyteDBOpsRequestTypeFinalizeUpgrade, "")
			}},
		})
	}
}

func TestYugabyteDBOpsConnectionManagerNeedsRestart(t *testing.T) {
	cm := func(restart opsapi.ReconfigureRestartType) func() *opsapi.YugabyteDBOpsRequest {
		return func() *opsapi.YugabyteDBOpsRequest {
			r := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeReconfigure)
			r.Spec.Configuration = &opsapi.YugabyteDBReconfigurationSpec{
				ReconfigurationSpec: opsapi.ReconfigurationSpec{Restart: restart},
				ConnectionManager:   &dbapi.YugabyteDBConnectionManager{Enabled: true},
			}
			return r
		}
	}
	runYBOpsCases(t, []ybOpsCase{
		{name: "enable, restart auto", req: cm(opsapi.ReconfigureRestartAuto)},
		{name: "enable, restart unset", req: cm("")},
		{name: "enable, restart false", req: cm(opsapi.ReconfigureRestartFalse), want: "must not be false"},
	})
}

func TestYugabyteDBOpsDeleteResumes(t *testing.T) {
	paused := func() *dbapi.YugabyteDB {
		db := ybOpsDB(nil, false)
		db.Status.Conditions = []kmapi.Condition{{Type: "Paused", Status: metav1.ConditionTrue}}
		return db
	}
	isPaused := func(t *testing.T, w *YugabyteDBOpsRequestCustomWebhook) bool {
		t.Helper()
		got := &dbapi.YugabyteDB{}
		if err := w.DefaultClient.Get(context.TODO(), client.ObjectKey{Namespace: "demo", Name: "yb"}, got); err != nil {
			t.Fatal(err)
		}
		for _, c := range got.Status.Conditions {
			if c.Type == "Paused" {
				return true
			}
		}
		return false
	}
	cases := []struct {
		name       string
		phase      opsapi.OpsRequestPhase
		dryRun     bool
		wantPaused bool
	}{
		{"progressing", opsapi.OpsRequestPhaseProgressing, false, false},
		{"progressing, dry run", opsapi.OpsRequestPhaseProgressing, true, true},
		{"pending", opsapi.OpsRequestPhasePending, false, true},
		{"successful", opsapi.OpsRequestPhaseSuccessful, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := ybOpsWebhook(t, paused())
			req := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart)
			req.Status.Phase = c.phase
			ctx := admission.NewContextWithRequest(context.TODO(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{DryRun: ptr.To(c.dryRun)}})
			if _, err := w.ValidateDelete(ctx, req); err != nil {
				t.Fatal(err)
			}
			if got := isPaused(t, w); got != c.wantPaused {
				t.Fatalf("paused = %v, want %v", got, c.wantPaused)
			}
		})
	}
	// the database is gone: nothing to resume
	w := ybOpsWebhook(t)
	req := ybOpsReq(opsapi.YugabyteDBOpsRequestTypeRestart)
	req.Status.Phase = opsapi.OpsRequestPhaseProgressing
	if _, err := w.ValidateDelete(context.TODO(), req); err != nil {
		t.Fatal(err)
	}
}
