package util

import (
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateOrPatchRedis(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.Redis) *api.Redis) (*api.Redis, kutil.VerbType, error) {
	cur, err := c.Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Redis %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Redises(meta.Namespace).Create(transform(&api.Redis{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Redis",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchRedis(c, cur, transform)
}

func PatchRedis(c cs.KubedbV1alpha1Interface, cur *api.Redis, transform func(*api.Redis) *api.Redis) (*api.Redis, kutil.VerbType, error) {
	return PatchRedisObject(c, cur, transform(cur.DeepCopy()))
}

func PatchRedisObject(c cs.KubedbV1alpha1Interface, cur, mod *api.Redis) (*api.Redis, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(curJson, modJson, curJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching Redis %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.Redises(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateRedis(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.Redis) *api.Redis) (result *api.Redis, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {

			result, e2 = c.Redises(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Redis %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Redis %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateRedisStatus(
	c cs.KubedbV1alpha1Interface,
	in *api.Redis,
	transform func(*api.RedisStatus) *api.RedisStatus, useSubresource ...bool,
) (result *api.Redis, err error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	apply := func(x *api.Redis, copy bool) *api.Redis {
		out := &api.Redis{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
		}
		if copy {
			out.Status = *transform(in.Status.DeepCopy())
		} else {
			out.Status = *transform(&in.Status)
		}
		return out
	}

	if len(useSubresource) == 1 && useSubresource[0] {
		attempt := 0
		cur := in.DeepCopy()
		err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
			attempt++
			var e2 error
			result, e2 = c.Redises(in.Namespace).UpdateStatus(apply(cur, false))
			if kerr.IsConflict(e2) {
				latest, e3 := c.Redises(in.Namespace).Get(in.Name, metav1.GetOptions{})
				switch {
				case e3 == nil:
					cur = latest
					return false, nil
				case kutil.IsRequestRetryable(e3):
					return false, nil
				default:
					return false, e3
				}
			} else if !kutil.IsRequestRetryable(e2) {
				return false, e2
			}
			return e2 == nil, nil
		})

		if err != nil {
			err = fmt.Errorf("failed to update status of Redis %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
		}
		return
	}

	result, _, err = PatchRedisObject(c, in, apply(in, true))
	return
}
