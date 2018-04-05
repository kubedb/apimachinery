package dormantdatabase

import (
	"time"

	"github.com/appscode/go/log"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	kubedb_informers "github.com/kubedb/apimachinery/client/informers/externalversions/kubedb/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) initWatcher() {
	c.ddbInformer = c.kubedbInformerFactory.InformerFor(&api.DormantDatabase{}, func(client cs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return kubedb_informers.NewFilteredDormantDatabaseInformer(
			client,
			c.watchNamespace, // need to provide namespace
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			func(options *metav1.ListOptions) {
				options.LabelSelector = labels.SelectorFromSet(c.labelMap).String()
			},
		)
	})
	c.ddbQueue = queue.New("MongoDB", c.maxNumRequests, c.numThreads, c.runDormantDatabase)
	c.ddbInformer.AddEventHandler(queue.NewEventHandler(c.ddbQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		oldObj, ok := old.(*api.DormantDatabase)
		if !ok {
			log.Errorln("Invalid DormantDatabase object")
			return false
		}
		newObj, ok := new.(*api.DormantDatabase)
		if !ok {
			log.Errorln("Invalid DormantDatabase object")
			return false
		}

		if !dormantDatabaseEqual(oldObj, newObj) {
			return true
		}
		return false
	}))
	c.ddbLister = c.kubedbInformerFactory.Kubedb().V1alpha1().DormantDatabases().Lister()
}

func dormantDatabaseEqual(old, new *api.DormantDatabase) bool {
	if !meta_util.Equal(old.Spec, new.Spec) {
		diff := meta_util.Diff(old.Spec, new.Spec)
		log.Debugf("DormantDatabase %s/%s has changed. Diff: %s\n", new.Namespace, new.Name, diff)
		return false
	}
	return true
}

func (c *Controller) runDormantDatabase(key string) error {
	log.Debugf("started processing, key: %v\n", key)
	obj, exists, err := c.ddbInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		log.Debugf("DormantDatabase %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a DormantDatabase was recreated with the same name
		dormantDatabase := obj.(*api.DormantDatabase).DeepCopy()
		util.AssignTypeKind(dormantDatabase)
		if err := c.create(dormantDatabase); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
