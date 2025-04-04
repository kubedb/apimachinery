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
	v1alpha1 "kubedb.dev/apimachinery/apis/gitops/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MemcachedLister helps list Memcacheds.
// All objects returned here must be treated as read-only.
type MemcachedLister interface {
	// List lists all Memcacheds in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Memcached, err error)
	// Memcacheds returns an object that can list and get Memcacheds.
	Memcacheds(namespace string) MemcachedNamespaceLister
	MemcachedListerExpansion
}

// memcachedLister implements the MemcachedLister interface.
type memcachedLister struct {
	indexer cache.Indexer
}

// NewMemcachedLister returns a new MemcachedLister.
func NewMemcachedLister(indexer cache.Indexer) MemcachedLister {
	return &memcachedLister{indexer: indexer}
}

// List lists all Memcacheds in the indexer.
func (s *memcachedLister) List(selector labels.Selector) (ret []*v1alpha1.Memcached, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Memcached))
	})
	return ret, err
}

// Memcacheds returns an object that can list and get Memcacheds.
func (s *memcachedLister) Memcacheds(namespace string) MemcachedNamespaceLister {
	return memcachedNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MemcachedNamespaceLister helps list and get Memcacheds.
// All objects returned here must be treated as read-only.
type MemcachedNamespaceLister interface {
	// List lists all Memcacheds in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Memcached, err error)
	// Get retrieves the Memcached from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Memcached, error)
	MemcachedNamespaceListerExpansion
}

// memcachedNamespaceLister implements the MemcachedNamespaceLister
// interface.
type memcachedNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Memcacheds in the indexer for a given namespace.
func (s memcachedNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Memcached, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Memcached))
	})
	return ret, err
}

// Get retrieves the Memcached from the indexer for a given namespace and name.
func (s memcachedNamespaceLister) Get(name string) (*v1alpha1.Memcached, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("memcached"), name)
	}
	return obj.(*v1alpha1.Memcached), nil
}
