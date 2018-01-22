package job

import (
	"time"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Job interface {
	GetDatabase(*api.Snapshot) (runtime.Object, error)
}

type Controller struct {
	*amc.Controller
	// Job interface
	job Job
	// Watcher selector
	selector labels.Selector
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
	selector labels.Selector,
	syncPeriod time.Duration,
) *Controller {

	// return new DormantDatabase Controller
	return &Controller{
		Controller:     controller,
		job:            job,
		selector:       selector,
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
