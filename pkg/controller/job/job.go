package job

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) completeJob(job *batch.Job) error {

	snapshotName, found := job.Annotations[api.AnnotationSnapshotName]
	if !found {
		return fmt.Errorf(`invalid Job "%s/%s" to handle`, job.Namespace, job.Name)
	}
	snapshot, err := c.ExtClient.Snapshots(job.Namespace).Get(snapshotName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	jobSucceeded := job.Status.Succeeded > 0

	_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
		t := metav1.Now()
		if jobSucceeded {
			in.Status.Phase = api.SnapshotPhaseSucceeded
		} else {
			in.Status.Phase = api.SnapshotPhaseFailed
		}
		in.Status.CompletionTime = &t
		delete(in.Labels, api.LabelSnapshotStatus)
		return in
	})
	if err != nil {
		c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	dbRuntineObject, err := c.job.GetDatabase(snapshot)
	if err != nil {
		return err
	}

	c.DeleteJobResources(snapshot, job, dbRuntineObject, c.eventRecorder)

	return nil
}
