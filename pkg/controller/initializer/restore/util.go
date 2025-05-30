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

package restore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
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
	addonapi "kubestash.dev/apimachinery/apis/addons/v1alpha1"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
	sapis "stash.appscode.dev/apimachinery/apis"
	"stash.appscode.dev/apimachinery/apis/stash"
	"stash.appscode.dev/apimachinery/apis/stash/v1alpha1"
	"stash.appscode.dev/apimachinery/apis/stash/v1beta1"
	"stash.appscode.dev/apimachinery/pkg/invoker"
)

func (c *Controller) extractRestoreInfo(inv interface{}) (*restoreInfo, error) {
	ri := &restoreInfo{
		do: dmcond.DynamicOptions{
			Client: c.DynamicClient,
		},
	}
	var err error
	switch inv := inv.(type) {
	case *v1beta1.RestoreSession:
		// invoker information
		ri.invoker.APIGroup = pointer.StringP(stash.GroupName)
		ri.invoker.Kind = inv.Kind
		ri.invoker.Name = inv.Name

		// database information
		ri.do.Namespace = inv.Namespace
		ri.invokerUID = inv.UID

		// stash information
		ri.stash = &stashInfo{
			target: inv.Spec.Target,
			phase:  inv.Status.Phase,
		}
	case *v1beta1.RestoreBatch:
		// invoker information
		ri.invoker.APIGroup = pointer.StringP(stash.GroupName)
		ri.invoker.Kind = inv.Kind
		ri.invoker.Name = inv.Name

		// database information
		ri.do.Namespace = inv.Namespace
		ri.invokerUID = inv.UID

		// stash information
		// RestoreBatch can have multiple targets. In this case, only the database related target's phase does matter.
		info := &stashInfo{}
		if info.target, err = c.identifyTarget(inv.Spec.Members, ri.do.Namespace); err != nil {
			return nil, err
		}

		// restore status
		// RestoreBatch can have multiple targets. In this case, finding the appropriate target is necessary.
		info.phase = getTargetPhase(inv.Status, info.target)
		ri.stash = info
	case *coreapi.RestoreSession:
		// invoker information
		ri.invoker.APIGroup = pointer.StringP(storageapi.GroupVersion.Group)
		ri.invoker.Kind = inv.Kind
		ri.invoker.Name = inv.Name

		// database information
		ri.do.Namespace = inv.Namespace
		ri.invokerUID = inv.UID

		// kubestash information
		ri.kubestash = &kubestashInfo{
			phase:  inv.Status.Phase,
			target: inv.Spec.Target,
		}
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
	if ri.stash == nil && ri.kubestash == nil {
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
		_, dbCond := cutil.GetCondition(conditions, kubedb.DatabaseDataRestored)
		if dbCond == nil {
			dbCond = &kmapi.Condition{
				Type: kubedb.DatabaseDataRestored,
			}
		}
		if dbCond.Status != metav1.ConditionFalse {
			dbCond.Status = metav1.ConditionFalse
			dbCond.Reason = kubedb.DataRestoreInterrupted
			dbCond.Message = fmt.Sprintf("Data initializer %s %s/%s with UID %s has been deleted",
				ri.invoker.Kind,
				ri.do.Namespace,
				ri.invoker.Name,
				ri.invokerUID,
			)
		}

		conditions = cutil.RemoveCondition(conditions, kubedb.DatabaseDataRestoreStarted)
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
	if ri.stash == nil && ri.kubestash == nil {
		return fmt.Errorf("couldn't identify the restore target from invoker: %s/%s/%s", *ri.invoker.APIGroup, ri.invoker.Kind, ri.invoker.Name)
	}

	// If restore is successful or failed,
	// Remove: condition.Type="DataRestoreStarted"
	// Add: condition.Type="DataRestored" --> true/false
	if isDataRestoreCompleted(ri) {
		dbCond := kmapi.Condition{
			Type: kubedb.DatabaseDataRestored,
		}
		if (ri.stash != nil && ri.stash.phase == v1beta1.RestoreSucceeded) || (ri.kubestash != nil && ri.kubestash.phase == coreapi.RestoreSucceeded) {
			dbCond.Status = metav1.ConditionTrue
			dbCond.Reason = kubedb.DatabaseSuccessfullyRestored
			dbCond.Message = fmt.Sprintf("Successfully restored data by initializer %s %s/%s with UID %s",
				ri.invoker.Kind,
				ri.do.Namespace,
				ri.invoker.Name,
				ri.invokerUID,
			)
		} else {
			dbCond.Status = metav1.ConditionFalse
			dbCond.Reason = kubedb.FailedToRestoreData
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
		conditions = cutil.RemoveCondition(conditions, kubedb.DatabaseDataRestoreStarted)
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
		Type:    kubedb.DatabaseDataRestoreStarted,
		Status:  metav1.ConditionTrue,
		Reason:  kubedb.DataRestoreStartedByExternalInitializer,
		Message: fmt.Sprintf("Data restore started by initializer: %s/%s/%s with UID %s.", *ri.invoker.APIGroup, ri.invoker.Kind, ri.invoker.Name, ri.invokerUID),
	}
	_, conditions, err := ri.do.ReadConditions()
	if err != nil {
		return fmt.Errorf("failed to read conditions with %s", err.Error())
	}
	conditions = cutil.RemoveCondition(conditions, kubedb.DatabaseDataRestored)
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
	return wait.PollUntilContextCancel(wait.ContextForChannel(stopCh), time.Second*10, true, func(ctx context.Context) (bool, error) {
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
	})
}

// waitUntilKubeStashInstalled waits for KubeStash operator to be installed. It check whether all the CRDs that are necessary for backup KubeDB database,
// is present in the cluster or not. It wait until all the CRDs are found.
func (c *Controller) waitUntilKubeStashInstalled(stopCh <-chan struct{}) error {
	klog.Infoln("Looking for the KubeStash operator.......")

	return wait.PollUntilContextCancel(wait.ContextForChannel(stopCh), time.Second*10, true, func(ctx context.Context) (bool, error) {
		return discovery.ExistsGroupKinds(c.Client.Discovery(),
			schema.GroupKind{Group: storageapi.GroupVersion.Group, Kind: storageapi.ResourceKindBackupStorage},
			schema.GroupKind{Group: storageapi.GroupVersion.Group, Kind: storageapi.ResourceKindRepository},
			schema.GroupKind{Group: storageapi.GroupVersion.Group, Kind: storageapi.ResourceKindSnapshot},
			schema.GroupKind{Group: storageapi.GroupVersion.Group, Kind: storageapi.ResourceKindRetentionPolicy},
			schema.GroupKind{Group: coreapi.GroupVersion.Group, Kind: coreapi.ResourceKindBackupBatch},
			schema.GroupKind{Group: coreapi.GroupVersion.Group, Kind: coreapi.ResourceKindBackupBlueprint},
			schema.GroupKind{Group: coreapi.GroupVersion.Group, Kind: coreapi.ResourceKindBackupConfiguration},
			schema.GroupKind{Group: coreapi.GroupVersion.Group, Kind: coreapi.ResourceKindBackupSession},
			schema.GroupKind{Group: coreapi.GroupVersion.Group, Kind: coreapi.ResourceKindRestoreSession},
			schema.GroupKind{Group: addonapi.GroupVersion.Group, Kind: addonapi.ResourceKindAddon},
			schema.GroupKind{Group: addonapi.GroupVersion.Group, Kind: addonapi.ResourceKindFunction},
		), nil
	})
}

func (c *Controller) extractDatabaseInfo(ri *restoreInfo) error {
	if ri == nil {
		return fmt.Errorf("invalid restoreInfo. It must not be nil")
	}

	// It is guaranteed that if either ri.stash or ri.kubestash is initialized, then 'target' field is also initialized.
	if ri.stash == nil && ri.kubestash == nil {
		return fmt.Errorf("invalid target. It must not be nil")
	}

	var owner *metav1.OwnerReference
	if ri.stash != nil {
		if matched, err := targetOfGroupKind(ri.stash.target.Ref, appcat.GroupName, ab.ResourceKindApp); err == nil && matched {
			appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(ri.do.Namespace).Get(context.TODO(), ri.stash.target.Ref.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			owner = metav1.GetControllerOf(appBinding)
		} else if matched, err := targetOfGroupKind(ri.stash.target.Ref, apps.GroupName, sapis.KindStatefulSet); err == nil && matched {
			sts, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(ri.do.Namespace).Get(context.TODO(), ri.stash.target.Ref.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			owner = metav1.GetControllerOf(sts)
		}
	} else {
		appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(ri.kubestash.target.Namespace).Get(context.TODO(), ri.kubestash.target.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		owner = metav1.GetControllerOf(appBinding)
		ri.do.Namespace = appBinding.Namespace
	}

	if owner == nil {
		return fmt.Errorf("failed to extract database information from the target info. Reason: target does not have controlling owner")
	}
	ri.do.Name = owner.Name

	gvk := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
	mapping, err := c.KBClient.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}
	ri.do.GVR = mapping.Resource

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
	if ri.stash != nil {
		return ri.stash.phase == v1beta1.RestoreSucceeded ||
			ri.stash.phase == v1beta1.RestoreFailed ||
			ri.stash.phase == v1beta1.RestorePhaseUnknown
	} else {
		return ri.kubestash.phase == coreapi.RestoreSucceeded ||
			ri.kubestash.phase == coreapi.RestoreFailed ||
			ri.kubestash.phase == coreapi.RestorePhaseUnknown
	}
}
