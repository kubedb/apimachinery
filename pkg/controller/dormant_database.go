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

type DormantDbController struct {
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

// NewDormantDbController creates a new DormantDatabase Controller
func NewDormantDbController(
	client clientset.Interface,
	extClient tcs.ExtensionInterface,
	deleter Deleter,
	lw *cache.ListWatch,
	syncPeriod time.Duration,
) *DormantDbController {
	// return new DormantDatabase Controller
	return &DormantDbController{
		client:        client,
		extClient:     extClient,
		deleter:       deleter,
		lw:            lw,
		eventRecorder: eventer.NewEventRecorder(client, "DormantDatabase Controller"),
		syncPeriod:    syncPeriod,
	}
}

func (c *DormantDbController) Run() {
	// Ensure DormantDatabase TPR
	c.ensureThirdPartyResource()
	// Watch DormantDatabase with provided ListerWatcher
	c.watch()
}

// Ensure DormantDatabase ThirdPartyResource
func (c *DormantDbController) ensureThirdPartyResource() {
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

func (c *DormantDbController) watch() {
	_, cacheController := cache.NewInformer(c.lw,
		&tapi.DormantDatabase{},
		c.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				dormantDb := obj.(*tapi.DormantDatabase)
				if dormantDb.Status.CreationTime == nil {
					if err := c.create(dormantDb); err != nil {
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
				oldDormantDb, ok := old.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				newDormantDb, ok := new.(*tapi.DormantDatabase)
				if !ok {
					return
				}
				// TODO: Find appropriate checking
				// Only allow if Spec varies
				if !reflect.DeepEqual(oldDormantDb.Spec, newDormantDb.Spec) {
					if err := c.update(oldDormantDb, newDormantDb); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	cacheController.Run(wait.NeverStop)
}

func (c *DormantDbController) create(dormantDb *tapi.DormantDatabase) error {

	var err error
	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleting
	t := unversioned.Now()
	dormantDb.Status.CreationTime = &t
	if _, err := c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
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
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name); err != nil {
			c.eventRecorder.Eventf(
				dormantDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleting
	t = unversioned.Now()
	dormantDb.Status.Phase = tapi.DormantDatabasePhaseStopping
	if _, err = c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(dormantDb, kapi.EventTypeNormal, eventer.EventReasonDeleting, "Deleting Database")

	// Delete Database workload
	if err := c.deleter.DeleteDatabase(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			"Failed to delete. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(
		dormantDb,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulDelete,
		"Successfully deleted Database workload",
	)

	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleted
	t = unversioned.Now()
	dormantDb.Status.DeletionTime = &t
	dormantDb.Status.Phase = tapi.DormantDatabasePhaseStopped
	if _, err = c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	return nil
}

func (c *DormantDbController) delete(dormantDb *tapi.DormantDatabase) error {
	if dormantDb.Status.Phase != tapi.DormantDatabasePhaseWipedOut {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToDelete,
			`DormantDatabase "%v" is not %v.`,
			dormantDb.Name,
			tapi.DormantDatabasePhaseWipedOut,
		)

		_dormantDb := &tapi.DormantDatabase{
			ObjectMeta: kapi.ObjectMeta{
				Name:        dormantDb.Name,
				Namespace:   dormantDb.Namespace,
				Labels:      dormantDb.Labels,
				Annotations: dormantDb.Annotations,
			},
			Spec:   dormantDb.Spec,
			Status: dormantDb.Status,
		}

		if _, err := c.extClient.DormantDatabases(_dormantDb.Namespace).Create(_dormantDb); err != nil {
			c.eventRecorder.Eventf(
				dormantDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate DormantDatabase: "%v". Reason: %v`,
				dormantDb.Name,
				err,
			)
			return err
		}
	}
	return nil
}

func (c *DormantDbController) update(oldDormantDb, updatedDormantDb *tapi.DormantDatabase) error {
	if oldDormantDb.Spec.WipeOut != updatedDormantDb.Spec.WipeOut && updatedDormantDb.Spec.WipeOut {
		return c.wipeOut(updatedDormantDb)
	}

	if oldDormantDb.Spec.Resume != updatedDormantDb.Spec.Resume && updatedDormantDb.Spec.Resume {
		if oldDormantDb.Status.Phase == tapi.DormantDatabasePhaseStopped {
			return c.recover(updatedDormantDb)
		} else {
			message := "Failed to recover Database. " +
				"Only DormantDatabase of \"Deleted\" Phase can be recovered"
			c.eventRecorder.Event(
				updatedDormantDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				message,
			)
		}
	}
	return nil
}

func (c *DormantDbController) wipeOut(dormantDb *tapi.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
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
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			message,
		)

		// Delete DormantDatabase object
		if err := c.extClient.DormantDatabases(dormantDb.Namespace).Delete(dormantDb.Name); err != nil {
			c.eventRecorder.Eventf(
				dormantDb,
				kapi.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete DormantDatabase. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
		return errors.New(message)
	}

	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Wiping out
	t := unversioned.Now()
	dormantDb.Status.Phase = tapi.DormantDatabasePhaseWipingOut

	if _, err := c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	// Wipe out Database workload
	c.eventRecorder.Event(dormantDb, kapi.EventTypeNormal, eventer.EventReasonWipingOut, "Wiping out Database")
	if err := c.deleter.WipeOutDatabase(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToWipeOut,
			"Failed to wipeOut. Reason: %v",
			err,
		)
		return err
	}

	c.eventRecorder.Event(
		dormantDb,
		kapi.EventTypeNormal,
		eventer.EventReasonSuccessfulWipeOut,
		"Successfully wiped out Database workload",
	)

	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	// Set DormantDatabase Phase: Deleted
	t = unversioned.Now()
	dormantDb.Status.WipeOutTime = &t
	dormantDb.Status.Phase = tapi.DormantDatabasePhaseWipedOut
	if _, err = c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	return nil
}

func (c *DormantDbController) recover(dormantDb *tapi.DormantDatabase) error {
	// Check if DB TPR object exists
	found, err := c.deleter.Exists(&dormantDb.ObjectMeta)
	if err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
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
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToRecover,
			message,
		)
		return errors.New(message)
	}

	if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
		return err
	}

	dormantDb.Status.Phase = tapi.DormantDatabasePhaseResuming
	if _, err = c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			"Failed to update DormantDatabase. Reason: %v",
			err,
		)
		return err
	}

	if err = c.deleter.ResumeDatabase(dormantDb); err != nil {
		c.eventRecorder.Eventf(
			dormantDb,
			kapi.EventTypeWarning,
			eventer.EventReasonFailedToRecover,
			"Failed to recover Database. Reason: %v",
			err,
		)

		if dormantDb, err = c.extClient.DormantDatabases(dormantDb.Namespace).Get(dormantDb.Name); err != nil {
			return err
		}

		dormantDb.Status.Phase = tapi.DormantDatabasePhaseStopped
		if _, err = c.extClient.DormantDatabases(dormantDb.Namespace).Update(dormantDb); err != nil {
			c.eventRecorder.Eventf(
				dormantDb,
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
