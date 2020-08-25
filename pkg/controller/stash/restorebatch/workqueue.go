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

package restorebatch

import (
	"github.com/appscode/go/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

func (c *Controller) addEventHandlerBatch(selector labels.Selector) {
	c.RBQueue = queue.New("RestoreBatch", c.MaxNumRequeues, c.NumThreads, c.runRestoreBatch)
	c.rbLister = c.StashInformerFactory.Stash().V1beta1().RestoreBatches().Lister()
	c.RBInformer.AddEventHandler(queue.NewFilteredHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rb := obj.(*v1beta1.RestoreBatch)
			if rb.Status.Phase == v1beta1.RestoreSucceeded || rb.Status.Phase == v1beta1.RestoreFailed {
				queue.Enqueue(c.RBQueue.GetQueue(), obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldObj := old.(*v1beta1.RestoreBatch)
			newObj := new.(*v1beta1.RestoreBatch)
			if newObj.Status.Phase != oldObj.Status.Phase && (newObj.Status.Phase == v1beta1.RestoreSucceeded || newObj.Status.Phase == v1beta1.RestoreFailed) {
				queue.Enqueue(c.RBQueue.GetQueue(), newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	}, selector))
}

func (c *Controller) runRestoreBatch(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.RBInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("RestoreBatch %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Job was recreated with the same name
		rb := obj.(*v1beta1.RestoreBatch).DeepCopy()
		if err := c.handleRestoreBatch(rb); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
