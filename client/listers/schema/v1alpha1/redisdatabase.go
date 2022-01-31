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

package v1alpha1

import (
	v1alpha1 "kubedb.dev/apimachinery/apis/schema/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// RedisDatabaseLister helps list RedisDatabases.
// All objects returned here must be treated as read-only.
type RedisDatabaseLister interface {
	// List lists all RedisDatabases in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.RedisDatabase, err error)
	// RedisDatabases returns an object that can list and get RedisDatabases.
	RedisDatabases(namespace string) RedisDatabaseNamespaceLister
	RedisDatabaseListerExpansion
}

// redisDatabaseLister implements the RedisDatabaseLister interface.
type redisDatabaseLister struct {
	indexer cache.Indexer
}

// NewRedisDatabaseLister returns a new RedisDatabaseLister.
func NewRedisDatabaseLister(indexer cache.Indexer) RedisDatabaseLister {
	return &redisDatabaseLister{indexer: indexer}
}

// List lists all RedisDatabases in the indexer.
func (s *redisDatabaseLister) List(selector labels.Selector) (ret []*v1alpha1.RedisDatabase, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.RedisDatabase))
	})
	return ret, err
}

// RedisDatabases returns an object that can list and get RedisDatabases.
func (s *redisDatabaseLister) RedisDatabases(namespace string) RedisDatabaseNamespaceLister {
	return redisDatabaseNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// RedisDatabaseNamespaceLister helps list and get RedisDatabases.
// All objects returned here must be treated as read-only.
type RedisDatabaseNamespaceLister interface {
	// List lists all RedisDatabases in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.RedisDatabase, err error)
	// Get retrieves the RedisDatabase from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.RedisDatabase, error)
	RedisDatabaseNamespaceListerExpansion
}

// redisDatabaseNamespaceLister implements the RedisDatabaseNamespaceLister
// interface.
type redisDatabaseNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all RedisDatabases in the indexer for a given namespace.
func (s redisDatabaseNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.RedisDatabase, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.RedisDatabase))
	})
	return ret, err
}

// Get retrieves the RedisDatabase from the indexer for a given namespace and name.
func (s redisDatabaseNamespaceLister) Get(name string) (*v1alpha1.RedisDatabase, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("redisdatabase"), name)
	}
	return obj.(*v1alpha1.RedisDatabase), nil
}