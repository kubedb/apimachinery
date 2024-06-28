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

package phase

import (
	"testing"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

var (
	lastTransactionTime        = metav1.Now()
	lastTransactionTimePlusOne = metav1.NewTime(lastTransactionTime.Add(1 * time.Minute))
)

func TestPhaseForCondition(t *testing.T) {
	testCases := []struct {
		name          string
		conditions    []kmapi.Condition
		expectedPhase api.DatabasePhase
	}{
		{
			name:          "No condition present yet",
			conditions:    nil,
			expectedPhase: "",
		},
		{
			name: "Provisioning just started",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "Some replicas are not ready",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionFalse,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "All replicas are ready but no other conditions present yet",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "Database is not accepting connection",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionFalse,
				},
				{
					Type:   kubedb.DatabaseProvisioned,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseNotReady,
		},
		{
			name: "Database is accepting connection",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "1st restore: didn't completed yet",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestoreStarted,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
			},
			expectedPhase: api.DatabasePhaseDataRestoring,
		},
		{
			name: "1st restore: completed successfully",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "1st restore: failed to complete",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionFalse,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
			},
			expectedPhase: api.DatabasePhaseNotReady,
		},
		{
			name: "Database is not ready",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   kubedb.DatabaseProvisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReady,
					Status: metav1.ConditionFalse,
				},
			},
			expectedPhase: api.DatabasePhaseCritical,
		},
		{
			name: "Database is ready",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   kubedb.DatabaseProvisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReady,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseReady,
		},
		{
			name: "Database is ready but not accepting connection",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionFalse,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   kubedb.DatabaseProvisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReady,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseNotReady,
		},
		{
			name: "With conditions that does not have effect on phase",
			conditions: []kmapi.Condition{
				{
					Type:   kubedb.DatabaseProvisioningStarted,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseReplicaReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseAcceptingConnection,
					Status: metav1.ConditionTrue,
				},
				{
					Type:               kubedb.DatabaseDataRestored,
					Status:             metav1.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   kubedb.DatabaseReady,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabaseProvisioned,
					Status: metav1.ConditionTrue,
				},
				{
					Type:   kubedb.DatabasePaused,
					Status: metav1.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseReady,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PhaseFromCondition(tc.conditions); got != tc.expectedPhase {
				t.Errorf("Expected: %s Found: %s", tc.expectedPhase, got)
			}
		})
	}
}

func TestCompareLastTransactionTime(t *testing.T) {
	testCases := []struct {
		name       string
		conditions []kmapi.Condition
		expected   int32
	}{
		{
			name:       "Both conditions does not exist",
			conditions: nil,
			expected:   0,
		},
		{
			name: "Only first condition exist",
			conditions: []kmapi.Condition{
				{
					Type: "type-1",
				},
			},
			expected: 1,
		},
		{
			name: "Only second condition exist",
			conditions: []kmapi.Condition{
				{
					Type: "type-2",
				},
			},
			expected: -1,
		},
		{
			name: "Both condition was created at the same time",
			conditions: []kmapi.Condition{
				{
					Type:               "type-1",
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               "type-2",
					LastTransitionTime: lastTransactionTime,
				},
			},
			expected: 0,
		},
		{
			name: "First condition is older",
			conditions: []kmapi.Condition{
				{
					Type:               "type-1",
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               "type-2",
					LastTransitionTime: lastTransactionTimePlusOne,
				},
			},
			expected: -1,
		},
		{
			name: "Second condition is older",
			conditions: []kmapi.Condition{
				{
					Type:               "type-1",
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:               "type-2",
					LastTransitionTime: lastTransactionTime,
				},
			},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareLastTransactionTime(tc.conditions, "type-1", "type-2"); got != tc.expected {
				t.Errorf("Expected: %d Found: %d", tc.expected, got)
			}
		})
	}
}
