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
	// Druids returns a DruidInformer.
	Druids() DruidInformer
	// Elasticsearches returns a ElasticsearchInformer.
	Elasticsearches() ElasticsearchInformer
	// Etcds returns a EtcdInformer.
	Etcds() EtcdInformer
	// Kafkas returns a KafkaInformer.
	Kafkas() KafkaInformer
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
	// Pgpools returns a PgpoolInformer.
	Pgpools() PgpoolInformer
	// Postgreses returns a PostgresInformer.
	Postgreses() PostgresInformer
	// ProxySQLs returns a ProxySQLInformer.
	ProxySQLs() ProxySQLInformer
	// RabbitMQs returns a RabbitMQInformer.
	RabbitMQs() RabbitMQInformer
	// Redises returns a RedisInformer.
	Redises() RedisInformer
	// RedisSentinels returns a RedisSentinelInformer.
	RedisSentinels() RedisSentinelInformer
	// Singlestores returns a SinglestoreInformer.
	Singlestores() SinglestoreInformer
	// ZooKeepers returns a ZooKeeperInformer.
	ZooKeepers() ZooKeeperInformer
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

// Druids returns a DruidInformer.
func (v *version) Druids() DruidInformer {
	return &druidInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Elasticsearches returns a ElasticsearchInformer.
func (v *version) Elasticsearches() ElasticsearchInformer {
	return &elasticsearchInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Etcds returns a EtcdInformer.
func (v *version) Etcds() EtcdInformer {
	return &etcdInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Kafkas returns a KafkaInformer.
func (v *version) Kafkas() KafkaInformer {
	return &kafkaInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
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

// Pgpools returns a PgpoolInformer.
func (v *version) Pgpools() PgpoolInformer {
	return &pgpoolInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Postgreses returns a PostgresInformer.
func (v *version) Postgreses() PostgresInformer {
	return &postgresInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ProxySQLs returns a ProxySQLInformer.
func (v *version) ProxySQLs() ProxySQLInformer {
	return &proxySQLInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RabbitMQs returns a RabbitMQInformer.
func (v *version) RabbitMQs() RabbitMQInformer {
	return &rabbitMQInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Redises returns a RedisInformer.
func (v *version) Redises() RedisInformer {
	return &redisInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RedisSentinels returns a RedisSentinelInformer.
func (v *version) RedisSentinels() RedisSentinelInformer {
	return &redisSentinelInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Singlestores returns a SinglestoreInformer.
func (v *version) Singlestores() SinglestoreInformer {
	return &singlestoreInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ZooKeepers returns a ZooKeeperInformer.
func (v *version) ZooKeepers() ZooKeeperInformer {
	return &zooKeeperInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
