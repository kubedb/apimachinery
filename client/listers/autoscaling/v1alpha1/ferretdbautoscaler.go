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

// FerretDBAutoscalerLister helps list FerretDBAutoscalers.
// All objects returned here must be treated as read-only.
type FerretDBAutoscalerLister interface {
	// List lists all FerretDBAutoscalers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.FerretDBAutoscaler, err error)
	// FerretDBAutoscalers returns an object that can list and get FerretDBAutoscalers.
	FerretDBAutoscalers(namespace string) FerretDBAutoscalerNamespaceLister
	FerretDBAutoscalerListerExpansion
}

// ferretDBAutoscalerLister implements the FerretDBAutoscalerLister interface.
type ferretDBAutoscalerLister struct {
	indexer cache.Indexer
}

// NewFerretDBAutoscalerLister returns a new FerretDBAutoscalerLister.
func NewFerretDBAutoscalerLister(indexer cache.Indexer) FerretDBAutoscalerLister {
	return &ferretDBAutoscalerLister{indexer: indexer}
}

// List lists all FerretDBAutoscalers in the indexer.
func (s *ferretDBAutoscalerLister) List(selector labels.Selector) (ret []*v1alpha1.FerretDBAutoscaler, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.FerretDBAutoscaler))
	})
	return ret, err
}

// FerretDBAutoscalers returns an object that can list and get FerretDBAutoscalers.
func (s *ferretDBAutoscalerLister) FerretDBAutoscalers(namespace string) FerretDBAutoscalerNamespaceLister {
	return ferretDBAutoscalerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// FerretDBAutoscalerNamespaceLister helps list and get FerretDBAutoscalers.
// All objects returned here must be treated as read-only.
type FerretDBAutoscalerNamespaceLister interface {
	// List lists all FerretDBAutoscalers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.FerretDBAutoscaler, err error)
	// Get retrieves the FerretDBAutoscaler from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.FerretDBAutoscaler, error)
	FerretDBAutoscalerNamespaceListerExpansion
}

// ferretDBAutoscalerNamespaceLister implements the FerretDBAutoscalerNamespaceLister
// interface.
type ferretDBAutoscalerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all FerretDBAutoscalers in the indexer for a given namespace.
func (s ferretDBAutoscalerNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.FerretDBAutoscaler, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.FerretDBAutoscaler))
	})
	return ret, err
}

// Get retrieves the FerretDBAutoscaler from the indexer for a given namespace and name.
func (s ferretDBAutoscalerNamespaceLister) Get(name string) (*v1alpha1.FerretDBAutoscaler, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("ferretdbautoscaler"), name)
	}
	return obj.(*v1alpha1.FerretDBAutoscaler), nil
}
