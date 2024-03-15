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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha2

import (
	v1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// KafkaLister helps list Kafkas.
// All objects returned here must be treated as read-only.
type KafkaLister interface {
	// List lists all Kafkas in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.Kafka, err error)
	// Kafkas returns an object that can list and get Kafkas.
	Kafkas(namespace string) KafkaNamespaceLister
	KafkaListerExpansion
}

// kafkaLister implements the KafkaLister interface.
type kafkaLister struct {
	indexer cache.Indexer
}

// NewKafkaLister returns a new KafkaLister.
func NewKafkaLister(indexer cache.Indexer) KafkaLister {
	return &kafkaLister{indexer: indexer}
}

// List lists all Kafkas in the indexer.
func (s *kafkaLister) List(selector labels.Selector) (ret []*v1alpha2.Kafka, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Kafka))
	})
	return ret, err
}

// Kafkas returns an object that can list and get Kafkas.
func (s *kafkaLister) Kafkas(namespace string) KafkaNamespaceLister {
	return kafkaNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// KafkaNamespaceLister helps list and get Kafkas.
// All objects returned here must be treated as read-only.
type KafkaNamespaceLister interface {
	// List lists all Kafkas in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.Kafka, err error)
	// Get retrieves the Kafka from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.Kafka, error)
	KafkaNamespaceListerExpansion
}

// kafkaNamespaceLister implements the KafkaNamespaceLister
// interface.
type kafkaNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Kafkas in the indexer for a given namespace.
func (s kafkaNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.Kafka, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Kafka))
	})
	return ret, err
}

// Get retrieves the Kafka from the indexer for a given namespace and name.
func (s kafkaNamespaceLister) Get(name string) (*v1alpha2.Kafka, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("kafka"), name)
	}
	return obj.(*v1alpha2.Kafka), nil
}
