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

package petset

import (
	"fmt"

	"kubedb.dev/apimachinery/apis/kubedb"
	db_cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	petsetapps "kubeops.dev/petset/apis/apps/v1"
	petsetcs "kubeops.dev/petset/client/clientset/versioned"
)

type Controller struct {
	*amc.Config

	// Kubernetes client
	Client kubernetes.Interface
	// KubeDB client
	DBClient db_cs.Interface
	// PetSet client
	PSClient petsetcs.Interface
	// Dynamic client
	DynamicClient dynamic.Interface
}

func NewController(
	config *amc.Config,
	client kubernetes.Interface,
	dbClient db_cs.Interface,
	dmClient dynamic.Interface,
	psClient petsetcs.Interface,
) *Controller {
	return &Controller{
		Config:        config,
		Client:        client,
		DBClient:      dbClient,
		DynamicClient: dmClient,
		PSClient:      psClient,
	}
}

func (c *Controller) InitPetSetWatcher() {
	klog.Infoln("Initializing PetSet watcher.....")
	// Initialize PetSet Watcher
	c.PSInformer = c.PetSetInformerFactory.Apps().V1().PetSets().Informer()
	c.PSQueue = queue.New(kubedb.ResourceKindPetSet, c.MaxNumRequeues, c.NumThreads, c.processPetSet)
	c.PSLister = c.PetSetInformerFactory.Apps().V1().PetSets().Lister()
	_, _ = c.PSInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if ps, ok := obj.(*petsetapps.PetSet); ok {
				if c.RestrictToNamespace != core.NamespaceAll {
					if ps.GetNamespace() != c.RestrictToNamespace {
						klog.Infof("Skipping PetSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", ps.GetNamespace(), ps.GetName(), c.RestrictToNamespace)
						return
					}
				}

				c.enqueueOnlyKubeDBPS(ps)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if ps, ok := newObj.(*petsetapps.PetSet); ok {
				if c.RestrictToNamespace != core.NamespaceAll {
					if ps.GetNamespace() != c.RestrictToNamespace {
						klog.Infof("Skipping PetSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", ps.GetNamespace(), ps.GetName(), c.RestrictToNamespace)
						return
					}
				}

				c.enqueueOnlyKubeDBPS(ps)
			}
		},
		DeleteFunc: func(obj interface{}) {
			var ps *petsetapps.PetSet
			var ok bool
			if ps, ok = obj.(*petsetapps.PetSet); !ok {
				tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					klog.V(5).Info("error decoding object, invalid type")
					return
				}
				ps, ok = tombstone.Obj.(*petsetapps.PetSet)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
					return
				}
				klog.V(5).Infof("Recovered deleted object '%v' from tombstone", tombstone.Obj.(metav1.Object).GetName())
			}

			if c.RestrictToNamespace != core.NamespaceAll {
				if ps.GetNamespace() != c.RestrictToNamespace {
					klog.Infof("Skipping PetSet %s/%s. Only %s namespace is supported for Community Edition. Please upgrade to Enterprise to use any namespace.", ps.GetNamespace(), ps.GetName(), c.RestrictToNamespace)
					return
				}
			}

			ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(ps), kubedb.GroupName)
			if err != nil || !ok {
				klog.Warningln(err)
				return
			}
			dbInfo, err := c.extractDatabaseInfo(ps)
			if err != nil {
				if !kerr.IsNotFound(err) {
					klog.Warningf("failed to extract database info from PetSet: %s/%s. Reason: %v", ps.Namespace, ps.Name, err)
				}
				return
			}
			if dbInfo == nil {
				return
			}
			err = c.ensureReadyReplicasCond(dbInfo)
			if err != nil {
				klog.Warningf("failed to update ReadyReplicas condition. Reason: %v", err)
				return
			}
		},
	})
}

func (c *Controller) enqueueOnlyKubeDBPS(ps *petsetapps.PetSet) {
	// only enqueue if the controlling owner is a KubeDB resource
	ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(ps), kubedb.GroupName)
	if err != nil {
		klog.Warningf("failed to enqueue PetSet: %s/%s. Reason: %v", ps.Namespace, ps.Name, err)
		return
	}
	if ok {
		queue.Enqueue(c.PSQueue.GetQueue(), cache.ExplicitKey(ps.Namespace+"/"+ps.Name))
	}
}

func (c *Controller) processPetSet(key string) error {
	obj, exists, err := c.PSInformer.GetIndexer().GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.V(5).Infof("PetSet %s does not exist anymore", key)
	} else {
		ps := obj.(*petsetapps.PetSet).DeepCopy()
		dbInfo, err := c.extractDatabaseInfo(ps)
		if err != nil {
			return fmt.Errorf("failed to extract database info from PetSet: %s/%s. Reason: %v", ps.Namespace, ps.Name, err)
		}
		if dbInfo == nil {
			return nil
		}
		return c.ensureReadyReplicasCond(dbInfo)
	}
	return nil
}
