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

// Package appbinding holds the AppBinding logic shared by every kubedb operator. The AppBinding
// is registered in each operator's controller-runtime scheme, so it is written through KBClient
// (clientutil.CreateOrPatch) rather than the typed appcatalog clientset.
//
// The helper fills the envelope common to all DBs (owner ref, labels, filtered annotations, type,
// AppRef, version, and the auth Secret ref). Everything DB-specific — the ClientConfig
// (scheme/port/URL/CABundle/query), the TLSSecret ref, and Spec.Parameters — is layered on by the
// caller's Customize callback, which runs inside the CreateOrPatch mutate function.
package appbinding

import (
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	clientutil "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DB is the minimal contract needed to reconcile an AppBinding. Every kubedb DB type satisfies
// it: AppBindingMeta and OffshootLabels are type-level helpers and the metav1.Object /
// runtime.Object accessors come from client.Object.
type DB interface {
	client.Object
	AppBindingMeta() appcat.AppBindingMeta
	OffshootLabels() map[string]string
}

// Options reconciles the AppBinding for a database through KBClient. The helper fills the common
// envelope and defers everything DB-specific to Customize.
type Options struct {
	KBClient client.Client
	DB       DB
	Owner    *metav1.OwnerReference

	// Kind is the DB resource kind (e.g. ResourceKindRabbitmq / ResourceKindPostgres) used for
	// Spec.AppRef.
	Kind string
	// Version is the resolved catalog version string written to Spec.Version.
	Version string
	// Secret, when set, is written to Spec.Secret (the auth secret ref). Left nil by DBs that
	// have no auth secret (e.g. security disabled).
	Secret *appcat.TypedLocalObjectReference
	// Customize layers the DB-specific ClientConfig / TLSSecret / Parameters onto the envelope the
	// helper has already filled. It runs inside the CreateOrPatch mutate function.
	Customize func(in *appcat.AppBinding)
	// Recorder, when set, emits a Normal event on the DB after a create/patch.
	Recorder record.EventRecorder
}

// Ensure creates or patches the AppBinding, filling the common envelope and calling Customize for
// the DB-specific remainder.
func (o Options) Ensure(ctx context.Context) (kutil.VerbType, error) {
	appmeta := o.DB.AppBindingMeta()
	ab := &appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appmeta.Name(),
			Namespace: o.DB.GetNamespace(),
		},
	}

	vt, err := clientutil.CreateOrPatch(ctx, o.KBClient, ab, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*appcat.AppBinding)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.Owner)
		in.Labels = o.DB.OffshootLabels()
		in.Annotations = meta_util.FilterKeys(kubedb.GroupName, nil, o.DB.GetAnnotations())

		in.Spec.Type = appmeta.Type()
		in.Spec.AppRef = &kmapi.TypedObjectReference{
			APIGroup:  kubedb.GroupName,
			Kind:      o.Kind,
			Namespace: o.DB.GetNamespace(),
			Name:      o.DB.GetName(),
		}
		in.Spec.Version = o.Version
		if o.Secret != nil {
			in.Spec.Secret = o.Secret
		}

		if o.Customize != nil {
			o.Customize(in)
		}
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	if vt != kutil.VerbUnchanged && o.Recorder != nil {
		o.Recorder.Eventf(o.DB, core.EventTypeNormal, "Successful", "Successfully %s appbinding", vt)
	}
	return vt, nil
}
