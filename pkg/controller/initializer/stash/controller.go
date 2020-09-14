package stash

import (
	"time"

	amc "kubedb.dev/apimachinery/pkg/controller"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"kmodules.xyz/client-go/tools/queue"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	stashinformer "stash.appscode.dev/apimachinery/client/informers/externalversions"
)

type Controller struct {
	amc.Config
	*amc.Controller
	*amc.StashInitializer
	// SnapshotDoer interface
	snapshotter amc.DBHelper
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
}

func NewController(
	cfg amc.Config,
	ctrl *amc.Controller,
	initializer *amc.StashInitializer,
	snapshotter amc.DBHelper,
	tweakOptions func(*metav1.ListOptions),
	recorder record.EventRecorder,
) *Controller {
	return &Controller{
		Config:           cfg,
		Controller:       ctrl,
		StashInitializer: initializer,
		snapshotter:      snapshotter,
		tweakListOptions: tweakOptions,
		eventRecorder:    recorder,
	}
}

type restoreInfo struct {
	invoker      core.TypedLocalObjectReference
	namespace    string
	target       *v1beta1.RestoreTarget
	phase        v1beta1.RestorePhase
	targetDBKind string
}

func Configure(cfg *rest.Config, s *amc.StashInitializer, resyncPeriod time.Duration) error {
	var err error
	if s.StashClient, err = scs.NewForConfig(cfg); err != nil {
		return err
	}
	s.StashInformerFactory = stashinformer.NewSharedInformerFactory(s.StashClient, resyncPeriod)
	return nil
}

func (c *Controller) InitWatcher(selector labels.Selector) {
	// Initialize RestoreSession Watcher
	c.RSInformer = c.restoreSessionInformer()
	c.RSQueue = queue.New(v1beta1.ResourceKindRestoreSession, c.MaxNumRequeues, c.NumThreads, c.processRestoreSession)
	c.RSLister = c.StashInformerFactory.Stash().V1beta1().RestoreSessions().Lister()
	c.RSInformer.AddEventHandler(c.restoreSessionEventHandler(selector))

	// Initialize RestoreBatch Watcher
	c.RBInformer = c.restoreBatchInformer()
	c.RBQueue = queue.New(v1beta1.ResourceKindRestoreBatch, c.MaxNumRequeues, c.NumThreads, c.processRestoreBatch)
	c.RBLister = c.StashInformerFactory.Stash().V1beta1().RestoreBatches().Lister()
	c.RBInformer.AddEventHandler(c.restoreBatchEventHandler(selector))
}

func (c Controller) StartController(stopCh <-chan struct{}) {
	// start StashInformerFactory only if stash crds (ie, "restoreSession") are available.
	if err := c.waitUntilStashInstalled(stopCh); err != nil {
		log.Errorln("error while waiting for restoreSession.", err)
		return
	}

	// start informer factory
	c.StashInformerFactory.Start(stopCh)
	// wait for cache to sync
	for t, v := range c.StashInformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync", t)
			return
		}
	}
	// run the queues
	c.RSQueue.Run(stopCh)
	c.RBQueue.Run(stopCh)
}
