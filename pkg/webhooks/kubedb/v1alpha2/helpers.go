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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// Defaulting & validation must be skipped once the object is being deleted,
// otherwise updates like finalizer removal get rejected.
func isDeletionInProgress(obj runtime.Object) bool {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return false
	}
	return acc.GetDeletionTimestamp() != nil
}
