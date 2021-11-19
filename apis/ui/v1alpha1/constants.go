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

package v1alpha1

import kubeDBv1alpha2 "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

const (
	GetNATSURLAPIPath = "192.168.1.118:3333/api/get-nats-url"
)

type DBEnvironment string

var (
	KubeDBGroup   = kubeDBv1alpha2.SchemeGroupVersion.Group
	KubeDBVersion = kubeDBv1alpha2.SchemeGroupVersion.Version
)

const (
	DBEnvironmentProduction  DBEnvironment = "PROD"
	DBEnvironmentDevelopment DBEnvironment = "DEV"
	DBEnvironmentQA          DBEnvironment = "QA"
)

var SupportedDatabases = []DatabaseKindPlural{DatabaseKindPluralMongoDB, DatabaseKindPluralElasticsearch, DatabaseKindPluralPostgresSQL, DatabaseKindPluralMySQL, DatabaseKindPluralMariaDB, DatabaseKindPluralRedis}

type DatabaseKindPlural string

const (
	DatabaseKindPluralMongoDB       = kubeDBv1alpha2.ResourcePluralMongoDB
	DatabaseKindPluralElasticsearch = kubeDBv1alpha2.ResourcePluralElasticsearch
	DatabaseKindPluralPostgresSQL   = kubeDBv1alpha2.ResourcePluralPostgres
	DatabaseKindPluralMySQL         = kubeDBv1alpha2.ResourcePluralMySQL
	DatabaseKindPluralMariaDB       = kubeDBv1alpha2.ResourcePluralMariaDB
	DatabaseKindPluralRedis         = kubeDBv1alpha2.ResourcePluralRedis
)

type DBStatus string

const (
	DBStatusReady         DBStatus = "Ready"
	DBStatusDataRestoring DBStatus = "DataRestoring"
	DBStatusCritical      DBStatus = "Critical"
	DBStatusProvisioning  DBStatus = "Provisioning"
	DBStatusHalted        DBStatus = "Halted"
	DBStatusNotReady      DBStatus = "NotReady"
)

type DBNodeStatus string

const (
	DBNodeStatusHealthy   DBNodeStatus = "Healthy"
	DBNodeStatusUnhealthy DBNodeStatus = "Unhealthy"
)

type MongoDBMode string

const (
	MongoDBModeStandalone MongoDBMode = "Standalone"
	MongoDBModeReplicaSet MongoDBMode = "ReplicaSet"
	MongoDBModeSharded    MongoDBMode = "Sharded"
)

type MongoDBNodeType string

const (
	MongoDBNodeTypeShard        MongoDBNodeType = "Shard"
	MongoDBNodeTypeConfigServer MongoDBNodeType = "ConfigServer"
	MongoDBNodeTypeMongos       MongoDBNodeType = "Mongos"
	MongoDBNodeTypeReplicaSet   MongoDBNodeType = "ReplicaSet"
	MongoDBNodeTypeStandalone   MongoDBNodeType = "Standalone"
)

type DatabaseRole string

const (
	DatabaseRolePrimary   DatabaseRole = "Primary"
	DatabaseRoleSecondary DatabaseRole = "Secondary"
)

type MongoDBOperation string

const (
	MongoDBOperationQuery   MongoDBOperation = "QUERY"
	MongoDBOperationInsert  MongoDBOperation = "INSERT"
	MongoDBOperationUpdate  MongoDBOperation = "UPDATE"
	MongoDBOperationDelete  MongoDBOperation = "DELETE"
	MongoDBOperationGetMore MongoDBOperation = "GETMORE"
)

type GrafanaTheme string

const (
	GrafanaThemeDark  GrafanaTheme = "dark"
	GrafanaThemeLight GrafanaTheme = "light"
)

type GrafanaChartURL string

const (
	GrafanaChartURLTemplate GrafanaChartURL = "http://{{.URL}}/d-solo/{{.BoardUID}}/{{.DashboardName}}?orgId={{.OrgID}}&from={{.From}}&theme={{.Theme}}&panelId="
)

const (
	// MongoDB NATS Subject

	NATSSubjectStatus = "k8s.%s.products.kubedb-monitoring-agent.status"

	NATSSubjectMongoDBOverview       = "k8s.%s.products.kubedb-monitoring-agent.MONGODB.INFO"
	NATSSubjectMongoDBSlowQueries    = "k8s.%s.products.kubedb-monitoring-agent.MONGODB.SLOW.QUERIES"
	NATSSubjectMongoDBTopCollections = "k8s.%s.products.kubedb-monitoring-agent.MONGODB.TOP.COLLECTIONS"

	// Postgres NATS Subject

	NATSSubjectPostgresOverview    = "k8s.%s.products.kubedb-monitoring-agent.POSTGRES.OVERVIEW"
	NATSSubjectPostgresSlowQueries = "k8s.%s.products.kubedb-monitoring-agent.POSTGRES.SLOW.QUERIES"
	NATSSubjectPostgresSettings    = "k8s.%s.products.kubedb-monitoring-agent.POSTGRES.SETTINGS"
	NATSSubjectPostgresTableInfo   = "k8s.%s.products.kubedb-monitoring-agent.POSTGRES.TABLE.INFO"

	// Elasticsearch NATS Subject

	NATSSubjectElasticsearchOverview   = "k8s.%s.products.kubedb-monitoring-agent.ELASTICSEARCH.OVERVIEW"
	NATSSubjectElasticsearchNodesStats = "k8s.%s.products.kubedb-monitoring-agent.ELASTICSEARCH.NODES.STATS"
	NATSSubjectElasticsearchIndices    = "k8s.%s.products.kubedb-monitoring-agent.ELASTICSEARCH.INDICES"

	// MySQL NATS Subject

	NATSSubjectMySQLOverview    = "k8s.%s.products.kubedb-monitoring-agent.MYSQL.OVERVIEW"
	NATSSubjectMySQLSlowQueries = "k8s.%s.products.kubedb-monitoring-agent.MYSQL.SLOW.QUERIES"
	NATSSubjectMySQLTableInfo   = "k8s.%s.products.kubedb-monitoring-agent.MYSQL.TABLE.INFO"

	// MariaDB NATS Subject

	NATSSubjectMariaDBOverview    = "k8s.%s.products.kubedb-monitoring-agent.MARIADB.OVERVIEW"
	NATSSubjectMariaDBSlowQueries = "k8s.%s.products.kubedb-monitoring-agent.MARIADB.SLOW.QUERIES"
	NATSSubjectMariaDBTableInfo   = "k8s.%s.products.kubedb-monitoring-agent.MARIADB.TABLE.INFO"

	// Redis NATS Subject

	NATSSubjectRedisOverview    = "k8s.%s.products.kubedb-monitoring-agent.REDIS.OVERVIEW"
	NATSSubjectRedisSlowQueries = "k8s.%s.products.kubedb-monitoring-agent.REDIS.SLOW.QUERIES"
	NATSSubjectRedisDBInfo      = "k8s.%s.products.kubedb-monitoring-agent.REDIS.DBINFO"
)
