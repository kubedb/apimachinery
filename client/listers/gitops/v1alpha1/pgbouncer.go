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

// PgBouncerLister helps list PgBouncers.
// All objects returned here must be treated as read-only.
type PgBouncerLister interface {
	// List lists all PgBouncers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PgBouncer, err error)
	// PgBouncers returns an object that can list and get PgBouncers.
	PgBouncers(namespace string) PgBouncerNamespaceLister
	PgBouncerListerExpansion
}

// pgBouncerLister implements the PgBouncerLister interface.
type pgBouncerLister struct {
	indexer cache.Indexer
}

// NewPgBouncerLister returns a new PgBouncerLister.
func NewPgBouncerLister(indexer cache.Indexer) PgBouncerLister {
	return &pgBouncerLister{indexer: indexer}
}

// List lists all PgBouncers in the indexer.
func (s *pgBouncerLister) List(selector labels.Selector) (ret []*v1alpha1.PgBouncer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PgBouncer))
	})
	return ret, err
}

// PgBouncers returns an object that can list and get PgBouncers.
func (s *pgBouncerLister) PgBouncers(namespace string) PgBouncerNamespaceLister {
	return pgBouncerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PgBouncerNamespaceLister helps list and get PgBouncers.
// All objects returned here must be treated as read-only.
type PgBouncerNamespaceLister interface {
	// List lists all PgBouncers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PgBouncer, err error)
	// Get retrieves the PgBouncer from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.PgBouncer, error)
	PgBouncerNamespaceListerExpansion
}

// pgBouncerNamespaceLister implements the PgBouncerNamespaceLister
// interface.
type pgBouncerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PgBouncers in the indexer for a given namespace.
func (s pgBouncerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.PgBouncer, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PgBouncer))
	})
	return ret, err
}

// Get retrieves the PgBouncer from the indexer for a given namespace and name.
func (s pgBouncerNamespaceLister) Get(name string) (*v1alpha1.PgBouncer, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("pgbouncer"), name)
	}
	return obj.(*v1alpha1.PgBouncer), nil
}
