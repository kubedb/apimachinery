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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

const (
	ResourceKindRedisMigration     = "RedisMigration"
	ResourceSingularRedisMigration = "redismigration"
	ResourcePluralRedisMigrations  = "redismigrations"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=redismigrations,singular=redismigration,shortName=rdmig,categories={kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Stage",type="string",JSONPath=".status.progress.info.Stage"
// +kubebuilder:printcolumn:name="Lag",type="string",JSONPath=".status.progress.info.Lag"
// +kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress.info.Progress"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type RedisMigration struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of RedisMigration
	// +required
	Spec RedisMigrationSpec `json:"spec"`

	// status defines the observed state of RedisMigration.
	// +optional
	Status MigrationStatus `json:"status,omitzero"`
}

// RedisMigrationSpec defines the desired state of RedisMigration
type RedisMigrationSpec struct {
	// Source defines the source Redis database configuration
	Source RedisSource `json:"source"`

	// Target defines the target Redis database configuration
	Target RedisTarget `json:"target"`

	// JobDefaults specifies default settings for migration jobs
	// +optional
	JobDefaults *JobDefaults `json:"jobDefaults,omitempty"`

	// JobTemplate specifies runtime configurations for the migration Job
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`
}

// RedisMigrationList contains a list of RedisMigration

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedisMigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []RedisMigration `json:"items"`
}

// RedisMode mirrors kubedb.com/v1 RedisMode; redeclared here to avoid an import cycle.
// +kubebuilder:validation:Enum=Standalone;Cluster;Sentinel
type RedisMode string

const (
	RedisModeStandalone RedisMode = "Standalone"
	RedisModeCluster    RedisMode = "Cluster"
	RedisModeSentinel   RedisMode = "Sentinel"
)

// +kubebuilder:validation:Enum=sync;scan
type RedisShakeReader string

const (
	RedisShakeReaderSync RedisShakeReader = "sync"
	RedisShakeReaderScan RedisShakeReader = "scan"
)

// +kubebuilder:validation:Enum=all;full
type RedisShakeSyncMode string

const (
	RedisShakeSyncModeAll  RedisShakeSyncMode = "all"
	RedisShakeSyncModeFull RedisShakeSyncMode = "full"
)

// +kubebuilder:validation:Enum=panic;rewrite;skip
type RedisShakeKeyExistsBehavior string

type RedisSource struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`

	// Mode is the topology of the source. When an AppBinding is used, the operator fills it from the Redis object.
	// +optional
	Mode RedisMode `yaml:"mode" json:"mode,omitempty"`

	// Sentinel is required when Mode is Sentinel and the master address must be resolved through sentinels.
	// +optional
	Sentinel *RedisSentinelSource `yaml:"sentinel" json:"sentinel,omitempty"`

	// +optional
	RedisShake *RedisShake `yaml:"redisShake" json:"redisShake,omitempty"`
}

type RedisSentinelSource struct {
	// Addresses of the sentinels in host:port form
	Addresses  []string `yaml:"addresses" json:"addresses"`
	MasterName string   `yaml:"masterName" json:"masterName"`
	// +optional
	Username string `yaml:"username" json:"username,omitempty"`
	// +optional
	Password string `yaml:"password" json:"password,omitempty"`
}

type RedisTarget struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`

	// Mode is the topology of the target. When an AppBinding is used, the operator fills it from the Redis object.
	// +optional
	Mode RedisMode `yaml:"mode" json:"mode,omitempty"`
}

type RedisShake struct {
	// Reader selects how data is read from the source.
	// sync uses PSYNC replication (full + incremental); scan uses SCAN + DUMP (full only, for services that block SYNC/PSYNC).
	// +optional
	Reader RedisShakeReader `yaml:"reader" json:"reader,omitempty"`

	// SyncMode is all (full + incremental) or full (snapshot only, the job exits when done).
	// +optional
	SyncMode RedisShakeSyncMode `yaml:"syncMode" json:"syncMode,omitempty"`

	// +optional
	Filter *RedisShakeFilter `yaml:"filter" json:"filter,omitempty"`

	// KeyExistsBehavior maps to advanced.rdb_restore_command_behavior.
	// +optional
	KeyExistsBehavior RedisShakeKeyExistsBehavior `yaml:"keyExistsBehavior" json:"keyExistsBehavior,omitempty"`

	// EmptyTargetBeforeSync flushes the target before syncing.
	// +optional
	EmptyTargetBeforeSync bool `yaml:"emptyTargetBeforeSync" json:"emptyTargetBeforeSync,omitempty"`

	// ExtraConfiguration holds raw "section.key" overrides, e.g. "advanced.target_redis_max_qps": "100000".
	// +optional
	ExtraConfiguration map[string]string `yaml:"extraConfiguration" json:"extraConfiguration,omitempty"`
}

type RedisShakeFilter struct {
	// +optional
	AllowKeys []string `yaml:"allowKeys" json:"allowKeys,omitempty"`
	// +optional
	AllowKeyPrefix []string `yaml:"allowKeyPrefix" json:"allowKeyPrefix,omitempty"`
	// +optional
	AllowKeySuffix []string `yaml:"allowKeySuffix" json:"allowKeySuffix,omitempty"`
	// +optional
	AllowKeyRegex []string `yaml:"allowKeyRegex" json:"allowKeyRegex,omitempty"`
	// +optional
	BlockKeys []string `yaml:"blockKeys" json:"blockKeys,omitempty"`
	// +optional
	BlockKeyPrefix []string `yaml:"blockKeyPrefix" json:"blockKeyPrefix,omitempty"`
	// +optional
	BlockKeySuffix []string `yaml:"blockKeySuffix" json:"blockKeySuffix,omitempty"`
	// +optional
	BlockKeyRegex []string `yaml:"blockKeyRegex" json:"blockKeyRegex,omitempty"`
	// +optional
	AllowDB []int `yaml:"allowDB" json:"allowDB,omitempty"`
	// +optional
	BlockDB []int `yaml:"blockDB" json:"blockDB,omitempty"`
	// +optional
	AllowCommand []string `yaml:"allowCommand" json:"allowCommand,omitempty"`
	// +optional
	BlockCommand []string `yaml:"blockCommand" json:"blockCommand,omitempty"`
	// +optional
	AllowCommandGroup []string `yaml:"allowCommandGroup" json:"allowCommandGroup,omitempty"`
	// +optional
	BlockCommandGroup []string `yaml:"blockCommandGroup" json:"blockCommandGroup,omitempty"`
}

func init() {
	SchemeBuilder.Register(&RedisMigration{}, &RedisMigrationList{})
}
