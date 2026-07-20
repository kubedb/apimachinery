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

package appbinding

import (
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	kmapi "kmodules.xyz/client-go/api/v1"
	clientutil "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DB interface {
	client.Object
	AppBindingMeta() appcat.AppBindingMeta
	OffshootLabels() map[string]string
	AsOwner() *metav1.OwnerReference
}

type Options struct {
	KBClient client.Client
	DB       DB
	Version  string
	// Customize runs inside the CreateOrPatch mutate function to layer on everything DB-specific:
	// ClientConfig, Secret, TLSSecret, Parameters.
	Customize func(in *appcat.AppBinding)
}

func (o Options) Ensure(ctx context.Context) (kutil.VerbType, error) {
	gvks, _, err := o.KBClient.Scheme().ObjectKinds(o.DB)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	appmeta := o.DB.AppBindingMeta()
	ab := &appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appmeta.Name(),
			Namespace: o.DB.GetNamespace(),
		},
	}

	return clientutil.CreateOrPatch(ctx, o.KBClient, ab, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*appcat.AppBinding)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.DB.AsOwner())
		in.Labels = o.DB.OffshootLabels()
		in.Annotations = meta_util.FilterKeys(kubedb.GroupName, nil, o.DB.GetAnnotations())

		in.Spec.Type = appmeta.Type()
		in.Spec.AppRef = &kmapi.TypedObjectReference{
			APIGroup:  kubedb.GroupName,
			Kind:      gvks[0].Kind,
			Namespace: o.DB.GetNamespace(),
			Name:      o.DB.GetName(),
		}
		in.Spec.Version = o.Version

		if o.Customize != nil {
			o.Customize(in)
		}
		return in
	})
}
