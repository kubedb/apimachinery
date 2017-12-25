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

type DormantDbController struct {
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
	MaxNumRequeues int
}

// NewDormantDbController creates a new DormantDatabase Controller
func NewDormantDbController(
	controller *amc.Controller,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DormantDbController {
	// return new DormantDatabase Controller
	return &DormantDbController{
		Controller: controller,
		deleter:    deleter,
		lw:         lw,
		recorder:   eventer.NewEventRecorder(controller.Client, "DormantDatabase Controller"),
		syncPeriod: syncPeriod,
	}
}

func (c *DormantDbController) Setup() error {
	crds := []*crd_api.CustomResourceDefinition{
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return apiext_util.RegisterCRDs(c.ApiExtKubeClient, crds)
}

func (c *DormantDbController) Run() {
	// Watch DormantDatabase with provided ListerWatcher
	c.watchDormantDatabase()
}

func (c *DormantDbController) watchDormantDatabase() {

	c.initWatcher()

	stop := make(chan struct{})
	defer close(stop)

	c.runWatcher(1, stop)
	select {}
}
