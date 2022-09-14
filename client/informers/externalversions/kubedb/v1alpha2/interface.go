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

package v1alpha2

import (
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// Elasticsearches returns a ElasticsearchInformer.
	Elasticsearches() ElasticsearchInformer
	// Etcds returns a EtcdInformer.
	Etcds() EtcdInformer
	// MariaDBs returns a MariaDBInformer.
	MariaDBs() MariaDBInformer
	// Memcacheds returns a MemcachedInformer.
	Memcacheds() MemcachedInformer
	// MongoDBs returns a MongoDBInformer.
	MongoDBs() MongoDBInformer
	// MySQLs returns a MySQLInformer.
	MySQLs() MySQLInformer
	// PerconaXtraDBs returns a PerconaXtraDBInformer.
	PerconaXtraDBs() PerconaXtraDBInformer
	// PgBouncers returns a PgBouncerInformer.
	PgBouncers() PgBouncerInformer
	// Postgreses returns a PostgresInformer.
	Postgreses() PostgresInformer
	// ProxySQLs returns a ProxySQLInformer.
	ProxySQLs() ProxySQLInformer
	// Publishers returns a PublisherInformer.
	Publishers() PublisherInformer
	// Redises returns a RedisInformer.
	Redises() RedisInformer
	// RedisSentinels returns a RedisSentinelInformer.
	RedisSentinels() RedisSentinelInformer
	// Subscribers returns a SubscriberInformer.
	Subscribers() SubscriberInformer
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

// Elasticsearches returns a ElasticsearchInformer.
func (v *version) Elasticsearches() ElasticsearchInformer {
	return &elasticsearchInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Etcds returns a EtcdInformer.
func (v *version) Etcds() EtcdInformer {
	return &etcdInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MariaDBs returns a MariaDBInformer.
func (v *version) MariaDBs() MariaDBInformer {
	return &mariaDBInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Memcacheds returns a MemcachedInformer.
func (v *version) Memcacheds() MemcachedInformer {
	return &memcachedInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MongoDBs returns a MongoDBInformer.
func (v *version) MongoDBs() MongoDBInformer {
	return &mongoDBInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MySQLs returns a MySQLInformer.
func (v *version) MySQLs() MySQLInformer {
	return &mySQLInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PerconaXtraDBs returns a PerconaXtraDBInformer.
func (v *version) PerconaXtraDBs() PerconaXtraDBInformer {
	return &perconaXtraDBInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PgBouncers returns a PgBouncerInformer.
func (v *version) PgBouncers() PgBouncerInformer {
	return &pgBouncerInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Postgreses returns a PostgresInformer.
func (v *version) Postgreses() PostgresInformer {
	return &postgresInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ProxySQLs returns a ProxySQLInformer.
func (v *version) ProxySQLs() ProxySQLInformer {
	return &proxySQLInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Publishers returns a PublisherInformer.
func (v *version) Publishers() PublisherInformer {
	return &publisherInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Redises returns a RedisInformer.
func (v *version) Redises() RedisInformer {
	return &redisInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RedisSentinels returns a RedisSentinelInformer.
func (v *version) RedisSentinels() RedisSentinelInformer {
	return &redisSentinelInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Subscribers returns a SubscriberInformer.
func (v *version) Subscribers() SubscriberInformer {
	return &subscriberInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
