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
