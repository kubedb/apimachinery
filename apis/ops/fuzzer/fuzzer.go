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

	fuzz "github.com/google/gofuzz"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

// Funcs returns the fuzzer functions for this api group.
var Funcs = func(codecs runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		func(s *v1alpha1.ElasticsearchOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.EtcdOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MariaDBOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MemcachedOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MongoDBOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.MySQLOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PerconaXtraDBOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.PostgresOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.ProxySQLOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.RedisOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
		func(s *v1alpha1.RabbitMQOpsRequest, c fuzz.Continue) {
			c.FuzzNoCustom(s) // fuzz self without calling this function again
		},
	}
}
