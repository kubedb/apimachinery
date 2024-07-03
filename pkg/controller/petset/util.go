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

	"kubedb.dev/apimachinery/apis/kubedb"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

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
	if gv.Group != olddbapi.SchemeGroupVersion.Group {
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

	case olddbapi.ResourceKindDruid:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralDruid
		dr, err := c.DBClient.KubedbV1alpha2().Druids(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = dr.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindElasticsearch:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralElasticsearch
		es, err := c.DBClient.KubedbV1().Elasticsearches(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = es.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindFerretDB:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralFerretDB
		fr, err := c.DBClient.KubedbV1alpha2().FerretDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = fr.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}
	case olddbapi.ResourceKindKafka:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralKafka
		kf, err := c.DBClient.KubedbV1().Kafkas(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = kf.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindMariaDB:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralMariaDB
		md, err := c.DBClient.KubedbV1().MariaDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = md.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindMemcached:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralMemcached
		mc, err := c.DBClient.KubedbV1().Memcacheds(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mc.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindMongoDB:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralMongoDB
		mg, err := c.DBClient.KubedbV1().MongoDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mg.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindMSSQLServer:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralMSSQLServer
		ms, err := c.DBClient.KubedbV1alpha2().MSSQLServers(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = ms.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindMySQL:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralMySQL
		my, err := c.DBClient.KubedbV1().MySQLs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = my.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindPerconaXtraDB:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralPerconaXtraDB
		px, err := c.DBClient.KubedbV1().PerconaXtraDBs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = px.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindPgBouncer:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralPgBouncer
		pp, err := c.DBClient.KubedbV1().PgBouncers(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pp.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindPgpool:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralPgpool
		pp, err := c.DBClient.KubedbV1alpha2().Pgpools(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pp.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindPostgres:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralPostgres
		pg, err := c.DBClient.KubedbV1().Postgreses(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = pg.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindProxySQL:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralProxySQL
		psq, err := c.DBClient.KubedbV1().ProxySQLs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = psq.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindRabbitmq:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralRabbitmq
		mq, err := c.DBClient.KubedbV1alpha2().RabbitMQs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = mq.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case dbapi.ResourceKindRedis:
		dbInfo.opts.GVR.Resource = dbapi.ResourcePluralRedis
		rd, err := c.DBClient.KubedbV1().Redises(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rd.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindRedisSentinel:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralRedisSentinel
		rds, err := c.DBClient.KubedbV1().RedisSentinels(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = rds.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindSinglestore:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralSinglestore
		sdb, err := c.DBClient.KubedbV1alpha2().Singlestores(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = sdb.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case kubedb.ResourceKindSolr:
		dbInfo.opts.GVR.Resource = kubedb.ResourcePluralSolr
		sl, err := c.DBClient.KubedbV1alpha2().Solrs(dbInfo.opts.Namespace).Get(context.TODO(), dbInfo.opts.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		dbInfo.replicasReady, dbInfo.msg, err = sl.ReplicasAreReady(c.PSLister)
		if err != nil {
			return nil, err
		}

	case olddbapi.ResourceKindZooKeeper:
		dbInfo.opts.GVR.Resource = olddbapi.ResourcePluralZooKeeper
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
		Type:    kubedb.DatabaseReplicaReady,
		Message: dbInfo.msg,
	}

	if dbInfo.replicasReady {
		dbCond.Status = metav1.ConditionTrue
		dbCond.Reason = kubedb.AllReplicasAreReady
	} else {
		dbCond.Status = metav1.ConditionFalse
		dbCond.Reason = kubedb.SomeReplicasAreNotReady
	}

	// Add "ReplicasReady" condition to the respective database CR
	return dbInfo.opts.SetCondition(dbCond)
}
