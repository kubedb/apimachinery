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

package secret

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	dbsecret "kubedb.dev/apimachinery/pkg/secret"

	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	cu "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DB interface {
	client.Object
	OffshootLabels() map[string]string
	AsOwner() *metav1.OwnerReference
	ResourceFQN() string
	GetAuthSecretName() string
}

type Options struct {
	KBClient client.Client
	DB       DB
	// DefaultUsername is the username stamped into a kubedb-generated auth secret.
	DefaultUsername string
	// EnforceUsername gates the username policy for user-supplied (externally
	// managed) secrets: when true, a secret whose username differs from
	// DefaultUsername is rejected; when false, only the presence of the
	// username/password keys is checked.
	EnforceUsername bool
}

func (o Options) EnsureAuthSecret(ctx context.Context) error {
	ref := o.readAuthSecretRef()
	name := o.DB.GetAuthSecretName()
	isVirtual := ref.apiGroup == vsecretapi.GroupName

	if ref.externallyManaged {
		return o.ensureExternalAuthSecret(ctx, name, isVirtual)
	}
	return o.ensureInternalAuthSecret(ctx, name, isVirtual, ref.secretStoreName)
}

func (o Options) ensureExternalAuthSecret(ctx context.Context, name string, isVirtual bool) error {
	data, annotations, err := dbsecret.Get(ctx, o.KBClient, o.DB.GetNamespace(), name, isVirtual)
	if err != nil {
		return err
	}
	if err := o.validateAuthData(data); err != nil {
		return err
	}
	activeFrom, err := activationTime(annotations)
	if err != nil {
		return err
	}
	return o.patchAuthSecretRef(ctx, name, activeFrom)
}

func (o Options) ensureInternalAuthSecret(ctx context.Context, name string, isVirtual bool, storeName string) error {
	data, _, err := dbsecret.Get(ctx, o.KBClient, o.DB.GetNamespace(), name, isVirtual)
	switch {
	case kerr.IsNotFound(err):
		if err := o.createAuthSecret(ctx, name, isVirtual, storeName); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if err := o.validateAuthData(data); err != nil {
			return err
		}
		if err := o.syncOwnedLabels(ctx, name, isVirtual); err != nil {
			return err
		}
	}

	activeFrom, err := o.manageActiveFrom(ctx, name, isVirtual)
	if err != nil {
		return err
	}
	return o.patchAuthSecretRef(ctx, name, activeFrom)
}

func (o Options) createAuthSecret(ctx context.Context, name string, isVirtual bool, storeName string) error {
	ns := o.DB.GetNamespace()
	if isVirtual {
		obj := &vsecretapi.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
		_, err := cu.CreateOrPatch(ctx, o.KBClient, obj, func(in client.Object, _ bool) client.Object {
			s := in.(*vsecretapi.Secret)
			s.Labels = meta_util.OverwriteKeys(s.Labels, o.DB.OffshootLabels())
			core_util.EnsureOwnerReference(&s.ObjectMeta, o.DB.AsOwner())
			s.SecretStoreName = storeName
			s.Type = core.SecretTypeBasicAuth
			if len(s.Data) == 0 {
				s.Data = o.generatedData()
			}
			return s
		})
		return err
	}
	obj := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
	_, err := cu.CreateOrPatch(ctx, o.KBClient, obj, func(in client.Object, _ bool) client.Object {
		s := in.(*core.Secret)
		s.Labels = meta_util.OverwriteKeys(s.Labels, o.DB.OffshootLabels())
		core_util.EnsureOwnerReference(&s.ObjectMeta, o.DB.AsOwner())
		s.Type = core.SecretTypeBasicAuth
		if len(s.Data) == 0 {
			s.Data = o.generatedData()
		}
		return s
	})
	return err
}

func (o Options) syncOwnedLabels(ctx context.Context, name string, isVirtual bool) error {
	ns := o.DB.GetNamespace()
	if isVirtual {
		obj := &vsecretapi.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
		_, err := cu.CreateOrPatchE(ctx, o.KBClient, obj, func(in client.Object, _ bool) (client.Object, error) {
			s := in.(*vsecretapi.Secret)
			if err := o.validateOwnerLabels(s.Labels); err != nil {
				return nil, err
			}
			s.Labels = meta_util.OverwriteKeys(s.Labels, o.DB.OffshootLabels())
			return s, nil
		})
		return err
	}
	obj := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}}
	_, err := cu.CreateOrPatchE(ctx, o.KBClient, obj, func(in client.Object, _ bool) (client.Object, error) {
		s := in.(*core.Secret)
		if err := o.validateOwnerLabels(s.Labels); err != nil {
			return nil, err
		}
		s.Labels = meta_util.OverwriteKeys(s.Labels, o.DB.OffshootLabels())
		return s, nil
	})
	return err
}

func (o Options) validateOwnerLabels(existing map[string]string) error {
	if v, ok := existing[meta_util.NameLabelKey]; ok && v != o.DB.ResourceFQN() {
		return fmt.Errorf("auth secret is owned by a different resource kind %q, expected %q", v, o.DB.ResourceFQN())
	}
	if v, ok := existing[meta_util.InstanceLabelKey]; ok && v != o.DB.GetName() {
		return fmt.Errorf("auth secret is owned by a different instance %q, expected %q", v, o.DB.GetName())
	}
	return nil
}

func (o Options) manageActiveFrom(ctx context.Context, name string, isVirtual bool) (*metav1.Time, error) {
	_, annotations, err := dbsecret.Get(ctx, o.KBClient, o.DB.GetNamespace(), name, isVirtual)
	if err != nil {
		return nil, err
	}
	if t, err := activationTime(annotations); err != nil {
		return nil, err
	} else if t != nil {
		return t, nil
	}

	now := metav1.Now()
	err = dbsecret.CreateOrPatch(ctx, o.KBClient, o.DB.GetNamespace(), name, isVirtual,
		func(ann map[string]string, data map[string][]byte) (map[string]string, map[string][]byte) {
			if ann == nil {
				ann = map[string]string{}
			}
			ann[kubedb.AuthActiveFromAnnotation] = now.Format(time.RFC3339)
			return ann, data
		})
	if err != nil {
		return nil, err
	}
	return &now, nil
}

func (o Options) validateAuthData(data map[string][]byte) error {
	if len(data[core.BasicAuthUsernameKey]) == 0 || len(data[core.BasicAuthPasswordKey]) == 0 {
		return fmt.Errorf("auth secret must contain non-empty %q and %q keys", core.BasicAuthUsernameKey, core.BasicAuthPasswordKey)
	}
	if o.EnforceUsername && string(data[core.BasicAuthUsernameKey]) != o.DefaultUsername {
		return fmt.Errorf("auth secret username must be %q", o.DefaultUsername)
	}
	return nil
}

func (o Options) generatedData() map[string][]byte {
	return map[string][]byte{
		core.BasicAuthUsernameKey: []byte(o.DefaultUsername),
		core.BasicAuthPasswordKey: []byte(rand.String(kubedb.DefaultPasswordLength)),
	}
}

func activationTime(annotations map[string]string) (*metav1.Time, error) {
	val, ok := annotations[kubedb.AuthActiveFromAnnotation]
	if !ok {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil, err
	}
	return &metav1.Time{Time: t}, nil
}

type secretRefView struct {
	apiGroup          string
	secretStoreName   string
	externallyManaged bool
}

func (o Options) readAuthSecretRef() secretRefView {
	var v secretRefView
	f := authSecretField(reflect.ValueOf(o.DB))
	if !f.IsValid() || f.IsNil() {
		return v
	}
	ref := f.Elem()
	if x := ref.FieldByName("APIGroup"); x.IsValid() {
		v.apiGroup = x.String()
	}
	if x := ref.FieldByName("SecretStoreName"); x.IsValid() {
		v.secretStoreName = x.String()
	}
	if x := ref.FieldByName("ExternallyManaged"); x.IsValid() {
		v.externallyManaged = x.Bool()
	}
	return v
}

func (o Options) patchAuthSecretRef(ctx context.Context, name string, activeFrom *metav1.Time) error {
	_, err := cu.CreateOrPatch(ctx, o.KBClient, o.DB, func(in client.Object, _ bool) client.Object {
		setAuthSecretRef(in, name, activeFrom)
		return in
	})
	return err
}

func setAuthSecretRef(db client.Object, name string, activeFrom *metav1.Time) {
	f := authSecretField(reflect.ValueOf(db))
	if !f.IsValid() || !f.CanSet() {
		return
	}
	if f.IsNil() {
		f.Set(reflect.New(f.Type().Elem()))
	}
	ref := f.Elem()
	if x := ref.FieldByName("Name"); x.IsValid() && x.CanSet() {
		x.SetString(name)
	}
	if activeFrom != nil {
		if x := ref.FieldByName("ActiveFrom"); x.IsValid() && x.CanSet() {
			x.Set(reflect.ValueOf(activeFrom))
		}
	}
}

// Spec.AuthSecret is a version-specific SecretReference type (kubedb/v1 vs
// v1alpha2), so it is accessed via reflection to keep this package version-agnostic.
func authSecretField(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	spec := v.FieldByName("Spec")
	if !spec.IsValid() {
		return reflect.Value{}
	}
	return spec.FieldByName("AuthSecret")
}
