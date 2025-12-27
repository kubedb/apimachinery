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

package fuzzer

import (
	"kubedb.dev/apimachinery/apis/gitops/v1alpha1"

	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/randfill"
)

// Funcs returns the fuzzer functions for this api group.
var Funcs = func(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(s *v1alpha1.Elasticsearch, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MariaDB, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.Memcached, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MongoDB, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MySQL, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PerconaXtraDB, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PgBouncer, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.Postgres, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.ProxySQL, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.Redis, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
	}
}
