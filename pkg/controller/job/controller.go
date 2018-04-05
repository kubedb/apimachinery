package job

import (
	"time"

	"github.com/appscode/kutil/tools/queue"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"k8s.io/client-go/informers"
	batch_listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	// SnapshotDoer interface
	snapshotter amc.Snapshotter
	// ListOptions for watcher
	labelMap map[string]string
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Max number requests for retries
	maxNumRequests int
	// threadiness of Job handler
	numThreads     int
	watchNamespace string

	// Informer factory
	kubeInformerFactory   informers.SharedInformerFactory
	kubedbInformerFactory kubedbinformers.SharedInformerFactory

	// DormantDatabase
	jobQueue    *queue.Worker
	jobInformer cache.SharedIndexInformer
	jobLister   batch_listers.JobLister
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	kubeInformerFactory informers.SharedInformerFactory,
	kubedbInformerFactory kubedbinformers.SharedInformerFactory,
	watchNamespace string,
	labelmap map[string]string,
	syncPeriod time.Duration,
	maxNumRequests int,
	numThreads int,
) *Controller {
	// return new DormantDatabase Controller
	return &Controller{
		Controller:            controller,
		snapshotter:           snapshotter,
		kubedbInformerFactory: kubedbInformerFactory,
		kubeInformerFactory:   kubeInformerFactory,
		watchNamespace:        watchNamespace,
		labelMap:              labelmap,
		eventRecorder:         eventer.NewEventRecorder(controller.Client, "Job Controller"),
		syncPeriod:            syncPeriod,
		maxNumRequests:        maxNumRequests,
		numThreads:            numThreads,
	}
}

func InitJobWatcher(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	kubeInformerFactory informers.SharedInformerFactory,
	kubedbInformerFactory kubedbinformers.SharedInformerFactory,
	watchNamespace string,
	labelmap map[string]string,
	syncPeriod time.Duration,
	maxNumRequests int,
	numThreads int,
) *queue.Worker {

	ctrl := NewController(controller, snapshotter, kubeInformerFactory, kubedbInformerFactory, watchNamespace, labelmap, syncPeriod, maxNumRequests, numThreads)
	ctrl.initWatcher()
	return ctrl.jobQueue
}
