package util

import (
	"testing"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

var lastTransactionTime = metav1.Now()
var lastTransactionTimePlusOne = metav1.NewTime(lastTransactionTime.Add(1 * time.Minute))

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
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "Some replicas are not ready",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionFalse,
				},
			},
			expectedPhase: api.DatabasePhaseCritical,
		},
		{
			name: "All replicas are ready but no other conditions present yet",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "Database is not accepting connection",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionFalse,
				},
			},
			expectedPhase: api.DatabasePhaseNotReady,
		},
		{
			name: "Database is accepting connection",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "1st restore: didn't completed yet",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
			},
			expectedPhase: api.DatabasePhaseDataRestoring,
		},
		{
			name: "1st restore: completed successfully",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
			},
			expectedPhase: api.DatabasePhaseProvisioning,
		},
		{
			name: "1st restore: failed to complete",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionFalse,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
			},
			expectedPhase: api.DatabasePhaseNotReady,
		},
		{
			name: "2nd restore: not completed yet (previous one succeeded)",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
			},
			expectedPhase: api.DatabasePhaseDataRestoring,
		},
		{
			name: "2nd restore: not completed yet (previous one failed)",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionFalse,
					LastTransitionTime: lastTransactionTime,
				},
			},
			expectedPhase: api.DatabasePhaseDataRestoring,
		},
		{
			name: "Database is not ready",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   api.DatabaseReady,
					Status: kmapi.ConditionFalse,
				},
			},
			expectedPhase: api.DatabasePhaseCritical,
		},
		{
			name: "Database is ready",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   api.DatabaseReady,
					Status: kmapi.ConditionTrue,
				},
			},
			expectedPhase: api.DatabasePhaseReady,
		},
		{
			name: "With conditions that does not have effect on phase",
			conditions: []kmapi.Condition{
				{
					Type:   api.DatabaseProvisioningStarted,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseReplicaReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseAcceptingConnection,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:               api.DatabaseDataRestoreStarted,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTime,
				},
				{
					Type:               api.DatabaseDataRestored,
					Status:             kmapi.ConditionTrue,
					LastTransitionTime: lastTransactionTimePlusOne,
				},
				{
					Type:   api.DatabaseReady,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabaseProvisioned,
					Status: kmapi.ConditionTrue,
				},
				{
					Type:   api.DatabasePaused,
					Status: kmapi.ConditionTrue,
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
			if got := CompareLastTransactionTime(tc.conditions, "type-1", "type-2"); got != tc.expected {
				t.Errorf("Expected: %d Found: %d", tc.expected, got)
			}
		})
	}
}
