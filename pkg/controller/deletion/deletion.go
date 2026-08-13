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

package deletion

import (
	"context"
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"

	"github.com/pkg/errors"
	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Untyped string consts so this package stays decoupled from the v1/v1alpha2 DeletionPolicy types.
const (
	DeletionPolicyHalt           = "Halt"
	DeletionPolicyDelete         = "Delete"
	DeletionPolicyWipeOut        = "WipeOut"
	DeletionPolicyDoNotTerminate = "DoNotTerminate"
)

var (
	secretGVR        = core.SchemeGroupVersion.WithResource("secrets")
	pvcGVR           = core.SchemeGroupVersion.WithResource("persistentvolumeclaims")
	virtualSecretGVR = vsecretapi.SchemeGroupVersion.WithResource(vsecretapi.ResourceSecrets)
)

// OffshootSelectors is intentionally not part of this interface: its signature differs across
// API versions (v1 non-variadic, v1alpha2 variadic), so callers pass the map via Options.Selectors.
type DBInterface interface {
	client.Object
	GetPersistentSecrets() []string
	GetDeletionPolicy() string
}

type Options struct {
	KBClient      client.Client
	DynamicClient dynamic.Interface
	DB            DBInterface
	// Selectors is passed in because OffshootSelectors has an incompatible signature across API versions.
	Selectors map[string]string
	// PeerList is an empty typed list of the same kind (e.g. &api.MongoDBList{}); used to find
	// which secrets are still referenced by sibling DBs before wiping them.
	PeerList client.ObjectList
}

type authSecretReferrer interface {
	GetAuthSecretName() string
}

// Safe to run unconditionally: when the auth secret is an ordinary core secret (or the DB has
// none), the virtual-secrets object won't exist and the owner-reference helpers skip it.
func virtualAuthSecretNames(db DBInterface) []string {
	if r, ok := db.(authSecretReferrer); ok {
		if name := r.GetAuthSecretName(); name != "" {
			return []string{name}
		}
	}
	return nil
}

// Do runs the DeletionPolicy owner reference sync from the operator's terminate path.
//
//	Halt           -> keep PVCs and secrets (remove owner reference).
//	DoNotTerminate -> treated like Halt as a safety net: the webhook normally blocks deletion,
//	                  but if it's bypassed or down, this still keeps PVCs/secrets.
//	Delete         -> delete PVCs (add owner reference), keep secrets.
//	WipeOut        -> delete PVCs and unused kubedb-owned secrets.
func Do(ctx context.Context, opts Options) error {
	policy := opts.DB.GetDeletionPolicy()
	if policy == DeletionPolicyHalt || policy == DeletionPolicyDoNotTerminate {
		return removeOwnerRefsFromOffshoots(ctx, opts)
	}
	return ensureOwnerRefsOnOffshoots(ctx, opts)
}

func removeOwnerRefsFromOffshoots(ctx context.Context, opts Options) error {
	ns := opts.DB.GetNamespace()
	selector := labels.SelectorFromSet(opts.Selectors)

	if err := dynamic_util.RemoveOwnerReferenceForSelector(ctx, opts.DynamicClient, pvcGVR, ns, selector, opts.DB); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, secretGVR, ns, opts.DB.GetPersistentSecrets(), opts.DB); err != nil {
		return err
	}
	return dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, virtualSecretGVR, ns, virtualAuthSecretNames(opts.DB), opts.DB)
}

func ensureOwnerRefsOnOffshoots(ctx context.Context, opts Options) error {
	owner, err := buildOwnerRef(opts)
	if err != nil {
		return err
	}
	ns := opts.DB.GetNamespace()
	selector := labels.SelectorFromSet(opts.Selectors)

	if opts.DB.GetDeletionPolicy() == DeletionPolicyWipeOut {
		if err := wipeOut(ctx, opts, owner); err != nil {
			return errors.Wrap(err, "error in wiping out database")
		}
	} else {
		// Delete: keep secrets intact by removing their owner reference.
		if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, secretGVR, ns, opts.DB.GetPersistentSecrets(), opts.DB); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(ctx, opts.DynamicClient, virtualSecretGVR, ns, virtualAuthSecretNames(opts.DB), opts.DB); err != nil {
			return err
		}
	}
	// Delete PVCs for both WipeOut and Delete by making the DB their owner.
	return dynamic_util.EnsureOwnerReferenceForSelector(ctx, opts.DynamicClient, pvcGVR, ns, selector, owner)
}

// wipeOut makes the DB the owner of every persistent secret not shared with a peer DB and
// managed by kubedb, so garbage collection removes them.
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
	// The auth secret may be a virtual-secret; make the DB its owner too. Harmless when the
	// auth secret is an ordinary core secret (the virtual-secrets object won't exist).
	return dynamic_util.EnsureOwnerReferenceForItems(ctx, opts.DynamicClient, virtualSecretGVR, ns, virtualAuthSecretNames(opts.DB), owner)
}

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

func buildOwnerRef(opts Options) (*metav1.OwnerReference, error) {
	gvks, _, err := opts.KBClient.Scheme().ObjectKinds(opts.DB)
	if err != nil {
		return nil, err
	}
	if len(gvks) == 0 {
		return nil, fmt.Errorf("no registered GVK for %T", opts.DB)
	}
	return metav1.NewControllerRef(opts.DB, gvks[0]), nil
}
