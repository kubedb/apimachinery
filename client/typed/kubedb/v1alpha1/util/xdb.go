package util

import (
	"encoding/json"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	aci "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/wait"
)

func EnsureXdb(c tcs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *aci.Xdb) *aci.Xdb) (*aci.Xdb, error) {
	return CreateOrPatchXdb(c, meta, transform)
}

func CreateOrPatchXdb(c tcs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *aci.Xdb) *aci.Xdb) (*aci.Xdb, error) {
	cur, err := c.Xdbs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Xdb %s/%s.", meta.Namespace, meta.Name)
		return c.Xdbs(meta.Namespace).Create(transform(&aci.Xdb{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Xdb",
				APIVersion: aci.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
	} else if err != nil {
		return nil, err
	}
	return PatchXdb(c, cur, transform)
}

func PatchXdb(c tcs.KubedbV1alpha1Interface, cur *aci.Xdb, transform func(*aci.Xdb) *aci.Xdb) (*aci.Xdb, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	modJson, err := json.Marshal(transform(cur))
	if err != nil {
		return nil, err
	}

	patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(curJson, modJson, curJson)
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, nil
	}
	glog.V(3).Infof("Patching Xdb %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	result, err := c.Xdbs(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return result, err
}

func TryPatchXdb(c tcs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*aci.Xdb) *aci.Xdb) (result *aci.Xdb, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Xdbs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = PatchXdb(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch Xdb %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to patch Xdb %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func TryUpdateXdb(c tcs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*aci.Xdb) *aci.Xdb) (result *aci.Xdb, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Xdbs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Xdbs(cur.Namespace).Update(transform(cur))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Xdb %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Xdb %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}
