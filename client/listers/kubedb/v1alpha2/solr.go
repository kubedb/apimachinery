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

// SolrLister helps list Solrs.
// All objects returned here must be treated as read-only.
type SolrLister interface {
	// List lists all Solrs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.Solr, err error)
	// Solrs returns an object that can list and get Solrs.
	Solrs(namespace string) SolrNamespaceLister
	SolrListerExpansion
}

// solrLister implements the SolrLister interface.
type solrLister struct {
	indexer cache.Indexer
}

// NewSolrLister returns a new SolrLister.
func NewSolrLister(indexer cache.Indexer) SolrLister {
	return &solrLister{indexer: indexer}
}

// List lists all Solrs in the indexer.
func (s *solrLister) List(selector labels.Selector) (ret []*v1alpha2.Solr, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Solr))
	})
	return ret, err
}

// Solrs returns an object that can list and get Solrs.
func (s *solrLister) Solrs(namespace string) SolrNamespaceLister {
	return solrNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SolrNamespaceLister helps list and get Solrs.
// All objects returned here must be treated as read-only.
type SolrNamespaceLister interface {
	// List lists all Solrs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.Solr, err error)
	// Get retrieves the Solr from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.Solr, error)
	SolrNamespaceListerExpansion
}

// solrNamespaceLister implements the SolrNamespaceLister
// interface.
type solrNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Solrs in the indexer for a given namespace.
func (s solrNamespaceLister) List(selector labels.Selector) (ret []*v1alpha2.Solr, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha2.Solr))
	})
	return ret, err
}

// Get retrieves the Solr from the indexer for a given namespace and name.
func (s solrNamespaceLister) Get(name string) (*v1alpha2.Solr, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha2.Resource("solr"), name)
	}
	return obj.(*v1alpha2.Solr), nil
}
