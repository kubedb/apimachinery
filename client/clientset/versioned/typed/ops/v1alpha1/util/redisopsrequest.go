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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	api "kubedb.dev/apimachinery/apis/ops/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/ops/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchRedisOpsRequest(ctx context.Context, c cs.OpsV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.RedisOpsRequest) *api.RedisOpsRequest, opts metav1.PatchOptions) (*api.RedisOpsRequest, kutil.VerbType, error) {
	cur, err := c.RedisOpsRequests(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating RedisOpsRequest %s/%s.", meta.Namespace, meta.Name)
		out, err := c.RedisOpsRequests(meta.Namespace).Create(ctx, transform(&api.RedisOpsRequest{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RedisOpsRequest",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}), metav1.CreateOptions{
			DryRun:       opts.DryRun,
			FieldManager: opts.FieldManager,
		})
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchRedisOpsRequest(ctx, c, cur, transform, opts)
}

func PatchRedisOpsRequest(ctx context.Context, c cs.OpsV1alpha1Interface, cur *api.RedisOpsRequest, transform func(*api.RedisOpsRequest) *api.RedisOpsRequest, opts metav1.PatchOptions) (*api.RedisOpsRequest, kutil.VerbType, error) {
	return PatchRedisOpsRequestObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchRedisOpsRequestObject(ctx context.Context, c cs.OpsV1alpha1Interface, cur, mod *api.RedisOpsRequest, opts metav1.PatchOptions) (*api.RedisOpsRequest, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	klog.V(3).Infof("Patching RedisOpsRequest %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.RedisOpsRequests(cur.Namespace).Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdateRedisOpsRequest(ctx context.Context, c cs.OpsV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.RedisOpsRequest) *api.RedisOpsRequest, opts metav1.UpdateOptions) (result *api.RedisOpsRequest, err error) {
	attempt := 0
	err = wait.PollUntilContextTimeout(ctx, kutil.RetryInterval, kutil.RetryTimeout, true, func(ctx context.Context) (bool, error) {
		attempt++
		cur, e2 := c.RedisOpsRequests(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.RedisOpsRequests(cur.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		klog.Errorf("Attempt %d failed to update RedisOpsRequest %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})
	if err != nil {
		err = fmt.Errorf("failed to update RedisOpsRequest %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateRedisOpsRequestStatus(
	ctx context.Context,
	c cs.OpsV1alpha1Interface,
	meta metav1.ObjectMeta,
	transform func(*api.OpsRequestStatus) (types.UID, *api.OpsRequestStatus),
	opts metav1.UpdateOptions,
) (result *api.RedisOpsRequest, err error) {
	apply := func(x *api.RedisOpsRequest) *api.RedisOpsRequest {
		uid, updatedStatus := transform(x.Status.DeepCopy())
		// Ignore status update when uid does not match
		if uid != "" && uid != x.UID {
			return x
		}
		return &api.RedisOpsRequest{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *updatedStatus,
		}
	}

	attempt := 0
	cur, err := c.RedisOpsRequests(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = wait.PollUntilContextTimeout(ctx, kutil.RetryInterval, kutil.RetryTimeout, true, func(ctx context.Context) (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.RedisOpsRequests(meta.Namespace).UpdateStatus(ctx, apply(cur), opts)
		if kerr.IsConflict(e2) {
			latest, e3 := c.RedisOpsRequests(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})
	if err != nil {
		err = fmt.Errorf("failed to update status of RedisOpsRequest %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
