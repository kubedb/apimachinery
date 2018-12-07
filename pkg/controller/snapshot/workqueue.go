package snapshot

import (
	"github.com/appscode/go/log"
	core_util "github.com/appscode/kutil/core/v1"
	"github.com/appscode/kutil/tools/queue"
	"github.com/kubedb/apimachinery/apis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Controller) addEventHandler(selector labels.Selector) {
	c.SnapQueue = queue.New("Snapshot", c.MaxNumRequeues, c.NumThreads, c.runSnapshot, c.pushFailureEvent)
	c.snLister = c.KubedbInformerFactory.Kubedb().V1alpha1().Snapshots().Lister()
	c.SnapInformer.AddEventHandler(queue.NewFilteredHandler(queue.NewEventHandler(c.SnapQueue.GetQueue(), func(old interface{}, new interface{}) bool {
		snapshot := new.(*api.Snapshot)
		return snapshot.DeletionTimestamp != nil
	}), selector))
}

func (c *Controller) runSnapshot(key string) error {
	log.Debugf("started processing, key: %v", key)
	obj, exists, err := c.SnapInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("Snapshot %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Snapshot was recreated with the same name
		snapshot := obj.(*api.Snapshot).DeepCopy()
		if snapshot.DeletionTimestamp != nil {
			if core_util.HasFinalizer(snapshot.ObjectMeta, api.GenericKey) {
				if err := c.delete(snapshot); err != nil {
					log.Errorln(err)
					return err
				}
				snapshot, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			snapshot, _, err = util.PatchSnapshot(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.Snapshot) *api.Snapshot {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err := c.create(snapshot); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}
	return nil
}

func (c *Controller) pushFailureEvent(key string, reason string) {
	obj, exists, err := c.SnapInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return
	}
	if !exists {
		log.Debugf("Snapshot %s does not exist anymore", key)
		return
	}

	snapshot := obj.(*api.Snapshot).DeepCopy()
	// write failure event to snapshot
	c.eventRecorder.Eventf(
		snapshot,
		core.EventTypeWarning,
		eventer.EventReasonFailedToStart,
		`Snapshot %v/%v failed. Reason: %v`,
		snapshot.Namespace, snapshot.Name,
		reason,
	)

	// write failure event to database crd object
	if runtimeObj, err := c.snapshotter.GetDatabase(metav1.ObjectMeta{
		Name:      snapshot.Spec.DatabaseName,
		Namespace: snapshot.Namespace,
	}); err == nil && runtimeObj != nil {
		c.eventRecorder.Eventf(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Snapshot %v/%v failed. Reason: %v`,
			snapshot.Namespace, snapshot.Name,
			reason,
		)
	}

	snap, err := util.UpdateSnapshotStatus(c.ExtClient.KubedbV1alpha1(), snapshot, func(in *api.SnapshotStatus) *api.SnapshotStatus {
		in.Phase = api.SnapshotPhaseFailed
		in.Reason = reason
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		c.eventRecorder.Eventf(
			snapshot,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
	}
	snapshot.Status = snap.Status
}
