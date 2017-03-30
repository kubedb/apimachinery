package controller

import (
	"time"

	"github.com/appscode/go/wait"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kbatch "k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type Snapshotter interface {
	Validate(*tapi.DatabaseSnapshot) error
	GetObjects(*tapi.DatabaseSnapshot) (*kbatch.Job, *kapi.PersistentVolumeClaim, error)
	Destroy(*tapi.DatabaseSnapshot) error
}

type DatabaseSnapshotController struct {
	// Kubernetes client
	client clientset.Interface
	// ThirdPartyExtension client
	extClient tcs.ExtensionInterface
	// Deleter interface
	snapshoter Snapshotter
	// ListerWatcher
	lw *cache.ListWatch
	// sync time to sync the list.
	syncPeriod time.Duration
}

// NewDatabaseSnapshotController creates a new NewDatabaseSnapshot Controller
func NewDatabaseSnapshotController(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	snapshoter Snapshotter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DatabaseSnapshotController {
	// return new DeletedDatabase Controller
	return &DatabaseSnapshotController{
		client:     client,
		extClient:  extClient,
		snapshoter: snapshoter,
		lw:         lw,
		syncPeriod: syncPeriod,
	}
}

func (c *DatabaseSnapshotController) Run() {
	// Ensure DeletedDatabase TPR
	c.ensureThirdPartyResource()
	// Watch DeletedDatabase with provided ListerWatcher
	c.watch()
}

// Ensure DatabaseSnapshot ThirdPartyResource
func (w *DatabaseSnapshotController) ensureThirdPartyResource() {
	log.Infoln("Ensuring DatabaseSnapshot ThirdPartyResource")

	resourceName := tapi.ResourceNameDatabaseSnapshot + "." + tapi.V1beta1SchemeGroupVersion.Group

	if _, err := w.client.Extensions().ThirdPartyResources().Get(resourceName); err != nil {
		if !k8serr.IsNotFound(err) {
			log.Fatalln(err)
		}
	} else {
		thirdPartyResource := &extensions.ThirdPartyResource{
			TypeMeta: unversioned.TypeMeta{
				APIVersion: "extensions/v1beta1",
				Kind:       "ThirdPartyResource",
			},
			ObjectMeta: kapi.ObjectMeta{
				Name: resourceName,
			},
			Versions: []extensions.APIVersion{
				{
					Name: tapi.V1beta1SchemeGroupVersion.Version,
				},
			},
		}
		if _, err := w.client.Extensions().ThirdPartyResources().Create(thirdPartyResource); err != nil {
			log.Fatalln(err)
		}
	}
}

func (c *DatabaseSnapshotController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.DatabaseSnapshot{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dbSnapshot := obj.(*tapi.DatabaseSnapshot)
				/*
					TODO: set appropriate checking
					We do not want to handle same TPR objects multiple times
				*/
				if true {
					c.create(dbSnapshot)
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *DatabaseSnapshotController) create(dbSnapshot *tapi.DatabaseSnapshot) {
	if err := c.snapshoter.Validate(dbSnapshot); err != nil {

	}
}
