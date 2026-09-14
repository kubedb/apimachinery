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

package validator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"kubedb.dev/apimachinery/apis/kubedb"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The same database is served under multiple versions; querying the highest one is enough.
var kubedbVersionPriority = map[string]int{
	"v1":       3,
	"v1alpha2": 2,
	"v1alpha1": 1,
}

var clientObjectType = reflect.TypeOf((*client.Object)(nil)).Elem()

func ValidateNameUniqueness(ctx context.Context, kc client.Client, obj runtime.Object) error {
	db, ok := obj.(client.Object)
	if !ok {
		return fmt.Errorf("expected a client.Object but got a %T", obj)
	}
	gvk, err := kc.GroupVersionKindFor(db)
	if err != nil {
		return err
	}

	for _, candidate := range otherKubeDBKinds(kc.Scheme(), gvk.Kind) {
		existing := unstructured.Unstructured{}
		existing.SetGroupVersionKind(candidate)
		err := kc.Get(ctx, client.ObjectKey{Namespace: db.GetNamespace(), Name: db.GetName()}, &existing)
		if err != nil {
			if kerr.IsNotFound(err) || meta.IsNoMatchError(err) {
				continue
			}
			return err
		}
		return fmt.Errorf("cannot create %s %s/%s, a %s already exists with the same name in this namespace",
			gvk.Kind, db.GetNamespace(), db.GetName(), candidate.Kind)
	}
	return nil
}

func otherKubeDBKinds(scheme *runtime.Scheme, skipKind string) []schema.GroupVersionKind {
	preferred := make(map[string]schema.GroupVersionKind)
	for gvk, typ := range scheme.AllKnownTypes() {
		if gvk.Group != kubedb.GroupName || gvk.Kind == skipKind {
			continue
		}
		// filters out the list kinds & the option kinds registered by metav1.AddToGroupVersion
		if !reflect.PointerTo(typ).Implements(clientObjectType) {
			continue
		}
		if cur, found := preferred[gvk.Kind]; found &&
			kubedbVersionPriority[cur.Version] >= kubedbVersionPriority[gvk.Version] {
			continue
		}
		preferred[gvk.Kind] = gvk
	}

	kinds := make([]schema.GroupVersionKind, 0, len(preferred))
	for _, gvk := range preferred {
		kinds = append(kinds, gvk)
	}
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].Kind < kinds[j].Kind
	})
	return kinds
}
