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

package v1alpha1

import (
	"net/http"

	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type CatalogV1alpha1Interface interface {
	RESTClient() rest.Interface
	CassandraVersionsGetter
	ClickHouseVersionsGetter
	DruidVersionsGetter
	ElasticsearchVersionsGetter
	EtcdVersionsGetter
	FerretDBVersionsGetter
	HazelcastVersionsGetter
	IgniteVersionsGetter
	KafkaConnectorVersionsGetter
	KafkaVersionsGetter
	MSSQLServerVersionsGetter
	MariaDBVersionsGetter
	MemcachedVersionsGetter
	MongoDBVersionsGetter
	MySQLVersionsGetter
	OracleVersionsGetter
	PerconaXtraDBVersionsGetter
	PgBouncerVersionsGetter
	PgpoolVersionsGetter
	PostgresVersionsGetter
	ProxySQLVersionsGetter
	RabbitMQVersionsGetter
	RedisVersionsGetter
	SchemaRegistryVersionsGetter
	SinglestoreVersionsGetter
	SolrVersionsGetter
	ZooKeeperVersionsGetter
}

// CatalogV1alpha1Client is used to interact with features provided by the catalog.kubedb.com group.
type CatalogV1alpha1Client struct {
	restClient rest.Interface
}

func (c *CatalogV1alpha1Client) CassandraVersions() CassandraVersionInterface {
	return newCassandraVersions(c)
}

func (c *CatalogV1alpha1Client) ClickHouseVersions() ClickHouseVersionInterface {
	return newClickHouseVersions(c)
}

func (c *CatalogV1alpha1Client) DruidVersions() DruidVersionInterface {
	return newDruidVersions(c)
}

func (c *CatalogV1alpha1Client) ElasticsearchVersions() ElasticsearchVersionInterface {
	return newElasticsearchVersions(c)
}

func (c *CatalogV1alpha1Client) EtcdVersions() EtcdVersionInterface {
	return newEtcdVersions(c)
}

func (c *CatalogV1alpha1Client) FerretDBVersions() FerretDBVersionInterface {
	return newFerretDBVersions(c)
}

func (c *CatalogV1alpha1Client) HazelcastVersions() HazelcastVersionInterface {
	return newHazelcastVersions(c)
}

func (c *CatalogV1alpha1Client) IgniteVersions() IgniteVersionInterface {
	return newIgniteVersions(c)
}

func (c *CatalogV1alpha1Client) KafkaConnectorVersions() KafkaConnectorVersionInterface {
	return newKafkaConnectorVersions(c)
}

func (c *CatalogV1alpha1Client) KafkaVersions() KafkaVersionInterface {
	return newKafkaVersions(c)
}

func (c *CatalogV1alpha1Client) MSSQLServerVersions() MSSQLServerVersionInterface {
	return newMSSQLServerVersions(c)
}

func (c *CatalogV1alpha1Client) MariaDBVersions() MariaDBVersionInterface {
	return newMariaDBVersions(c)
}

func (c *CatalogV1alpha1Client) MemcachedVersions() MemcachedVersionInterface {
	return newMemcachedVersions(c)
}

func (c *CatalogV1alpha1Client) MongoDBVersions() MongoDBVersionInterface {
	return newMongoDBVersions(c)
}

func (c *CatalogV1alpha1Client) MySQLVersions() MySQLVersionInterface {
	return newMySQLVersions(c)
}

func (c *CatalogV1alpha1Client) OracleVersions() OracleVersionInterface {
	return newOracleVersions(c)
}

func (c *CatalogV1alpha1Client) PerconaXtraDBVersions() PerconaXtraDBVersionInterface {
	return newPerconaXtraDBVersions(c)
}

func (c *CatalogV1alpha1Client) PgBouncerVersions() PgBouncerVersionInterface {
	return newPgBouncerVersions(c)
}

func (c *CatalogV1alpha1Client) PgpoolVersions() PgpoolVersionInterface {
	return newPgpoolVersions(c)
}

func (c *CatalogV1alpha1Client) PostgresVersions() PostgresVersionInterface {
	return newPostgresVersions(c)
}

func (c *CatalogV1alpha1Client) ProxySQLVersions() ProxySQLVersionInterface {
	return newProxySQLVersions(c)
}

func (c *CatalogV1alpha1Client) RabbitMQVersions() RabbitMQVersionInterface {
	return newRabbitMQVersions(c)
}

func (c *CatalogV1alpha1Client) RedisVersions() RedisVersionInterface {
	return newRedisVersions(c)
}

func (c *CatalogV1alpha1Client) SchemaRegistryVersions() SchemaRegistryVersionInterface {
	return newSchemaRegistryVersions(c)
}

func (c *CatalogV1alpha1Client) SinglestoreVersions() SinglestoreVersionInterface {
	return newSinglestoreVersions(c)
}

func (c *CatalogV1alpha1Client) SolrVersions() SolrVersionInterface {
	return newSolrVersions(c)
}

func (c *CatalogV1alpha1Client) ZooKeeperVersions() ZooKeeperVersionInterface {
	return newZooKeeperVersions(c)
}

// NewForConfig creates a new CatalogV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*CatalogV1alpha1Client, error) {
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

// NewForConfigAndClient creates a new CatalogV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*CatalogV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &CatalogV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new CatalogV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *CatalogV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new CatalogV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *CatalogV1alpha1Client {
	return &CatalogV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
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
func (c *CatalogV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
