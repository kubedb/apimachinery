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

	"kubedb.dev/apimachinery/apis/courier/fuzzer"
	"kubedb.dev/apimachinery/apis/courier/v1alpha1"

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/apiextensions"
	crdfuzz "kmodules.xyz/crd-schema-fuzz"
)

func TestPruneTypes(t *testing.T) {
	Install(clientsetscheme.Scheme)

	// Migration is the duck type and is not served as its own CRD; the served
	// resources are the per-engine {DB}Migration kinds below.
	crds := []*apiextensions.CustomResourceDefinition{
		(v1alpha1.PostgresMigration{}).CustomResourceDefinition(),
		(v1alpha1.MySQLMigration{}).CustomResourceDefinition(),
		(v1alpha1.MariaDBMigration{}).CustomResourceDefinition(),
		(v1alpha1.MongoDBMigration{}).CustomResourceDefinition(),
		(v1alpha1.MSSQLServerMigration{}).CustomResourceDefinition(),
	}

	// CRD v1
	for _, crd := range crds {
		if crd.V1 != nil {
			crdfuzz.SchemaFuzzTestForV1CRD(t, clientsetscheme.Scheme, crd.V1, fuzzer.Funcs)
		}
	}
}
