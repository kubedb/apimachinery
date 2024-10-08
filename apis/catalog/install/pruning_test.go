/*
Copyright AppsCode Inc. and Contributors

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

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

func TestPruneTypes(t *testing.T) {
	Install(clientsetscheme.Scheme)

	// CRD v1
	// if crd := (v1alpha1.ElasticsearchVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.EtcdVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.MariaDBVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.MemcachedVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.MongoDBVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.MySQLVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.PerconaXtraDBVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.PostgresVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.ProxySQLVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
	// if crd := (v1alpha1.RedisVersion{}).CustomResourceDefinition(); crd.V1 != nil {
	// 	crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
	// }
}
