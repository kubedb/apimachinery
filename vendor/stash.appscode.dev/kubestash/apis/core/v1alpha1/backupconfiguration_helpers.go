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
	"stash.appscode.dev/kubestash/crds"

	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
)

func (_ BackupConfiguration) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupConfiguration))
}

func (b *BackupConfiguration) CalculatePhase() BackupInvokerPhase {
	if kmapi.IsConditionFalse(b.Status.Conditions, TypeValidationPassed) {
		return BackupInvokerInvalid
	}

	if b.isReady() {
		return BackupInvokerReady
	}

	return BackupInvokerNotReady
}

func (b *BackupConfiguration) isReady() bool {
	if b.Status.TargetFound == nil || !*b.Status.TargetFound {
		return false
	}

	if !b.backendsReady() {
		return false
	}

	if !b.sessionsReady() {
		return false
	}

	return true
}

func (b *BackupConfiguration) sessionsReady() bool {
	if len(b.Status.Sessions) != len(b.Spec.Sessions) {
		return false
	}

	for _, status := range b.Status.Sessions {
		if !kmapi.IsConditionTrue(status.Conditions, TypeSchedulerEnsured) {
			return false
		}
	}

	return true
}

func (b *BackupConfiguration) backendsReady() bool {
	if len(b.Status.Backends) != len(b.Spec.Backends) {
		return false
	}

	for _, backend := range b.Status.Backends {
		if !backend.Ready {
			return false
		}
	}

	return true
}

func (b *BackupConfiguration) GetStorageRef(backend string) kmapi.TypedObjectReference {
	for _, b := range b.Spec.Backends {
		if b.Name == backend {
			return b.StorageRef
		}
	}
	return kmapi.TypedObjectReference{}
}
