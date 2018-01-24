package job

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) completeJob(job *batch.Job) error {
	jobType := job.Annotations[api.AnnotationJobType]
	if jobType == api.SnapshotProcessBackup {
		if err := c.handleBackupJob(job); err != nil {
			return err
		}
	} else if jobType == api.SnapshotProcessRestore {
		if err := c.handleRestoreJob(job); err != nil {
			return err
		}
	}

	deletePolicy := metav1.DeletePropagationBackground
	err := c.Client.BatchV1().Jobs(job.Namespace).Delete(job.Name, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil && !kerr.IsNotFound(err) {
		return fmt.Errorf("failed to delete job: %s, reason: %s", job.Name, err)
	}

	return nil
}

func (c *Controller) handleBackupJob(job *batch.Job) error {
	for _, o := range job.OwnerReferences {
		if o.Kind == api.ResourceKindSnapshot {
			snapshot, err := c.ExtClient.Snapshots(job.Namespace).Get(o.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			jobSucceeded := job.Status.Succeeded > 0

			_, _, err = util.PatchSnapshot(c.ExtClient, snapshot, func(in *api.Snapshot) *api.Snapshot {
				if jobSucceeded {
					in.Status.Phase = api.SnapshotPhaseSucceeded
				} else {
					in.Status.Phase = api.SnapshotPhaseFailed
				}
				t := metav1.Now()
				in.Status.CompletionTime = &t
				delete(in.Labels, api.LabelSnapshotStatus)
				return in
			})
			if err != nil {
				c.eventRecorder.Eventf(snapshot.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
				return err
			}
			return nil
		}
	}
	return fmt.Errorf(`resource Job "%s/%s" doesn't have OwnerReference for Snapshot`, job.Namespace, job.Name)
}

func (c *Controller) handleRestoreJob(job *batch.Job) error {
	for _, o := range job.OwnerReferences {
		if o.Kind == job.Labels[api.LabelDatabaseKind] {
			var phase api.DatabasePhase
			var reason string
			if job.Status.Succeeded > 0 {
				phase = api.DatabasePhaseRunning
			} else {
				phase = api.DatabasePhaseFailed
				reason = "Failed to complete initialization"
			}
			err := c.snapshotDoer.SetDatabaseStatus(
				metav1.ObjectMeta{Name: o.Name, Namespace: job.Namespace},
				phase,
				reason,
			)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf(`resource Job "%s/%s" doesn't have OwnerReference for %s`, job.Namespace, job.Name, job.Labels[api.LabelDatabaseKind])
}
