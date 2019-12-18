/*
Copyright The KubeDB Authors.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

// ProxySQLLister helps list ProxySQLs.
type ProxySQLLister interface {
	// List lists all ProxySQLs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.ProxySQL, err error)
	// ProxySQLs returns an object that can list and get ProxySQLs.
	ProxySQLs(namespace string) ProxySQLNamespaceLister
	ProxySQLListerExpansion
}

// proxySQLLister implements the ProxySQLLister interface.
type proxySQLLister struct {
	indexer cache.Indexer
}

// NewProxySQLLister returns a new ProxySQLLister.
func NewProxySQLLister(indexer cache.Indexer) ProxySQLLister {
	return &proxySQLLister{indexer: indexer}
}

// List lists all ProxySQLs in the indexer.
func (s *proxySQLLister) List(selector labels.Selector) (ret []*v1alpha1.ProxySQL, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ProxySQL))
	})
	return ret, err
}

// ProxySQLs returns an object that can list and get ProxySQLs.
func (s *proxySQLLister) ProxySQLs(namespace string) ProxySQLNamespaceLister {
	return proxySQLNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ProxySQLNamespaceLister helps list and get ProxySQLs.
type ProxySQLNamespaceLister interface {
	// List lists all ProxySQLs in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.ProxySQL, err error)
	// Get retrieves the ProxySQL from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.ProxySQL, error)
	ProxySQLNamespaceListerExpansion
}

// proxySQLNamespaceLister implements the ProxySQLNamespaceLister
// interface.
type proxySQLNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ProxySQLs in the indexer for a given namespace.
func (s proxySQLNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ProxySQL, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ProxySQL))
	})
	return ret, err
}

// Get retrieves the ProxySQL from the indexer for a given namespace and name.
func (s proxySQLNamespaceLister) Get(name string) (*v1alpha1.ProxySQL, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("proxysql"), name)
	}
	return obj.(*v1alpha1.ProxySQL), nil
}
