/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	"time"

	"stash.appscode.dev/kubestash/apis/storage/v1alpha1"
	"stash.appscode.dev/kubestash/crds"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/meta"
)

func (_ BackupSession) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupSession))
}

func (b *BackupSession) CalculatePhase() BackupSessionPhase {
	if kmapi.IsConditionTrue(b.Status.Conditions, TypeBackupSkipped) {
		return BackupSessionSkipped
	}

	if b.failedToEnsurebackupExecutor() ||
		b.failedToEnsureSnapshots() ||
		b.failedToExecuteHooks() ||
		b.failedToApplyRetentionPolicy() ||
		b.verificationsFailed() ||
		b.sessionHistoryCleanupFailed() {
		return BackupSessionFailed
	}

	componentsPhase := b.calculateBackupSessionPhaseFromSnapshots()
	if componentsPhase == BackupSessionPending || b.FinalStepExecuted() {
		return componentsPhase
	}

	return BackupSessionRunning
}

func (b *BackupSession) sessionHistoryCleanupFailed() bool {
	return kmapi.IsConditionFalse(b.Status.Conditions, TypeSessionHistoryCleaned)
}

func (b *BackupSession) failedToEnsureSnapshots() bool {
	return !kmapi.HasCondition(b.Status.Conditions, TypeSnapshotsEnsured) ||
		kmapi.IsConditionFalse(b.Status.Conditions, TypeSnapshotsEnsured)
}

func (b *BackupSession) failedToEnsurebackupExecutor() bool {
	return !kmapi.HasCondition(b.Status.Conditions, TypeBackupExecutorEnsured) ||
		kmapi.IsConditionFalse(b.Status.Conditions, TypeBackupExecutorEnsured)
}

func (b *BackupSession) FinalStepExecuted() bool {
	return kmapi.HasCondition(b.Status.Conditions, TypeSessionHistoryCleaned)
}

func (b *BackupSession) failedToApplyRetentionPolicy() bool {
	for _, status := range b.Status.RetentionPolicies {
		if status.Phase == RetentionPolicyFailedToApply {
			return true
		}
	}

	return false
}

func (b *BackupSession) failedToExecuteHooks() bool {
	for _, h := range b.Status.Hooks {
		if h.Phase == HookExecutionFailed {
			return true
		}
	}

	return false
}

func (b *BackupSession) verificationsFailed() bool {
	for _, v := range b.Status.Verifications {
		if v.Phase == VerificationFailed {
			return true
		}
	}

	return false
}

func (b *BackupSession) calculateBackupSessionPhaseFromSnapshots() BackupSessionPhase {
	status := b.Status.Snapshots
	if len(status) == 0 {
		return BackupSessionPending
	}

	pending := 0
	failed := 0
	succeeded := 0

	for _, s := range status {
		if s.Phase == v1alpha1.SnapshotFailed {
			failed++
		}
		if s.Phase == v1alpha1.SnapshotPending {
			pending++
		}
		if s.Phase == v1alpha1.SnapshotSucceeded {
			succeeded++
		}
	}

	if pending == len(status) {
		return BackupSessionPending
	}

	if succeeded+failed != len(status) {
		return BackupSessionRunning
	}

	if failed > 0 {
		return BackupSessionFailed
	}

	return BackupSessionSucceeded
}

func GenerateBackupSessionName(invokerName, sessionName string) string {
	return meta.ValidNameWithPrefixNSuffix(invokerName, sessionName, fmt.Sprintf("%d", time.Now().Unix()))
}
