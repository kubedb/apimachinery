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
	v1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SolrOpsRequestLister helps list SolrOpsRequests.
// All objects returned here must be treated as read-only.
type SolrOpsRequestLister interface {
	// List lists all SolrOpsRequests in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SolrOpsRequest, err error)
	// SolrOpsRequests returns an object that can list and get SolrOpsRequests.
	SolrOpsRequests(namespace string) SolrOpsRequestNamespaceLister
	SolrOpsRequestListerExpansion
}

// solrOpsRequestLister implements the SolrOpsRequestLister interface.
type solrOpsRequestLister struct {
	indexer cache.Indexer
}

// NewSolrOpsRequestLister returns a new SolrOpsRequestLister.
func NewSolrOpsRequestLister(indexer cache.Indexer) SolrOpsRequestLister {
	return &solrOpsRequestLister{indexer: indexer}
}

// List lists all SolrOpsRequests in the indexer.
func (s *solrOpsRequestLister) List(selector labels.Selector) (ret []*v1alpha1.SolrOpsRequest, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SolrOpsRequest))
	})
	return ret, err
}

// SolrOpsRequests returns an object that can list and get SolrOpsRequests.
func (s *solrOpsRequestLister) SolrOpsRequests(namespace string) SolrOpsRequestNamespaceLister {
	return solrOpsRequestNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SolrOpsRequestNamespaceLister helps list and get SolrOpsRequests.
// All objects returned here must be treated as read-only.
type SolrOpsRequestNamespaceLister interface {
	// List lists all SolrOpsRequests in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SolrOpsRequest, err error)
	// Get retrieves the SolrOpsRequest from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SolrOpsRequest, error)
	SolrOpsRequestNamespaceListerExpansion
}

// solrOpsRequestNamespaceLister implements the SolrOpsRequestNamespaceLister
// interface.
type solrOpsRequestNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SolrOpsRequests in the indexer for a given namespace.
func (s solrOpsRequestNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SolrOpsRequest, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SolrOpsRequest))
	})
	return ret, err
}

// Get retrieves the SolrOpsRequest from the indexer for a given namespace and name.
func (s solrOpsRequestNamespaceLister) Get(name string) (*v1alpha1.SolrOpsRequest, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("solropsrequest"), name)
	}
	return obj.(*v1alpha1.SolrOpsRequest), nil
}
