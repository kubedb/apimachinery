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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"kubestash.dev/apimachinery/apis"
	stashcoreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
)

const (
	ResourceKindOracleArchiver     = "OracleArchiver"
	ResourceSingularOracleArchiver = "oraclearchiver"
	ResourcePluralOracleArchiver   = "oraclearchivers"
)

// OracleArchiver backs up the Oracle databases it selects with RMAN, and keeps
// them restorable to any point in time inside the retained full backups.
//
// Every full backup is an RMAN INCREMENTAL LEVEL 0. Between full backups an
// incremental-backup loop takes LEVEL 1 backups and ships the archived redo
// logs every minute. A restore to a timestamp restores the LEVEL 0 whose
// window covers it, applies its LEVEL 1 backups and replays the logs.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=oraclearchivers,singular=oraclearchiver,shortName=oraarchiver,categories={archiver,kubedb,appscode}
// +kubebuilder:subresource:status
type OracleArchiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OracleArchiverSpec   `json:"spec,omitempty"`
	Status OracleArchiverStatus `json:"status,omitempty"`
}

// OracleArchiverSpec defines the desired state of OracleArchiver
type OracleArchiverSpec struct {
	// Databases define which Oracle databases are allowed to consume this archiver
	Databases *dbapi.AllowedConsumers `json:"databases"`
	// Pause defines if the backup process should be paused or not
	// +optional
	Pause bool `json:"pause,omitempty"`
	// RetentionPolicy field is the RetentionPolicy of the backupConfiguration's backend.
	// It also bounds the point-in-time recovery window: the LEVEL 1 and archived
	// log backups of a full backup are deleted together with it.
	// +optional
	RetentionPolicy *kmapi.ObjectReference `json:"retentionPolicy"`
	// FullBackup defines the sessionConfig of the full (LEVEL 0) backup.
	// This options will eventually go to the full-backup job's yaml
	// +optional
	FullBackup *OracleFullBackupOptions `json:"fullBackup"`
	// IncrementalBackup configures the incremental-backup loop that takes RMAN
	// LEVEL 1 backups and ships the archived redo logs between full backups.
	// +optional
	IncrementalBackup *OracleIncrementalBackupOptions `json:"incrementalBackup"`
	// ManifestBackup defines the sessionConfig of the manifestBackup
	// This options will eventually go to the manifest-backup job's yaml
	// +optional
	ManifestBackup *ManifestBackupOptions `json:"manifestBackup"`
	// +optional
	EncryptionSecret *kmapi.ObjectReference `json:"encryptionSecret"`
	// BackupStorage is the backend storageRef of the BackupConfiguration
	// +optional
	BackupStorage *BackupStorage `json:"backupStorage"`
	// DeletionPolicy defines the created repository's deletionPolicy
	// +optional
	DeletionPolicy *storageapi.BackupConfigDeletionPolicy `json:"deletionPolicy"`
}

type OracleFullBackupOptions struct {
	// Driver specifies the mechanism used to perform a full backup.
	// +kubebuilder:default:=OracleBackup
	Driver apis.Driver `json:"driver"`

	// Task specifies the task configuration for a full backup.
	// +optional
	Task *Task `json:"task,omitempty"`

	// Scheduler specifies the scheduling configuration for full backups.
	// +optional
	Scheduler *SchedulerOptions `json:"scheduler,omitempty"`

	// ContainerRuntimeSettings specifies the runtime settings for the full backup containers.
	// +optional
	ContainerRuntimeSettings *ofst.ContainerRuntimeSettings `json:"containerRuntimeSettings,omitempty"`

	// JobTemplate specifies the pod template for a full backup job.
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`

	// RetryConfig specifies the retry policy for a failed full backup.
	// +optional
	RetryConfig *stashcoreapi.RetryConfig `json:"retryConfig,omitempty"`

	// Timeout specifies the maximum duration allowed for a full backup.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^([0-9]+(ns|us|ms|s|m|h))+$')",message="must be a Go duration such as 30s, 5m or 1h30m"
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// SessionHistoryLimit specifies the number of full backup session histories to retain.
	// +optional
	SessionHistoryLimit int32 `json:"sessionHistoryLimit,omitempty"`
}

// OracleIncrementalBackupOptions configures the incremental-backup loop.
type OracleIncrementalBackupOptions struct {
	// BackupInterval is the time between RMAN LEVEL 1 backups. A LEVEL 1 only
	// shortens how much redo a restore replays; the archived logs are what make
	// any second restorable.
	// +kubebuilder:default="1h"
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^([0-9]+(ns|us|ms|s|m|h))+$')",message="must be a Go duration such as 30s, 5m or 1h30m"
	BackupInterval *metav1.Duration `json:"backupInterval,omitempty"`

	// ArchivelogInterval is the time between archived log cycles, each of which
	// ships the archived logs written since the last cycle and a control file.
	// Zero disables them, leaving only LEVEL 1 boundaries restorable.
	// +kubebuilder:default="1m"
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^([0-9]+(ns|us|ms|s|m|h))+$')",message="must be a Go duration such as 30s, 5m or 1h30m"
	ArchivelogInterval *metav1.Duration `json:"archivelogInterval,omitempty"`

	// ArchiveLagTarget sets the database's ARCHIVE_LAG_TARGET, so it switches
	// logs at least this often while it writes redo. With ArchivelogInterval it
	// bounds how much data a restore can lose. Zero leaves the database setting
	// alone; otherwise 60s to 2h. Ignored on DataGuard, set it on the primary.
	// +kubebuilder:default="60s"
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^([0-9]+(ns|us|ms|s|m|h))+$')",message="must be a Go duration such as 30s, 5m or 1h30m"
	ArchiveLagTarget *metav1.Duration `json:"archiveLagTarget,omitempty"`

	// ForceLogging turns on FORCE LOGGING, so NOLOGGING writes still produce redo;
	// without it those blocks are corrupt after a point-in-time restore.
	// +kubebuilder:default=true
	// +optional
	ForceLogging *bool `json:"forceLogging,omitempty"`

	// MinControlFileRecordKeepDays raises CONTROL_FILE_RECORD_KEEP_TIME to at
	// least this many days; it is never lowered. Keep it above the retention
	// window, or the control file forgets backups a restore needs. Zero leaves
	// the database setting alone.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=365
	// +optional
	MinControlFileRecordKeepDays int32 `json:"minControlFileRecordKeepDays,omitempty"`

	// Cumulative makes every LEVEL 1 cover all changes since the LEVEL 0 instead
	// of since the previous LEVEL 1.
	// +optional
	Cumulative bool `json:"cumulative,omitempty"`

	// BlockChangeTracking enables Oracle Block Change Tracking, so a LEVEL 1 reads
	// a bitmap instead of scanning every datafile. Enterprise Edition only;
	// ignored on DataGuard.
	// +kubebuilder:default=true
	// +optional
	BlockChangeTracking *bool `json:"blockChangeTracking,omitempty"`

	// RMANTimeout is how long one RMAN run of the loop may take.
	// +kubebuilder:default="6h"
	// +optional
	// +kubebuilder:validation:XValidation:rule="self.matches('^([0-9]+(ns|us|ms|s|m|h))+$')",message="must be a Go duration such as 30s, 5m or 1h30m"
	RMANTimeout *metav1.Duration `json:"rmanTimeout,omitempty"`

	// RuntimeSettings specifies the runtime settings of the loop's pod.
	// +optional
	RuntimeSettings *ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`

	// SuccessfulLogHistoryLimit defines the number of successful cycles the incremental snapshot will retain.
	// +kubebuilder:default=5
	// +optional
	SuccessfulLogHistoryLimit int32 `json:"successfulLogHistoryLimit,omitempty"`

	// FailedLogHistoryLimit defines the number of failed cycles the incremental snapshot will retain for debugging purposes.
	// +kubebuilder:default=5
	// +optional
	FailedLogHistoryLimit int32 `json:"failedLogHistoryLimit,omitempty"`
}

// OracleArchiverStatus defines the observed state of OracleArchiver
type OracleArchiverStatus struct {
	// Specifies the information of all the databases managed by this archiver
	// +optional
	DatabaseRefs []ArchiverDatabaseRef `json:"databaseRefs,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type OracleArchiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OracleArchiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OracleArchiver{}, &OracleArchiverList{})
}
