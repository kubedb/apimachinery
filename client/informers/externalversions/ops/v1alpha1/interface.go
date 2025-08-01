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
	// CassandraOpsRequests returns a CassandraOpsRequestInformer.
	CassandraOpsRequests() CassandraOpsRequestInformer
	// ClickHouseOpsRequests returns a ClickHouseOpsRequestInformer.
	ClickHouseOpsRequests() ClickHouseOpsRequestInformer
	// DruidOpsRequests returns a DruidOpsRequestInformer.
	DruidOpsRequests() DruidOpsRequestInformer
	// ElasticsearchOpsRequests returns a ElasticsearchOpsRequestInformer.
	ElasticsearchOpsRequests() ElasticsearchOpsRequestInformer
	// EtcdOpsRequests returns a EtcdOpsRequestInformer.
	EtcdOpsRequests() EtcdOpsRequestInformer
	// FerretDBOpsRequests returns a FerretDBOpsRequestInformer.
	FerretDBOpsRequests() FerretDBOpsRequestInformer
	// HazelcastOpsRequests returns a HazelcastOpsRequestInformer.
	HazelcastOpsRequests() HazelcastOpsRequestInformer
	// IgniteOpsRequests returns a IgniteOpsRequestInformer.
	IgniteOpsRequests() IgniteOpsRequestInformer
	// KafkaOpsRequests returns a KafkaOpsRequestInformer.
	KafkaOpsRequests() KafkaOpsRequestInformer
	// MSSQLServerOpsRequests returns a MSSQLServerOpsRequestInformer.
	MSSQLServerOpsRequests() MSSQLServerOpsRequestInformer
	// MariaDBOpsRequests returns a MariaDBOpsRequestInformer.
	MariaDBOpsRequests() MariaDBOpsRequestInformer
	// MemcachedOpsRequests returns a MemcachedOpsRequestInformer.
	MemcachedOpsRequests() MemcachedOpsRequestInformer
	// MongoDBOpsRequests returns a MongoDBOpsRequestInformer.
	MongoDBOpsRequests() MongoDBOpsRequestInformer
	// MySQLOpsRequests returns a MySQLOpsRequestInformer.
	MySQLOpsRequests() MySQLOpsRequestInformer
	// PerconaXtraDBOpsRequests returns a PerconaXtraDBOpsRequestInformer.
	PerconaXtraDBOpsRequests() PerconaXtraDBOpsRequestInformer
	// PgBouncerOpsRequests returns a PgBouncerOpsRequestInformer.
	PgBouncerOpsRequests() PgBouncerOpsRequestInformer
	// PgpoolOpsRequests returns a PgpoolOpsRequestInformer.
	PgpoolOpsRequests() PgpoolOpsRequestInformer
	// PostgresOpsRequests returns a PostgresOpsRequestInformer.
	PostgresOpsRequests() PostgresOpsRequestInformer
	// ProxySQLOpsRequests returns a ProxySQLOpsRequestInformer.
	ProxySQLOpsRequests() ProxySQLOpsRequestInformer
	// RabbitMQOpsRequests returns a RabbitMQOpsRequestInformer.
	RabbitMQOpsRequests() RabbitMQOpsRequestInformer
	// RedisOpsRequests returns a RedisOpsRequestInformer.
	RedisOpsRequests() RedisOpsRequestInformer
	// RedisSentinelOpsRequests returns a RedisSentinelOpsRequestInformer.
	RedisSentinelOpsRequests() RedisSentinelOpsRequestInformer
	// SinglestoreOpsRequests returns a SinglestoreOpsRequestInformer.
	SinglestoreOpsRequests() SinglestoreOpsRequestInformer
	// SolrOpsRequests returns a SolrOpsRequestInformer.
	SolrOpsRequests() SolrOpsRequestInformer
	// ZooKeeperOpsRequests returns a ZooKeeperOpsRequestInformer.
	ZooKeeperOpsRequests() ZooKeeperOpsRequestInformer
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

// CassandraOpsRequests returns a CassandraOpsRequestInformer.
func (v *version) CassandraOpsRequests() CassandraOpsRequestInformer {
	return &cassandraOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ClickHouseOpsRequests returns a ClickHouseOpsRequestInformer.
func (v *version) ClickHouseOpsRequests() ClickHouseOpsRequestInformer {
	return &clickHouseOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// DruidOpsRequests returns a DruidOpsRequestInformer.
func (v *version) DruidOpsRequests() DruidOpsRequestInformer {
	return &druidOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ElasticsearchOpsRequests returns a ElasticsearchOpsRequestInformer.
func (v *version) ElasticsearchOpsRequests() ElasticsearchOpsRequestInformer {
	return &elasticsearchOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// EtcdOpsRequests returns a EtcdOpsRequestInformer.
func (v *version) EtcdOpsRequests() EtcdOpsRequestInformer {
	return &etcdOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// FerretDBOpsRequests returns a FerretDBOpsRequestInformer.
func (v *version) FerretDBOpsRequests() FerretDBOpsRequestInformer {
	return &ferretDBOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// HazelcastOpsRequests returns a HazelcastOpsRequestInformer.
func (v *version) HazelcastOpsRequests() HazelcastOpsRequestInformer {
	return &hazelcastOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// IgniteOpsRequests returns a IgniteOpsRequestInformer.
func (v *version) IgniteOpsRequests() IgniteOpsRequestInformer {
	return &igniteOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// KafkaOpsRequests returns a KafkaOpsRequestInformer.
func (v *version) KafkaOpsRequests() KafkaOpsRequestInformer {
	return &kafkaOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MSSQLServerOpsRequests returns a MSSQLServerOpsRequestInformer.
func (v *version) MSSQLServerOpsRequests() MSSQLServerOpsRequestInformer {
	return &mSSQLServerOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MariaDBOpsRequests returns a MariaDBOpsRequestInformer.
func (v *version) MariaDBOpsRequests() MariaDBOpsRequestInformer {
	return &mariaDBOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MemcachedOpsRequests returns a MemcachedOpsRequestInformer.
func (v *version) MemcachedOpsRequests() MemcachedOpsRequestInformer {
	return &memcachedOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MongoDBOpsRequests returns a MongoDBOpsRequestInformer.
func (v *version) MongoDBOpsRequests() MongoDBOpsRequestInformer {
	return &mongoDBOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MySQLOpsRequests returns a MySQLOpsRequestInformer.
func (v *version) MySQLOpsRequests() MySQLOpsRequestInformer {
	return &mySQLOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PerconaXtraDBOpsRequests returns a PerconaXtraDBOpsRequestInformer.
func (v *version) PerconaXtraDBOpsRequests() PerconaXtraDBOpsRequestInformer {
	return &perconaXtraDBOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PgBouncerOpsRequests returns a PgBouncerOpsRequestInformer.
func (v *version) PgBouncerOpsRequests() PgBouncerOpsRequestInformer {
	return &pgBouncerOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PgpoolOpsRequests returns a PgpoolOpsRequestInformer.
func (v *version) PgpoolOpsRequests() PgpoolOpsRequestInformer {
	return &pgpoolOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// PostgresOpsRequests returns a PostgresOpsRequestInformer.
func (v *version) PostgresOpsRequests() PostgresOpsRequestInformer {
	return &postgresOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ProxySQLOpsRequests returns a ProxySQLOpsRequestInformer.
func (v *version) ProxySQLOpsRequests() ProxySQLOpsRequestInformer {
	return &proxySQLOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RabbitMQOpsRequests returns a RabbitMQOpsRequestInformer.
func (v *version) RabbitMQOpsRequests() RabbitMQOpsRequestInformer {
	return &rabbitMQOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RedisOpsRequests returns a RedisOpsRequestInformer.
func (v *version) RedisOpsRequests() RedisOpsRequestInformer {
	return &redisOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// RedisSentinelOpsRequests returns a RedisSentinelOpsRequestInformer.
func (v *version) RedisSentinelOpsRequests() RedisSentinelOpsRequestInformer {
	return &redisSentinelOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// SinglestoreOpsRequests returns a SinglestoreOpsRequestInformer.
func (v *version) SinglestoreOpsRequests() SinglestoreOpsRequestInformer {
	return &singlestoreOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// SolrOpsRequests returns a SolrOpsRequestInformer.
func (v *version) SolrOpsRequests() SolrOpsRequestInformer {
	return &solrOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ZooKeeperOpsRequests returns a ZooKeeperOpsRequestInformer.
func (v *version) ZooKeeperOpsRequests() ZooKeeperOpsRequestInformer {
	return &zooKeeperOpsRequestInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
