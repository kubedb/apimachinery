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

// MSSQLServerArchiverLister helps list MSSQLServerArchivers.
// All objects returned here must be treated as read-only.
type MSSQLServerArchiverLister interface {
	// List lists all MSSQLServerArchivers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MSSQLServerArchiver, err error)
	// MSSQLServerArchivers returns an object that can list and get MSSQLServerArchivers.
	MSSQLServerArchivers(namespace string) MSSQLServerArchiverNamespaceLister
	MSSQLServerArchiverListerExpansion
}

// mSSQLServerArchiverLister implements the MSSQLServerArchiverLister interface.
type mSSQLServerArchiverLister struct {
	indexer cache.Indexer
}

// NewMSSQLServerArchiverLister returns a new MSSQLServerArchiverLister.
func NewMSSQLServerArchiverLister(indexer cache.Indexer) MSSQLServerArchiverLister {
	return &mSSQLServerArchiverLister{indexer: indexer}
}

// List lists all MSSQLServerArchivers in the indexer.
func (s *mSSQLServerArchiverLister) List(selector labels.Selector) (ret []*v1alpha1.MSSQLServerArchiver, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MSSQLServerArchiver))
	})
	return ret, err
}

// MSSQLServerArchivers returns an object that can list and get MSSQLServerArchivers.
func (s *mSSQLServerArchiverLister) MSSQLServerArchivers(namespace string) MSSQLServerArchiverNamespaceLister {
	return mSSQLServerArchiverNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MSSQLServerArchiverNamespaceLister helps list and get MSSQLServerArchivers.
// All objects returned here must be treated as read-only.
type MSSQLServerArchiverNamespaceLister interface {
	// List lists all MSSQLServerArchivers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.MSSQLServerArchiver, err error)
	// Get retrieves the MSSQLServerArchiver from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.MSSQLServerArchiver, error)
	MSSQLServerArchiverNamespaceListerExpansion
}

// mSSQLServerArchiverNamespaceLister implements the MSSQLServerArchiverNamespaceLister
// interface.
type mSSQLServerArchiverNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all MSSQLServerArchivers in the indexer for a given namespace.
func (s mSSQLServerArchiverNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.MSSQLServerArchiver, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.MSSQLServerArchiver))
	})
	return ret, err
}

// Get retrieves the MSSQLServerArchiver from the indexer for a given namespace and name.
func (s mSSQLServerArchiverNamespaceLister) Get(name string) (*v1alpha1.MSSQLServerArchiver, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("mssqlserverarchiver"), name)
	}
	return obj.(*v1alpha1.MSSQLServerArchiver), nil
}
