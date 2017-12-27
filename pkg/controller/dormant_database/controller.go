package dormant_database

import (
	"time"

	apiext_util "github.com/appscode/kutil/apiextensions/v1beta1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Deleter interface {
	// Check Database CRD
	Exists(*metav1.ObjectMeta) (bool, error)
	// Pause operation
	PauseDatabase(*api.DormantDatabase) error
	// Wipe out operation
	WipeOutDatabase(*api.DormantDatabase) error
	// Resume operation
	ResumeDatabase(*api.DormantDatabase) error
}

type controller struct {
	*amc.Controller
	// Deleter interface
	deleter Deleter
	// ListerWatcher
	lw *cache.ListWatch
	// Event Recorder
	recorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Workqueue
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	//Max number requests for retries
	maxNumRequeues int
}

// NewController creates a new DormantDatabase controller
func NewController(
	controller *amc.Controller,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *controller {
	// return new DormantDatabase controller
	return &controller{
		Controller:     controller,
		deleter:        deleter,
		lw:             lw,
		recorder:       eventer.NewEventRecorder(controller.Client, "DormantDatabase controller"),
		syncPeriod:     syncPeriod,
		maxNumRequeues: 5,
	}
}

func (c *controller) Setup() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(c.ApiExtKubeClient, crd)
}

func (c *controller) Run() {
	// Watch DormantDatabase with provided ListerWatcher
	c.watchDormantDatabase()
}

func (c *controller) watchDormantDatabase() {

	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(1, stop)
	select {}
}
