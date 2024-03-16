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
	"kubestash.dev/apimachinery/apis"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindAddon     = "Addon"
	ResourceSingularAddon = "addon"
	ResourcePluralAddon   = "addons"
)

// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=addons,singular=addon,scope=Cluster,categories={kubestash,appscode,all}
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Addon specifies the backup and restore capabilities for a particular resource.
// For example, MySQL addon specifies the backup and restore capabilities of MySQL database where
// Postgres addon specifies backup and restore capabilities for PostgreSQL database.
// An Addon CR defines the backup and restore tasks that can be performed by this addon.
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AddonSpec `json:"spec,omitempty"`
}

// AddonSpec defines the specification for backup and restore tasks.
type AddonSpec struct {
	// BackupTasks specifies a list of backup tasks that can be performed by the addon.
	BackupTasks []Task `json:"backupTasks,omitempty"`

	// RestoreTasks specifies a list of restore tasks that can be performed by the addon.
	RestoreTasks []Task `json:"restoreTasks,omitempty"`
}

// Task defines the specification of a backup/restore task.
type Task struct {
	// Name specifies the name of the task. The name of a Task should indicate what
	// this task does. For example, a name LogicalBackup indicate that this task performs
	// a logical backup of a database.
	Name string `json:"name,omitempty"`

	// Function specifies the name of a Function CR that defines a container definition
	// which will execute the backup/restore logic for a particular application.
	Function string `json:"function,omitempty"`

	// Driver specifies the underlying tool that will be used to upload the data to the backend storage.
	// Valid values are:
	// - "Restic": The underlying tool is [restic](https://restic.net/).
	// - "WalG": The underlying tool is [wal-g](https://github.com/wal-g/wal-g).
	// +kubebuilder:validation:Enum=Restic;WalG;VolumeSnapshotter;
	Driver apis.Driver `json:"driver,omitempty"`

	// Executor specifies the type of entity that will execute the task. For example, it can be a Job,
	// a sidecar container, an ephemeral container, or a Job that creates additional Jobs/Pods
	// for executing the backup/restore logic.
	// Valid values are:
	// - "Job": Stash will create a Job to execute the backup/restore task.
	// - "Sidecar": Stash will inject a sidecar container into the application to execute the backup/restore task.
	// - "EphemeralContainer": Stash will attach an ephemeral container to the respective Pods to execute the backup/restore task.
	// - "MultiLevelJob": Stash will create a Job that will create additional Jobs/Pods to execute the backup/restore task.
	Executor TaskExecutor `json:"executor,omitempty"`

	// Singleton specifies whether this task will be executed on a single job or across multiple jobs.
	Singleton bool `json:"singleton,omitempty"`

	// Parameters defines a list of parameters that is used by the task to execute its logic.
	// +optional
	Parameters []apis.ParameterDefinition `json:"parameters,omitempty"`

	// VolumeTemplate specifies a list of volume templates that is used by the respective backup/restore
	// Job to execute its logic.
	// User can overwrite these volume templates using `addonVolumes` field of BackupConfiguration/BackupBatch.
	// +optional
	VolumeTemplate []VolumeTemplate `json:"volumeTemplate,omitempty"`

	// VolumeMounts specifies the mount path of the volumes specified in the VolumeTemplate section.
	// These volumes will be mounted directly on the Job/Container created/injected by Stash operator.
	// If the volume type is VolumeClaimTemplate, then Stash operator is responsible for creating the volume.
	// +optional
	VolumeMounts []core.VolumeMount `json:"volumeMounts,omitempty"`

	// PassThroughMounts specifies a list of volume mount for the VolumeTemplates that should be mounted
	// on second level Jobs/Pods created by the first level executor Job.
	// If the volume needs to be mounted on both first level and second level Jobs/Pods, then specify the
	// mount in both VolumeMounts and PassThroughMounts section.
	// If the volume type is VolumeClaimTemplate, then the first level job is responsible for creating the volume.
	// +optional
	PassThroughMounts []core.VolumeMount `json:"passThroughMounts,omitempty"`
}

// TaskExecutor defines the type of the executor that will execute the backup/restore task.
// +kubebuilder:validation:Enum=Job;Sidecar;EphemeralContainer;MultiLevelJob
type TaskExecutor string

const (
	ExecutorJob                TaskExecutor = "Job"
	ExecutorSidecar            TaskExecutor = "Sidecar"
	ExecutorEphemeralContainer TaskExecutor = "EphemeralContainer"
	ExecutorMultiLevelJob      TaskExecutor = "MultiLevelJob"
)

// VolumeTemplate specifies the name, usage, and the source of volume that will be used by the
// addon to execute it's backup/restore task.
type VolumeTemplate struct {
	// Name specifies the name of the volume
	Name string `json:"name,omitempty"`

	// Usage specifies the usage of the volume.
	// +optional
	Usage string `json:"usage,omitempty"`

	// Source specifies the source of this volume.
	Source *apis.VolumeSource `json:"source,omitempty"`
}

//+kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
