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

package stash

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	"gomodules.xyz/pointer"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/reference"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	cutil "kmodules.xyz/client-go/conditions"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog"
	ab "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	sapis "stash.appscode.dev/apimachinery/apis"
	"stash.appscode.dev/apimachinery/apis/stash"
	"stash.appscode.dev/apimachinery/apis/stash/v1alpha1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	"stash.appscode.dev/apimachinery/pkg/invoker"
)

func (c *Controller) extractRestoreInfo(inv interface{}) (*restoreInfo, error) {
	ri := &restoreInfo{
		invoker: core.TypedLocalObjectReference{
			APIGroup: pointer.StringP(stash.GroupName),
		},
		do: dmcond.DynamicOptions{
			Client: c.DynamicClient,
		},
	}
	var err error
	switch inv := inv.(type) {
	case *v1beta1.RestoreSession:
		// invoker information
		ri.invoker.Kind = inv.Kind
		ri.invoker.Name = inv.Name
		// target information
		ri.target = inv.Spec.Target
		// restore status
		ri.phase = inv.Status.Phase
		// database information
		ri.do.Namespace = inv.Namespace
		ri.invokerUID = inv.UID
	case *v1beta1.RestoreBatch:
		// invoker information
		ri.invoker.Kind = inv.Kind
		ri.invoker.Name = inv.Name
		// target information
		// RestoreBatch can have multiple targets. In this case, only the database related target's phase does matter.
		ri.target, err = c.identifyTarget(inv.Spec.Members, ri.do.Namespace)
		if err != nil {
			return ri, err
		}
		// restore status
		// RestoreBatch can have multiple targets. In this case, finding the appropriate target is necessary.
		ri.phase = getTargetPhase(inv.Status, ri.target)
		// database information
		ri.do.Namespace = inv.Namespace
		ri.invokerUID = inv.UID
	default:
		return ri, fmt.Errorf("unknown restore invoker type")
	}
	// Now, extract the respective database group,version,resource
	err = c.extractDatabaseInfo(ri)
	if err != nil {
		return nil, err
	}
	return ri, nil
}

func (c *Controller) handleTerminateEvent(ri *restoreInfo) error {
	if ri == nil {
		return fmt.Errorf("invalid restore information. it must not be nil")
	}
	// If the target could not be identified properly, we can't process further.
	if ri.target == nil {
		return fmt.Errorf("couldn't identify the restore target from invoker: %s/%s/%s", *ri.invoker.APIGroup, ri.invoker.Kind, ri.invoker.Name)
	}

	// If the RestoreSession is deleted before completion,
	// Set the DB's "DataRestored" condition status to "False".
	// If already "False", no need to update the reason and message.
	// Also remove the "DataRestoreStarted" condition, if any exists.
	if !isDataRestoreCompleted(ri) {
		_, conditions, err := ri.do.ReadConditions()
		if err != nil {
			return fmt.Errorf("failed to read conditions with %s", err.Error())
		}
		_, dbCond := cutil.GetCondition(conditions, api.DatabaseDataRestored)
		if dbCond == nil {
			dbCond = &kmapi.Condition{
				Type: api.DatabaseDataRestored,
			}
		}
		if dbCond.Status != metav1.ConditionFalse {
			dbCond.Status = metav1.ConditionFalse
			dbCond.Reason = api.DataRestoreInterrupted
			dbCond.Message = fmt.Sprintf("Data initializer %s %s/%s with UID %s has been deleted",
				ri.invoker.Kind,
				ri.do.Namespace,
				ri.invoker.Name,
				ri.invokerUID,
			)
		}

		conditions = cutil.RemoveCondition(conditions, api.DatabaseDataRestoreStarted)
		conditions = cutil.SetCondition(conditions, *dbCond)
		return ri.do.UpdateConditions(conditions)
	}
	return nil
}

func (c *Controller) handleRestoreInvokerEvent(ri *restoreInfo) error {
	if ri == nil {
		return fmt.Errorf("invalid restore information. it must not be nil")
	}
	// If the target could not be identified properly, we can't process further.
	if ri.target == nil {
		return fmt.Errorf("couldn't identify the restore target from invoker: %s/%s/%s", *ri.invoker.APIGroup, ri.invoker.Kind, ri.invoker.Name)
	}

	// If restore is successful or failed,
	// Remove: condition.Type="DataRestoreStarted"
	// Add: condition.Type="DataRestored" --> true/false
	if isDataRestoreCompleted(ri) {
		dbCond := kmapi.Condition{
			Type: api.DatabaseDataRestored,
		}
		if ri.phase == v1beta1.RestoreSucceeded {
			dbCond.Status = metav1.ConditionTrue
			dbCond.Reason = api.DatabaseSuccessfullyRestored
			dbCond.Message = fmt.Sprintf("Successfully restored data by initializer %s %s/%s with UID %s",
				ri.invoker.Kind,
				ri.do.Namespace,
				ri.invoker.Name,
				ri.invokerUID,
			)
		} else {
			dbCond.Status = metav1.ConditionFalse
			dbCond.Reason = api.FailedToRestoreData
			dbCond.Message = fmt.Sprintf("Failed to restore data by initializer %s %s/%s with UID %s."+
				"\nRun 'kubectl describe %s %s -n %s' for more details.",
				ri.invoker.Kind,
				ri.do.Namespace,
				ri.invoker.Name,
				ri.invokerUID,
				strings.ToLower(ri.invoker.Kind),
				ri.invoker.Name,
				ri.do.Namespace,
			)
		}

		_, conditions, err := ri.do.ReadConditions()
		if err != nil {
			return fmt.Errorf("failed to read conditions with %s", err.Error())
		}
		conditions = cutil.RemoveCondition(conditions, api.DatabaseDataRestoreStarted)
		conditions = cutil.SetCondition(conditions, dbCond)
		err = ri.do.UpdateConditions(conditions)
		if err != nil {
			return fmt.Errorf("failed to update conditions with %s", err.Error())
		}

		// Write data restore completion event to the respective database CR
		return c.writeRestoreCompletionEvent(ri.do, dbCond)
	}

	// Restore process has started
	// Add: "DataRestoreStarted" condition to the respective database CR.
	// Remove: "DataRestored" condition, if any.
	dbCond := kmapi.Condition{
		Type:    api.DatabaseDataRestoreStarted,
		Status:  metav1.ConditionTrue,
		Reason:  api.DataRestoreStartedByExternalInitializer,
		Message: fmt.Sprintf("Data restore started by initializer: %s/%s/%s with UID %s.", *ri.invoker.APIGroup, ri.invoker.Kind, ri.invoker.Name, ri.invokerUID),
	}
	_, conditions, err := ri.do.ReadConditions()
	if err != nil {
		return fmt.Errorf("failed to read conditions with %s", err.Error())
	}
	conditions = cutil.RemoveCondition(conditions, api.DatabaseDataRestored)
	conditions = cutil.SetCondition(conditions, dbCond)
	return ri.do.UpdateConditions(conditions)
}

func (c *Controller) identifyTarget(members []v1beta1.RestoreTargetSpec, namespace string) (*v1beta1.RestoreTarget, error) {
	// check if there is any AppBinding as target. if there any, this is the desired target.
	for i, m := range members {
		if m.Target != nil {
			ok, err := targetOfGroupKind(m.Target.Ref, appcat.GroupName, ab.ResourceKindApp)
			if err != nil {
				return nil, err
			}
			if ok {
				return members[i].Target, nil
			}
		}
	}
	// no AppBinding has found as target. the target might be resulting workload (i.e. StatefulSet or Deployment(for memcached)).
	// we should check the workload's owner reference to be sure.
	for i, m := range members {
		if m.Target != nil {
			ok, err := targetOfGroupKind(m.Target.Ref, apps.GroupName, sapis.KindStatefulSet)
			if err != nil {
				return nil, err
			}
			if ok {
				sts, err := c.Client.AppsV1().StatefulSets(namespace).Get(context.Background(), m.Target.Ref.Name, metav1.GetOptions{})
				if err != nil {
					return nil, err
				}
				// if the controller owner is a KubeDB resource, then this StatefulSet must be the desired target
				ok, _, err := core_util.IsOwnerOfGroup(metav1.GetControllerOf(sts), kubedb.GroupName)
				if err != nil {
					return nil, err
				}
				if ok {
					return members[i].Target, nil
				}
			}
		}
	}
	return nil, nil
}

func getTargetPhase(status v1beta1.RestoreBatchStatus, target *v1beta1.RestoreTarget) v1beta1.RestorePhase {
	if target != nil {
		for _, m := range status.Members {
			if invoker.TargetMatched(m.Ref, target.Ref) {
				return v1beta1.RestorePhase(m.Phase)
			}
		}
	}
	return status.Phase
}

// waitUntilStashInstalled waits for Stash operator to be installed. It check whether all the CRDs that are necessary for backup KubeDB database,
// is present in the cluster or not. It wait until all the CRDs are found.
func (c *Controller) waitUntilStashInstalled(stopCh <-chan struct{}) error {
	klog.Infoln("Looking for the Stash operator.......")
	return wait.PollImmediateUntil(time.Second*10, func() (bool, error) {
		return discovery.ExistsGroupKinds(c.Client.Discovery(),
			schema.GroupKind{Group: stash.GroupName, Kind: v1alpha1.ResourceKindRepository},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindBackupConfiguration},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindBackupSession},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindBackupBlueprint},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindRestoreSession},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindRestoreBatch},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindTask},
			schema.GroupKind{Group: stash.GroupName, Kind: v1beta1.ResourceKindFunction},
		), nil
	}, stopCh)
}

func (c *Controller) extractDatabaseInfo(ri *restoreInfo) error {
	if ri == nil {
		return fmt.Errorf("invalid restoreInfo. It must not be nil")
	}
	if ri.target == nil {
		return fmt.Errorf("invalid target. It must not be nil")
	}
	var owner *metav1.OwnerReference
	if matched, err := targetOfGroupKind(ri.target.Ref, appcat.GroupName, ab.ResourceKindApp); err == nil && matched {
		appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(ri.do.Namespace).Get(context.TODO(), ri.target.Ref.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		owner = metav1.GetControllerOf(appBinding)
	} else if matched, err := targetOfGroupKind(ri.target.Ref, apps.GroupName, sapis.KindStatefulSet); err == nil && matched {
		sts, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(ri.do.Namespace).Get(context.TODO(), ri.target.Ref.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		owner = metav1.GetControllerOf(sts)
	}
	if owner == nil {
		return fmt.Errorf("failed to extract database information from the target info. Reason: target does not have controlling owner")
	}
	ri.do.Name = owner.Name

	gvr, err := c.Mapper.GVR(schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind))
	if err != nil {
		return err
	}
	ri.do.GVR = gvr

	return nil
}

func targetOfGroupKind(target v1beta1.TargetRef, group, kind string) (bool, error) {
	gv, err := schema.ParseGroupVersion(target.APIVersion)
	if err != nil {
		return false, err
	}
	return gv.Group == group && target.Kind == kind, nil
}

func (c *Controller) writeRestoreCompletionEvent(do dmcond.DynamicOptions, cond kmapi.Condition) error {
	// Get the database CR
	resp, err := do.Client.Resource(do.GVR).Namespace(do.Namespace).Get(context.TODO(), do.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Create database CR's reference
	ref, err := reference.GetReference(scheme.Scheme, resp)
	if err != nil {
		return err
	}

	eventType := core.EventTypeNormal
	if cond.Status != metav1.ConditionTrue {
		eventType = core.EventTypeWarning
	}
	// create event
	c.Recorder.Eventf(ref, eventType, cond.Reason, cond.Message)
	return nil
}

func isDataRestoreCompleted(ri *restoreInfo) bool {
	return ri.phase == v1beta1.RestoreSucceeded ||
		ri.phase == v1beta1.RestoreFailed ||
		ri.phase == v1beta1.RestorePhaseUnknown
}
