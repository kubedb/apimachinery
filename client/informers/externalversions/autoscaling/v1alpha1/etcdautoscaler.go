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

	autoscalingv1alpha1 "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	versioned "kubedb.dev/apimachinery/client/clientset/versioned"
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kubedb.dev/apimachinery/client/listers/autoscaling/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EtcdAutoscalerInformer provides access to a shared informer and lister for
// EtcdAutoscalers.
type EtcdAutoscalerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.EtcdAutoscalerLister
}

type etcdAutoscalerInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEtcdAutoscalerInformer constructs a new informer for EtcdAutoscaler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEtcdAutoscalerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEtcdAutoscalerInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEtcdAutoscalerInformer constructs a new informer for EtcdAutoscaler type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEtcdAutoscalerInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AutoscalingV1alpha1().EtcdAutoscalers(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AutoscalingV1alpha1().EtcdAutoscalers(namespace).Watch(context.TODO(), options)
			},
		},
		&autoscalingv1alpha1.EtcdAutoscaler{},
		resyncPeriod,
		indexers,
	)
}

func (f *etcdAutoscalerInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEtcdAutoscalerInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *etcdAutoscalerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&autoscalingv1alpha1.EtcdAutoscaler{}, f.defaultInformer)
}

func (f *etcdAutoscalerInformer) Lister() v1alpha1.EtcdAutoscalerLister {
	return v1alpha1.NewEtcdAutoscalerLister(f.Informer().GetIndexer())
}
