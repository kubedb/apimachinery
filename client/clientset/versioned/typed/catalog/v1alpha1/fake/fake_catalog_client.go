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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/catalog/v1alpha1"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCatalogV1alpha1 struct {
	*testing.Fake
}

func (c *FakeCatalogV1alpha1) CassandraVersions() v1alpha1.CassandraVersionInterface {
	return &FakeCassandraVersions{c}
}

func (c *FakeCatalogV1alpha1) ClickHouseVersions() v1alpha1.ClickHouseVersionInterface {
	return &FakeClickHouseVersions{c}
}

func (c *FakeCatalogV1alpha1) DruidVersions() v1alpha1.DruidVersionInterface {
	return &FakeDruidVersions{c}
}

func (c *FakeCatalogV1alpha1) ElasticsearchVersions() v1alpha1.ElasticsearchVersionInterface {
	return &FakeElasticsearchVersions{c}
}

func (c *FakeCatalogV1alpha1) EtcdVersions() v1alpha1.EtcdVersionInterface {
	return &FakeEtcdVersions{c}
}

func (c *FakeCatalogV1alpha1) FerretDBVersions() v1alpha1.FerretDBVersionInterface {
	return &FakeFerretDBVersions{c}
}

func (c *FakeCatalogV1alpha1) HazelcastVersions() v1alpha1.HazelcastVersionInterface {
	return &FakeHazelcastVersions{c}
}

func (c *FakeCatalogV1alpha1) IgniteVersions() v1alpha1.IgniteVersionInterface {
	return &FakeIgniteVersions{c}
}

func (c *FakeCatalogV1alpha1) KafkaConnectorVersions() v1alpha1.KafkaConnectorVersionInterface {
	return &FakeKafkaConnectorVersions{c}
}

func (c *FakeCatalogV1alpha1) KafkaVersions() v1alpha1.KafkaVersionInterface {
	return &FakeKafkaVersions{c}
}

func (c *FakeCatalogV1alpha1) MSSQLServerVersions() v1alpha1.MSSQLServerVersionInterface {
	return &FakeMSSQLServerVersions{c}
}

func (c *FakeCatalogV1alpha1) MariaDBVersions() v1alpha1.MariaDBVersionInterface {
	return &FakeMariaDBVersions{c}
}

func (c *FakeCatalogV1alpha1) MemcachedVersions() v1alpha1.MemcachedVersionInterface {
	return &FakeMemcachedVersions{c}
}

func (c *FakeCatalogV1alpha1) MongoDBVersions() v1alpha1.MongoDBVersionInterface {
	return &FakeMongoDBVersions{c}
}

func (c *FakeCatalogV1alpha1) MySQLVersions() v1alpha1.MySQLVersionInterface {
	return &FakeMySQLVersions{c}
}

func (c *FakeCatalogV1alpha1) OracleVersions() v1alpha1.OracleVersionInterface {
	return &FakeOracleVersions{c}
}

func (c *FakeCatalogV1alpha1) PerconaXtraDBVersions() v1alpha1.PerconaXtraDBVersionInterface {
	return &FakePerconaXtraDBVersions{c}
}

func (c *FakeCatalogV1alpha1) PgBouncerVersions() v1alpha1.PgBouncerVersionInterface {
	return &FakePgBouncerVersions{c}
}

func (c *FakeCatalogV1alpha1) PgpoolVersions() v1alpha1.PgpoolVersionInterface {
	return &FakePgpoolVersions{c}
}

func (c *FakeCatalogV1alpha1) PostgresVersions() v1alpha1.PostgresVersionInterface {
	return &FakePostgresVersions{c}
}

func (c *FakeCatalogV1alpha1) ProxySQLVersions() v1alpha1.ProxySQLVersionInterface {
	return &FakeProxySQLVersions{c}
}

func (c *FakeCatalogV1alpha1) RabbitMQVersions() v1alpha1.RabbitMQVersionInterface {
	return &FakeRabbitMQVersions{c}
}

func (c *FakeCatalogV1alpha1) RedisVersions() v1alpha1.RedisVersionInterface {
	return &FakeRedisVersions{c}
}

func (c *FakeCatalogV1alpha1) SchemaRegistryVersions() v1alpha1.SchemaRegistryVersionInterface {
	return &FakeSchemaRegistryVersions{c}
}

func (c *FakeCatalogV1alpha1) SinglestoreVersions() v1alpha1.SinglestoreVersionInterface {
	return &FakeSinglestoreVersions{c}
}

func (c *FakeCatalogV1alpha1) SolrVersions() v1alpha1.SolrVersionInterface {
	return &FakeSolrVersions{c}
}

func (c *FakeCatalogV1alpha1) ZooKeeperVersions() v1alpha1.ZooKeeperVersionInterface {
	return &FakeZooKeeperVersions{c}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCatalogV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
