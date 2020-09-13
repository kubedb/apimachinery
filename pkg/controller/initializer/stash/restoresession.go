package stash

import (
	"time"

	"github.com/appscode/go/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformers "stash.appscode.dev/apimachinery/client/informers/externalversions/stash/v1beta1"
)

func (s *Stash) restoreSessionInformer() cache.SharedIndexInformer {
	return s.InformerFactory.InformerFor(&v1beta1.RestoreSession{}, func(client scs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return stashinformers.NewFilteredRestoreSessionInformer(
			client,
			s.WatchNamespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			s.tweakListOptions,
		)
	})
}

func (s Stash) restoreSessionEventHandler(selector labels.Selector) cache.ResourceEventHandler {
	return queue.NewFilteredHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rs := obj.(*v1beta1.RestoreSession)
			if rs.Status.Phase == v1beta1.RestoreSucceeded || rs.Status.Phase == v1beta1.RestoreFailed {
				queue.Enqueue(s.RSQueue.GetQueue(), obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldObj := old.(*v1beta1.RestoreSession)
			newObj := new.(*v1beta1.RestoreSession)
			if newObj.Status.Phase != oldObj.Status.Phase && (newObj.Status.Phase == v1beta1.RestoreSucceeded || newObj.Status.Phase == v1beta1.RestoreFailed) {
				queue.Enqueue(s.RSQueue.GetQueue(), newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	}, selector)
}

func (s Stash) processRestoreSession(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := s.RSInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("RestoreSession %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Job was recreated with the same name
		rs := obj.(*v1beta1.RestoreSession).DeepCopy()
		ri, err := s.extractRestoreInfo(rs)
		if err != nil {
			log.Errorln(err)
			return err
		}
		return s.syncDatabasePhase(ri)
	}
	return nil
}
