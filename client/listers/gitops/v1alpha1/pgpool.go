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

// PgpoolLister helps list Pgpools.
// All objects returned here must be treated as read-only.
type PgpoolLister interface {
	// List lists all Pgpools in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Pgpool, err error)
	// Pgpools returns an object that can list and get Pgpools.
	Pgpools(namespace string) PgpoolNamespaceLister
	PgpoolListerExpansion
}

// pgpoolLister implements the PgpoolLister interface.
type pgpoolLister struct {
	indexer cache.Indexer
}

// NewPgpoolLister returns a new PgpoolLister.
func NewPgpoolLister(indexer cache.Indexer) PgpoolLister {
	return &pgpoolLister{indexer: indexer}
}

// List lists all Pgpools in the indexer.
func (s *pgpoolLister) List(selector labels.Selector) (ret []*v1alpha1.Pgpool, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Pgpool))
	})
	return ret, err
}

// Pgpools returns an object that can list and get Pgpools.
func (s *pgpoolLister) Pgpools(namespace string) PgpoolNamespaceLister {
	return pgpoolNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PgpoolNamespaceLister helps list and get Pgpools.
// All objects returned here must be treated as read-only.
type PgpoolNamespaceLister interface {
	// List lists all Pgpools in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Pgpool, err error)
	// Get retrieves the Pgpool from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Pgpool, error)
	PgpoolNamespaceListerExpansion
}

// pgpoolNamespaceLister implements the PgpoolNamespaceLister
// interface.
type pgpoolNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Pgpools in the indexer for a given namespace.
func (s pgpoolNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Pgpool, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Pgpool))
	})
	return ret, err
}

// Get retrieves the Pgpool from the indexer for a given namespace and name.
func (s pgpoolNamespaceLister) Get(name string) (*v1alpha1.Pgpool, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("pgpool"), name)
	}
	return obj.(*v1alpha1.Pgpool), nil
}
