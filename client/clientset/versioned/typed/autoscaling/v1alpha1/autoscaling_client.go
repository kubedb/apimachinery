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

	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	rest "k8s.io/client-go/rest"
)

type AutoscalingV1alpha1Interface interface {
	RESTClient() rest.Interface
	ElasticsearchAutoscalersGetter
	EtcdAutoscalersGetter
	KafkaAutoscalersGetter
	MariaDBAutoscalersGetter
	MemcachedAutoscalersGetter
	MongoDBAutoscalersGetter
	MySQLAutoscalersGetter
	PerconaXtraDBAutoscalersGetter
	PgBouncerAutoscalersGetter
	PgpoolAutoscalersGetter
	PostgresAutoscalersGetter
	ProxySQLAutoscalersGetter
	RabbitMQAutoscalersGetter
	RedisAutoscalersGetter
	RedisSentinelAutoscalersGetter
	SinglestoreAutoscalersGetter
}

// AutoscalingV1alpha1Client is used to interact with features provided by the autoscaling.kubedb.com group.
type AutoscalingV1alpha1Client struct {
	restClient rest.Interface
}

func (c *AutoscalingV1alpha1Client) ElasticsearchAutoscalers(namespace string) ElasticsearchAutoscalerInterface {
	return newElasticsearchAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) EtcdAutoscalers(namespace string) EtcdAutoscalerInterface {
	return newEtcdAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) KafkaAutoscalers(namespace string) KafkaAutoscalerInterface {
	return newKafkaAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) MariaDBAutoscalers(namespace string) MariaDBAutoscalerInterface {
	return newMariaDBAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) MemcachedAutoscalers(namespace string) MemcachedAutoscalerInterface {
	return newMemcachedAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) MongoDBAutoscalers(namespace string) MongoDBAutoscalerInterface {
	return newMongoDBAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) MySQLAutoscalers(namespace string) MySQLAutoscalerInterface {
	return newMySQLAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) PerconaXtraDBAutoscalers(namespace string) PerconaXtraDBAutoscalerInterface {
	return newPerconaXtraDBAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) PgBouncerAutoscalers(namespace string) PgBouncerAutoscalerInterface {
	return newPgBouncerAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) PgpoolAutoscalers(namespace string) PgpoolAutoscalerInterface {
	return newPgpoolAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) PostgresAutoscalers(namespace string) PostgresAutoscalerInterface {
	return newPostgresAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) ProxySQLAutoscalers(namespace string) ProxySQLAutoscalerInterface {
	return newProxySQLAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) RabbitMQAutoscalers(namespace string) RabbitMQAutoscalerInterface {
	return newRabbitMQAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) RedisAutoscalers(namespace string) RedisAutoscalerInterface {
	return newRedisAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) RedisSentinelAutoscalers(namespace string) RedisSentinelAutoscalerInterface {
	return newRedisSentinelAutoscalers(c, namespace)
}

func (c *AutoscalingV1alpha1Client) SinglestoreAutoscalers(namespace string) SinglestoreAutoscalerInterface {
	return newSinglestoreAutoscalers(c, namespace)
}

// NewForConfig creates a new AutoscalingV1alpha1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*AutoscalingV1alpha1Client, error) {
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

// NewForConfigAndClient creates a new AutoscalingV1alpha1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*AutoscalingV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &AutoscalingV1alpha1Client{client}, nil
}

// NewForConfigOrDie creates a new AutoscalingV1alpha1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *AutoscalingV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new AutoscalingV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *AutoscalingV1alpha1Client {
	return &AutoscalingV1alpha1Client{c}
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
func (c *AutoscalingV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
