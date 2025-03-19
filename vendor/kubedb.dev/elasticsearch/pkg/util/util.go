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

package util

import (
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/pkg/errors"
)

func ServiceURL(db *api.Elasticsearch) string {
	return fmt.Sprintf("%v://%s.%s.svc:%d", db.GetConnectionScheme(), db.ServiceName(), db.GetNamespace(), kubedb.ElasticsearchRestPort)
}

func K8sChainOpts(db *api.Elasticsearch) *k8schain.Options {
	opts := &k8schain.Options{
		Namespace: db.Namespace,
	}
	if db.Spec.PodTemplate.Spec.ServiceAccountName == "" {
		opts.ServiceAccountName = db.OffshootName()
	} else {
		opts.ServiceAccountName = db.Spec.PodTemplate.Spec.ServiceAccountName
	}
	if db.Spec.PodTemplate.Spec.ImagePullSecrets != nil {
		for _, ims := range db.Spec.PodTemplate.Spec.ImagePullSecrets {
			opts.ImagePullSecrets = append(opts.ImagePullSecrets, ims.Name)
		}
	}
	return opts
}

func AppendError(err error, newError error) error {
	if err == nil {
		return newError
	} else {
		return errors.Wrap(err, newError.Error())
	}
}
