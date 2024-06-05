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

package v1

import (
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PostgresLister helps list Postgreses.
// All objects returned here must be treated as read-only.
type PostgresLister interface {
	// List lists all Postgreses in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Postgres, err error)
	// Postgreses returns an object that can list and get Postgreses.
	Postgreses(namespace string) PostgresNamespaceLister
	PostgresListerExpansion
}

// postgresLister implements the PostgresLister interface.
type postgresLister struct {
	indexer cache.Indexer
}

// NewPostgresLister returns a new PostgresLister.
func NewPostgresLister(indexer cache.Indexer) PostgresLister {
	return &postgresLister{indexer: indexer}
}

// List lists all Postgreses in the indexer.
func (s *postgresLister) List(selector labels.Selector) (ret []*v1.Postgres, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Postgres))
	})
	return ret, err
}

// Postgreses returns an object that can list and get Postgreses.
func (s *postgresLister) Postgreses(namespace string) PostgresNamespaceLister {
	return postgresNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PostgresNamespaceLister helps list and get Postgreses.
// All objects returned here must be treated as read-only.
type PostgresNamespaceLister interface {
	// List lists all Postgreses in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.Postgres, err error)
	// Get retrieves the Postgres from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.Postgres, error)
	PostgresNamespaceListerExpansion
}

// postgresNamespaceLister implements the PostgresNamespaceLister
// interface.
type postgresNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Postgreses in the indexer for a given namespace.
func (s postgresNamespaceLister) List(selector labels.Selector) (ret []*v1.Postgres, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.Postgres))
	})
	return ret, err
}

// Get retrieves the Postgres from the indexer for a given namespace and name.
func (s postgresNamespaceLister) Get(name string) (*v1.Postgres, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("postgres"), name)
	}
	return obj.(*v1.Postgres), nil
}
