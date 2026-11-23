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
	ResourceKindYugabyteDBArchiver     = "YugabyteDBArchiver"
	ResourceSingularYugabyteDBArchiver = "yugabytedbarchiver"
	ResourcePluralYugabyteDBArchiver   = "yugabytedbarchivers"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=yugabytedbarchivers,singular=yugabytedbarchiver,shortName=ybarchiver,categories={archiver,kubedb,appscode}
// +kubebuilder:subresource:status
type YugabyteDBArchiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YugabyteDBArchiverSpec   `json:"spec,omitempty"`
	Status YugabyteDBArchiverStatus `json:"status,omitempty"`
}

// YugabyteDBArchiverSpec defines the desired state of YugabyteDBArchiver.
//
// YugabyteDB has no write-ahead log to ship. The history needed for point-in-time restore lives
// inside the snapshots of the database's snapshot schedule (spec.pitr of the YugabyteDB): every
// schedule snapshot keeps the history back to the previous one. The archiver therefore has no
// logBackup. Its full backup exports every snapshot the schedule took since the last run, oldest
// first, so the exported chain covers every moment since the first export. The schedules stay
// owned by spec.pitr; the archiver never creates or deletes them.
type YugabyteDBArchiverSpec struct {
	// Databases define which YugabyteDB databases are allowed to consume this archiver
	Databases *dbapi.AllowedConsumers `json:"databases"`
	// Pause defines if the backup process should be paused or not
	// +optional
	Pause bool `json:"pause,omitempty"`
	// RetentionPolicy field is the RetentionPolicy of the backupConfiguration's backend
	// +optional
	RetentionPolicy *kmapi.ObjectReference `json:"retentionPolicy"`
	// FullBackup defines the sessionConfig of the full backup, which exports the pending snapshots
	// of the snapshot schedules. Its schedule is the export cadence: a schedule snapshot must still
	// exist when the next run starts, so the retentionMinutes of every spec.pitr entry must cover
	// at least two runs plus the backup timeout.
	// This options will eventually go to the full-backup job's yaml
	// +optional
	FullBackup *YugabyteDBFullBackupOptions `json:"fullBackup"`
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

// YugabyteDBFullBackupOptions is FullBackupOptions with the Restic driver: the snapshot files are
// streamed into restic, not taken as volume snapshots.
type YugabyteDBFullBackupOptions struct {
	// +kubebuilder:validation:Enum=Restic
	// +kubebuilder:default:=Restic
	Driver apis.Driver `json:"driver"`
	// +optional
	Task *Task `json:"task,omitempty"`
	// +optional
	Scheduler *SchedulerOptions `json:"scheduler,omitempty"`
	// +optional
	ContainerRuntimeSettings *ofst.ContainerRuntimeSettings `json:"containerRuntimeSettings,omitempty"`
	// +optional
	JobTemplate *ofst.PodTemplateSpec `json:"jobTemplate,omitempty"`
	// +optional
	RetryConfig *stashcoreapi.RetryConfig `json:"retryConfig,omitempty"`
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// +optional
	SessionHistoryLimit int32 `json:"sessionHistoryLimit,omitempty"`
}

// YugabyteDBArchiverStatus defines the observed state of YugabyteDBArchiver
type YugabyteDBArchiverStatus struct {
	// Specifies the information of all the databases managed by this archiver
	// +optional
	DatabaseRefs []ArchiverDatabaseRef `json:"databaseRefs,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type YugabyteDBArchiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YugabyteDBArchiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YugabyteDBArchiver{}, &YugabyteDBArchiverList{})
}
