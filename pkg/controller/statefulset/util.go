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
	"context"
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kmapi "kmodules.xyz/client-go/api/v1"
	dmcond "kmodules.xyz/client-go/dynamic/conditions"
)

type databaseInfo struct {
	opts          dmcond.DynamicOptions
	replicasReady bool
	msg           string
}

func (c *Controller) extractDatabaseInfo(sts *apps.StatefulSet) (*databaseInfo, error) {
	// read the controlling owner
	owner := metav1.GetControllerOf(sts)
	if owner == nil {
		return nil, fmt.Errorf("StatefulSet %s/%s has no controlling owner", sts.Namespace, sts.Name)
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
			Namespace: sts.Namespace,
		},
	}
	dbInfo.opts.GVR = schema.GroupVersionResource{
		Group:   gv.Group,
		Version: gv.Version,
	}
	switch owner.Kind {
	case api.ResourceKindDruid:
		dbInfo.opts.GVR.Resource = api.ResourcePluralDruid
		dr, err := c.DBClient.KubedbV1alpha2().Druids(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = dr.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindElasticsearch:
		dbInfo.opts.GVR.Resource = api.ResourcePluralElasticsearch
		es, err := c.DBClient.KubedbV1alpha2().Elasticsearches(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = es.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindEtcd:
		dbInfo.opts.GVR.Resource = api.ResourcePluralEtcd
		etcd, err := c.DBClient.KubedbV1alpha2().Etcds(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = etcd.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindFerretDB:
		dbInfo.opts.GVR.Resource = api.ResourcePluralFerretDB
		fr, err := c.DBClient.KubedbV1alpha2().FerretDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = fr.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindKafka:
		dbInfo.opts.GVR.Resource = api.ResourcePluralKafka
		kf, err := c.DBClient.KubedbV1alpha2().Kafkas(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = kf.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMariaDB:
		dbInfo.opts.GVR.Resource = api.ResourcePluralMariaDB
		mr, err := c.DBClient.KubedbV1alpha2().MariaDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mr.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMemcached:
		dbInfo.opts.GVR.Resource = api.ResourcePluralMemcached
		mc, err := c.DBClient.KubedbV1alpha2().Memcacheds(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mc.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMongoDB:
		dbInfo.opts.GVR.Resource = api.ResourcePluralMongoDB
		mg, err := c.DBClient.KubedbV1alpha2().MongoDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mg.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindMySQL:
		dbInfo.opts.GVR.Resource = api.ResourcePluralMySQL
		my, err := c.DBClient.KubedbV1alpha2().MySQLs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = my.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPerconaXtraDB:
		dbInfo.opts.GVR.Resource = api.ResourcePluralPerconaXtraDB
		px, err := c.DBClient.KubedbV1alpha2().PerconaXtraDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = px.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPgBouncer:
		dbInfo.opts.GVR.Resource = api.ResourcePluralPgBouncer
		pgb, err := c.DBClient.KubedbV1alpha2().PgBouncers(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pgb.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindPgpool:
		dbInfo.opts.GVR.Resource = api.ResourcePluralPgpool
		pp, err := c.DBClient.KubedbV1alpha2().Pgpools(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pp.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

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

	case api.ResourceKindProxySQL:
		dbInfo.opts.GVR.Resource = api.ResourcePluralProxySQL
		pxql, err := c.DBClient.KubedbV1alpha2().ProxySQLs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pxql.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindRabbitmq:
		dbInfo.opts.GVR.Resource = api.ResourcePluralRabbitmq
		rb, err := c.DBClient.KubedbV1alpha2().RabbitMQs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rb.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindRedis:
		dbInfo.opts.GVR.Resource = api.ResourcePluralRedis
		rd, err := c.DBClient.KubedbV1alpha2().Redises(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rd.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindRedisSentinel:
		dbInfo.opts.GVR.Resource = api.ResourcePluralRedisSentinel
		rd, err := c.DBClient.KubedbV1alpha2().RedisSentinels(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rd.ReplicasAreReady(c.StsLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindSinglestore:
		dbInfo.opts.GVR.Resource = api.ResourcePluralSinglestore
		ss, err := c.DBClient.KubedbV1alpha2().Singlestores(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = ss.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindSolr:
		dbInfo.opts.GVR.Resource = api.ResourcePluralSolr
		sl, err := c.DBClient.KubedbV1alpha2().Solrs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = sl.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case api.ResourceKindZooKeeper:
		dbInfo.opts.GVR.Resource = api.ResourcePluralZooKeeper
		zk, err := c.DBClient.KubedbV1alpha2().ZooKeepers(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = zk.ReplicasAreReady(c.PSLister)
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
