/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"
)

const (
	// Every feature gate should add method here following this template:
	//
	// MyFeature featuregate.Feature = "MyFeature"
	//
	// Feature gates should be listed in alphabetical, case-sensitive
	// (upper before any lower case character) order. This reduces the risk
	// of code conflicts because changes are more likely to be scattered
	// across the file.

	// Enables Cassandra operator.
	Cassandra featuregate.Feature = "Cassandra"

	// Enables ClickHouse operator.
	ClickHouse featuregate.Feature = "ClickHouse"

	// Enables Druid operator.
	Druid featuregate.Feature = "Druid"

	// Enables DuckDB operator.
	// DuckDB featuregate.Feature = "DuckDB"

	// Enables Elasticsearch operator.
	Elasticsearch featuregate.Feature = "Elasticsearch"

	// Enables Etcd operator.
	// Etcd featuregate.Feature = "Etcd"

	// Enables FerretDB operator.
	FerretDB featuregate.Feature = "FerretDB"

	// Enables Flink operator.
	// Flink featuregate.Feature = "Flink"

	// Enables FoundationDB operator.
	// FoundationDB featuregate.Feature = "FoundationDB"

	// Enables Kafka operator.
	Kafka featuregate.Feature = "Kafka"

	// Enables MariaDB operator.
	MariaDB featuregate.Feature = "MariaDB"

	// Enables Memcached operator.
	Memcached featuregate.Feature = "Memcached"

	// Enables MongoDB operator.
	MongoDB featuregate.Feature = "MongoDB"

	// Enables MSSQLServer operator.
	MSSQLServer featuregate.Feature = "MSSQLServer"
	// Enables MySQL operator.
	MySQL featuregate.Feature = "MySQL"

	// Enables NATS operator
	// NATS featuregate.Feature = "NATS"

	// Enables Oracle operator.
	// Oracle featuregate.Feature = "Oracle"

	// Enables PerconaXtraDB operator.
	PerconaXtraDB featuregate.Feature = "PerconaXtraDB"

	// Enables PgBouncer operator.
	PgBouncer featuregate.Feature = "PgBouncer"

	// Enables Pgpool operator.
	Pgpool featuregate.Feature = "Pgpool"

	// Enables Postgres operator.
	Postgres featuregate.Feature = "Postgres"

	// Enables ProxySQL operator.
	ProxySQL featuregate.Feature = "ProxySQL"

	// Enables Pulsar operator.
	// Pulsar featuregate.Feature = "Pulsar"

	// Enables RabbitMQ operator.
	RabbitMQ featuregate.Feature = "RabbitMQ"

	// Enables Redis operator.
	Redis featuregate.Feature = "Redis"

	// Enables Singlestore operator.
	Singlestore featuregate.Feature = "Singlestore"

	// Enables Solr operator.
	Solr featuregate.Feature = "Solr"

	// Enables ValKey operator.
	// ValKey featuregate.Feature = "ValKey"

	// Enables Vitess operator.
	// Vitess featuregate.Feature = "Vitess"

	// Enables YugabyteDB operator.
	// YugabyteDB featuregate.Feature = "YugabyteDB"

	// Enables ZooKeeper operator.
	ZooKeeper featuregate.Feature = "ZooKeeper"
)

func init() {
	runtime.Must(DefaultMutableFeatureGate.Add(defaultKubeDBFeatureGates))
}

// defaultKubeDBFeatureGates consists of all known KubeDB-specific feature keys.
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout KubeDB binaries.
var defaultKubeDBFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	Cassandra:     {Default: false, PreRelease: featuregate.Alpha},
	ClickHouse:    {Default: false, PreRelease: featuregate.Alpha},
	Druid:         {Default: false, PreRelease: featuregate.Alpha},
	Elasticsearch: {Default: true, PreRelease: featuregate.GA},
	// Etcd:               {Default: false, PreRelease: featuregate.Alpha, LockToDefault: true},
	FerretDB:    {Default: false, PreRelease: featuregate.Alpha},
	Kafka:       {Default: true, PreRelease: featuregate.Beta},
	MariaDB:     {Default: true, PreRelease: featuregate.GA},
	Memcached:   {Default: false, PreRelease: featuregate.Beta},
	MongoDB:     {Default: true, PreRelease: featuregate.GA},
	MSSQLServer: {Default: false, PreRelease: featuregate.Alpha},
	MySQL:       {Default: true, PreRelease: featuregate.GA},
	// Oracle:             {Default: false, PreRelease: featuregate.Alpha, LockToDefault: true},
	PerconaXtraDB: {Default: false, PreRelease: featuregate.Beta},
	PgBouncer:     {Default: false, PreRelease: featuregate.Beta},
	Pgpool:        {Default: false, PreRelease: featuregate.Alpha},
	Postgres:      {Default: true, PreRelease: featuregate.GA},
	ProxySQL:      {Default: false, PreRelease: featuregate.Beta},
	RabbitMQ:      {Default: false, PreRelease: featuregate.Alpha},
	Redis:         {Default: true, PreRelease: featuregate.GA},
	Singlestore:   {Default: false, PreRelease: featuregate.Alpha},
	Solr:          {Default: false, PreRelease: featuregate.Alpha},
	ZooKeeper:     {Default: false, PreRelease: featuregate.Alpha},
}
