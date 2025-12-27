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
	"kubedb.dev/apimachinery/apis/ops/v1alpha1"

	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/randfill"
)

// Funcs returns the fuzzer functions for this api group.
var Funcs = func(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(s *v1alpha1.ElasticsearchOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.EtcdOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MariaDBOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MemcachedOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MongoDBOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MySQLOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PerconaXtraDBOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PostgresOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.ProxySQLOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.RedisOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.RabbitMQOpsRequest, c randfill.Continue) {
			c.Fill(s) // fuzz self without calling this function again
		},
	}
}
