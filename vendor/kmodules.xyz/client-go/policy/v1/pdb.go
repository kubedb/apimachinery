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

package v1

import (
	"context"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	policy "k8s.io/api/policy/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchPodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	cur, err := c.PolicyV1().PodDisruptionBudgets(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		klog.V(3).Infof("Creating PodDisruptionBudget %s/%s.", meta.Namespace, meta.Name)
		out, err := c.PolicyV1().PodDisruptionBudgets(meta.Namespace).Create(ctx, transform(&policy.PodDisruptionBudget{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodDisruptionBudget",
				APIVersion: policy.SchemeGroupVersion.String(),
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
	return PatchPodDisruptionBudget(ctx, c, cur, transform, opts)
}

func PatchPodDisruptionBudget(ctx context.Context, c kubernetes.Interface, cur *policy.PodDisruptionBudget, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
	return PatchPodDisruptionBudgetObject(ctx, c, cur, transform(cur.DeepCopy()), opts)
}

func PatchPodDisruptionBudgetObject(ctx context.Context, c kubernetes.Interface, cur, mod *policy.PodDisruptionBudget, opts metav1.PatchOptions) (*policy.PodDisruptionBudget, kutil.VerbType, error) {
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
	klog.V(3).Infof("Patching PodDisruptionBudget %s with %s.", cur.Name, string(patch))
	out, err := c.PolicyV1().PodDisruptionBudgets(cur.Namespace).Patch(ctx, cur.Name, types.MergePatchType, patch, opts)
	return out, kutil.VerbPatched, err
}

func TryUpdatePodDisruptionBudget(ctx context.Context, c kubernetes.Interface, meta metav1.ObjectMeta, transform func(*policy.PodDisruptionBudget) *policy.PodDisruptionBudget, opts metav1.UpdateOptions) (result *policy.PodDisruptionBudget, err error) {
	attempt := 0
	err = wait.PollUntilContextTimeout(ctx, kutil.RetryInterval, kutil.RetryTimeout, true, func(ctx context.Context) (bool, error) {
		attempt++
		cur, e2 := c.PolicyV1().PodDisruptionBudgets(meta.Namespace).Get(ctx, meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.PolicyV1().PodDisruptionBudgets(meta.Namespace).Update(ctx, transform(cur.DeepCopy()), opts)
			return e2 == nil, nil
		}
		klog.Errorf("Attempt %d failed to update PodDisruptionBudget %s due to %v.", attempt, cur.Name, e2)
		return false, nil
	})
	if err != nil {
		err = errors.Errorf("failed to update PodDisruptionBudget %s after %d attempts due to %v", meta.Name, attempt, err)
	}
	return
}
