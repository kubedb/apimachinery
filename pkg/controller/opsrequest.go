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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	cutil "kmodules.xyz/client-go/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OpsRequestController struct {
	ParallelCtrl map[string]*ParallelismController
	Mux          sync.Mutex
	KBClient     client.Client
	Kind         string
}

type ParallelismController struct {
	cancelContext *context.CancelFunc
	*sync.Mutex
}

func NewOpsRequestController(kbClient client.Client, kind string) *OpsRequestController {
	return &OpsRequestController{
		ParallelCtrl: make(map[string]*ParallelismController),
		Mux:          sync.Mutex{},
		KBClient:     kbClient,
		Kind:         kind,
	}
}

func (c *OpsRequestController) KeyExists(key string) bool {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	_, ok := c.ParallelCtrl[key]
	return ok
}

func (c *OpsRequestController) GetParallelismController(key string) *ParallelismController {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	return c.ParallelCtrl[key]
}

func (c *OpsRequestController) SetParallelismController(key string, cancelFunc *context.CancelFunc) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.ParallelCtrl[key] = &ParallelismController{cancelContext: cancelFunc, Mutex: &sync.Mutex{}}
}

func (c *OpsRequestController) DeleteParallelismController(key string) {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	if c.ParallelCtrl[key] != nil {
		if c.ParallelCtrl[key].cancelContext != nil {
			(*c.ParallelCtrl[key].cancelContext)()
		}
		delete(c.ParallelCtrl, key)
	}
}

func (c *OpsRequestController) RemoveCancelFunc(key string) {
	pCtrl := c.GetParallelismController(key)
	if pCtrl == nil {
		return
	}
	c.Mux.Lock()
	defer c.Mux.Unlock()
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
	var ops opsapi.Accessor
	err := c.getObjectFromKey(key, ops)
	if kerr.IsNotFound(err) || ops.GetDeletionTimestamp() != nil {
		return true
	}
	if err != nil {
		return false
	}
	return cutil.IsConditionTrue(ops.GetStatus().Conditions, conditionType) || cutil.IsConditionTrue(ops.GetStatus().Conditions, opsapi.Successful) ||
		ops.GetStatus().Phase == opsapi.OpsRequestPhaseSuccessful || ops.GetStatus().Phase == opsapi.OpsRequestPhaseFailed ||
		ops.GetStatus().Phase == opsapi.OpsRequestPhaseSkipped
}

func (c *OpsRequestController) getObjectFromKey(key string, object client.Object) error {
	if object == nil {
		return fmt.Errorf("object is nil")
	}
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	obectKey := types.NamespacedName{Name: name, Namespace: namespace}
	err = c.KBClient.Get(context.TODO(), obectKey, object)

	uns := &unstructured.Unstructured{}
	uns.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "ops.kubedb.com",
		Version: "v1alpha1",
		Kind:    c.Kind,
	})

	err = c.KBClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, uns)
	return err
}
