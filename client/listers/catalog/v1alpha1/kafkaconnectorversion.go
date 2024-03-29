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

// KafkaConnectorVersionLister helps list KafkaConnectorVersions.
// All objects returned here must be treated as read-only.
type KafkaConnectorVersionLister interface {
	// List lists all KafkaConnectorVersions in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.KafkaConnectorVersion, err error)
	// Get retrieves the KafkaConnectorVersion from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.KafkaConnectorVersion, error)
	KafkaConnectorVersionListerExpansion
}

// kafkaConnectorVersionLister implements the KafkaConnectorVersionLister interface.
type kafkaConnectorVersionLister struct {
	indexer cache.Indexer
}

// NewKafkaConnectorVersionLister returns a new KafkaConnectorVersionLister.
func NewKafkaConnectorVersionLister(indexer cache.Indexer) KafkaConnectorVersionLister {
	return &kafkaConnectorVersionLister{indexer: indexer}
}

// List lists all KafkaConnectorVersions in the indexer.
func (s *kafkaConnectorVersionLister) List(selector labels.Selector) (ret []*v1alpha1.KafkaConnectorVersion, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.KafkaConnectorVersion))
	})
	return ret, err
}

// Get retrieves the KafkaConnectorVersion from the index for a given name.
func (s *kafkaConnectorVersionLister) Get(name string) (*v1alpha1.KafkaConnectorVersion, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("kafkaconnectorversion"), name)
	}
	return obj.(*v1alpha1.KafkaConnectorVersion), nil
}
