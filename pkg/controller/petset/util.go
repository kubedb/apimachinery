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

package petset

import (
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kmapi "kmodules.xyz/client-go/api/v1"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
	petsetapps "kubeops.dev/petset/apis/apps/v1"
)

type databaseInfo struct {
	opts          dmcond.DynamicOptions
	replicasReady bool
	msg           string
}

func (c *Controller) extractDatabaseInfo(ps *petsetapps.PetSet) (*databaseInfo, error) {
	// read the controlling owner
	owner := metav1.GetControllerOf(ps)
	if owner == nil {
		return nil, fmt.Errorf("PetSet %s/%s has no controlling owner", ps.Namespace, ps.Name)
	}
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return nil, err
	}
	if gv.Group != api.SchemeGroupVersion.Group {
		return nil, nil
	}
	dbInfo := &databaseInfo{
		opts: dmcond.DynamicOptions{
			Client:    c.DynamicClient,
			Kind:      owner.Kind,
			Name:      owner.Name,
			Namespace: ps.Namespace,
		},
	}
	dbInfo.opts.GVR = schema.GroupVersionResource{
		Group:   gv.Group,
		Version: gv.Version,
	}
	switch owner.Kind {

	case api.ResourceKindPostgres:
		dbInfo.opts.GVR.Resource = api.ResourcePluralPostgres
		pg, err := c.DBClient.KubedbV1alpha2().Postgreses(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pg.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown database kind: %s", owner.Kind)
	}
	return dbInfo, nil
}

func (c *Controller) ensureReadyReplicasCond(dbInfo *databaseInfo) error {
	dbCond := kmapi.Condition{
		Type:    api.DatabaseReplicaReady,
		Message: dbInfo.msg,
	}

	if dbInfo.replicasReady {
		dbCond.Status = metav1.ConditionTrue
		dbCond.Reason = api.AllReplicasAreReady
	} else {
		dbCond.Status = metav1.ConditionFalse
		dbCond.Reason = api.SomeReplicasAreNotReady
	}

	// Add "ReplicasReady" condition to the respective database CR
	return dbInfo.opts.SetCondition(dbCond)
}
