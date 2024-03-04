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
	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MemcachedAutoscalerLister helps list MemcachedAutoscalers.
// All objects returned here must be treated as read-only.
type MemcachedAutoscalerLister interface {
	// List lists all MemcachedAutoscalers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MemcachedAutoscaler, err error)
	// MemcachedAutoscalers returns an object that can list and get MemcachedAutoscalers.
	MemcachedAutoscalers(namespace string) MemcachedAutoscalerNamespaceLister
	MemcachedAutoscalerListerExpansion
}

// memcachedAutoscalerLister implements the MemcachedAutoscalerLister interface.
type memcachedAutoscalerLister struct {
	indexer cache.Indexer
}

// NewMemcachedAutoscalerLister returns a new MemcachedAutoscalerLister.
func NewMemcachedAutoscalerLister(indexer cache.Indexer) MemcachedAutoscalerLister {
	return &memcachedAutoscalerLister{indexer: indexer}
}

// List lists all MemcachedAutoscalers in the indexer.
func (s *memcachedAutoscalerLister) List(selector labels.Selector) (ret []*v1alpha1.MemcachedAutoscaler, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MemcachedAutoscaler))
	})
	return ret, err
}

// MemcachedAutoscalers returns an object that can list and get MemcachedAutoscalers.
func (s *memcachedAutoscalerLister) MemcachedAutoscalers(namespace string) MemcachedAutoscalerNamespaceLister {
	return memcachedAutoscalerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MemcachedAutoscalerNamespaceLister helps list and get MemcachedAutoscalers.
// All objects returned here must be treated as read-only.
type MemcachedAutoscalerNamespaceLister interface {
	// List lists all MemcachedAutoscalers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MemcachedAutoscaler, err error)
	// Get retrieves the MemcachedAutoscaler from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MemcachedAutoscaler, error)
	MemcachedAutoscalerNamespaceListerExpansion
}

// memcachedAutoscalerNamespaceLister implements the MemcachedAutoscalerNamespaceLister
// interface.
type memcachedAutoscalerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MemcachedAutoscalers in the indexer for a given namespace.
func (s memcachedAutoscalerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MemcachedAutoscaler, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MemcachedAutoscaler))
	})
	return ret, err
}

// Get retrieves the MemcachedAutoscaler from the indexer for a given namespace and name.
func (s memcachedAutoscalerNamespaceLister) Get(name string) (*v1alpha1.MemcachedAutoscaler, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("memcachedautoscaler"), name)
	}
	return obj.(*v1alpha1.MemcachedAutoscaler), nil
}
