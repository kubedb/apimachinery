/*
Copyright AppsCode Inc. and Contributors.

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

package v1

import (
	"context"
	time "time"

	appsv1 "kubeops.dev/petset/apis/apps/v1"
	versioned "kubeops.dev/petset/client/clientset/versioned"
	internalinterfaces "kubeops.dev/petset/client/informers/externalversions/internalinterfaces"
	v1 "kubeops.dev/petset/client/listers/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// PlacementPolicyInformer provides access to a shared informer and lister for
// PlacementPolicies.
type PlacementPolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.PlacementPolicyLister
}

type placementPolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewPlacementPolicyInformer constructs a new informer for PlacementPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewPlacementPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredPlacementPolicyInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredPlacementPolicyInformer constructs a new informer for PlacementPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredPlacementPolicyInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1().PlacementPolicies().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1().PlacementPolicies().Watch(context.TODO(), options)
			},
		},
		&appsv1.PlacementPolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *placementPolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredPlacementPolicyInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *placementPolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appsv1.PlacementPolicy{}, f.defaultInformer)
}

func (f *placementPolicyInformer) Lister() v1.PlacementPolicyLister {
	return v1.NewPlacementPolicyLister(f.Informer().GetIndexer())
}
