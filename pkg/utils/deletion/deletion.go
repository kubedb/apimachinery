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

// Package deletion holds the DB deletion logic shared by every kubedb operator:
// the DeletionPolicy owner-reference sync (Halt/Delete/WipeOut) and the spec.Halted
// Halt mode. Callers pass the DB object and its list type; everything else is derived.
package deletion

import (
	"context"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeletionPolicy values. Kept as untyped string consts so this package stays decoupled
// from the v1/v1alpha2 DeletionPolicy types (both are `type DeletionPolicy string`).
const (
	DeletionPolicyHalt    = "Halt"
	DeletionPolicyDelete  = "Delete"
	DeletionPolicyWipeOut = "WipeOut"
)

var (
	secretGVR = core.SchemeGroupVersion.WithResource("secrets")
	pvcGVR    = core.SchemeGroupVersion.WithResource("persistentvolumeclaims")
)

// DBInterface is the minimal contract every kubedb DB type satisfies. All accessors
// are type-level on every DB (see apis/kubedb/*/*_helpers.go).
type DBInterface interface {
	client.Object
	OffshootSelectors(extraSelectors ...map[string]string) map[string]string
	GetPersistentSecrets() []string
	GetDeletionPolicy() string
}

// ExtraSecrets lets a DB apply the same owner-reference handling to secrets living in a
// non-core resource (e.g. virtual-secrets). Optional; nil for most DBs.
type ExtraSecrets struct {
	GVR   schema.GroupVersionResource
	Names []string
}

// Options carries everything Do needs. The DB supplies only DB + PeerList (+ optional
// ExtraSecrets); namespace, name, selectors, secrets, policy and owner are all derived.
type Options struct {
	KBClient      client.Client
	DynamicClient dynamic.Interface
	DB            DBInterface
	// PeerList is an empty typed list of the same kind (e.g. &api.MongoDBList{}); used to
	// find which secrets are still referenced by sibling DBs before wiping them.
	PeerList     client.ObjectList
	ExtraSecrets []ExtraSecrets
}

// Do runs the DeletionPolicy owner-reference sync. Call it from the operator's terminate path.
//
//	Halt    -> keep PVCs and secrets (remove owner reference).
//	Delete  -> delete PVCs (add owner reference), keep secrets.
//	WipeOut -> delete PVCs and unused kubedb-owned secrets.
func Do(ctx context.Context, opts Options) error {
	if opts.DB.GetDeletionPolicy() == DeletionPolicyHalt {
		return removeOwnerReferenceFromOffshoots(ctx, opts)
	}
	return setOwnerReferenceToOffshoots(ctx, opts)
}

func removeOwnerReferenceFromOffshoots(ctx context.Context, opts Options) error {
	ns := opts.DB.GetNamespace()
	selector := labels.SelectorFromSet(opts.DB.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(ctx, opts.DynamicClient, pvcGVR, ns, selector, opts.DB); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, secretGVR, ns, opts.DB.GetPersistentSecrets(), opts.DB); err != nil {
		return err
	}
	for _, es := range opts.ExtraSecrets {
		if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, es.GVR, ns, es.Names, opts.DB); err != nil {
			return err
		}
	}
	return nil
}

func setOwnerReferenceToOffshoots(ctx context.Context, opts Options) error {
	owner, err := ownerRef(opts)
	if err != nil {
		return err
	}
	ns := opts.DB.GetNamespace()
	selector := labels.SelectorFromSet(opts.DB.OffshootSelectors())

	if opts.DB.GetDeletionPolicy() == DeletionPolicyWipeOut {
		if err := wipeOut(ctx, opts, owner); err != nil {
			return errors.Wrap(err, "error in wiping out database")
		}
	} else {
		// Delete: keep secrets intact by removing their owner reference.
		if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, secretGVR, ns, opts.DB.GetPersistentSecrets(), opts.DB); err != nil {
			return err
		}
		for _, es := range opts.ExtraSecrets {
			if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, es.GVR, ns, es.Names, opts.DB); err != nil {
				return err
			}
		}
	}
	// Delete PVCs for both WipeOut and Delete by making the DB their owner.
	return dynamic_util.EnsureOwnerReferenceForSelector(ctx, opts.DynamicClient, pvcGVR, ns, selector, owner)
}

// wipeOut makes the DB the owner of every persistent secret that is not shared with a peer
// DB and is managed by kubedb, so garbage collection removes them. ExtraSecrets are always
// wiped (they belong solely to this DB).
func wipeOut(ctx context.Context, opts Options, owner *metav1.OwnerReference) error {
	used, err := secretsUsedByPeers(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "error in getting used secret list")
	}
	ns := opts.DB.GetNamespace()
	unused := sets.New[string](opts.DB.GetPersistentSecrets()...).Difference(used)

	// Don't wipe secrets that are missing or not managed by kubedb.
	for _, name := range sets.List[string](unused) {
		secret := &core.Secret{}
		err := opts.KBClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, secret)
		if kerr.IsNotFound(err) {
			unused.Delete(name)
			continue
		}
		if err != nil {
			return errors.Wrap(err, "error in getting db secret")
		}
		if secret.Labels[meta_util.ManagedByLabelKey] != kubedb.GroupName {
			unused.Delete(name)
		}
	}

	if err := dynamic_util.EnsureOwnerReferenceForItems(ctx, opts.DynamicClient, secretGVR, ns, sets.List[string](unused), owner); err != nil {
		return err
	}
	for _, es := range opts.ExtraSecrets {
		if err := dynamic_util.EnsureOwnerReferenceForItems(ctx, opts.DynamicClient, es.GVR, ns, es.Names, owner); err != nil {
			return err
		}
	}
	return nil
}

// secretsUsedByPeers returns the set of secrets referenced by other DBs of the same kind
// in this namespace.
func secretsUsedByPeers(ctx context.Context, opts Options) (sets.Set[string], error) {
	used := sets.New[string]()
	if opts.PeerList == nil {
		return used, nil
	}
	if err := opts.KBClient.List(ctx, opts.PeerList, client.InNamespace(opts.DB.GetNamespace())); err != nil {
		return nil, err
	}
	items, err := meta.ExtractList(opts.PeerList)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		peer, ok := item.(DBInterface)
		if !ok {
			continue
		}
		if peer.GetName() == opts.DB.GetName() {
			continue
		}
		used.Insert(peer.GetPersistentSecrets()...)
	}
	return used, nil
}

// ownerRef builds a controller owner reference for the DB using its registered GVK.
func ownerRef(opts Options) (*metav1.OwnerReference, error) {
	gvks, _, err := opts.KBClient.Scheme().ObjectKinds(opts.DB)
	if err != nil {
		return nil, err
	}
	if len(gvks) == 0 {
		return nil, fmt.Errorf("no registered GVK for %T", opts.DB)
	}
	return metav1.NewControllerRef(opts.DB, gvks[0]), nil
}
