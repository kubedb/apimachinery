/*
Copyright 2018 The KubeDB Authors.

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

// This file was automatically generated by informer-gen

package v1alpha1

import (
	kubedb_v1alpha1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	client "github.com/kubedb/apimachinery/client"
	internalinterfaces "github.com/kubedb/apimachinery/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubedb/apimachinery/listers/kubedb/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// MemcachedInformer provides access to a shared informer and lister for
// Memcacheds.
type MemcachedInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MemcachedLister
}

type memcachedInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMemcachedInformer constructs a new informer for Memcached type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMemcachedInformer(client client.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMemcachedInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMemcachedInformer constructs a new informer for Memcached type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMemcachedInformer(client client.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubedbV1alpha1().Memcacheds(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubedbV1alpha1().Memcacheds(namespace).Watch(options)
			},
		},
		&kubedb_v1alpha1.Memcached{},
		resyncPeriod,
		indexers,
	)
}

func (f *memcachedInformer) defaultInformer(client client.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMemcachedInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *memcachedInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubedb_v1alpha1.Memcached{}, f.defaultInformer)
}

func (f *memcachedInformer) Lister() v1alpha1.MemcachedLister {
	return v1alpha1.NewMemcachedLister(f.Informer().GetIndexer())
}
