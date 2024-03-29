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

// MongoDBDatabaseLister helps list MongoDBDatabases.
// All objects returned here must be treated as read-only.
type MongoDBDatabaseLister interface {
	// List lists all MongoDBDatabases in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MongoDBDatabase, err error)
	// MongoDBDatabases returns an object that can list and get MongoDBDatabases.
	MongoDBDatabases(namespace string) MongoDBDatabaseNamespaceLister
	MongoDBDatabaseListerExpansion
}

// mongoDBDatabaseLister implements the MongoDBDatabaseLister interface.
type mongoDBDatabaseLister struct {
	indexer cache.Indexer
}

// NewMongoDBDatabaseLister returns a new MongoDBDatabaseLister.
func NewMongoDBDatabaseLister(indexer cache.Indexer) MongoDBDatabaseLister {
	return &mongoDBDatabaseLister{indexer: indexer}
}

// List lists all MongoDBDatabases in the indexer.
func (s *mongoDBDatabaseLister) List(selector labels.Selector) (ret []*v1alpha1.MongoDBDatabase, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MongoDBDatabase))
	})
	return ret, err
}

// MongoDBDatabases returns an object that can list and get MongoDBDatabases.
func (s *mongoDBDatabaseLister) MongoDBDatabases(namespace string) MongoDBDatabaseNamespaceLister {
	return mongoDBDatabaseNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MongoDBDatabaseNamespaceLister helps list and get MongoDBDatabases.
// All objects returned here must be treated as read-only.
type MongoDBDatabaseNamespaceLister interface {
	// List lists all MongoDBDatabases in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MongoDBDatabase, err error)
	// Get retrieves the MongoDBDatabase from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MongoDBDatabase, error)
	MongoDBDatabaseNamespaceListerExpansion
}

// mongoDBDatabaseNamespaceLister implements the MongoDBDatabaseNamespaceLister
// interface.
type mongoDBDatabaseNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MongoDBDatabases in the indexer for a given namespace.
func (s mongoDBDatabaseNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MongoDBDatabase, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MongoDBDatabase))
	})
	return ret, err
}

// Get retrieves the MongoDBDatabase from the indexer for a given namespace and name.
func (s mongoDBDatabaseNamespaceLister) Get(name string) (*v1alpha1.MongoDBDatabase, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mongodbdatabase"), name)
	}
	return obj.(*v1alpha1.MongoDBDatabase), nil
}
