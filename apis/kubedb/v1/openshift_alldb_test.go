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

package v1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// catalogID is the UID/GID every test catalog version carries. After defaulting with namespace
// IDs it must not appear anywhere in the database spec.
const catalogID = 999

var testNSIDs = &dbapi.NamespaceSecurityIDs{RunAsUser: 1001880000, RunAsGroup: 1001880000, FSGroup: 1001890000}

// catalogVersion returns a catalog version object of type T whose securityContext uses
// catalogID for every RunAs* field it has.
func catalogVersion[T any]() *T {
	v := new(T)
	sc := reflect.ValueOf(v).Elem().FieldByName("Spec").FieldByName("SecurityContext")
	for _, f := range []string{"RunAsUser", "RunAsGroup"} {
		if field := sc.FieldByName(f); field.IsValid() {
			field.Set(reflect.ValueOf(ptr.To(int64(catalogID))))
		}
	}
	reflect.ValueOf(v).Elem().FieldByName("Spec").FieldByName("Version").SetString("1.0.0")
	return v
}

// versionClient answers Get for the one catalog version a database's SetDefaults(kc) looks up.
// Any other call panics on the nil embedded client, which makes an unexpected lookup obvious.
type versionClient struct {
	client.Client
	version client.Object
}

func (c versionClient) Get(_ context.Context, _ client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	if reflect.TypeOf(obj) != reflect.TypeOf(c.version) {
		return fmt.Errorf("unexpected Get of %T", obj)
	}
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(c.version).Elem())
	return nil
}

// druidVersion is a Druid catalog version new enough (> 25) for Druid to default fsGroups.
func druidVersion() *catalog.DruidVersion {
	v := catalogVersion[catalog.DruidVersion]()
	v.Spec.Version = "28.0.1"
	return v
}

func meta() metav1.ObjectMeta { return metav1.ObjectMeta{Name: "db", Namespace: "demo"} }

// idCounts counts how many runAsUser/runAsGroup/fsGroup fields in the JSON of spec hold id.
func idCounts(t *testing.T, spec any, id int64) int {
	t.Helper()
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(fmt.Sprintf(`"(runAsUser|runAsGroup|fsGroup)":%d[,}]`, id))
	return len(re.FindAll(b, -1))
}

type dbCase struct {
	name string
	// run defaults a fresh database with the given namespace IDs and returns its spec.
	run func(nsIDs *dbapi.NamespaceSecurityIDs) any
}

// All databases served by a mutating webhook in kubedb/v1 and kubedb/v1alpha2, plus Aerospike.
func allDBCases() []dbCase {
	wrap := func(nsIDs *dbapi.NamespaceSecurityIDs, spec any, f func()) any {
		dbapi.SetDefaultsWithNamespaceSecurityIDs(spec, nsIDs, f)
		return spec
	}
	kcCase := func(name string, version client.Object, newDB func() (spec any, setDefaults func(kc client.Client))) dbCase {
		return dbCase{name: name, run: func(nsIDs *dbapi.NamespaceSecurityIDs) any {
			spec, sd := newDB()
			return wrap(nsIDs, spec, func() { sd(versionClient{version: version}) })
		}}
	}
	replicas := func() *int32 { return ptr.To(int32(3)) }

	return []dbCase{
		// kubedb/v1
		{"v1/Postgres", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.Postgres{ObjectMeta: meta(), Spec: dbapi.PostgresSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.PostgresVersion]()) })
		}},
		{"v1/MySQL", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.MySQL{ObjectMeta: meta(), Spec: dbapi.MySQLSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { _ = db.SetDefaults(catalogVersion[catalog.MySQLVersion]()) })
		}},
		{"v1/MariaDB", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.MariaDB{ObjectMeta: meta(), Spec: dbapi.MariaDBSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.MariaDBVersion]()) })
		}},
		{"v1/PerconaXtraDB", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.PerconaXtraDB{ObjectMeta: meta(), Spec: dbapi.PerconaXtraDBSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.PerconaXtraDBVersion]()) })
		}},
		{"v1/MongoDB", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.MongoDB{ObjectMeta: meta(), Spec: dbapi.MongoDBSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.MongoDBVersion]()) })
		}},
		{"v1/Elasticsearch", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.Elasticsearch{ObjectMeta: meta(), Spec: dbapi.ElasticsearchSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.ElasticsearchVersion]()) })
		}},
		{"v1/Redis", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.Redis{ObjectMeta: meta(), Spec: dbapi.RedisSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { _ = db.SetDefaults(catalogVersion[catalog.RedisVersion]()) })
		}},
		{"v1/RedisSentinel", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.RedisSentinel{ObjectMeta: meta(), Spec: dbapi.RedisSentinelSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { _ = db.SetDefaults(catalogVersion[catalog.RedisVersion]()) })
		}},
		{"v1/Memcached", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.Memcached{ObjectMeta: meta(), Spec: dbapi.MemcachedSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.MemcachedVersion]()) })
		}},
		{"v1/PgBouncer", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.PgBouncer{ObjectMeta: meta(), Spec: dbapi.PgBouncerSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.PgBouncerVersion](), false) })
		}},
		{"v1/ProxySQL", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &dbapi.ProxySQL{ObjectMeta: meta(), Spec: dbapi.ProxySQLSpec{Version: "1.0.0", Replicas: replicas(), Backend: &core.LocalObjectReference{Name: "mysql"}}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.ProxySQLVersion](), false) })
		}},
		kcCase("v1/Kafka", catalogVersion[catalog.KafkaVersion](), func() (any, func(client.Client)) {
			db := &dbapi.Kafka{ObjectMeta: meta(), Spec: dbapi.KafkaSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),

		// kubedb/v1alpha2 (databases that are not in v1)
		{"v1alpha2/Aerospike", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &olddbapi.Aerospike{ObjectMeta: meta(), Spec: olddbapi.AerospikeSpec{Version: "1.0.0"}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.AerospikeVersion]()) })
		}},
		{"v1alpha2/Etcd", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &olddbapi.Etcd{ObjectMeta: meta(), Spec: olddbapi.EtcdSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(catalogVersion[catalog.EtcdVersion]()) })
		}},
		{"v1alpha2/DocumentDB", func(n *dbapi.NamespaceSecurityIDs) any {
			db := &olddbapi.DocumentDB{ObjectMeta: meta(), Spec: olddbapi.DocumentDBSpec{Version: "1.0.0", Replicas: replicas()}}
			return wrap(n, &db.Spec, func() { db.SetDefaults(nil, *catalogVersion[catalog.DocumentDBVersion]()) })
		}},
		kcCase("v1alpha2/Cassandra", catalogVersion[catalog.CassandraVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Cassandra{ObjectMeta: meta(), Spec: olddbapi.CassandraSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/ClickHouse", catalogVersion[catalog.ClickHouseVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.ClickHouse{ObjectMeta: meta(), Spec: olddbapi.ClickHouseSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		// DB2 defaults no UID/GID/fsGroup at all, so OpenShift's SCC assigns them: nothing to test.
		kcCase("v1alpha2/Druid", druidVersion(), func() (any, func(client.Client)) {
			db := &olddbapi.Druid{ObjectMeta: meta(), Spec: olddbapi.DruidSpec{Version: "28.0.1", Topology: &olddbapi.DruidClusterTopology{}}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/HanaDB", catalogVersion[catalog.HanaDBVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.HanaDB{ObjectMeta: meta(), Spec: olddbapi.HanaDBSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Hazelcast", catalogVersion[catalog.HazelcastVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Hazelcast{ObjectMeta: meta(), Spec: olddbapi.HazelcastSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Ignite", catalogVersion[catalog.IgniteVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Ignite{ObjectMeta: meta(), Spec: olddbapi.IgniteSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Milvus", catalogVersion[catalog.MilvusVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Milvus{ObjectMeta: meta(), Spec: olddbapi.MilvusSpec{Version: "1.0.0"}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/MSSQLServer", catalogVersion[catalog.MSSQLServerVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.MSSQLServer{ObjectMeta: meta(), Spec: olddbapi.MSSQLServerSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Neo4j", catalogVersion[catalog.Neo4jVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Neo4j{ObjectMeta: meta(), Spec: olddbapi.Neo4jSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Oracle", catalogVersion[catalog.OracleVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Oracle{ObjectMeta: meta(), Spec: olddbapi.OracleSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Pgpool", catalogVersion[catalog.PgpoolVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Pgpool{ObjectMeta: meta(), Spec: olddbapi.PgpoolSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Qdrant", catalogVersion[catalog.QdrantVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Qdrant{ObjectMeta: meta(), Spec: olddbapi.QdrantSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/RabbitMQ", catalogVersion[catalog.RabbitMQVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.RabbitMQ{ObjectMeta: meta(), Spec: olddbapi.RabbitMQSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Singlestore", catalogVersion[catalog.SinglestoreVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Singlestore{ObjectMeta: meta(), Spec: olddbapi.SinglestoreSpec{Version: "1.0.0"}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Solr", catalogVersion[catalog.SolrVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Solr{ObjectMeta: meta(), Spec: olddbapi.SolrSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/Weaviate", catalogVersion[catalog.WeaviateVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.Weaviate{ObjectMeta: meta(), Spec: olddbapi.WeaviateSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
		kcCase("v1alpha2/ZooKeeper", catalogVersion[catalog.ZooKeeperVersion](), func() (any, func(client.Client)) {
			db := &olddbapi.ZooKeeper{ObjectMeta: meta(), Spec: olddbapi.ZooKeeperSpec{Version: "1.0.0", Replicas: replicas()}}
			return &db.Spec, db.SetDefaults
		}),
	}
}

// TestNamespaceSecurityIDsAllDatabases runs every database's real SetDefaults through
// SetDefaultsWithNamespaceSecurityIDs. Without namespace IDs the catalog ID must show up (so the
// database really defaults IDs from the catalog); with them no catalog ID may be left over.
func TestNamespaceSecurityIDsAllDatabases(t *testing.T) {
	for _, c := range allDBCases() {
		t.Run(c.name, func(t *testing.T) {
			plain := c.run(nil)
			if idCounts(t, plain, catalogID) == 0 {
				t.Fatalf("no catalog ID defaulted without namespace IDs; the test would prove nothing")
			}
			if n := idCounts(t, plain, testNSIDs.RunAsUser) + idCounts(t, plain, testNSIDs.FSGroup); n != 0 {
				t.Fatalf("%d namespace IDs set without namespace IDs", n)
			}

			got := c.run(testNSIDs)
			if n := idCounts(t, got, catalogID); n != 0 {
				b, _ := json.MarshalIndent(got, "", "  ")
				t.Fatalf("%d catalog IDs left after namespace defaulting:\n%s", n, b)
			}
			if idCounts(t, got, testNSIDs.RunAsUser)+idCounts(t, got, testNSIDs.FSGroup) == 0 {
				t.Fatalf("no namespace ID applied")
			}
		})
	}
}
