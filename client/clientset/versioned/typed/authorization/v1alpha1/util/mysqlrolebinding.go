package util

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/authorization/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/authorization/v1alpha1"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateOrPatchMySQLRoleBinding(c cs.AuthorizationV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.MySQLRoleBinding) *api.MySQLRoleBinding) (*api.MySQLRoleBinding, kutil.VerbType, error) {
	cur, err := c.MySQLRoleBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating MySQLRoleBinding %s/%s.", meta.Namespace, meta.Name)
		out, err := c.MySQLRoleBindings(meta.Namespace).Create(transform(&api.MySQLRoleBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindMySQLRoleBinding,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchMySQLRoleBinding(c, cur, transform)
}

func PatchMySQLRoleBinding(c cs.AuthorizationV1alpha1Interface, cur *api.MySQLRoleBinding, transform func(*api.MySQLRoleBinding) *api.MySQLRoleBinding) (*api.MySQLRoleBinding, kutil.VerbType, error) {
	return PatchMySQLRoleBindingObject(c, cur, transform(cur.DeepCopy()))
}

func PatchMySQLRoleBindingObject(c cs.AuthorizationV1alpha1Interface, cur, mod *api.MySQLRoleBinding) (*api.MySQLRoleBinding, kutil.VerbType, error) {
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
	glog.V(3).Infof("Patching MySQLRoleBinding %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.MySQLRoleBindings(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateMySQLRoleBinding(c cs.AuthorizationV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.MySQLRoleBinding) *api.MySQLRoleBinding) (result *api.MySQLRoleBinding, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.MySQLRoleBindings(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.MySQLRoleBindings(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update MySQLRoleBinding %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update MySQLRoleBinding %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateMySQLRoleBindingStatus(
	c cs.AuthorizationV1alpha1Interface,
	in *api.MySQLRoleBinding,
	transform func(*api.MySQLRoleBindingStatus) *api.MySQLRoleBindingStatus,
	useSubresource ...bool,
) (result *api.MySQLRoleBinding, err error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	apply := func(x *api.MySQLRoleBinding) *api.MySQLRoleBinding {
		return &api.MySQLRoleBinding{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	if len(useSubresource) == 1 && useSubresource[0] {
		attempt := 0
		cur := in.DeepCopy()
		err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
			attempt++
			var e2 error
			result, e2 = c.MySQLRoleBindings(in.Namespace).UpdateStatus(apply(cur))
			if kerr.IsConflict(e2) {
				latest, e3 := c.MySQLRoleBindings(in.Namespace).Get(in.Name, metav1.GetOptions{})
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
			err = fmt.Errorf("failed to update status of MySQLRoleBinding %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
		}
		return
	}

	result, _, err = PatchMySQLRoleBindingObject(c, in, apply(in))
	return
}
