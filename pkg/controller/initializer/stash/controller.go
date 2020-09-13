package stash

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	amc "kubedb.dev/apimachinery/pkg/controller"

	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"kmodules.xyz/client-go/tools/queue"
	dbcs "kubedb.dev/apimachinery/client/clientset/versioned"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	scs "stash.appscode.dev/apimachinery/client/clientset/versioned"
	informers "stash.appscode.dev/apimachinery/client/informers/externalversions"
	lister "stash.appscode.dev/apimachinery/client/listers/stash/v1beta1"
)

type Stash struct {
	KubeClient      kubernetes.Interface
	StashClient     scs.Interface
	DBClient        dbcs.Interface
	InformerFactory informers.SharedInformerFactory
	// Stash RestoreSession
	RSQueue    *queue.Worker
	RSInformer cache.SharedIndexInformer
	RSLister   lister.RestoreSessionLister

	// Stash RestoreBatch
	RBQueue    *queue.Worker
	RBInformer cache.SharedIndexInformer
	RBLister   lister.RestoreBatchLister
	// SnapshotDoer interface
	snapshotter amc.DBHelper
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
	// restoreSession Lister
	WatchNamespace string
}

type restoreInfo struct {
	invoker      core.TypedLocalObjectReference
	namespace    string
	target       *v1beta1.RestoreTarget
	phase        v1beta1.RestorePhase
	targetDBKind string
}

func (s *Stash) Configure(cfg *rest.Config, resyncPeriod time.Duration) error {
	var err error
	if s.StashClient, err = scs.NewForConfig(cfg); err != nil {
		return err
	}
	s.InformerFactory = informers.NewSharedInformerFactory(s.StashClient, resyncPeriod)
	return nil
}

func (s *Stash) InitWatcher(maxNumRequeues, numThreads int, selector labels.Selector) {
	// Initialize RestoreSession Watcher
	s.RSInformer = s.restoreSessionInformer()
	s.RSQueue = queue.New(v1beta1.ResourceKindRestoreSession, maxNumRequeues, numThreads, s.processRestoreSession)
	s.RSLister = s.InformerFactory.Stash().V1beta1().RestoreSessions().Lister()
	s.RSInformer.AddEventHandler(s.restoreSessionEventHandler(selector))

	// Initialize RestoreBatch Watcher
	s.RBInformer = s.restoreBatchInformer()
	s.RBQueue = queue.New(v1beta1.ResourceKindRestoreBatch, maxNumRequeues, numThreads, s.processRestoreBatch)
	s.RBLister = s.InformerFactory.Stash().V1beta1().RestoreBatches().Lister()
	s.RBInformer.AddEventHandler(s.restoreBatchEventHandler(selector))
}

func (s Stash) StartController(stopCh <-chan struct{}) {
	// start StashInformerFactory only if stash crds (ie, "restoreSession") are available.
	if err := s.waitUntilStashInstalled(stopCh); err != nil {
		log.Errorln("error while waiting for restoreSession.", err)
		return
	}

	// start informer factory
	s.InformerFactory.Start(stopCh)
	// wait for cache to sync
	for t, v := range s.InformerFactory.WaitForCacheSync(stopCh) {
		if !v {
			log.Fatalf("%v timed out waiting for caches to sync", t)
			return
		}
	}
	// run the queues
	s.RSQueue.Run(stopCh)
	s.RBQueue.Run(stopCh)
}
