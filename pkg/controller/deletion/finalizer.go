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

func EnsureFinalizer(ctx context.Context, kbClient client.Client, db client.Object) error {
	if controllerutil.ContainsFinalizer(db, apis.Finalizer) {
		return nil
	}
	_, err := clientutil.Patch(ctx, kbClient, db, func(obj client.Object) client.Object {
		controllerutil.AddFinalizer(obj, apis.Finalizer)
		return obj
	})
	if kerr.IsNotFound(err) {
		return nil
	}
	return errors.Wrap(err, "failed to add finalizer")
}

func RemoveFinalizer(ctx context.Context, kbClient client.Client, db client.Object) error {
	if !controllerutil.ContainsFinalizer(db, apis.Finalizer) {
		return nil
	}
	_, err := clientutil.Patch(ctx, kbClient, db, func(obj client.Object) client.Object {
		controllerutil.RemoveFinalizer(obj, apis.Finalizer)
		return obj
	})
	if kerr.IsNotFound(err) {
		return nil
	}
	return errors.Wrap(err, "failed to remove finalizer")
}
