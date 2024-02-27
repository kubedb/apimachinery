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
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
	v1alpha1 "kubedb.dev/apimachinery/client/clientset/versioned/typed/autoscaling/v1alpha1"
)

type FakeAutoscalingV1alpha1 struct {
	*testing.Fake
}

func (c *FakeAutoscalingV1alpha1) ElasticsearchAutoscalers(namespace string) v1alpha1.ElasticsearchAutoscalerInterface {
	return &FakeElasticsearchAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) EtcdAutoscalers(namespace string) v1alpha1.EtcdAutoscalerInterface {
	return &FakeEtcdAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) MariaDBAutoscalers(namespace string) v1alpha1.MariaDBAutoscalerInterface {
	return &FakeMariaDBAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) MemcachedAutoscalers(namespace string) v1alpha1.MemcachedAutoscalerInterface {
	return &FakeMemcachedAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) MongoDBAutoscalers(namespace string) v1alpha1.MongoDBAutoscalerInterface {
	return &FakeMongoDBAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) MySQLAutoscalers(namespace string) v1alpha1.MySQLAutoscalerInterface {
	return &FakeMySQLAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) PerconaXtraDBAutoscalers(namespace string) v1alpha1.PerconaXtraDBAutoscalerInterface {
	return &FakePerconaXtraDBAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) PgBouncerAutoscalers(namespace string) v1alpha1.PgBouncerAutoscalerInterface {
	return &FakePgBouncerAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) PostgresAutoscalers(namespace string) v1alpha1.PostgresAutoscalerInterface {
	return &FakePostgresAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) ProxySQLAutoscalers(namespace string) v1alpha1.ProxySQLAutoscalerInterface {
	return &FakeProxySQLAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) RedisAutoscalers(namespace string) v1alpha1.RedisAutoscalerInterface {
	return &FakeRedisAutoscalers{c, namespace}
}

func (c *FakeAutoscalingV1alpha1) RedisSentinelAutoscalers(namespace string) v1alpha1.RedisSentinelAutoscalerInterface {
	return &FakeRedisSentinelAutoscalers{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAutoscalingV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
