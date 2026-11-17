/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"context"
	"fmt"
	"sort"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	"github.com/Masterminds/semver/v3"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsManagedPersistence tells if the operator creates the database of the default store.
func (r *DurableWorkflow) IsManagedPersistence() bool {
	return r.Spec.Persistence.Managed != nil
}

// ManagedStoreName is the name of the Postgres or MySQL the operator creates in managed mode. The KubeDB
// operator publishes the AppBinding of that database under the same name.
func (r *DurableWorkflow) ManagedStoreName() string {
	return meta_util.NameWithSuffix(r.Name, "db")
}

// ManagedEngine is the engine of the managed store, Postgres if unset.
func (s DurableWorkflowManagedStore) ManagedEngine() DurableWorkflowManagedEngine {
	if s.Engine == "" {
		return DurableWorkflowManagedPostgres
	}
	return s.Engine
}

// ManagedReplicas is the number of database pods, 1 if unset.
func (s DurableWorkflowManagedStore) ManagedReplicas() int32 {
	return ptr.Deref(s.Replicas, 1)
}

// DefaultStore is the default store as the provisioner uses it: in managed mode, the AppBinding of the database the
// operator creates, otherwise spec.persistence.default.
func (r *DurableWorkflow) DefaultStore() DurableWorkflowStore {
	s := r.Spec.Persistence.Default
	if m := r.Spec.Persistence.Managed; m != nil {
		s.ObjectReference = &kmapi.ObjectReference{Namespace: r.Namespace, Name: r.ManagedStoreName()}
		s.Type = DurableWorkflowStoreType(m.ManagedEngine())
	}
	return s
}

// ResolveManagedVersion returns the name of the newest PostgresVersion or MySQLVersion that is official, neither
// deprecated nor end of life, and needs no license.
func ResolveManagedVersion(ctx context.Context, kc client.Client, engine DurableWorkflowManagedEngine) (string, error) {
	type candidate struct{ name, version string }
	var cands []candidate
	switch engine {
	case DurableWorkflowManagedPostgres:
		var list catalog.PostgresVersionList
		if err := kc.List(ctx, &list); err != nil {
			return "", err
		}
		for _, v := range list.Items {
			if v.Spec.Deprecated || v.Spec.EndOfLife || (v.Spec.License != nil && v.Spec.License.Required) {
				continue
			}
			if d := v.Spec.Distribution; d != "" && d != catalog.PostgresDistroOfficial {
				continue
			}
			cands = append(cands, candidate{v.Name, v.Spec.Version})
		}
	case DurableWorkflowManagedMySQL:
		var list catalog.MySQLVersionList
		if err := kc.List(ctx, &list); err != nil {
			return "", err
		}
		for _, v := range list.Items {
			if v.Spec.Deprecated || v.Spec.EndOfLife {
				continue
			}
			if d := v.Spec.Distribution; d != "" && d != catalog.MySQLDistroOfficial {
				continue
			}
			cands = append(cands, candidate{v.Name, v.Spec.Version})
		}
	default:
		return "", fmt.Errorf("engine %q can't be managed", engine)
	}
	if len(cands) == 0 {
		return "", fmt.Errorf("no usable %sVersion found in the catalog", engine)
	}
	sort.SliceStable(cands, func(i, j int) bool {
		a, errA := semver.NewVersion(cands[i].version)
		b, errB := semver.NewVersion(cands[j].version)
		switch {
		case errA != nil || errB != nil:
			return errA != nil && errB == nil // unparsable versions sort first, so they lose
		case a.Equal(b):
			return cands[i].name < cands[j].name
		}
		return a.LessThan(b)
	})
	return cands[len(cands)-1].name, nil
}
