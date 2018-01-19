package snapshot

import (
	"time"

	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	batch "k8s.io/api/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Job interface {
	DeleteJobResources(job *batch.Job) error
}

type Controller struct {
	*amc.Controller
	// Snapshotter interface
	job Job
	// ListerWatcher
	lw *cache.ListWatch
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Workqueue
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	//Max number requests for retries
	maxNumRequests int
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	job Job,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *Controller {

	// return new DormantDatabase Controller
	return &Controller{
		Controller:     controller,
		job:            job,
		lw:             lw,
		eventRecorder:  eventer.NewEventRecorder(controller.Client, "Job Controller"),
		syncPeriod:     syncPeriod,
		maxNumRequests: 2,
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

	c.runWatcher(5, stop)
	select {}
}
