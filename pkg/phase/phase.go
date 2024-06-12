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
	dapi "kubedb.dev/apimachinery/apis/elasticsearch/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	apiv1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	apiv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	kmapi "kmodules.xyz/client-go/api/v1"
	cutil "kmodules.xyz/client-go/conditions"
)

func DashboardPhaseFromCondition(conditions []kmapi.Condition) dapi.DashboardPhase {
	if !cutil.IsConditionTrue(conditions, string(dapi.DashboardConditionProvisioned)) {
		return dapi.DashboardPhaseProvisioning
	}

	if !cutil.IsConditionTrue(conditions, string(dapi.DashboardConditionAcceptingConnection)) {
		return dapi.DashboardPhaseNotReady
	}

	// TODO: implement deployment watcher to handle replica ready

	if cutil.HasCondition(conditions, string(dapi.DashboardConditionServerHealthy)) {

		if !cutil.IsConditionTrue(conditions, string(dapi.DashboardConditionServerHealthy)) {

			_, cond := cutil.GetCondition(conditions, string(dapi.DashboardConditionServerHealthy))

			if cond.Reason == dapi.DashboardStateRed {
				return dapi.DashboardPhaseNotReady
			} else {
				return dapi.DashboardPhaseCritical
			}
		}

		return dapi.DashboardPhaseReady
	}

	return dapi.DashboardPhaseNotReady
}

func PhaseFromCondition(conditions []kmapi.Condition) apiv1alpha2.DatabasePhase {
	// Generally, the conditions should maintain the following chronological order
	// For normal restore process:
	//   ProvisioningStarted --> ReplicaReady --> AcceptingConnection --> DataRestoreStarted --> DataRestored --> Ready --> Provisioned
	// For restoring the volumes (PerconaXtraDB):
	//	 ProvisioningStarted --> DataRestoreStarted --> DataRestored --> ReplicaReady --> AcceptingConnection --> Ready --> Provisioned

	// These are transitional conditions. They can update any time. So, their order may vary:
	// 1. ReplicaReady
	// 2. AcceptingConnection
	// 3. DataRestoreStarted
	// 4. DataRestored
	// 5. Ready
	// 6. Paused
	// 7. HealthCheckPaused

	var phase apiv1alpha2.DatabasePhase

	// ================================= Handling "HealthCheckPaused" condition ==========================
	// If the condition is present and its "true", then the phase should be "Unknown".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseHealthCheckPaused) {
		return apiv1alpha2.DatabasePhaseUnknown
	}

	// ==================================  Handling "ProvisioningStarted" condition  ========================
	// If the condition is present and its "true", then the phase should be "Provisioning".
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioningStarted) {
		phase = apiv1alpha2.DatabasePhaseProvisioning
	}

	// ================================== Handling "Halted" condition =======================================
	// The "Halted" condition has higher priority, that's why it is placed at the top.
	// If the condition is present and its "true", then the phase should be "Halted".
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseHalted) {
		return apiv1alpha2.DatabasePhaseHalted
	}

	// =================================== Handling "DataRestoreStarted" and "DataRestored" conditions  ==================================================
	// For data restoring, there could be the following scenarios:
	// 1. if condition["DataRestoreStarted"] = true, the phase should be "Restoring".
	//		And there will be no "false" status for "DataRestoreStarted" type.
	// 2. if condition["DataRestored"] = false, the phase should be "NotReady".
	//		if the status is "true", the phase should depend on the rest of checks.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseDataRestoreStarted) {
		// TODO:
		// 		- remove these conditions.
		//		- It is here for backward compatibility.
		//		- Just return "Restoring" in future.
		if cutil.HasCondition(conditions, kubedb.DatabaseDataRestored) {
			if cutil.IsConditionFalse(conditions, kubedb.DatabaseDataRestored) {
				return apiv1alpha2.DatabasePhaseNotReady
			}
		} else {
			return apiv1alpha2.DatabasePhaseDataRestoring
		}
	}
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseDataRestored) {
		return apiv1alpha2.DatabasePhaseNotReady
	}

	// ================================= Handling "AcceptingConnection" condition ==========================
	// If the condition is present and its "false", then the phase should be "NotReady".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseAcceptingConnection) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1alpha2.DatabasePhaseNotReady
	}

	// ================================= Handling "ReplicaReady" condition ==========================
	// If the condition is present and its "false", then the phase should be "Critical".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseReplicaReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1alpha2.DatabasePhaseCritical
	}

	// ================================= Handling "Ready" condition ==========================
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1alpha2.DatabasePhaseCritical
	}
	// Ready, if the database is provisioned and readinessProbe passed.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1alpha2.DatabasePhaseReady
	}

	// ================================= Handling "Provisioned" and "Paused" conditions ==========================
	// These conditions does not have any effect on the database phase. They are only for internal usage.
	// So, we don't have to do anything for them.
	return phase
}

func PhaseFromConditionV1(conditions []kmapi.Condition) apiv1.DatabasePhase {
	// Generally, the conditions should maintain the following chronological order
	// For normal restore process:
	//   ProvisioningStarted --> ReplicaReady --> AcceptingConnection --> DataRestoreStarted --> DataRestored --> Ready --> Provisioned
	// For restoring the volumes (PerconaXtraDB):
	//	 ProvisioningStarted --> DataRestoreStarted --> DataRestored --> ReplicaReady --> AcceptingConnection --> Ready --> Provisioned

	// These are transitional conditions. They can update any time. So, their order may vary:
	// 1. ReplicaReady
	// 2. AcceptingConnection
	// 3. DataRestoreStarted
	// 4. DataRestored
	// 5. Ready
	// 6. Paused
	// 7. HealthCheckPaused

	var phase apiv1.DatabasePhase

	// ================================= Handling "HealthCheckPaused" condition ==========================
	// If the condition is present and its "true", then the phase should be "Unknown".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseHealthCheckPaused) {
		return apiv1.DatabasePhaseUnknown
	}

	// ==================================  Handling "ProvisioningStarted" condition  ========================
	// If the condition is present and its "true", then the phase should be "Provisioning".
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioningStarted) {
		phase = apiv1.DatabasePhaseProvisioning
	}

	// ================================== Handling "Halted" condition =======================================
	// The "Halted" condition has higher priority, that's why it is placed at the top.
	// If the condition is present and its "true", then the phase should be "Halted".
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseHalted) {
		return apiv1.DatabasePhaseHalted
	}

	// =================================== Handling "DataRestoreStarted" and "DataRestored" conditions  ==================================================
	// For data restoring, there could be the following scenarios:
	// 1. if condition["DataRestoreStarted"] = true, the phase should be "Restoring".
	//		And there will be no "false" status for "DataRestoreStarted" type.
	// 2. if condition["DataRestored"] = false, the phase should be "NotReady".
	//		if the status is "true", the phase should depend on the rest of checks.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseDataRestoreStarted) {
		// TODO:
		// 		- remove these conditions.
		//		- It is here for backward compatibility.
		//		- Just return "Restoring" in future.
		if cutil.HasCondition(conditions, kubedb.DatabaseDataRestored) {
			if cutil.IsConditionFalse(conditions, kubedb.DatabaseDataRestored) {
				return apiv1.DatabasePhaseNotReady
			}
		} else {
			return apiv1.DatabasePhaseDataRestoring
		}
	}
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseDataRestored) {
		return apiv1.DatabasePhaseNotReady
	}

	// ================================= Handling "AcceptingConnection" condition ==========================
	// If the condition is present and its "false", then the phase should be "NotReady".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseAcceptingConnection) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1.DatabasePhaseNotReady
	}

	// ================================= Handling "ReplicaReady" condition ==========================
	// If the condition is present and its "false", then the phase should be "Critical".
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseReplicaReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1.DatabasePhaseCritical
	}

	// ================================= Handling "Ready" condition ==========================
	// Skip if the database isn't provisioned yet.
	if cutil.IsConditionFalse(conditions, kubedb.DatabaseReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1.DatabasePhaseCritical
	}
	// Ready, if the database is provisioned and readinessProbe passed.
	if cutil.IsConditionTrue(conditions, kubedb.DatabaseReady) && cutil.IsConditionTrue(conditions, kubedb.DatabaseProvisioned) {
		return apiv1.DatabasePhaseReady
	}

	// ================================= Handling "Provisioned" and "Paused" conditions ==========================
	// These conditions does not have any effect on the database phase. They are only for internal usage.
	// So, we don't have to do anything for them.
	return phase
}

// compareLastTransactionTime compare two condition's "LastTransactionTime" and return an integer based on the followings:
// 1. If both conditions does not exist, then return 0
// 2. If cond1 exist but cond2 does not, then return 1
// 3. If cond1 does not exist but cond2 exist, then return -1
// 3. If cond1.LastTransactionTime > cond2.LastTransactionTime, then return 1
// 4. If cond1.LastTransactionTime = cond2.LastTransactionTime, then return 0
// 5. If cond1.LastTransactionTime < cond2.LastTransactionTime, then return -1
func compareLastTransactionTime(conditions []kmapi.Condition, type1, type2 string) int32 {
	idx1, cond1 := cutil.GetCondition(conditions, type1)
	idx2, cond2 := cutil.GetCondition(conditions, type2)
	// both condition does not exist
	if idx1 == -1 && idx2 == -1 {
		return 0
	}
	// cond1 exist but cond2 does not
	if idx1 != -1 && idx2 == -1 {
		return 1
	}
	// cond2 does not exist but cond2 exist
	if idx1 == -1 && idx2 != -1 {
		return -1
	}

	if cond1.LastTransitionTime.After(cond2.LastTransitionTime.Time) {
		// cond1 is newer than cond2
		return 1
	} else if cond2.LastTransitionTime.After(cond1.LastTransitionTime.Time) {
		// cond1 is older than cond2
		return -1
	}
	return 0
}
