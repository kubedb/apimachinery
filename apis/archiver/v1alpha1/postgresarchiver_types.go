/*
Copyright 2022.

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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
)

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=postgresarchivers,singular=postgresarchiver,shortName=pgarchiver,categories={archiver,kubedb,appscode}
// +kubebuilder:subresource:status
type PostgresArchiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresArchiverSpec   `json:"spec,omitempty"`
	Status PostgresArchiverStatus `json:"status,omitempty"`
}

// PostgresArchiverSpec defines the desired state of PostgresArchiver
type PostgresArchiverSpec struct {
	Databases *dbapi.AllowedConsumers `json:"databases"`
	// +optional
	Pause bool `json:"pause,omitempty"`
	// +optional
	RetentionPolicy *kmapi.ObjectReference `json:"retentionPolicy"`
	// +optional
	FullBackup *FullBackupOptions `json:"fullBackup"`
	// +optional
	WalBackup *WalBackupOptions `json:"walBackup"`
	// +optional
	BackupStorage *BackupStorage `json:"backupStorage"`
	// +optional
	DeletionPolicy *DeletionPolicy `json:"deletionPolicy"`
}

// PostgresArchiverStatus defines the observed state of PostgresArchiver
type PostgresArchiverStatus struct {
	// Specifies the current phase of the archiver
	// +optional
	Phase ArchiverPhase `json:"phase,omitempty"`
	// Specifies the information of all the database managed by this DB
	// +optional
	DatabaseRefs []ArchiverDatabaseRef `json:"databaseRefs,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PostgresArchiverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresArchiver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgresArchiver{}, &PostgresArchiverList{})
}
