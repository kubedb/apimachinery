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
	"kubestash.dev/apimachinery/crds"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/meta"
)

func (_ Snapshot) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralSnapshot))
}

func (s *Snapshot) CalculatePhase() SnapshotPhase {
	if kmapi.IsConditionFalse(s.Status.Conditions, TypeBackendMetadataWritten) ||
		kmapi.IsConditionFalse(s.Status.Conditions, TypeRecentSnapshotListUpdated) {
		return SnapshotFailed
	}

	return s.GetComponentsPhase()
}

func (s *Snapshot) GetComponentsPhase() SnapshotPhase {
	failedComponent := 0
	successfulComponent := 0
	pendingComponent := 0

	for _, c := range s.Spec.Components {
		if c.Phase == ComponentPhaseSucceeded {
			successfulComponent++
		}
		if c.Phase == ComponentPhaseFailed {
			failedComponent++
		}
		if c.Phase == ComponentPhasePending {
			pendingComponent++
		}
	}

	totalComponents := len(s.Spec.Components)

	if pendingComponent == totalComponents {
		return SnapshotPending
	}

	if successfulComponent == totalComponents {
		return SnapshotSucceeded
	}

	if successfulComponent+failedComponent == totalComponents {
		return SnapshotFailed
	}

	return SnapshotRunning
}

func GenerateSnapshotName(repoName, backupSession string) string {
	return meta.ValidNameWithPrefix(repoName, backupSession)
}
