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

import "time"

type MSSQLServerSource struct {
	// ConnectionInfo refers to the source MSSQL Server database connection information.
	ConnectionInfo *MSSQLServerConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
	Schema         *MSSQLServerSchema         `yaml:"schema" json:"schema,omitempty"`
	Snapshot       *MSSQLServerSnapshot       `yaml:"snapshot" json:"snapshot,omitempty"`
	Streaming      *MSSQLServerStreaming      `yaml:"streaming" json:"streaming,omitempty"`
}

type MSSQLServerTarget struct {
	// ConnectionInfo refers to the target MSSQL Server database connection information.
	ConnectionInfo *MSSQLServerConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
}

type MSSQLServerSchema struct {
	// Enabled controls whether the Schema Phase should be executed.
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Database is the list of databases to migrate.
	// +optional
	Database []string `yaml:"database" json:"database,omitempty"`
	// Schema is the list of SQL Server schemas (e.g. "dbo") to include.
	// +optional
	Schema []string `yaml:"schema" json:"schema,omitempty"`
	// ExcludeSchema is the list of SQL Server schemas to exclude.
	// +optional
	ExcludeSchema []string `yaml:"excludeSchema" json:"excludeSchema,omitempty"`
	// Table is the list of schema-qualified tables (e.g. "dbo.Users") to include.
	// +optional
	Table []string `yaml:"table" json:"table,omitempty"`
	// ExcludeTable is the list of schema-qualified tables to exclude.
	// +optional
	ExcludeTable []string `yaml:"excludeTable" json:"excludeTable,omitempty"`
}

type MSSQLServerSnapshot struct {
	// Enabled controls whether the Snapshot Phase should be executed.
	// +optional
	Enabled  bool                         `yaml:"enabled" json:"enabled"`
	Pipeline *MSSQLServerSnapshotPipeline `yaml:"pipeline" json:"pipeline,omitempty"`
}

type MSSQLServerStreaming struct {
	// Enabled controls whether the CDC Streaming Phase should be executed.
	// +optional
	Enabled bool `yaml:"enabled" json:"enabled"`
	// PollInterval controls how often CDC changes are polled from the source.
	// +optional
	PollInterval time.Duration `yaml:"pollInterval" json:"pollInterval,omitempty"`
	// AutoEnableCDC enables Change Data Capture on the source database/tables
	// automatically when set to true. If false, CDC must be pre-enabled.
	// +optional
	AutoEnableCDC bool `yaml:"autoEnableCDC" json:"autoEnableCDC,omitempty"`
}

type MSSQLServerConnectionInfo struct {
	Address                string `yaml:"address" json:"address"`
	User                   string `yaml:"user" json:"user"`
	Password               string `yaml:"password" json:"password"`
	Database               string `yaml:"database" json:"database"`
	MaxConnections         int    `yaml:"maxConnections" json:"maxConnections,omitempty"`
	Encrypt                bool   `yaml:"encrypt" json:"encrypt,omitempty"`
	TrustServerCertificate bool   `yaml:"trustServerCertificate" json:"trustServerCertificate,omitempty"`
}

type MSSQLServerSnapshotPipeline struct {
	Workers        *int `yaml:"workers" json:"workers"`
	Sinkers        *int `yaml:"sinkers" json:"sinkers"`
	Buffer         *int `yaml:"buffer" json:"buffer"`
	ReadBatchSize  *int `yaml:"readBatchSize" json:"read_batch_size"`
	WriteBatchSize *int `yaml:"writeBatchSize" json:"write_batch_size"`
}
