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

// PostgresArchiverLister helps list PostgresArchivers.
// All objects returned here must be treated as read-only.
type PostgresArchiverLister interface {
	// List lists all PostgresArchivers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PostgresArchiver, err error)
	// PostgresArchivers returns an object that can list and get PostgresArchivers.
	PostgresArchivers(namespace string) PostgresArchiverNamespaceLister
	PostgresArchiverListerExpansion
}

// postgresArchiverLister implements the PostgresArchiverLister interface.
type postgresArchiverLister struct {
	indexer cache.Indexer
}

// NewPostgresArchiverLister returns a new PostgresArchiverLister.
func NewPostgresArchiverLister(indexer cache.Indexer) PostgresArchiverLister {
	return &postgresArchiverLister{indexer: indexer}
}

// List lists all PostgresArchivers in the indexer.
func (s *postgresArchiverLister) List(selector labels.Selector) (ret []*v1alpha1.PostgresArchiver, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PostgresArchiver))
	})
	return ret, err
}

// PostgresArchivers returns an object that can list and get PostgresArchivers.
func (s *postgresArchiverLister) PostgresArchivers(namespace string) PostgresArchiverNamespaceLister {
	return postgresArchiverNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PostgresArchiverNamespaceLister helps list and get PostgresArchivers.
// All objects returned here must be treated as read-only.
type PostgresArchiverNamespaceLister interface {
	// List lists all PostgresArchivers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PostgresArchiver, err error)
	// Get retrieves the PostgresArchiver from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.PostgresArchiver, error)
	PostgresArchiverNamespaceListerExpansion
}

// postgresArchiverNamespaceLister implements the PostgresArchiverNamespaceLister
// interface.
type postgresArchiverNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PostgresArchivers in the indexer for a given namespace.
func (s postgresArchiverNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.PostgresArchiver, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PostgresArchiver))
	})
	return ret, err
}

// Get retrieves the PostgresArchiver from the indexer for a given namespace and name.
func (s postgresArchiverNamespaceLister) Get(name string) (*v1alpha1.PostgresArchiver, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("postgresarchiver"), name)
	}
	return obj.(*v1alpha1.PostgresArchiver), nil
}
