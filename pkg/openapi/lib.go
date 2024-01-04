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

package openapi

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/runtime"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

func ConfigureOpenAPI(scheme *runtime.Scheme, serverConfig *genericapiserver.RecommendedConfig) {
	ignorePrefixes := []string{
		"/swaggerapi",

		"/apis/mutators.autoscaling.kubedb.com/v1alpha1",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/elasticsearchautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/mariadbautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/mongodbautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/mysqlautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/perconaxtradbautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/pgbouncerautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/postgresautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/proxysqlautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/redisautoscalerwebhooks",
		"/apis/mutators.autoscaling.kubedb.com/v1alpha1/redissentinelautoscalerwebhooks",

		"/apis/mutators.dashboard.kubedb.com/v1alpha1",
		"/apis/mutators.dashboard.kubedb.com/v1alpha1/elasticsearchdashboardwebhooks",

		"/apis/mutators.kubedb.com/v1alpha1",
		"/apis/mutators.kubedb.com/v1alpha1/elasticsearchwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/etcdwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/kafkawebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/mariadbwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/memcachedwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/mongodbwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/mysqlwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/perconaxtradbwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/pgbouncerwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/postgreswebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/proxysqlwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/redissentinelwebhooks",
		"/apis/mutators.kubedb.com/v1alpha1/rediswebhooks",

		"/apis/mutators.ops.kubedb.com/v1alpha1",
		"/apis/mutators.ops.kubedb.com/v1alpha1/mysqlopsrequestwebhooks",

		"/apis/mutators.schema.kubedb.com/v1alpha1",
		"/apis/mutators.schema.kubedb.com/v1alpha1/mongodbdatabasewebhooks",
		"/apis/mutators.schema.kubedb.com/v1alpha1/mysqldatabasewebhooks",

		"/apis/validators.dashboard.kubedb.com/v1alpha1",
		"/apis/validators.dashboard.kubedb.com/v1alpha1/elasticsearchdashboardwebhooks",

		"/apis/validators.kubedb.com/v1alpha1",
		"/apis/validators.kubedb.com/v1alpha1/elasticsearchwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/etcdwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/kafkawebhooks",
		"/apis/validators.kubedb.com/v1alpha1/mariadbwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/memcachedwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/mongodbwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/mysqlwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/namespacewebhooks",
		"/apis/validators.kubedb.com/v1alpha1/perconaxtradbwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/pgbouncerwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/postgreswebhooks",
		"/apis/validators.kubedb.com/v1alpha1/proxysqlwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/redissentinelwebhooks",
		"/apis/validators.kubedb.com/v1alpha1/rediswebhooks",

		"/apis/validators.ops.kubedb.com/v1alpha1",
		"/apis/validators.ops.kubedb.com/v1alpha1/elasticsearchopsrequestwebhooks",
		"/apis/validators.ops.kubedb.com/v1alpha1/kafkaopsrequestwebhooks",
		"/apis/validators.ops.kubedb.com/v1alpha1/mongodbopsrequestwebhooks",
		"/apis/validators.ops.kubedb.com/v1alpha1/mysqlopsrequestwebhooks",
		"/apis/validators.ops.kubedb.com/v1alpha1/redisopsrequestwebhooks",

		"/apis/validators.postgres.kubedb.com/v1alpha1",
		"/apis/validators.postgres.kubedb.com/v1alpha1/publisherwebhooks",
		"/apis/validators.postgres.kubedb.com/v1alpha1/subscriberwebhooks",

		"/apis/validators.schema.kubedb.com/v1alpha1",
		"/apis/validators.schema.kubedb.com/v1alpha1/mongodbdatabasewebhooks",
		"/apis/validators.schema.kubedb.com/v1alpha1/mysqldatabasewebhooks",
		"/apis/validators.schema.kubedb.com/v1alpha1/postgresdatabasewebhooks",
	}

	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(dbapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(scheme))
	serverConfig.OpenAPIConfig.Info.Title = "kubedb-webhook-server"
	serverConfig.OpenAPIConfig.Info.Version = dbapi.SchemeGroupVersion.Version
	serverConfig.OpenAPIConfig.IgnorePrefixes = ignorePrefixes

	serverConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(dbapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(scheme))
	serverConfig.OpenAPIV3Config.Info.Title = "kubedb-webhook-server"
	serverConfig.OpenAPIV3Config.Info.Version = dbapi.SchemeGroupVersion.Version
	serverConfig.OpenAPIV3Config.IgnorePrefixes = ignorePrefixes
}
