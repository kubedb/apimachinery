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

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	opsapi "kubedb.dev/apimachinery/apis/ops/v1alpha1"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	cutil "kmodules.xyz/client-go/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OpsRequestController struct {
	parallelCtrl map[string]*ParallelismController
	mux          sync.Mutex
	kbClient     client.Client
	kind         string
	scheme       *runtime.Scheme
}

type ParallelismController struct {
	cancelContext *context.CancelFunc
	*sync.Mutex
}

func NewOpsRequestController(kbClient client.Client, kind string, scm *runtime.Scheme) *OpsRequestController {
	return &OpsRequestController{
		parallelCtrl: make(map[string]*ParallelismController),
		mux:          sync.Mutex{},
		kbClient:     kbClient,
		kind:         kind,
		scheme:       scm,
	}
}

func (c *OpsRequestController) KeyExists(key string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, ok := c.parallelCtrl[key]
	return ok
}

func (c *OpsRequestController) GetParallelismController(key string) *ParallelismController {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.parallelCtrl[key]
}

func (c *OpsRequestController) SetParallelismController(key string, cancelFunc *context.CancelFunc) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.parallelCtrl[key] = &ParallelismController{cancelContext: cancelFunc, Mutex: &sync.Mutex{}}
}

func (c *OpsRequestController) DeleteParallelismController(key string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.parallelCtrl[key] != nil {
		if c.parallelCtrl[key].cancelContext != nil {
			(*c.parallelCtrl[key].cancelContext)()
		}
		delete(c.parallelCtrl, key)
	}
}

func (c *OpsRequestController) RemoveCancelFunc(key string) {
	pCtrl := c.GetParallelismController(key)
	if pCtrl == nil {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if pCtrl.cancelContext != nil {
		(*pCtrl.cancelContext)()
		pCtrl.cancelContext = nil
	}
}

func (c *OpsRequestController) AddCancelFunc(key string, cancelFunc *context.CancelFunc) {
	pCtrl := c.GetParallelismController(key)
	if pCtrl == nil {
		return
	}
	pCtrl.cancelContext = cancelFunc
}

const retryInterval = 5 * time.Second

func (c *OpsRequestController) ShouldProceed(key, conditionType string) bool {
	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()
	for range ticker.C {
		if c.IsCompleted(key, conditionType) {
			return false
		}
		pCtrl := c.GetParallelismController(key)
		// check if there is no running go routine
		if canLock := pCtrl.TryLock(); canLock {
			return true
		}
	}
	return false
}

func (c *OpsRequestController) IsCompleted(key, conditionType string) bool {
	ops, err := c.getOpsObjFromKey(key)

	if kerr.IsNotFound(err) || (ops != nil && ops.GetDeletionTimestamp() != nil) {
		return true
	}
	if err != nil {
		return false
	}

	return cutil.IsConditionTrue(ops.GetStatus().Conditions, conditionType) || cutil.IsConditionTrue(ops.GetStatus().Conditions, opsapi.Successful) ||
		ops.GetStatus().Phase == opsapi.OpsRequestPhaseSuccessful || ops.GetStatus().Phase == opsapi.OpsRequestPhaseFailed ||
		ops.GetStatus().Phase == opsapi.OpsRequestPhaseSkipped
}

func (c *OpsRequestController) getOpsObjFromKey(key string) (opsapi.Accessor, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, err
	}

	uns := &unstructured.Unstructured{}
	uns.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "ops.kubedb.com",
		Version: "v1alpha1",
		Kind:    c.kind,
	})

	err = c.kbClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, uns)
	if err != nil {
		return nil, err
	}

	return c.unstructuredToOpsAccessor(uns)
}

func (c *OpsRequestController) unstructuredToOpsAccessor(u *unstructured.Unstructured) (opsapi.Accessor, error) {
	if u == nil {
		return nil, fmt.Errorf("unstructured object is nil")
	}

	gvk := u.GroupVersionKind()
	obj, err := c.scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("failed to create object for GVK %v: %w", gvk, err)
	}

	// Convert unstructured -> typed
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.Object,
		obj,
	); err != nil {
		return nil, fmt.Errorf("failed to convert unstructured to %v: %w", gvk, err)
	}

	// Ensure it implements Accessor
	accessor, ok := obj.(opsapi.Accessor)
	if !ok {
		return nil, fmt.Errorf("object %T does not implement opsapi.Accessor", obj)
	}

	return accessor, nil
}
