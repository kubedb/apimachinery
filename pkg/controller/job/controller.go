package job

import (
	"time"

	"github.com/appscode/kutil/tools/queue"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	listOption metav1.ListOptions
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Max number requests for retries
	maxNumRequests int
	// threadiness of Job handler
	numThreads int

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
	listOption metav1.ListOptions,
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
		listOption:            listOption,
		eventRecorder:         eventer.NewEventRecorder(controller.Client, "Job Controller"),
		syncPeriod:            syncPeriod,
		maxNumRequests:        maxNumRequests,
		numThreads:            numThreads,
	}
}

func (c *Controller) Run() {
	// Watch DormantDatabase with provided ListerWatcher
	c.watchJob()
}

func (c *Controller) watchJob() {

	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(c.numThreads, stop)
	select {}
}
