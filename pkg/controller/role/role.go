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

// Package role holds the per-database RBAC logic shared by every kubedb operator: the
// ServiceAccount -> Role -> RoleBinding trio a database's pods run under. Only the PolicyRule set
// differs per DB, so callers supply Rules; everything else is reconciled here through KBClient.
package role

import (
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientutil "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DB is the minimal contract needed to reconcile RBAC. Every kubedb DB type satisfies it:
// OffshootLabels is a type-level helper and the metav1.Object accessors come from client.Object.
type DB interface {
	client.Object
	OffshootLabels() map[string]string
}

// Options reconciles the ServiceAccount + Role + RoleBinding for a database through KBClient.
// Names, rules and owner are supplied by the caller because they differ across API versions
// (e.g. rabbitmq uses DefaultPodRoleName/DefaultPodRoleBindingName, postgres uses OffshootName).
type Options struct {
	KBClient client.Client
	DB       DB
	Owner    *metav1.OwnerReference

	// ServiceAccountName is the SA the pods run as. When ManageServiceAccount is true it is
	// created/patched by this package; otherwise it must already exist (it is only verified).
	ServiceAccountName string
	// ManageServiceAccount reports whether this package owns the ServiceAccount's lifecycle.
	// Set it to false when the user supplied spec.PodTemplate.Spec.ServiceAccountName.
	ManageServiceAccount bool
	// SkipIfUnmanaged, when true, aborts the whole reconcile (SA, Role, RoleBinding) if the
	// ServiceAccount already exists and is not labelled managed-by kubedb — i.e. a user brought
	// their own SA under the default name. Only consulted when ManageServiceAccount is true.
	SkipIfUnmanaged bool

	RoleName        string
	RoleBindingName string
	Rules           []rbac.PolicyRule
}

// Ensure reconciles the ServiceAccount, Role and RoleBinding. See the field docs on Options for
// how the ServiceAccount is handled. It is idempotent and safe to call every reconcile.
func (o Options) Ensure(ctx context.Context) error {
	proceed, err := o.ensureServiceAccount(ctx)
	if err != nil {
		return err
	}
	if !proceed {
		// A user-managed ServiceAccount sits under our default name; leave RBAC untouched.
		return nil
	}
	if err := o.ensureRole(ctx); err != nil {
		return err
	}
	return o.ensureRoleBinding(ctx)
}

// ensureServiceAccount reconciles the ServiceAccount and reports whether the caller should go on
// to reconcile the Role and RoleBinding.
func (o Options) ensureServiceAccount(ctx context.Context) (bool, error) {
	key := types.NamespacedName{Namespace: o.DB.GetNamespace(), Name: o.ServiceAccountName}

	if !o.ManageServiceAccount {
		// User-provided ServiceAccount: verify it exists but never mutate it.
		sa := &core.ServiceAccount{}
		if err := o.KBClient.Get(ctx, key, sa); err != nil {
			return false, err
		}
		return true, nil
	}

	if o.SkipIfUnmanaged {
		sa := &core.ServiceAccount{}
		err := o.KBClient.Get(ctx, key, sa)
		if err == nil {
			if sa.Labels[meta_util.ManagedByLabelKey] != kubedb.GroupName {
				// The SA exists under our default name but the user owns it: do nothing.
				return false, nil
			}
		} else if !kerr.IsNotFound(err) {
			return false, err
		}
	}

	sa := &core.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Namespace: o.DB.GetNamespace(), Name: o.ServiceAccountName},
	}
	_, err := clientutil.CreateOrPatch(ctx, o.KBClient, sa, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.ServiceAccount)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.Owner)
		in.Labels = o.DB.OffshootLabels()
		return in
	})
	return true, err
}

func (o Options) ensureRole(ctx context.Context) error {
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{Namespace: o.DB.GetNamespace(), Name: o.RoleName},
	}
	_, err := clientutil.CreateOrPatch(ctx, o.KBClient, role, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*rbac.Role)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.Owner)
		in.Labels = o.DB.OffshootLabels()
		in.Rules = o.Rules
		return in
	})
	return err
}

func (o Options) ensureRoleBinding(ctx context.Context) error {
	rolebinding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Namespace: o.DB.GetNamespace(), Name: o.RoleBindingName},
	}
	_, err := clientutil.CreateOrPatch(ctx, o.KBClient, rolebinding, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*rbac.RoleBinding)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.Owner)
		in.Labels = o.DB.OffshootLabels()
		in.RoleRef = rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "Role",
			Name:     o.RoleName,
		}
		in.Subjects = []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Name:      o.ServiceAccountName,
				Namespace: o.DB.GetNamespace(),
			},
		}
		return in
	})
	return err
}
