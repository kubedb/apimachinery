package controller

import (
	"errors"
	"reflect"
	"time"

	"github.com/appscode/go/wait"
	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	kapi "k8s.io/kubernetes/pkg/api"
	k8serr "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/client/record"
)

type Deleter interface {
	// Check Database TPR
	Exists(*kapi.ObjectMeta) (bool, error)
	// Delete operation
	DeleteDatabase(*tapi.DormantDatabase) error
	// Wipe out operation
	WipeOutDatabase(*tapi.DormantDatabase) error
	// Resume operation
	ResumeDatabase(*tapi.DormantDatabase) error
}

type DormantDatabaseController struct {
	// Kubernetes client
	client clientset.Interface
	// ThirdPartyExtension client
	extClient tcs.ExtensionInterface
	// Deleter interface
	deleter Deleter
	// ListerWatcher
	lw *cache.ListWatch
	// Event Recorder
	eventRecorder record.EventRecorder
	// sync time to sync the list.
	syncPeriod time.Duration
}

// NewDeletedDbController creates a new DormantDatabase Controller
func NewDeletedDbController(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DormantDatabaseController {
	// return new DormantDatabase Controller
	return &DormantDatabaseController{
		client:        client,
		extClient:     extClient,
		deleter:       deleter,
		lw:            lw,
		eventRecorder: eventer.NewEventRecorder(client, "DormantDatabase Controller"),
		syncPeriod:    syncPeriod,
	}
}

func (c *DormantDatabaseController) Run() {
	// Ensure DormantDatabase TPR
	c.ensureThirdPartyResource()
	// Watch DormantDatabase with provided ListerWatcher
	c.watch()
}

// Ensure DormantDatabase ThirdPartyResource
func (c *DormantDatabaseController) ensureThirdPartyResource() {
	log.Infoln("Ensuring DormantDatabase ThirdPartyResource")

	resourceName := tapi.ResourceNameDormantDatabase + "." + tapi.V1beta1SchemeGroupVersion.Group
	var err error
	if _, err = c.client.Extensions().ThirdPartyResources().Get(resourceName); err == nil {
		return
	}
	if !k8serr.IsNotFound(err) {
		log.Fatalln(err)
	}

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

func (c *DormantDatabaseController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.DormantDatabase{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				deletedDb := obj.(*tapi.DormantDatabase)
				if deletedDb.Status.CreationTime == nil {
					if err := c.create(deletedDb); err != nil {
						log.Errorln(err)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if err := c.delete(obj.(*tapi.DormantDatabase)); err != nil {
					log.Errorln(err)
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldDeletedDb, ok := old.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				newDeletedDb, ok := new.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				// TODO: Find appropriate checking
				// Only allow if Spec varies
				if !reflect.DeepEqual(oldDeletedDb.Spec, newDeletedDb.Spec) {
					if err := c.update(oldDeletedDb, newDeletedDb); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *DormantDatabaseController) create(deletedDb *tapi.DormantDatabase) error {

	var err error
	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleting
	t := unversioned.Now()
	deletedDb.Status.CreationTime = &t
	if _, err := c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&deletedDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to delete Database. Delete Database TPR object first"
		c.eventRecorder.Event(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(deletedDb.Namespace).Delete(deletedDb.Name); err != nil {
			c.eventRecorder.Eventf(
				deletedDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleting
	t = unversioned.Now()
	deletedDb.Status.Phase = tapi.DormantDatabasePhaseDeleting
	if _, err = c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(deletedDb, kapi.EventTypeNormal, eventer.EventReasonDeleting, "Deleting Database")

	// Delete Database workload
	if err := c.deleter.DeleteDatabase(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(
		deletedDb,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulDelete,
		"Successfully deleted Database workload",
	)

	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleted
	t = unversioned.Now()
	deletedDb.Status.DeletionTime = &t
	deletedDb.Status.Phase = tapi.DormantDatabasePhaseDeleted
	if _, err = c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	return nil
}

func (c *DormantDatabaseController) delete(deletedDb *tapi.DormantDatabase) error {
	if deletedDb.Status.Phase != tapi.DormantDatabasePhaseWipedOut {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			`DormantDatabase "%v" is not %v.`,
			deletedDb.Name,
			tapi.DormantDatabasePhaseWipedOut,
		)

		_deletedDb := &tapi.DormantDatabase{
			ObjectMeta: kapi.ObjectMeta{
				Name:        deletedDb.Name,
				Namespace:   deletedDb.Namespace,
				Labels:      deletedDb.Labels,
				Annotations: deletedDb.Annotations,
			},
			Spec:   deletedDb.Spec,
			Status: deletedDb.Status,
		}

		if _, err := c.extClient.DormantDatabases(_deletedDb.Namespace).Create(_deletedDb); err != nil {
			c.eventRecorder.Eventf(
				deletedDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				deletedDb.Name,
				err,
			)
			return err
		}
	}
	return nil
}

func (c *DormantDatabaseController) update(oldDeletedDb, updatedDeletedDb *tapi.DormantDatabase) error {
	if oldDeletedDb.Spec.WipeOut != updatedDeletedDb.Spec.WipeOut && updatedDeletedDb.Spec.WipeOut {
		return c.wipeOut(updatedDeletedDb)
	}

	if oldDeletedDb.Spec.Resume != updatedDeletedDb.Spec.Resume && updatedDeletedDb.Spec.Resume {
		if oldDeletedDb.Status.Phase == tapi.DormantDatabasePhaseDeleted {
			return c.recover(updatedDeletedDb)
		} else {
			message := "Failed to recover Database. " +
				"Only DormantDatabase of \"Deleted\" Phase can be recovered"
			c.eventRecorder.Event(
				updatedDeletedDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				message,
			)
		}
	}
	return nil
}

func (c *DormantDatabaseController) wipeOut(deletedDb *tapi.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&deletedDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to wipeOut Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to wipeOut Database. Delete Database TPR object first"
		c.eventRecorder.Event(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(deletedDb.Namespace).Delete(deletedDb.Name); err != nil {
			c.eventRecorder.Eventf(
				deletedDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Wiping out
	t := unversioned.Now()
	deletedDb.Status.Phase = tapi.DormantDatabasePhaseWipingOut

	if _, err := c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	// Wipe out Database workload
	c.eventRecorder.Event(deletedDb, kapi.EventTypeNormal, eventer.EventReasonWipingOut, "Wiping out Database")
	if err := c.deleter.WipeOutDatabase(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			"Failed to wipeOut. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(
		deletedDb,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulWipeOut,
		"Successfully wiped out Database workload",
	)

	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleted
	t = unversioned.Now()
	deletedDb.Status.WipeOutTime = &t
	deletedDb.Status.Phase = tapi.DormantDatabasePhaseWipedOut
	if _, err = c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	return nil
}

func (c *DormantDatabaseController) recover(deletedDb *tapi.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&deletedDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToRecover,
			"Failed to recover Database. Reason: %v",
			err,
		)
		return err
	}

	if found {
		message := "Failed to recover Database. One Database TPR object exists with same name"
		c.eventRecorder.Event(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToRecover,
			message,
		)
		return errors.New(message)
	}

	if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
		return err
	}

	deletedDb.Status.Phase = tapi.DormantDatabasePhaseRecovering
	if _, err = c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if err = c.deleter.ResumeDatabase(deletedDb); err != nil {
		c.eventRecorder.Eventf(
			deletedDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToRecover,
			"Failed to recover Database. Reason: %v",
			err,
		)

		if deletedDb, err = c.extClient.DormantDatabases(deletedDb.Namespace).Get(deletedDb.Name); err != nil {
			return err
		}

		deletedDb.Status.Phase = tapi.DormantDatabasePhaseDeleted
		if _, err = c.extClient.DormantDatabases(deletedDb.Namespace).Update(deletedDb); err != nil {
			c.eventRecorder.Eventf(
				deletedDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}

		return err
	}

	return nil
}
