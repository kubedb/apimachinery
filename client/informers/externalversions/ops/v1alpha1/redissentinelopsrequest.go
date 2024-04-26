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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	opsv1alpha1 "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	versioned "kubedb.dev/apimachinery/client/clientset/versioned"
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kubedb.dev/apimachinery/client/listers/ops/v1alpha1"
)

// RedisSentinelOpsRequestInformer provides access to a shared informer and lister for
// RedisSentinelOpsRequests.
type RedisSentinelOpsRequestInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.RedisSentinelOpsRequestLister
}

type redisSentinelOpsRequestInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewRedisSentinelOpsRequestInformer constructs a new informer for RedisSentinelOpsRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewRedisSentinelOpsRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredRedisSentinelOpsRequestInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredRedisSentinelOpsRequestInformer constructs a new informer for RedisSentinelOpsRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredRedisSentinelOpsRequestInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpsV1alpha1().RedisSentinelOpsRequests(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OpsV1alpha1().RedisSentinelOpsRequests(namespace).Watch(context.TODO(), options)
			},
		},
		&opsv1alpha1.RedisSentinelOpsRequest{},
		resyncPeriod,
		indexers,
	)
}

func (f *redisSentinelOpsRequestInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredRedisSentinelOpsRequestInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *redisSentinelOpsRequestInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&opsv1alpha1.RedisSentinelOpsRequest{}, f.defaultInformer)
}

func (f *redisSentinelOpsRequestInformer) Lister() v1alpha1.RedisSentinelOpsRequestLister {
	return v1alpha1.NewRedisSentinelOpsRequestLister(f.Informer().GetIndexer())
}
