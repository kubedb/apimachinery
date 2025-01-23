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

package v1alpha2

import (
	"net/http"

	rest "k8s.io/client-go/rest"
	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

type KubedbV1alpha2Interface interface {
	RESTClient() rest.Interface
	CassandrasGetter
	ClickHousesGetter
	DruidsGetter
	ElasticsearchesGetter
	EtcdsGetter
	FerretDBsGetter
	KafkasGetter
	MSSQLServersGetter
	MariaDBsGetter
	MemcachedsGetter
	MongoDBsGetter
	MySQLsGetter
	PerconaXtraDBsGetter
	PgBouncersGetter
	PgpoolsGetter
	PostgresesGetter
	ProxySQLsGetter
	RabbitMQsGetter
	RedisesGetter
	RedisSentinelsGetter
	SinglestoresGetter
	SolrsGetter
	ZooKeepersGetter
}

// KubedbV1alpha2Client is used to interact with features provided by the kubedb.com group.
type KubedbV1alpha2Client struct {
	restClient rest.Interface
}

func (c *KubedbV1alpha2Client) Cassandras(namespace string) CassandraInterface {
	return newCassandras(c, namespace)
}

func (c *KubedbV1alpha2Client) ClickHouses(namespace string) ClickHouseInterface {
	return newClickHouses(c, namespace)
}

func (c *KubedbV1alpha2Client) Druids(namespace string) DruidInterface {
	return newDruids(c, namespace)
}

func (c *KubedbV1alpha2Client) Elasticsearches(namespace string) ElasticsearchInterface {
	return newElasticsearches(c, namespace)
}

func (c *KubedbV1alpha2Client) Etcds(namespace string) EtcdInterface {
	return newEtcds(c, namespace)
}

func (c *KubedbV1alpha2Client) FerretDBs(namespace string) FerretDBInterface {
	return newFerretDBs(c, namespace)
}

func (c *KubedbV1alpha2Client) Kafkas(namespace string) KafkaInterface {
	return newKafkas(c, namespace)
}

func (c *KubedbV1alpha2Client) MSSQLServers(namespace string) MSSQLServerInterface {
	return newMSSQLServers(c, namespace)
}

func (c *KubedbV1alpha2Client) MariaDBs(namespace string) MariaDBInterface {
	return newMariaDBs(c, namespace)
}

func (c *KubedbV1alpha2Client) Memcacheds(namespace string) MemcachedInterface {
	return newMemcacheds(c, namespace)
}

func (c *KubedbV1alpha2Client) MongoDBs(namespace string) MongoDBInterface {
	return newMongoDBs(c, namespace)
}

func (c *KubedbV1alpha2Client) MySQLs(namespace string) MySQLInterface {
	return newMySQLs(c, namespace)
}

func (c *KubedbV1alpha2Client) PerconaXtraDBs(namespace string) PerconaXtraDBInterface {
	return newPerconaXtraDBs(c, namespace)
}

func (c *KubedbV1alpha2Client) PgBouncers(namespace string) PgBouncerInterface {
	return newPgBouncers(c, namespace)
}

func (c *KubedbV1alpha2Client) Pgpools(namespace string) PgpoolInterface {
	return newPgpools(c, namespace)
}

func (c *KubedbV1alpha2Client) Postgreses(namespace string) PostgresInterface {
	return newPostgreses(c, namespace)
}

func (c *KubedbV1alpha2Client) ProxySQLs(namespace string) ProxySQLInterface {
	return newProxySQLs(c, namespace)
}

func (c *KubedbV1alpha2Client) RabbitMQs(namespace string) RabbitMQInterface {
	return newRabbitMQs(c, namespace)
}

func (c *KubedbV1alpha2Client) Redises(namespace string) RedisInterface {
	return newRedises(c, namespace)
}

func (c *KubedbV1alpha2Client) RedisSentinels(namespace string) RedisSentinelInterface {
	return newRedisSentinels(c, namespace)
}

func (c *KubedbV1alpha2Client) Singlestores(namespace string) SinglestoreInterface {
	return newSinglestores(c, namespace)
}

func (c *KubedbV1alpha2Client) Solrs(namespace string) SolrInterface {
	return newSolrs(c, namespace)
}

func (c *KubedbV1alpha2Client) ZooKeepers(namespace string) ZooKeeperInterface {
	return newZooKeepers(c, namespace)
}

// NewForConfig creates a new KubedbV1alpha2Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*KubedbV1alpha2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new KubedbV1alpha2Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*KubedbV1alpha2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &KubedbV1alpha2Client{client}, nil
}

// NewForConfigOrDie creates a new KubedbV1alpha2Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *KubedbV1alpha2Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new KubedbV1alpha2Client for the given RESTClient.
func New(c rest.Interface) *KubedbV1alpha2Client {
	return &KubedbV1alpha2Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha2.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *KubedbV1alpha2Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
