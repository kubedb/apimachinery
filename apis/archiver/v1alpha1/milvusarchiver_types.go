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
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"
)

const (
	ResourceKindMilvusArchiver     = "MilvusArchiver"
	ResourceSingularMilvusArchiver = "milvusarchiver"
	ResourcePluralMilvusArchiver   = "milvusarchivers"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=milvusarchivers,singular=milvusarchiver,shortName=mvarchiver,categories={archiver,kubedb,appscode}
// +kubebuilder:subresource:status
type MilvusArchiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MilvusArchiverSpec   `json:"spec,omitempty"`
	Status MilvusArchiverStatus `json:"status,omitempty"`
}

// MilvusArchiverSpec defines the desired state of MilvusArchiver.
//
// Milvus keeps its durable state in three places: the meta etcd, the object
// storage bucket and, for a standalone Milvus that uses RocksMQ, the data
// PVC. A FullBackup therefore captures the etcd metadata at a single revision
// together with the bucket objects, while a LogBackup continuously records
// every etcd change and every new object version so that the database can be
// restored to any point in time inside the recorded window.
//
// Point-in-time recovery requires the Woodpecker write-ahead log stored on the
// object storage. A standalone Milvus that keeps its WAL in RocksMQ on a PVC
// can only be restored to the time of a full backup.
type MilvusArchiverSpec struct {
	// Databases define which Milvus databases are allowed to consume this archiver
	Databases *dbapi.AllowedConsumers `json:"databases"`
	// Pause defines if the backup process should be paused or not
	// +optional
	Pause bool `json:"pause,omitempty"`
	// RetentionPolicy field is the RetentionPolicy of the backupConfiguration's backend
	// +optional
	RetentionPolicy *kmapi.ObjectReference `json:"retentionPolicy"`
	// FullBackup defines the sessionConfig of the fullBackup. The driver is either
	// Restic or VolumeSnapshotter; the latter snapshots the data PVC of a standalone
	// RocksMQ Milvus while metadata and objects are always archived by the plugin.
	// +optional
	FullBackup *FullBackupOptions `json:"fullBackup"`
	// LogBackup defines the sidekick configuration of the continuous archiver.
	// Without it only full backups are taken.
	// +optional
	LogBackup *LogBackupOptions `json:"logBackup"`
	// ManifestBackup defines the sessionConfig of the manifestBackup
	// +optional
	ManifestBackup *ManifestBackupOptions `json:"manifestBackup"`
	// EncryptionSecret holds the RESTIC_PASSWORD that encrypts the restic
	// repositories and the archived change log.
	// +optional
	EncryptionSecret *kmapi.ObjectReference `json:"encryptionSecret"`
	// BackupStorage is the backend storageRef of the BackupConfiguration
	// +optional
	BackupStorage *BackupStorage `json:"backupStorage"`
	// DeletionPolicy defines the created repository's deletionPolicy
	// +optional
	DeletionPolicy *storageapi.BackupConfigDeletionPolicy `json:"deletionPolicy"`
}

// MilvusArchiverStatus defines the observed state of MilvusArchiver
type MilvusArchiverStatus struct {
	// Specifies the information of all the databases managed by this archiver
	// +optional
	DatabaseRefs []ArchiverDatabaseRef `json:"databaseRefs,omitempty"`
	// Conditions of the archiver, e.g. PITRUnsupported.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MilvusArchiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MilvusArchiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MilvusArchiver{}, &MilvusArchiverList{})
}
