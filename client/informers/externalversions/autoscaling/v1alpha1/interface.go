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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ElasticsearchAutoscalers returns a ElasticsearchAutoscalerInformer.
	ElasticsearchAutoscalers() ElasticsearchAutoscalerInformer
	// EtcdAutoscalers returns a EtcdAutoscalerInformer.
	EtcdAutoscalers() EtcdAutoscalerInformer
	// MariaDBAutoscalers returns a MariaDBAutoscalerInformer.
	MariaDBAutoscalers() MariaDBAutoscalerInformer
	// MemcachedAutoscalers returns a MemcachedAutoscalerInformer.
	MemcachedAutoscalers() MemcachedAutoscalerInformer
	// MongoDBAutoscalers returns a MongoDBAutoscalerInformer.
	MongoDBAutoscalers() MongoDBAutoscalerInformer
	// MySQLAutoscalers returns a MySQLAutoscalerInformer.
	MySQLAutoscalers() MySQLAutoscalerInformer
	// PerconaXtraDBAutoscalers returns a PerconaXtraDBAutoscalerInformer.
	PerconaXtraDBAutoscalers() PerconaXtraDBAutoscalerInformer
	// PgBouncerAutoscalers returns a PgBouncerAutoscalerInformer.
	PgBouncerAutoscalers() PgBouncerAutoscalerInformer
	// PostgresAutoscalers returns a PostgresAutoscalerInformer.
	PostgresAutoscalers() PostgresAutoscalerInformer
	// ProxySQLAutoscalers returns a ProxySQLAutoscalerInformer.
	ProxySQLAutoscalers() ProxySQLAutoscalerInformer
	// RedisAutoscalers returns a RedisAutoscalerInformer.
	RedisAutoscalers() RedisAutoscalerInformer
	// VerticalPodAutopilots returns a VerticalPodAutopilotInformer.
	VerticalPodAutopilots() VerticalPodAutopilotInformer
	// VerticalPodAutopilotCheckpoints returns a VerticalPodAutopilotCheckpointInformer.
	VerticalPodAutopilotCheckpoints() VerticalPodAutopilotCheckpointInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ElasticsearchAutoscalers returns a ElasticsearchAutoscalerInformer.
func (v *version) ElasticsearchAutoscalers() ElasticsearchAutoscalerInformer {
	return &elasticsearchAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// EtcdAutoscalers returns a EtcdAutoscalerInformer.
func (v *version) EtcdAutoscalers() EtcdAutoscalerInformer {
	return &etcdAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MariaDBAutoscalers returns a MariaDBAutoscalerInformer.
func (v *version) MariaDBAutoscalers() MariaDBAutoscalerInformer {
	return &mariaDBAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MemcachedAutoscalers returns a MemcachedAutoscalerInformer.
func (v *version) MemcachedAutoscalers() MemcachedAutoscalerInformer {
	return &memcachedAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MongoDBAutoscalers returns a MongoDBAutoscalerInformer.
func (v *version) MongoDBAutoscalers() MongoDBAutoscalerInformer {
	return &mongoDBAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MySQLAutoscalers returns a MySQLAutoscalerInformer.
func (v *version) MySQLAutoscalers() MySQLAutoscalerInformer {
	return &mySQLAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PerconaXtraDBAutoscalers returns a PerconaXtraDBAutoscalerInformer.
func (v *version) PerconaXtraDBAutoscalers() PerconaXtraDBAutoscalerInformer {
	return &perconaXtraDBAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PgBouncerAutoscalers returns a PgBouncerAutoscalerInformer.
func (v *version) PgBouncerAutoscalers() PgBouncerAutoscalerInformer {
	return &pgBouncerAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PostgresAutoscalers returns a PostgresAutoscalerInformer.
func (v *version) PostgresAutoscalers() PostgresAutoscalerInformer {
	return &postgresAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ProxySQLAutoscalers returns a ProxySQLAutoscalerInformer.
func (v *version) ProxySQLAutoscalers() ProxySQLAutoscalerInformer {
	return &proxySQLAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RedisAutoscalers returns a RedisAutoscalerInformer.
func (v *version) RedisAutoscalers() RedisAutoscalerInformer {
	return &redisAutoscalerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// VerticalPodAutopilots returns a VerticalPodAutopilotInformer.
func (v *version) VerticalPodAutopilots() VerticalPodAutopilotInformer {
	return &verticalPodAutopilotInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// VerticalPodAutopilotCheckpoints returns a VerticalPodAutopilotCheckpointInformer.
func (v *version) VerticalPodAutopilotCheckpoints() VerticalPodAutopilotCheckpointInformer {
	return &verticalPodAutopilotCheckpointInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
