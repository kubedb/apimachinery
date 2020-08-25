/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restorebatch

import (
	"context"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
)

func (c *Controller) handleRestoreBatch(rb *v1beta1.RestoreBatch) error {
	if rb.Status.Phase != v1beta1.RestoreSucceeded && rb.Status.Phase != v1beta1.RestoreFailed {
		log.Debugf("restoreBatch %v/%v is not any of 'succeeded' or 'failed' ", rb.Namespace, rb.Name)
		return nil
	}

	if len(rb.Spec.Members) == 0 {
		log.Debugf("restoreBatch %v/%v does not have spec.member set. ", rb.Namespace, rb.Name)
		return nil
	}

	var meta metav1.ObjectMeta

	switch rb.Labels[api.LabelDatabaseKind] {
		case api.ResourceKindRedis:
			sts, err := c.Client.AppsV1().StatefulSets(rb.Namespace).Get(context.TODO(), rb.Spec.Members[0].Target.Ref.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			meta.Name = sts.OwnerReferences[0].Name

		default:
			meta = metav1.ObjectMeta{
				Name: rb.Spec.Members[0].Target.Ref.Name,
			}
	}

	meta.Namespace = rb.Namespace

	var phase api.DatabasePhase
	var reason string
	if rb.Status.Phase == v1beta1.RestoreSucceeded {
		phase = api.DatabasePhaseRunning
		if err := c.snapshotter.UpsertDatabaseAnnotation(meta, map[string]string{
			api.AnnotationInitialized: "",
		}); err != nil {
			return err
		}
	} else {
		phase = api.DatabasePhaseFailed
		reason = "Failed to complete initialization"
	}
	if err := c.snapshotter.SetDatabaseStatus(meta, phase, reason); err != nil {
		return err
	}

	runtimeObj, err := c.snapshotter.GetDatabase(meta)
	if err != nil {
		log.Errorln(err)
		return nil
	}
	if rb.Status.Phase == v1beta1.RestoreSucceeded {
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.eventRecorder.Event(
			runtimeObj,
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}
