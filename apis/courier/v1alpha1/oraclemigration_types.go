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

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	ResourceKindOracleMigration     = "OracleMigration"
	ResourceSingularOracleMigration = "oraclemigration"
	ResourcePluralOracleMigrations  = "oraclemigrations"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=oraclemigrations,singular=oraclemigration,shortName=ormig,categories={kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Stage",type="string",JSONPath=".status.progress.info.Stage"
// +kubebuilder:printcolumn:name="Lag",type="string",JSONPath=".status.progress.info.Lag"
// +kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress.info.Progress"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type OracleMigration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of OracleMigration
	// +required
	Spec OracleMigrationSpec `json:"spec"`

	// status defines the observed state of OracleMigration.
	// It reuses the shared MigrationStatus so that the Migration duck type can
	// project it and the operator's status patches replay onto it unchanged.
	// +optional
	Status MigrationStatus `json:"status,omitzero"`
}

// OracleMigrationSpec defines the desired state of OracleMigration
type OracleMigrationSpec struct {
	// Source defines the source Oracle database configuration
	Source OracleSource `json:"source"`

	// Target defines the target Oracle database configuration
	Target OracleTarget `json:"target"`

	// JobDefaults specifies default settings for migration jobs
	// +optional
	JobDefaults *JobDefaults `json:"jobDefaults,omitempty"`

	// JobTemplate specifies runtime configurations for the migration Job
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`
}

// OracleMigrationList contains a list of OracleMigration

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OracleMigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []OracleMigration `json:"items"`
}

type OracleSource struct {
	// ConnectionInfo refers to the source Oracle database connection information.
	// Oracle reuses the shared ConnectionInfo because the go-ora driver is
	// URL-based (URL + PDB dbName + TCPS TLS fit directly).
	ConnectionInfo *ConnectionInfo  `yaml:"connectionInfo" json:"connectionInfo"`
	Schema         *OracleSchema    `yaml:"schema" json:"schema,omitempty"`
	Snapshot       *OracleSnapshot  `yaml:"snapshot" json:"snapshot,omitempty"`
	Streaming      *OracleStreaming `yaml:"streaming" json:"streaming,omitempty"`
}

type OracleTarget struct {
	// ConnectionInfo refers to the target Oracle database connection information.
	ConnectionInfo *ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
}

type OracleSchema struct {
	// Enabled controls whether the Schema Phase should be executed.
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Schema is the list of Oracle owners (a schema is a user) to migrate.
	// +optional
	Schema []string `yaml:"schema" json:"schema,omitempty"`
	// ExcludeSchema is the list of Oracle owners to exclude.
	// +optional
	ExcludeSchema []string `yaml:"excludeSchema" json:"excludeSchema,omitempty"`
	// Table is the list of OWNER.TABLE qualified tables to include.
	// +optional
	Table []string `yaml:"table" json:"table,omitempty"`
	// ExcludeTable is the list of OWNER.TABLE qualified tables to exclude.
	// +optional
	ExcludeTable []string `yaml:"excludeTable" json:"excludeTable,omitempty"`
	// TablespaceMap optionally rewrites source tablespace names to target ones.
	// When empty, physical (tablespace) attributes are stripped and objects land
	// in the target user's default tablespace.
	// +optional
	TablespaceMap map[string]string `yaml:"tablespaceMap" json:"tablespaceMap,omitempty"`
}

type OracleSnapshot struct {
	// Enabled controls whether the Snapshot Phase should be executed.
	// +optional
	Enabled  bool                    `yaml:"enabled" json:"enabled"`
	Pipeline *OracleSnapshotPipeline `yaml:"pipeline" json:"pipeline,omitempty"`
	// LOBBatchSize overrides the pipeline ReadBatchSize for tables that carry
	// LOB columns, which go-ora materializes fully in memory.
	// +optional
	LOBBatchSize *int `yaml:"lobBatchSize" json:"lobBatchSize,omitempty"`
}

type OracleStreaming struct {
	// Enabled controls whether the CDC (LogMiner) Streaming Phase should be executed.
	// +optional
	Enabled bool `yaml:"enabled" json:"enabled"`
	// PollInterval is the idle sleep between LogMiner mining windows.
	// +optional
	PollInterval time.Duration `yaml:"pollInterval" json:"pollInterval,omitempty"`
	// SCNWindowSize is the number of SCNs mined per LogMiner window.
	// +optional
	SCNWindowSize *int64 `yaml:"scnWindowSize" json:"scnWindowSize,omitempty"`
	// AutoEnablePrereqs, when true, lets the CLI enable supplemental logging on
	// the source (needs privileges). ARCHIVELOG is never auto-enabled (it needs
	// a MOUNT-state restart).
	// +optional
	AutoEnablePrereqs bool `yaml:"autoEnablePrereqs" json:"autoEnablePrereqs,omitempty"`
	// LargeTxSpillMB is the per-XID client-side buffer spill-to-disk threshold.
	// +optional
	LargeTxSpillMB *int `yaml:"largeTxSpillMB" json:"largeTxSpillMB,omitempty"`
}

type OracleSnapshotPipeline struct {
	Workers        *int `yaml:"workers" json:"workers"`
	Sinkers        *int `yaml:"sinkers" json:"sinkers"`
	Buffer         *int `yaml:"buffer" json:"buffer"`
	ReadBatchSize  *int `yaml:"readBatchSize" json:"read_batch_size"`
	WriteBatchSize *int `yaml:"writeBatchSize" json:"write_batch_size"`
}

func init() {
	SchemeBuilder.Register(&OracleMigration{}, &OracleMigrationList{})
}
