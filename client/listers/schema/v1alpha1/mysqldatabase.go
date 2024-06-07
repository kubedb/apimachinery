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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	v1alpha1 "kubedb.dev/apimachinery/apis/schema/v1alpha1"
)

// MySQLDatabaseLister helps list MySQLDatabases.
// All objects returned here must be treated as read-only.
type MySQLDatabaseLister interface {
	// List lists all MySQLDatabases in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MySQLDatabase, err error)
	// MySQLDatabases returns an object that can list and get MySQLDatabases.
	MySQLDatabases(namespace string) MySQLDatabaseNamespaceLister
	MySQLDatabaseListerExpansion
}

// mySQLDatabaseLister implements the MySQLDatabaseLister interface.
type mySQLDatabaseLister struct {
	indexer cache.Indexer
}

// NewMySQLDatabaseLister returns a new MySQLDatabaseLister.
func NewMySQLDatabaseLister(indexer cache.Indexer) MySQLDatabaseLister {
	return &mySQLDatabaseLister{indexer: indexer}
}

// List lists all MySQLDatabases in the indexer.
func (s *mySQLDatabaseLister) List(selector labels.Selector) (ret []*v1alpha1.MySQLDatabase, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MySQLDatabase))
	})
	return ret, err
}

// MySQLDatabases returns an object that can list and get MySQLDatabases.
func (s *mySQLDatabaseLister) MySQLDatabases(namespace string) MySQLDatabaseNamespaceLister {
	return mySQLDatabaseNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MySQLDatabaseNamespaceLister helps list and get MySQLDatabases.
// All objects returned here must be treated as read-only.
type MySQLDatabaseNamespaceLister interface {
	// List lists all MySQLDatabases in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MySQLDatabase, err error)
	// Get retrieves the MySQLDatabase from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MySQLDatabase, error)
	MySQLDatabaseNamespaceListerExpansion
}

// mySQLDatabaseNamespaceLister implements the MySQLDatabaseNamespaceLister
// interface.
type mySQLDatabaseNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MySQLDatabases in the indexer for a given namespace.
func (s mySQLDatabaseNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MySQLDatabase, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MySQLDatabase))
	})
	return ret, err
}

// Get retrieves the MySQLDatabase from the indexer for a given namespace and name.
func (s mySQLDatabaseNamespaceLister) Get(name string) (*v1alpha1.MySQLDatabase, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mysqldatabase"), name)
	}
	return obj.(*v1alpha1.MySQLDatabase), nil
}
