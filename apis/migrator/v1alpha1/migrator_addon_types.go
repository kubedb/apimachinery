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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigratorAddon holds the global configuration for the migrator addons

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,path=migratoraddons,singular=migratoraddon,shortName=mgaddon,categories={kubedb,appscode,all}
type MigratorAddon struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of MigratorAddon
	// +required
	Spec MigratorAddonSpec `json:"spec"`
}

// MigratorAddonSpec defines the configuration for migrator addon
type MigratorAddonSpec struct {
	// MigratorImages contains the migrator CLI images for different database types
	MigratorImages MigratorImages `json:"migratorImages,omitempty"`

	// ImagePullSecrets specifies the secrets to pull images from a private registry
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// ImagePullPolicy specifies the image pull policy for all images
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default=IfNotPresent
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// JobDefaults defines default settings for migration jobs
type JobDefaults struct {
	// BackoffLimit specifies the number of retries before marking the job as failed
	// +kubebuilder:default=0
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// TTLSecondsAfterFinished specifies the TTL for completed jobs
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// ActiveDeadlineSeconds specifies the duration in seconds relative to the startTime
	// that the job may be active before the system tries to terminate it
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
}

// MigratorImages holds the migrator CLI images for different database types
type MigratorImages struct {
	// Postgres specifies the migrator CLI image for PostgreSQL migration
	// +optional
	Postgres string `json:"postgres,omitempty"`

	// MySQL specifies the migrator CLI image for MySQL migration
	// +optional
	MySQL string `json:"mysql,omitempty"`

	// MariaDB specifies the migrator CLI image for MariaDB migration
	// +optional
	MariaDB string `json:"mariadb,omitempty"`

	// MSSQLServer specifies the migrator CLI image for Microsoft SQL Server migration
	// +optional
	MSSQLServer string `json:"mssqlserver,omitempty"`

	// MongoDB specifies the migrator CLI image for MongoDB migration
	// +optional
	MongoDB string `json:"mongodb,omitempty"`

	// Sidecar specifies the status reporter sidecar image that runs as a gRPC client
	// to receive migration progress updates and update the Migrator CR status
	// +optional
	Sidecar string `json:"sidecar,omitempty"`
}

// MigratorAddonList contains a list of MigratorAddon

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MigratorAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []MigratorAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MigratorAddon{}, &MigratorAddonList{})
}

// GetImage returns the migrator CLI image based on the Migrator's source/target configuration
func (m *MigratorAddon) GetImage(migrator *Migrator) string {
	switch {
	case migrator.Spec.Source.Postgres != nil && migrator.Spec.Target.Postgres != nil:
		return m.Spec.MigratorImages.Postgres
		// case migrator.Spec.Source.MySQL != nil && migrator.Spec.Target.MySQL != nil:
		//	return m.MigratorImages.MySQL
		// case migrator.Spec.Source.MariaDB != nil && migrator.Spec.Target.MariaDB != nil:
		//	return m.MigratorImages.MariaDB
		// case migrator.Spec.Source.MSSQLServer != nil && migrator.Spec.Target.MSSQLServer != nil:
		//	return m.MigratorImages.MSSQLServer
		// case migrator.Spec.Source.MongoDB != nil && migrator.Spec.Target.MongoDB != nil:
		//	return m.MigratorImages.MongoDB
	}
	return ""
}

func (m *MigratorAddon) GetDatabase(migrator *Migrator) string {
	switch {
	case migrator.Spec.Source.Postgres != nil && migrator.Spec.Target.Postgres != nil:
		return "postgres"
		// case migrator.Spec.Source.MySQL != nil && migrator.Spec.Target.MySQL != nil:
		//	return m.MigratorImages.MySQL
		// case migrator.Spec.Source.MariaDB != nil && migrator.Spec.Target.MariaDB != nil:
		//	return m.MigratorImages.MariaDB
		// case migrator.Spec.Source.MSSQLServer != nil && migrator.Spec.Target.MSSQLServer != nil:
		//	return m.MigratorImages.MSSQLServer
		// case migrator.Spec.Source.MongoDB != nil && migrator.Spec.Target.MongoDB != nil:
		//	return m.MigratorImages.MongoDB
	}
	return ""
}
