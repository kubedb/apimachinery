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
	v1alpha1 "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
)

// MongoDBArchiverLister helps list MongoDBArchivers.
// All objects returned here must be treated as read-only.
type MongoDBArchiverLister interface {
	// List lists all MongoDBArchivers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MongoDBArchiver, err error)
	// MongoDBArchivers returns an object that can list and get MongoDBArchivers.
	MongoDBArchivers(namespace string) MongoDBArchiverNamespaceLister
	MongoDBArchiverListerExpansion
}

// mongoDBArchiverLister implements the MongoDBArchiverLister interface.
type mongoDBArchiverLister struct {
	indexer cache.Indexer
}

// NewMongoDBArchiverLister returns a new MongoDBArchiverLister.
func NewMongoDBArchiverLister(indexer cache.Indexer) MongoDBArchiverLister {
	return &mongoDBArchiverLister{indexer: indexer}
}

// List lists all MongoDBArchivers in the indexer.
func (s *mongoDBArchiverLister) List(selector labels.Selector) (ret []*v1alpha1.MongoDBArchiver, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MongoDBArchiver))
	})
	return ret, err
}

// MongoDBArchivers returns an object that can list and get MongoDBArchivers.
func (s *mongoDBArchiverLister) MongoDBArchivers(namespace string) MongoDBArchiverNamespaceLister {
	return mongoDBArchiverNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MongoDBArchiverNamespaceLister helps list and get MongoDBArchivers.
// All objects returned here must be treated as read-only.
type MongoDBArchiverNamespaceLister interface {
	// List lists all MongoDBArchivers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MongoDBArchiver, err error)
	// Get retrieves the MongoDBArchiver from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MongoDBArchiver, error)
	MongoDBArchiverNamespaceListerExpansion
}

// mongoDBArchiverNamespaceLister implements the MongoDBArchiverNamespaceLister
// interface.
type mongoDBArchiverNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MongoDBArchivers in the indexer for a given namespace.
func (s mongoDBArchiverNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MongoDBArchiver, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MongoDBArchiver))
	})
	return ret, err
}

// Get retrieves the MongoDBArchiver from the indexer for a given namespace and name.
func (s mongoDBArchiverNamespaceLister) Get(name string) (*v1alpha1.MongoDBArchiver, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mongodbarchiver"), name)
	}
	return obj.(*v1alpha1.MongoDBArchiver), nil
}
