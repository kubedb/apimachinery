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

package statefulset

import (
	"fmt"

	"github.com/appscode/go/log"
	appsv1 "k8s.io/api/apps/v1"
)

func (c *Controller) processStatefulSet(key string) error {
	log.Infof("Started processing, key: %v", key)
	obj, exists, err := c.StsInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("StatefulSet %s does not exist anymore", key)
	} else {
		sts := obj.(*appsv1.StatefulSet).DeepCopy()
		dbInfo, err := c.extractDatabaseInfo(sts)
		if err != nil {
			return fmt.Errorf("failed to extract database info from StatefulSet: %s/%s. Reason: %v", sts.Namespace, sts.Name, err)
		}
		return c.ensureReadyReplicasCond(dbInfo)
	}
	return nil
}
