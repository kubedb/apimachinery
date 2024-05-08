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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	opsv1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	versioned "kubedb.dev/apimachinery/client/clientset/versioned"
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kubedb.dev/apimachinery/client/listers/ops/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// KafkaOpsRequestInformer provides access to a shared informer and lister for
// KafkaOpsRequests.
type KafkaOpsRequestInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.KafkaOpsRequestLister
}

type kafkaOpsRequestInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewKafkaOpsRequestInformer constructs a new informer for KafkaOpsRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewKafkaOpsRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredKafkaOpsRequestInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredKafkaOpsRequestInformer constructs a new informer for KafkaOpsRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredKafkaOpsRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpsV1alpha1().KafkaOpsRequests(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpsV1alpha1().KafkaOpsRequests(namespace).Watch(context.TODO(), options)
			},
		},
		&opsv1alpha1.KafkaOpsRequest{},
		resyncPeriod,
		indexers,
	)
}

func (f *kafkaOpsRequestInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredKafkaOpsRequestInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *kafkaOpsRequestInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&opsv1alpha1.KafkaOpsRequest{}, f.defaultInformer)
}

func (f *kafkaOpsRequestInformer) Lister() v1alpha1.KafkaOpsRequestLister {
	return v1alpha1.NewKafkaOpsRequestLister(f.Informer().GetIndexer())
}
