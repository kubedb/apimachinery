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
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"reflect"
)

type Deleter interface {
	// Check Database TPR
	Exists(*tapi.DeletedDatabase) bool
	// Delete operation
	Delete(*tapi.DeletedDatabase) error
	// Destroy operation
	Destroy(*tapi.DeletedDatabase) error
}

type DeletedDatabaseController struct {
	// Kubernetes client
	client clientset.Interface
	// ThirdPartyExtension client
	extClient tcs.ExtensionInterface
	// Deleter interface
	deleter Deleter
	// ListerWatcher
	lw *cache.ListWatch
	// sync time to sync the list.
	syncPeriod time.Duration
}

// NewDeletedDbController creates a new DeletedDatabase Controller
func NewDeletedDbController(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DeletedDatabaseController {
	// return new DeletedDatabase Controller
	return &DeletedDatabaseController{
		client:     client,
		extClient:  extClient,
		deleter:    deleter,
		lw:         lw,
		syncPeriod: syncPeriod,
	}
}

func (c *DeletedDatabaseController) Run() {
	// Ensure DeletedDatabase TPR
	c.ensureThirdPartyResource()
	// Watch DeletedDatabase with provided ListerWatcher
	c.watch()
}

// Ensure DeletedDatabase ThirdPartyResource
func (c *DeletedDatabaseController) ensureThirdPartyResource() {
	log.Infoln("Ensuring DeletedDatabase ThirdPartyResource")

	resourceName := tapi.ResourceNameDeletedDatabase + "." + tapi.V1beta1SchemeGroupVersion.Group

	if _, err := c.client.Extensions().ThirdPartyResources().Get(resourceName); err != nil {
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
		if _, err := c.client.Extensions().ThirdPartyResources().Create(thirdPartyResource); err != nil {
			log.Fatalln(err)
		}
	}
}

func (c *DeletedDatabaseController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.DeletedDatabase{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				deletedDb := obj.(*tapi.DeletedDatabase)
				/*
					TODO: set appropriate checking
					We do not want to handle same TPR objects multiple times
				*/
				if true {
					c.create(deletedDb)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldDeletedDb, ok := old.(*tapi.DeletedDatabase)
				if !ok {
					return
				}
				newDeletedDb, ok := new.(*tapi.DeletedDatabase)
				if !ok {
					return
				}
				// TODO: Find appropriate checking
				// Only allow if Spec varies
				if !reflect.DeepEqual(oldDeletedDb.Spec, newDeletedDb.Spec) {
					c.update(newDeletedDb)
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *DeletedDatabaseController) create(deletedDb *tapi.DeletedDatabase) {
	// Check if DB TPR object exists
	if c.deleter.Exists(deletedDb) {
		/*
			TODO: Record event in DB TPR
			// Failed to delete. Reason: Delete DB TPR.
		*/

		// Delete DeletedDatabase object
		if err := c.extClient.DeletedDatabases(deletedDb.Namespace).Delete(deletedDb.Name); err != nil {
			// TODO: Do we need event for this?? ask @admin
			log.Errorln(err)
		}
		return
	}

	/*
		TODO: Record event in DDB TPR
		// Deleting
	*/

	// Delete Database workload
	if err := c.deleter.Delete(deletedDb); err != nil {
		/*
			TODO: Record event in DDB TPR
			// Failed to delete. Reason: err
		*/
		log.Errorln(err)
		return
	}

	/*
		TODO: Record event in DDB TPR
		// Successfully deleted
	*/

	/*
		TODO: Discuss with @admin
		// I think we do not need to check destroy.
		// Because, user can't create DeletedDatabase obbject manually.
		// It will always be created with Destroy=False
	*/
	// Destroy Database workload
	if deletedDb.Spec.Destroy {
		/*
			TODO: Record event in DDB TPR
			// Destroying
		*/
		if err := c.deleter.Destroy(deletedDb); err != nil {
			/*
				TODO: Record event in DDB TPR
				// Failed to destroy. Reason: err
			*/
			log.Errorln(err)
			return
		}
		/*
			TODO: Record event in DDB TPR
			// Successfully destroyed
		*/
	}

	return
}

func (c *DeletedDatabaseController) update(deletedDb *tapi.DeletedDatabase) {
	// Check if DB TPR object exists
	if c.deleter.Exists(deletedDb) {
		// TODO: Record event in DB TPR
		// Delete DeletedDatabase object
		if err := c.extClient.DeletedDatabases(deletedDb.Namespace).Delete(deletedDb.Name); err != nil {
			// TODO: Do we need event for this?? ask @admin
			log.Errorln(err)
		}
		return
	}

	// Destroy Database workload
	if deletedDb.Spec.Destroy {
		/*
			TODO: Record event in DDB TPR
			// Destroying
		*/
		if err := c.deleter.Destroy(deletedDb); err != nil {
			/*
				TODO: Record event in DDB TPR
				// Failed to destroy. Reason: err
			*/
			log.Errorln(err)
			return
		}
		/*
			TODO: Record event in DDB TPR
			// Successfully destroyed
		*/
	}
}
