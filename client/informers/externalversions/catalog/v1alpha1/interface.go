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
	// ElasticsearchVersions returns a ElasticsearchVersionInformer.
	ElasticsearchVersions() ElasticsearchVersionInformer
	// EtcdVersions returns a EtcdVersionInformer.
	EtcdVersions() EtcdVersionInformer
	// KafkaVersions returns a KafkaVersionInformer.
	KafkaVersions() KafkaVersionInformer
	// MariaDBVersions returns a MariaDBVersionInformer.
	MariaDBVersions() MariaDBVersionInformer
	// MemcachedVersions returns a MemcachedVersionInformer.
	MemcachedVersions() MemcachedVersionInformer
	// MongoDBVersions returns a MongoDBVersionInformer.
	MongoDBVersions() MongoDBVersionInformer
	// MySQLVersions returns a MySQLVersionInformer.
	MySQLVersions() MySQLVersionInformer
	// PerconaXtraDBVersions returns a PerconaXtraDBVersionInformer.
	PerconaXtraDBVersions() PerconaXtraDBVersionInformer
	// PgBouncerVersions returns a PgBouncerVersionInformer.
	PgBouncerVersions() PgBouncerVersionInformer
	// PostgresVersions returns a PostgresVersionInformer.
	PostgresVersions() PostgresVersionInformer
	// ProxySQLVersions returns a ProxySQLVersionInformer.
	ProxySQLVersions() ProxySQLVersionInformer
	// RedisVersions returns a RedisVersionInformer.
	RedisVersions() RedisVersionInformer
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

// ElasticsearchVersions returns a ElasticsearchVersionInformer.
func (v *version) ElasticsearchVersions() ElasticsearchVersionInformer {
	return &elasticsearchVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// EtcdVersions returns a EtcdVersionInformer.
func (v *version) EtcdVersions() EtcdVersionInformer {
	return &etcdVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// KafkaVersions returns a KafkaVersionInformer.
func (v *version) KafkaVersions() KafkaVersionInformer {
	return &kafkaVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// MariaDBVersions returns a MariaDBVersionInformer.
func (v *version) MariaDBVersions() MariaDBVersionInformer {
	return &mariaDBVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// MemcachedVersions returns a MemcachedVersionInformer.
func (v *version) MemcachedVersions() MemcachedVersionInformer {
	return &memcachedVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// MongoDBVersions returns a MongoDBVersionInformer.
func (v *version) MongoDBVersions() MongoDBVersionInformer {
	return &mongoDBVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// MySQLVersions returns a MySQLVersionInformer.
func (v *version) MySQLVersions() MySQLVersionInformer {
	return &mySQLVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// PerconaXtraDBVersions returns a PerconaXtraDBVersionInformer.
func (v *version) PerconaXtraDBVersions() PerconaXtraDBVersionInformer {
	return &perconaXtraDBVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// PgBouncerVersions returns a PgBouncerVersionInformer.
func (v *version) PgBouncerVersions() PgBouncerVersionInformer {
	return &pgBouncerVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// PostgresVersions returns a PostgresVersionInformer.
func (v *version) PostgresVersions() PostgresVersionInformer {
	return &postgresVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// ProxySQLVersions returns a ProxySQLVersionInformer.
func (v *version) ProxySQLVersions() ProxySQLVersionInformer {
	return &proxySQLVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// RedisVersions returns a RedisVersionInformer.
func (v *version) RedisVersions() RedisVersionInformer {
	return &redisVersionInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
