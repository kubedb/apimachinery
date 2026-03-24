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

type MySQLSource struct {
	// ConnectionInfo refers to the source MySQL database connection information.
	ConnectionInfo *MySQLConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
	Schema         *MySQLSchema         `yaml:"schema" json:"schema,omitempty"`
	Snapshot       *MySQLSnapshot       `yaml:"snapshot" json:"snapshot,omitempty"`
	Replication    *MySQLReplication    `yaml:"replication" json:"replication,omitempty"`
}

type MySQLTarget struct {
	// ConnectionInfo refers to the target MySQL database connection information.
	ConnectionInfo *MySQLConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
}

type MySQLSchema struct {
	// Enabled controls whether the Schema Phase should be executed.
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Database is the list of databases to migrate.
	// +optional
	Database []string `yaml:"database" json:"database,omitempty"`
	// ExcludeDatabase is the list of databases to exclude from migration.
	// +optional
	ExcludeDatabase []string `yaml:"excludeDatabase" json:"excludeDatabase,omitempty"`
}

type MySQLSnapshot struct {
	// Enabled controls whether the Snapshot Phase should be executed.
	// +optional
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type MySQLReplication struct {
	// Enabled controls whether the Logical Replication Phase should be executed.
	// +optional
	Enabled bool `yaml:"enabled" json:"enabled"`
}

type MySQLConnectionInfo struct {
	Address        string `yaml:"address" json:"address"`
	User           string `yaml:"user" json:"user"`
	Password       string `yaml:"password" json:"password"`
	DBName         string `yaml:"dbName" json:"dbName"`
	MaxConnections int    `yaml:"maxConnections" json:"maxConnections,omitempty"`
}
