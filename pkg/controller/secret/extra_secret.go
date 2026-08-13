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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cu "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExtraSecret struct {
	Name         string
	Type         core.SecretType
	RequiredKeys []string
	Generate     func() map[string][]byte
}

func (o Options) EnsureExtraSecret(ctx context.Context, spec ExtraSecret) (*core.Secret, error) {
	obj := &core.Secret{ObjectMeta: metav1.ObjectMeta{Name: spec.Name, Namespace: o.DB.GetNamespace()}}
	_, err := cu.CreateOrPatch(ctx, o.KBClient, obj, func(in client.Object, createOp bool) client.Object {
		s := in.(*core.Secret)
		s.Labels = meta_util.OverwriteKeys(s.Labels, o.DB.OffshootLabels())
		core_util.EnsureOwnerReference(&s.ObjectMeta, o.DB.AsOwner())
		if s.Type == "" {
			s.Type = spec.Type
		}
		if len(s.Data) == 0 && spec.Generate != nil {
			s.Data = spec.Generate()
		}
		return s
	})
	if err != nil {
		return nil, err
	}
	if err := requireKeys(obj.Data, spec.RequiredKeys); err != nil {
		return nil, fmt.Errorf("secret %s/%s: %w", obj.Namespace, obj.Name, err)
	}
	return obj, nil
}

func requireKeys(data map[string][]byte, keys []string) error {
	for _, k := range keys {
		if len(data[k]) == 0 {
			return fmt.Errorf("missing required key %q", k)
		}
	}
	return nil
}
