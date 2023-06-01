package v1alpha1

import (
	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "kmodules.xyz/client-go/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	stashcoreapi "kubestash.dev/apimachinery/apis/core/v1alpha1"
)

//type DatabasesSelector struct {
//	Selector   *metav1.LabelSelector
//	Namespaces dbapi.ConsumerNamespaces
//}

// +kubebuilder:validation:Enum=CSISnapshotter
type BackupDriver string

const (
	// Takes full backup using CSISnapshotter driver
	BackupDriverCSISnapshotter BackupDriver = "CSISnapshotter"
)

type FullBackupOptions struct {
	Driver BackupDriver `json:"driver"`
	// +optional
	CSISnapshotter *CSISnapshotterOptions `json:"csiSnapshotter,omitempty"`
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

type ManifestBackupOptions struct {
	// +optional
	EncryptionSecret *v1.ObjectReference `json:"encryptionSecret"`
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

type WalBackupOptions struct {
	// +optional
	RuntimeSettings *ofst.RuntimeSettings `json:"runtimeSettings,omitempty"`
	// +optional
	ConfigSecret *GenericSecretReference `json:"configSecret,omitempty"`
}

type CSISnapshotterOptions struct {
	VolumeSnapshotClassName string `json:"volumeSnapshotClassName"`
}

type BackupStorage struct {
	Ref *v1.TypedObjectReference `json:"ref,omitempty"`
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// +kubebuilder:validation:Enum=Delete;WipeOut;DoNotDelete
type DeletionPolicy string

const (
	// Deletes archiver, removes the backup jobs and walg sidecar containers, but keeps the backup data
	DeletionPolicyDelete DeletionPolicy = "Delete"
	// Deletes everything including the backup data
	DeletionPolicyWipeOut DeletionPolicy = "WipeOut"
	// Rejects attempt to delete archiver using ValidationWebhook
	DeletionPolicyDoNotDelete DeletionPolicy = "DoNotDelete"
)

type SchedulerOptions struct {
	Schedule string `json:"schedule"`
	// +optional
	ConcurrencyPolicy batch.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`
	// +optional
	JobTemplate stashcoreapi.JobTemplate `json:"jobTemplate"`
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

// +kubebuilder:validation:Enum=Current
type ArchiverPhase string

const (
	ArchiverPhaseCurrent ArchiverPhase = "Current"
)

type ArchiverDatabaseRef struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type GenericSecretReference struct {
	// Name of the provider secret
	Name string `json:"name"`

	EnvToSecretKey map[string]string `json:"envToSecretKey"`
}
