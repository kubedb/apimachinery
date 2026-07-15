/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deletion

import (
	"context"

	"kubedb.dev/apimachinery/apis"

	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	clientutil "kmodules.xyz/client-go/client"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// EnsureFinalizer adds the kubedb finalizer to the DB if it is not already present.
// Call it from the operator's reconcile path once the DB is admitted.
func EnsureFinalizer(ctx context.Context, kbClient client.Client, db client.Object) error {
	if controllerutil.ContainsFinalizer(db, apis.Finalizer) {
		return nil
	}
	_, err := clientutil.Patch(ctx, kbClient, db, func(obj client.Object) client.Object {
		controllerutil.AddFinalizer(obj, apis.Finalizer)
		return obj
	})
	if kerr.IsNotFound(err) {
		// The DB was deleted between the caller's check and this patch; nothing to add a
		// finalizer to.
		return err
	}
	return errors.Wrap(err, "failed to add finalizer")
}

// RemoveFinalizer removes the kubedb finalizer from the DB if present, allowing the object
// to be garbage collected. Call it from the operator's terminate path after Do has run.
func RemoveFinalizer(ctx context.Context, kbClient client.Client, db client.Object) error {
	if !controllerutil.ContainsFinalizer(db, apis.Finalizer) {
		return nil
	}
	_, err := clientutil.Patch(ctx, kbClient, db, func(obj client.Object) client.Object {
		controllerutil.RemoveFinalizer(obj, apis.Finalizer)
		return obj
	})
	if kerr.IsNotFound(err) {
		// Already gone; the finalizer's job of unblocking deletion is done.
		return nil
	}
	return errors.Wrap(err, "failed to remove finalizer")
}
