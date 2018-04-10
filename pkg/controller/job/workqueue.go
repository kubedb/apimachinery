package job

import (
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil/tools/queue"
	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	batchinformer "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) initWatcher() {
	c.jobInformer = c.KubeInformerFactory.InformerFor(&batch.Job{}, func(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return batchinformer.NewFilteredJobInformer(
			client,
			c.WatchNamespace, // need to provide namespace
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			func(options *metav1.ListOptions) {
				options.LabelSelector = labels.SelectorFromSet(c.labelMap).String()
			},
		)
	})
	c.jobQueue = queue.New("Job", c.MaxNumRequeues, c.NumThreads, c.runJob)
	c.jobLister = c.KubeInformerFactory.Batch().V1().Jobs().Lister()
	c.jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job, ok := obj.(*batch.Job)
			if !ok {
				log.Errorln("Invalid Job object")
				return
			}

			if job.Status.Succeeded > 0 || job.Status.Failed > types.Int32(job.Spec.BackoffLimit) {
				// IndexerInformer uses a delta queue, therefore for deletes we have to use this
				// key function.
				queue.Enqueue(c.jobQueue.GetQueue(), obj)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldObj, ok := old.(*batch.Job)
			if !ok {
				log.Errorln("Invalid Job object")
				return
			}
			newObj, ok := new.(*batch.Job)
			if !ok {
				log.Errorln("Invalid Job object")
				return
			}
			if isJobCompleted(oldObj, newObj) {
				queue.Enqueue(c.jobQueue.GetQueue(), new)
			}
		},
		DeleteFunc: func(obj interface{}) {
			job, ok := obj.(*batch.Job)
			if !ok {
				log.Errorln("Invalid Job object")
				return
			}

			if job.Status.Succeeded == 0 && job.Status.Failed <= types.Int32(job.Spec.BackoffLimit) {
				// IndexerInformer uses a delta queue, therefore for deletes we have to use this
				// key function.
				queue.Enqueue(c.jobQueue.GetQueue(), obj)
			}
		},
	})
}

func isJobCompleted(old, new *batch.Job) bool {
	if old.Status.Succeeded == 0 && new.Status.Succeeded > 0 {
		return true
	}
	if old.Status.Failed < types.Int32(old.Spec.BackoffLimit) && new.Status.Failed >= types.Int32(new.Spec.BackoffLimit) {
		return true
	}
	return false
}

func (c *Controller) runJob(key string) error {
	log.Debugf("started processing, key: %v\n", key)
	obj, exists, err := c.jobInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if !exists {
		log.Debugf("Job %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Job was recreated with the same name
		job := obj.(*batch.Job).DeepCopy()
		if err := c.completeJob(job); err != nil {
			log.Errorln(err)
			return err
		}
	}
	return nil
}
