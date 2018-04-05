package snapshot

import (
	"time"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	api_listers "github.com/kubedb/apimachinery/client/listers/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	jobc "github.com/kubedb/apimachinery/pkg/controller/job"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	// Snapshotter interface
	snapshotter amc.Snapshotter
	// ListOptions for watcher
	labelMap map[string]string
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Max number requests for retries
	maxNumRequests int
	// threadiness of Snapshot handler
	numThreads     int
	watchNamespace string

	// Informer factory
	kubeInformerFactory   informers.SharedInformerFactory
	kubedbInformerFactory kubedbinformers.SharedInformerFactory

	// DormantDatabase
	snQueue    *queue.Worker
	snInformer cache.SharedIndexInformer
	snLister   api_listers.SnapshotLister
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
		eventRecorder:         eventer.NewEventRecorder(controller.Client, "Snapshot Controller"),
		syncPeriod:            syncPeriod,
		maxNumRequests:        maxNumRequests,
		numThreads:            numThreads,
	}
}

func (c *Controller) Setup() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.Snapshot{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

// InitSnapshotWatcher ensures snapshot watcher and returns queue.Worker.
// So, it is possible to start queue.run from other package/repositories
func InitSnapshotWatcher(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	kubeInformerFactory informers.SharedInformerFactory,
	kubedbInformerFactory kubedbinformers.SharedInformerFactory,
	watchNamespace string,
	labelmap map[string]string,
	syncPeriod time.Duration,
	maxNumRequests int,
	numThreads int,
) (*queue.Worker, *queue.Worker) {

	ctrl := NewController(controller, snapshotter, kubeInformerFactory, kubedbInformerFactory, watchNamespace, labelmap, syncPeriod, maxNumRequests, numThreads)
	ctrl.initWatcher()

	jobQueue := jobc.InitJobWatcher(controller, snapshotter, kubeInformerFactory, kubedbInformerFactory, watchNamespace, labelmap, syncPeriod, maxNumRequests, numThreads)

	return ctrl.snQueue, jobQueue
}
