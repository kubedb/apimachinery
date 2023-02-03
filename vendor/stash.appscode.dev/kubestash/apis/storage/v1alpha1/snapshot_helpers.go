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
	"strconv"
	"time"

	"stash.appscode.dev/kubestash/crds"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kmodules.xyz/client-go/meta"
)

func (_ Snapshot) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralSnapshot))
}

func (s *Snapshot) IsCompleted() bool {
	return s.Status.Phase == SnapshotFailed || s.Status.Phase == SnapshotSucceeded
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

	for _, c := range s.Status.Components {
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

	totalComponents := len(s.Status.Components)

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

func GenerateSnapshotName(repoName string) string {
	return meta.ValidNameWithPrefix(repoName, strconv.FormatInt(time.Now().Unix(), 10))
}
