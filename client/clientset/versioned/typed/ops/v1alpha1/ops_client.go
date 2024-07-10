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

	rest "k8s.io/client-go/rest"
	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"
)

type OpsV1alpha1Interface interface {
	RESTClient() rest.Interface
	DruidOpsRequestsGetter
	ElasticsearchOpsRequestsGetter
	EtcdOpsRequestsGetter
	KafkaOpsRequestsGetter
	MariaDBOpsRequestsGetter
	MemcachedOpsRequestsGetter
	MongoDBOpsRequestsGetter
	MySQLOpsRequestsGetter
	PerconaXtraDBOpsRequestsGetter
	PgBouncerOpsRequestsGetter
	PgpoolOpsRequestsGetter
	PostgresOpsRequestsGetter
	ProxySQLOpsRequestsGetter
	RabbitMQOpsRequestsGetter
	RedisOpsRequestsGetter
	RedisSentinelOpsRequestsGetter
	SinglestoreOpsRequestsGetter
	SolrOpsRequestsGetter
}

// OpsV1alpha1Client is used to interact with features provided by the ops.kubedb.com group.
type OpsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *OpsV1alpha1Client) DruidOpsRequests(namespace string) DruidOpsRequestInterface {
	return newDruidOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) ElasticsearchOpsRequests(namespace string) ElasticsearchOpsRequestInterface {
	return newElasticsearchOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) EtcdOpsRequests(namespace string) EtcdOpsRequestInterface {
	return newEtcdOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) KafkaOpsRequests(namespace string) KafkaOpsRequestInterface {
	return newKafkaOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) MariaDBOpsRequests(namespace string) MariaDBOpsRequestInterface {
	return newMariaDBOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) MemcachedOpsRequests(namespace string) MemcachedOpsRequestInterface {
	return newMemcachedOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) MongoDBOpsRequests(namespace string) MongoDBOpsRequestInterface {
	return newMongoDBOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) MySQLOpsRequests(namespace string) MySQLOpsRequestInterface {
	return newMySQLOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) PerconaXtraDBOpsRequests(namespace string) PerconaXtraDBOpsRequestInterface {
	return newPerconaXtraDBOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) PgBouncerOpsRequests(namespace string) PgBouncerOpsRequestInterface {
	return newPgBouncerOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) PgpoolOpsRequests(namespace string) PgpoolOpsRequestInterface {
	return newPgpoolOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) PostgresOpsRequests(namespace string) PostgresOpsRequestInterface {
	return newPostgresOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) ProxySQLOpsRequests(namespace string) ProxySQLOpsRequestInterface {
	return newProxySQLOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) RabbitMQOpsRequests(namespace string) RabbitMQOpsRequestInterface {
	return newRabbitMQOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) RedisOpsRequests(namespace string) RedisOpsRequestInterface {
	return newRedisOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) RedisSentinelOpsRequests(namespace string) RedisSentinelOpsRequestInterface {
	return newRedisSentinelOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) SinglestoreOpsRequests(namespace string) SinglestoreOpsRequestInterface {
	return newSinglestoreOpsRequests(c, namespace)
}

func (c *OpsV1alpha1Client) SolrOpsRequests(namespace string) SolrOpsRequestInterface {
	return newSolrOpsRequests(c, namespace)
}

// NewForConfig creates a new OpsV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*OpsV1alpha1Client, error) {
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

// NewForConfigAndClient creates a new OpsV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*OpsV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &OpsV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new OpsV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *OpsV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new OpsV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *OpsV1alpha1Client {
	return &OpsV1alpha1Client{c}
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
func (c *OpsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
