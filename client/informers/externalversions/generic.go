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

package externalversions

import (
	"fmt"

	v1alpha1 "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
	autoscalingv1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	catalogv1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	elasticsearchv1alpha1 "kubedb.dev/apimachinery/apis/elasticsearch/v1alpha1"
	gitopsv1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"
	kafkav1alpha1 "kubedb.dev/apimachinery/apis/kafka/v1alpha1"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	opsv1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	postgresv1alpha1 "kubedb.dev/apimachinery/apis/postgres/v1alpha1"
	schemav1alpha1 "kubedb.dev/apimachinery/apis/schema/v1alpha1"

	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=archiver.kubedb.com, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("mssqlserverarchivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Archiver().V1alpha1().MSSQLServerArchivers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("mariadbarchivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Archiver().V1alpha1().MariaDBArchivers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("mongodbarchivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Archiver().V1alpha1().MongoDBArchivers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("mysqlarchivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Archiver().V1alpha1().MySQLArchivers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("postgresarchivers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Archiver().V1alpha1().PostgresArchivers().Informer()}, nil

		// Group=autoscaling.kubedb.com, Version=v1alpha1
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("cassandraautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().CassandraAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("clickhouseautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().ClickHouseAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("druidautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().DruidAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("elasticsearchautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().ElasticsearchAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("etcdautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().EtcdAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("ferretdbautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().FerretDBAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("hazelcastautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().HazelcastAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("kafkaautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().KafkaAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("mssqlserverautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().MSSQLServerAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("mariadbautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().MariaDBAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("memcachedautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().MemcachedAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("mongodbautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().MongoDBAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("mysqlautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().MySQLAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("perconaxtradbautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().PerconaXtraDBAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("pgbouncerautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().PgBouncerAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("pgpoolautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().PgpoolAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("postgresautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().PostgresAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("proxysqlautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().ProxySQLAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("rabbitmqautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().RabbitMQAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("redisautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().RedisAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("redissentinelautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().RedisSentinelAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("singlestoreautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().SinglestoreAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("solrautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().SolrAutoscalers().Informer()}, nil
	case autoscalingv1alpha1.SchemeGroupVersion.WithResource("zookeeperautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().ZooKeeperAutoscalers().Informer()}, nil

		// Group=catalog.kubedb.com, Version=v1alpha1
	case catalogv1alpha1.SchemeGroupVersion.WithResource("cassandraversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().CassandraVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("clickhouseversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().ClickHouseVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("druidversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().DruidVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("elasticsearchversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().ElasticsearchVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("etcdversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().EtcdVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("ferretdbversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().FerretDBVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("hazelcastversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().HazelcastVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("igniteversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().IgniteVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("kafkaconnectorversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().KafkaConnectorVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("kafkaversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().KafkaVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("mssqlserverversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().MSSQLServerVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("mariadbversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().MariaDBVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("memcachedversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().MemcachedVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("mongodbversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().MongoDBVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("mysqlversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().MySQLVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("oracleversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().OracleVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("perconaxtradbversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().PerconaXtraDBVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("pgbouncerversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().PgBouncerVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("pgpoolversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().PgpoolVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("postgresversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().PostgresVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("proxysqlversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().ProxySQLVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("rabbitmqversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().RabbitMQVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("redisversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().RedisVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("schemaregistryversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().SchemaRegistryVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("singlestoreversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().SinglestoreVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("solrversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().SolrVersions().Informer()}, nil
	case catalogv1alpha1.SchemeGroupVersion.WithResource("zookeeperversions"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Catalog().V1alpha1().ZooKeeperVersions().Informer()}, nil

		// Group=elasticsearch.kubedb.com, Version=v1alpha1
	case elasticsearchv1alpha1.SchemeGroupVersion.WithResource("elasticsearchdashboards"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Elasticsearch().V1alpha1().ElasticsearchDashboards().Informer()}, nil

		// Group=gitops.kubedb.com, Version=v1alpha1
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("druids"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Druids().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("elasticsearches"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Elasticsearches().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("ferretdbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().FerretDBs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("kafkas"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Kafkas().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("mssqlservers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().MSSQLServers().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("mariadbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().MariaDBs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("memcacheds"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Memcacheds().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("mongodbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().MongoDBs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("mysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().MySQLs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("perconaxtradbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().PerconaXtraDBs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("pgbouncers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().PgBouncers().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("pgpools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Pgpools().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("postgreses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Postgreses().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("proxysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().ProxySQLs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("rabbitmqs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().RabbitMQs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("redises"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Redises().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("redissentinels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().RedisSentinels().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("singlestores"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Singlestores().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("solrs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().Solrs().Informer()}, nil
	case gitopsv1alpha1.SchemeGroupVersion.WithResource("zookeepers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Gitops().V1alpha1().ZooKeepers().Informer()}, nil

		// Group=kafka.kubedb.com, Version=v1alpha1
	case kafkav1alpha1.SchemeGroupVersion.WithResource("connectclusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kafka().V1alpha1().ConnectClusters().Informer()}, nil
	case kafkav1alpha1.SchemeGroupVersion.WithResource("connectors"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kafka().V1alpha1().Connectors().Informer()}, nil
	case kafkav1alpha1.SchemeGroupVersion.WithResource("restproxies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kafka().V1alpha1().RestProxies().Informer()}, nil
	case kafkav1alpha1.SchemeGroupVersion.WithResource("schemaregistries"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kafka().V1alpha1().SchemaRegistries().Informer()}, nil

		// Group=kubedb.com, Version=v1
	case v1.SchemeGroupVersion.WithResource("elasticsearches"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().Elasticsearches().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("kafkas"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().Kafkas().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("mariadbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().MariaDBs().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("memcacheds"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().Memcacheds().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("mongodbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().MongoDBs().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("mysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().MySQLs().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("perconaxtradbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().PerconaXtraDBs().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("pgbouncers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().PgBouncers().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("postgreses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().Postgreses().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("proxysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().ProxySQLs().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("redises"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().Redises().Informer()}, nil
	case v1.SchemeGroupVersion.WithResource("redissentinels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1().RedisSentinels().Informer()}, nil

		// Group=kubedb.com, Version=v1alpha2
	case v1alpha2.SchemeGroupVersion.WithResource("cassandras"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Cassandras().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("clickhouses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().ClickHouses().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("druids"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Druids().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("elasticsearches"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Elasticsearches().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("etcds"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Etcds().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("ferretdbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().FerretDBs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("hazelcasts"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Hazelcasts().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("ignites"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Ignites().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("kafkas"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Kafkas().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("mssqlservers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().MSSQLServers().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("mariadbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().MariaDBs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("memcacheds"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Memcacheds().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("mongodbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().MongoDBs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("mysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().MySQLs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("oracles"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Oracles().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("perconaxtradbs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().PerconaXtraDBs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("pgbouncers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().PgBouncers().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("pgpools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Pgpools().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("postgreses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Postgreses().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("proxysqls"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().ProxySQLs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("rabbitmqs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().RabbitMQs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("redises"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Redises().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("redissentinels"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().RedisSentinels().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("singlestores"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Singlestores().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("solrs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().Solrs().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("zookeepers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Kubedb().V1alpha2().ZooKeepers().Informer()}, nil

		// Group=ops.kubedb.com, Version=v1alpha1
	case opsv1alpha1.SchemeGroupVersion.WithResource("cassandraopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().CassandraOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("clickhouseopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().ClickHouseOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("druidopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().DruidOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("elasticsearchopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().ElasticsearchOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("etcdopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().EtcdOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("ferretdbopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().FerretDBOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("hazelcastopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().HazelcastOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("igniteopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().IgniteOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("kafkaopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().KafkaOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("mssqlserveropsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().MSSQLServerOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("mariadbopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().MariaDBOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("memcachedopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().MemcachedOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("mongodbopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().MongoDBOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("mysqlopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().MySQLOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("perconaxtradbopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().PerconaXtraDBOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("pgbounceropsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().PgBouncerOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("pgpoolopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().PgpoolOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("postgresopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().PostgresOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("proxysqlopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().ProxySQLOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("rabbitmqopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().RabbitMQOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("redisopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().RedisOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("redissentinelopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().RedisSentinelOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("singlestoreopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().SinglestoreOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("solropsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().SolrOpsRequests().Informer()}, nil
	case opsv1alpha1.SchemeGroupVersion.WithResource("zookeeperopsrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Ops().V1alpha1().ZooKeeperOpsRequests().Informer()}, nil

		// Group=postgres.kubedb.com, Version=v1alpha1
	case postgresv1alpha1.SchemeGroupVersion.WithResource("publishers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Postgres().V1alpha1().Publishers().Informer()}, nil
	case postgresv1alpha1.SchemeGroupVersion.WithResource("subscribers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Postgres().V1alpha1().Subscribers().Informer()}, nil

		// Group=schema.kubedb.com, Version=v1alpha1
	case schemav1alpha1.SchemeGroupVersion.WithResource("mariadbdatabases"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Schema().V1alpha1().MariaDBDatabases().Informer()}, nil
	case schemav1alpha1.SchemeGroupVersion.WithResource("mongodbdatabases"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Schema().V1alpha1().MongoDBDatabases().Informer()}, nil
	case schemav1alpha1.SchemeGroupVersion.WithResource("mysqldatabases"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Schema().V1alpha1().MySQLDatabases().Informer()}, nil
	case schemav1alpha1.SchemeGroupVersion.WithResource("postgresdatabases"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Schema().V1alpha1().PostgresDatabases().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
