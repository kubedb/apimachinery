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
	"encoding/json"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

// Branched-mode helpers, shared by every KubeDB Database provisioner.
//
// A Database created by the KubeDB Courier Branch operator carries the
// BranchedFromAnnotation. Its provisioner then enters "branched mode": instead of
// provisioning empty storage, it waits for the Courier-cloned data PVCs and the
// auth/config secrets, then adopts them. These helpers are engine-agnostic so
// postgres, mysql, mongodb, ... reuse one implementation; each engine keeps only
// the reconcile-gate placement and its own volume-name/replica inputs.

// BranchedFrom is the parsed value of BranchedFromAnnotation. It records the
// provenance of a branch: the source cluster and the source Database (as
// "namespace/name").
type BranchedFrom struct {
	// Cluster is the source cluster name (empty for a same-cluster/Local branch).
	Cluster string `json:"cluster,omitempty"`
	// Source is the source Database as "namespace/name".
	Source string `json:"source,omitempty"`
}

// IsBranched reports whether obj is a KubeDB Courier branch, i.e. it carries the
// BranchedFromAnnotation. Nil-safe.
func IsBranched(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	_, ok := obj.GetAnnotations()[kubedb.BranchedFromAnnotation]
	return ok
}

// ParseBranchedFrom parses the BranchedFromAnnotation value on obj. ok is false
// when the annotation is absent; err is non-nil only when the value is present
// but not valid JSON.
func ParseBranchedFrom(obj metav1.Object) (bf BranchedFrom, ok bool, err error) {
	if obj == nil {
		return BranchedFrom{}, false, nil
	}
	v, ok := obj.GetAnnotations()[kubedb.BranchedFromAnnotation]
	if !ok {
		return BranchedFrom{}, false, nil
	}
	if err := json.Unmarshal([]byte(v), &bf); err != nil {
		return BranchedFrom{}, true, fmt.Errorf("parse %s: %w", kubedb.BranchedFromAnnotation, err)
	}
	return bf, true, nil
}

// BranchedPrerequisitesReady reports whether the resources the Courier Branch
// operator delivers — the auth/config secrets and the cloned data PVCs — already
// exist. Until they do, the provisioner must not proceed: it would otherwise
// generate a fresh auth secret (breaking the cloned credentials) or let the
// PetSet create empty PVCs. Callers should wait/requeue while this returns false.
//
// secretNames and pvcNames are engine-supplied. PVC naming differs per engine, so
// each Database exposes its own GetBranchedDataPVCNames() method (e.g.
// Postgres.GetBranchedDataPVCNames()) — the caller passes its result here.
func BranchedPrerequisitesReady(dc dynamic.Interface, namespace string, secretNames, pvcNames []string) (bool, error) {
	// Secrets first: the clone must boot with the source's credentials rather
	// than a freshly generated password.
	ok, err := dynamic_util.ResourcesExists(
		dc,
		core.SchemeGroupVersion.WithResource("secrets"),
		namespace,
		secretNames...,
	)
	if err != nil || !ok {
		return false, err
	}

	// The cloned data PVCs must exist so the PetSet adopts them by name.
	return dynamic_util.ResourcesExists(
		dc,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		namespace,
		pvcNames...,
	)
}
