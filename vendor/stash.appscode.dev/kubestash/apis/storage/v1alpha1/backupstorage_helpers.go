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

func (_ BackupStorage) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourcePluralBackupStorage))
}

func (b *BackupStorage) CalculatePhase() BackupStoragePhase {
	if kmapi.IsConditionTrue(b.Status.Conditions, TypeBackendInitialized) &&
		kmapi.IsConditionTrue(b.Status.Conditions, TypeRepositorySynced) {
		return BackupStorageReady
	}
	return BackupStorageNotReady
}
