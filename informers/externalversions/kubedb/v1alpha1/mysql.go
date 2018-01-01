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

// MySQLInformer provides access to a shared informer and lister for
// MySQLs.
type MySQLInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MySQLLister
}

type mySQLInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewMySQLInformer constructs a new informer for MySQL type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMySQLInformer(client client.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.KubedbV1alpha1().MySQLs(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.KubedbV1alpha1().MySQLs(namespace).Watch(options)
			},
		},
		&kubedb_v1alpha1.MySQL{},
		resyncPeriod,
		indexers,
	)
}

func defaultMySQLInformer(client client.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewMySQLInformer(client, v1.NamespaceAll, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *mySQLInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubedb_v1alpha1.MySQL{}, defaultMySQLInformer)
}

func (f *mySQLInformer) Lister() v1alpha1.MySQLLister {
	return v1alpha1.NewMySQLLister(f.Informer().GetIndexer())
}
