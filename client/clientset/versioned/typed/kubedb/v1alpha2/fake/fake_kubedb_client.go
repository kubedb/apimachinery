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
	v1alpha2 "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2"

	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeKubedbV1alpha2 struct {
	*testing.Fake
}

func (c *FakeKubedbV1alpha2) Druids(namespace string) v1alpha2.DruidInterface {
	return &FakeDruids{c, namespace}
}

func (c *FakeKubedbV1alpha2) Elasticsearches(namespace string) v1alpha2.ElasticsearchInterface {
	return &FakeElasticsearches{c, namespace}
}

func (c *FakeKubedbV1alpha2) Etcds(namespace string) v1alpha2.EtcdInterface {
	return &FakeEtcds{c, namespace}
}

func (c *FakeKubedbV1alpha2) Kafkas(namespace string) v1alpha2.KafkaInterface {
	return &FakeKafkas{c, namespace}
}

func (c *FakeKubedbV1alpha2) MariaDBs(namespace string) v1alpha2.MariaDBInterface {
	return &FakeMariaDBs{c, namespace}
}

func (c *FakeKubedbV1alpha2) Memcacheds(namespace string) v1alpha2.MemcachedInterface {
	return &FakeMemcacheds{c, namespace}
}

func (c *FakeKubedbV1alpha2) MongoDBs(namespace string) v1alpha2.MongoDBInterface {
	return &FakeMongoDBs{c, namespace}
}

func (c *FakeKubedbV1alpha2) MySQLs(namespace string) v1alpha2.MySQLInterface {
	return &FakeMySQLs{c, namespace}
}

func (c *FakeKubedbV1alpha2) PerconaXtraDBs(namespace string) v1alpha2.PerconaXtraDBInterface {
	return &FakePerconaXtraDBs{c, namespace}
}

func (c *FakeKubedbV1alpha2) PgBouncers(namespace string) v1alpha2.PgBouncerInterface {
	return &FakePgBouncers{c, namespace}
}

func (c *FakeKubedbV1alpha2) Pgpools(namespace string) v1alpha2.PgpoolInterface {
	return &FakePgpools{c, namespace}
}

func (c *FakeKubedbV1alpha2) Postgreses(namespace string) v1alpha2.PostgresInterface {
	return &FakePostgreses{c, namespace}
}

func (c *FakeKubedbV1alpha2) ProxySQLs(namespace string) v1alpha2.ProxySQLInterface {
	return &FakeProxySQLs{c, namespace}
}

func (c *FakeKubedbV1alpha2) Redises(namespace string) v1alpha2.RedisInterface {
	return &FakeRedises{c, namespace}
}

func (c *FakeKubedbV1alpha2) RedisSentinels(namespace string) v1alpha2.RedisSentinelInterface {
	return &FakeRedisSentinels{c, namespace}
}

func (c *FakeKubedbV1alpha2) Singlestores(namespace string) v1alpha2.SinglestoreInterface {
	return &FakeSinglestores{c, namespace}
}

func (c *FakeKubedbV1alpha2) ZooKeepers(namespace string) v1alpha2.ZooKeeperInterface {
	return &FakeZooKeepers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeKubedbV1alpha2) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
