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
	v1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// SolrAutoscalerLister helps list SolrAutoscalers.
// All objects returned here must be treated as read-only.
type SolrAutoscalerLister interface {
	// List lists all SolrAutoscalers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SolrAutoscaler, err error)
	// SolrAutoscalers returns an object that can list and get SolrAutoscalers.
	SolrAutoscalers(namespace string) SolrAutoscalerNamespaceLister
	SolrAutoscalerListerExpansion
}

// solrAutoscalerLister implements the SolrAutoscalerLister interface.
type solrAutoscalerLister struct {
	indexer cache.Indexer
}

// NewSolrAutoscalerLister returns a new SolrAutoscalerLister.
func NewSolrAutoscalerLister(indexer cache.Indexer) SolrAutoscalerLister {
	return &solrAutoscalerLister{indexer: indexer}
}

// List lists all SolrAutoscalers in the indexer.
func (s *solrAutoscalerLister) List(selector labels.Selector) (ret []*v1alpha1.SolrAutoscaler, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SolrAutoscaler))
	})
	return ret, err
}

// SolrAutoscalers returns an object that can list and get SolrAutoscalers.
func (s *solrAutoscalerLister) SolrAutoscalers(namespace string) SolrAutoscalerNamespaceLister {
	return solrAutoscalerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// SolrAutoscalerNamespaceLister helps list and get SolrAutoscalers.
// All objects returned here must be treated as read-only.
type SolrAutoscalerNamespaceLister interface {
	// List lists all SolrAutoscalers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.SolrAutoscaler, err error)
	// Get retrieves the SolrAutoscaler from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.SolrAutoscaler, error)
	SolrAutoscalerNamespaceListerExpansion
}

// solrAutoscalerNamespaceLister implements the SolrAutoscalerNamespaceLister
// interface.
type solrAutoscalerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all SolrAutoscalers in the indexer for a given namespace.
func (s solrAutoscalerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.SolrAutoscaler, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.SolrAutoscaler))
	})
	return ret, err
}

// Get retrieves the SolrAutoscaler from the indexer for a given namespace and name.
func (s solrAutoscalerNamespaceLister) Get(name string) (*v1alpha1.SolrAutoscaler, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("solrautoscaler"), name)
	}
	return obj.(*v1alpha1.SolrAutoscaler), nil
}
