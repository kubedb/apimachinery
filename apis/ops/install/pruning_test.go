/*
Copyright The KubeDB Authors.

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

	"kubedb.dev/apimachinery/apis/ops/fuzzer"
	"kubedb.dev/apimachinery/apis/ops/v1alpha1"

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	crdfuzz "kmodules.xyz/crd-schema-fuzz"
)

func TestPruneTypes(t *testing.T) {
	Install(clientsetscheme.Scheme)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.ElasticsearchOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.EtcdOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MemcachedOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MongoDBOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.MySQLOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.PerconaXtraDBOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.PostgresOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.ProxySQLOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
	crdfuzz.SchemaFuzzTestForV1beta1CRD(t, clientsetscheme.Scheme, v1alpha1.RedisOpsRequest{}.CustomResourceDefinition(), fuzzer.Funcs)
}
