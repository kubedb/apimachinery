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

	"kubedb.dev/apimachinery/apis/kubedb"

	"k8s.io/klog/v2"
	coreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// RestoreSessionReconciler reconciles a RestoreSession object
type RestoreSessionReconciler struct {
	ctrl *Controller
}

func (r *RestoreSessionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling: " + req.String())
	c := r.ctrl.KBClient

	rs := &coreapi.RestoreSession{}
	if err := c.Get(ctx, req.NamespacedName, rs); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if shouldIgnoreReconcile(rs) {
		return ctrl.Result{}, nil
	}

	ri, err := r.ctrl.extractRestoreInfo(rs)
	if err != nil {
		klog.Errorln("failed to extract kubeStash invoker info. Reason: ", err)
		return ctrl.Result{}, err
	}
	if rs.DeletionTimestamp != nil {
		return ctrl.Result{}, r.ctrl.handleTerminateEvent(ri)
	}

	return ctrl.Result{}, r.ctrl.handleRestoreInvokerEvent(ri)
}

func shouldIgnoreReconcile(rs *coreapi.RestoreSession) bool {
	if rs.Spec.Target != nil {
		return false
	}

	return !rs.IsApplicationLevelRestore()
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestoreSessionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&coreapi.RestoreSession{}, builder.WithPredicates(
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				rs, ok := object.(*coreapi.RestoreSession)
				if !ok {
					return false
				}
				return rs.Spec.Target != nil && rs.Spec.Target.APIGroup == kubedb.GroupName
			}),
		)).
		Named(coreapi.GroupVersion.WithKind(coreapi.ResourceKindRestoreSession).GroupKind().String()).
		Complete(r)
}
