package dormantdatabase

import (
	"time"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	kubedbinformers "github.com/kubedb/apimachinery/client/informers/externalversions"
	api_listers "github.com/kubedb/apimachinery/client/listers/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	// Deleter interface
	deleter amc.Deleter
	// ListerWatcher
	labelMap map[string]string
	// Event Recorder
	recorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
	// Max number requests for retries
	maxNumRequests int
	// threadiness of DormantDB handler
	numThreads     int
	watchNamespace string

	// Informer factory
	kubeInformerFactory   informers.SharedInformerFactory
	kubedbInformerFactory kubedbinformers.SharedInformerFactory

	// DormantDatabase
	ddbQueue    *queue.Worker
	ddbInformer cache.SharedIndexInformer
	ddbLister   api_listers.DormantDatabaseLister
}

// NewController creates a new DormantDatabase Controller
func NewController(
	controller *amc.Controller,
	deleter amc.Deleter,
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
		deleter:               deleter,
		kubedbInformerFactory: kubedbInformerFactory,
		kubeInformerFactory:   kubeInformerFactory,
		watchNamespace:        watchNamespace,
		labelMap:              labelmap,
		recorder:              eventer.NewEventRecorder(controller.Client, "DormantDatabase Controller"),
		syncPeriod:            syncPeriod,
		maxNumRequests:        maxNumRequests,
		numThreads:            numThreads,
	}
}

func (c *Controller) EnsureCustomResourceDefinitions() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

func InitDormantDatabaseWatcher(
	controller *amc.Controller,
	deleter amc.Deleter,
	kubeInformerFactory informers.SharedInformerFactory,
	kubedbInformerFactory kubedbinformers.SharedInformerFactory,
	watchNamespace string,
	labelmap map[string]string,
	syncPeriod time.Duration,
	maxNumRequests int,
	numThreads int,
) *queue.Worker {
	ctrl := NewController(controller, deleter, kubeInformerFactory, kubedbInformerFactory, watchNamespace, labelmap, syncPeriod, maxNumRequests, numThreads)
	ctrl.initWatcher()
	return ctrl.ddbQueue
}
