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

package manifestwork

import (
	"fmt"
	"sync"

	amc "kubedb.dev/apimachinery/pkg/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/meta"
	client_meta "kmodules.xyz/client-go/meta"
	kubesliceapi "open-cluster-management.io/api/work/v1"
)

type Store struct {
	m  map[string]func(*kubesliceapi.ManifestWork)
	mu sync.RWMutex
}

func NewStore() *Store {
	s := &Store{}
	s.Init()
	return s
}

func (s *Store) Init() *Store {
	s.m = make(map[string]func(*kubesliceapi.ManifestWork))
	return s
}

func (s *Store) Add(name string, f func(*kubesliceapi.ManifestWork)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[name] = f
}

func (s *Store) Get(name string) (func(*kubesliceapi.ManifestWork), bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[name]
	return v, ok
}

func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, name)
}

type ManifestWorkWatcher struct {
	*amc.Config
	Store *Store
}

func NewManifestWorkWatcher(
	config *amc.Config,
) *ManifestWorkWatcher {
	return &ManifestWorkWatcher{
		Config: config,
		Store:  NewStore(),
	}
}

func (c *ManifestWorkWatcher) InitManifestWorkWatcher() {
	klog.Infoln("Initializing ManifestWork watcher.....")
	// Initialize ManifestWork Watcher
	c.MWInformer = c.ManifestInformerFactory.Work().V1().ManifestWorks().Informer()
	c.MWLister = c.ManifestInformerFactory.Work().V1().ManifestWorks().Lister()
	_, _ = c.MWInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			mw, ok := obj.(*kubesliceapi.ManifestWork)
			if !ok {
				return
			}
			if !HasRequiredLabels(mw.Labels) {
				klog.V(4).Infof("%v, %v, %v, %v labels are required for manifestWork %v/%v", meta.InstanceLabelKey, meta.NameLabelKey, meta.ManagedByLabelKey, client_meta.NamespaceLabelKey, mw.Namespace, mw.Name)
				return
			}
			c.addManifestWork(mw)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if mw, ok := newObj.(*kubesliceapi.ManifestWork); ok {

				if !HasRequiredLabels(mw.Labels) {
					klog.V(4).Infof("%v, %v, %v, %v labels are required for manifestWork %v/%v", meta.InstanceLabelKey, meta.NameLabelKey, meta.ManagedByLabelKey, client_meta.NamespaceLabelKey, mw.Namespace, mw.Name)
					return
				}
				c.updateManifestWork(mw)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var mw *kubesliceapi.ManifestWork
			var ok bool
			if mw, ok = obj.(*kubesliceapi.ManifestWork); !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.V(5).Info("error decoding object, invalid type")
					return
				}
				mw, ok = tombstone.Obj.(*kubesliceapi.ManifestWork)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
					return
				}
				klog.V(5).Infof("Recovered deleted object '%v' from tombstone", tombstone.Obj.(metav1.Object).GetName())
			}

			if !HasRequiredLabels(mw.Labels) {
				klog.V(4).Infof("%v, %v, %v, %v labels are required for manifestWork %v/%v", meta.InstanceLabelKey, meta.NameLabelKey, meta.ManagedByLabelKey, client_meta.NamespaceLabelKey, mw.Namespace, mw.Name)
				return
			}
			c.deleteManifestWork(mw)
		},
	})
}

// addManifest adds the ManifestWork for the manifestwork to the sync queue
func (c *ManifestWorkWatcher) addManifestWork(mw *kubesliceapi.ManifestWork) {
	if mw.DeletionTimestamp != nil {
		klog.Infoln("need to delete manifestwork", mw.Name)
		// on a restart of the controller manager, it's possible a new pod shows up in a state that
		// is already pending deletion. Prevent the pod from being a creation observation.
		c.deleteManifestWork(mw)
		return
	}

	name := mw.Labels[meta.NameLabelKey]
	fn, exists := c.Store.Get(name)
	if !exists {
		klog.Errorf("no eventHandler found for manifestWork %s/%s with %v: %v", mw.Namespace, mw.Name, meta.NameLabelKey, mw.Labels[meta.NameLabelKey])
		return
	}
	fn(mw)
}

// updateManifestWork adds the ManifestWork for the current and ManifestWork pods to the sync queue.
func (c *ManifestWorkWatcher) updateManifestWork(curMW *kubesliceapi.ManifestWork) {
	name := curMW.Labels[meta.NameLabelKey]
	fn, exists := c.Store.Get(name)
	if !exists {
		klog.Errorf("no controller function found for manifestWork %s/%s", curMW.Namespace, curMW.Name)
		return
	}
	fn(curMW)
}

// deleteManifestWork enqueues the ManifestWork for the ManifestWork accounting for deletion tombstones.
func (c *ManifestWorkWatcher) deleteManifestWork(mw *kubesliceapi.ManifestWork) {
	name := mw.Labels[meta.NameLabelKey]
	fn, exists := c.Store.Get(name)
	if !exists {
		klog.Errorf("no controller function found for manifestWork %s/%s", mw.Namespace, mw.Name)
		return
	}
	fn(mw)
}

func HasRequiredLabels(labels map[string]string) bool {
	_, ok1 := labels[meta.InstanceLabelKey]
	_, ok2 := labels[meta.NameLabelKey]
	_, ok3 := labels[meta.ManagedByLabelKey]
	_, ok4 := labels[client_meta.NamespaceLabelKey]
	return ok1 && ok2 && ok3 && ok4
}
