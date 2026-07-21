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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Owner is retained as a thin alias for AsOwner. The accessor was renamed in
// apimachinery, but the released database operators still call Owner(), so
// dropping it outright breaks every consumer that vendors them at a pinned
// version. Delete this file once those operators have migrated to AsOwner.

// Deprecated: use AsOwner instead.
func (e *ElasticsearchDashboard) Owner() *metav1.OwnerReference {
	return e.AsOwner()
}
