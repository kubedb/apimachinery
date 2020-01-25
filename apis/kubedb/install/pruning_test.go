/*
Copyright The KubeVault Authors.

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

package install

import (
	"testing"

	"kubedb.dev/apimachinery/apis/kubedb/fuzzer"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	crdfuzz "kmodules.xyz/crd-schema-fuzz"
)

func TestPruneTypes(t *testing.T) {
	Install(clientsetscheme.Scheme)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.Elasticsearch{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.Etcd{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MariaDB{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.Memcached{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MongoDB{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MySQL{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.PerconaXtraDB{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.PgBouncer{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.Postgres{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.ProxySQL{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.Redis{}.CustomResourceDefinition(), fuzzer.Funcs)
}
