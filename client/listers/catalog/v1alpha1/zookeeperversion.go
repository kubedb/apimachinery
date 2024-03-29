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
	v1alpha1 "kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ZooKeeperVersionLister helps list ZooKeeperVersions.
// All objects returned here must be treated as read-only.
type ZooKeeperVersionLister interface {
	// List lists all ZooKeeperVersions in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ZooKeeperVersion, err error)
	// Get retrieves the ZooKeeperVersion from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ZooKeeperVersion, error)
	ZooKeeperVersionListerExpansion
}

// zooKeeperVersionLister implements the ZooKeeperVersionLister interface.
type zooKeeperVersionLister struct {
	indexer cache.Indexer
}

// NewZooKeeperVersionLister returns a new ZooKeeperVersionLister.
func NewZooKeeperVersionLister(indexer cache.Indexer) ZooKeeperVersionLister {
	return &zooKeeperVersionLister{indexer: indexer}
}

// List lists all ZooKeeperVersions in the indexer.
func (s *zooKeeperVersionLister) List(selector labels.Selector) (ret []*v1alpha1.ZooKeeperVersion, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ZooKeeperVersion))
	})
	return ret, err
}

// Get retrieves the ZooKeeperVersion from the index for a given name.
func (s *zooKeeperVersionLister) Get(name string) (*v1alpha1.ZooKeeperVersion, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("zookeeperversion"), name)
	}
	return obj.(*v1alpha1.ZooKeeperVersion), nil
}
