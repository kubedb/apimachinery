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

// MariaDBLister helps list MariaDBs.
type MariaDBLister interface {
	// List lists all MariaDBs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha2.MariaDB, err error)
	// MariaDBs returns an object that can list and get MariaDBs.
	MariaDBs(namespace string) MariaDBNamespaceLister
	MariaDBListerExpansion
}

// mariaDBLister implements the MariaDBLister interface.
type mariaDBLister struct {
	indexer cache.Indexer
}

// NewMariaDBLister returns a new MariaDBLister.
func NewMariaDBLister(indexer cache.Indexer) MariaDBLister {
	return &mariaDBLister{indexer: indexer}
}

// List lists all MariaDBs in the indexer.
func (s *mariaDBLister) List(selector labels.Selector) (ret []*v1alpha2.MariaDB, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.MariaDB))
	})
	return ret, err
}

// MariaDBs returns an object that can list and get MariaDBs.
func (s *mariaDBLister) MariaDBs(namespace string) MariaDBNamespaceLister {
	return mariaDBNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MariaDBNamespaceLister helps list and get MariaDBs.
type MariaDBNamespaceLister interface {
	// List lists all MariaDBs in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha2.MariaDB, err error)
	// Get retrieves the MariaDB from the indexer for a given namespace and name.
	Get(name string) (*v1alpha2.MariaDB, error)
	MariaDBNamespaceListerExpansion
}

// mariaDBNamespaceLister implements the MariaDBNamespaceLister
// interface.
type mariaDBNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MariaDBs in the indexer for a given namespace.
func (s mariaDBNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.MariaDB, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.MariaDB))
	})
	return ret, err
}

// Get retrieves the MariaDB from the indexer for a given namespace and name.
func (s mariaDBNamespaceLister) Get(name string) (*v1alpha2.MariaDB, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("mariadb"), name)
	}
	return obj.(*v1alpha2.MariaDB), nil
}
