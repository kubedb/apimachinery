package api

import (
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type DatabaseSnapshot struct {
	unversioned.TypeMeta `json:",inline,omitempty"`
	api.ObjectMeta       `json:"metadata,omitempty"`
	Spec                 DatabaseSnapshotSpec   `json:"spec,omitempty"`
	Status               DatabaseSnapshotStatus `json:"status,omitempty"`
}

type DatabaseSnapshotSpec struct {
	// Database name
	DatabaseName string `json:"databaseName,omitempty"`
	// Cloud credential secret
	CredSecret *api.SecretVolumeSource `json:"credSecret,omitempty"`
	// Database authentication secret
	// +optional
	AuthSecret *api.SecretVolumeSource `json:"authSecret,omitempty"`
	// Cloud bucket name
	BucketName string `json:"bucketName,omitempty"`
	// Database snapshot id
	// +optional
	SnapshotID string `json:"snapshotID,omitempty"`
}

type DatabaseSnapshotStatus struct {
	Message string    `json:"message,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Success time.Time `json:"success,omitempty"`
}

type DatabaseSnapshotList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DatabaseSnapshot TPR objects
	Items []DatabaseSnapshot `json:"items,omitempty"`
}

type DeletedDatabase struct {
	unversioned.TypeMeta `json:",inline,omitempty"`
	api.ObjectMeta       `json:"metadata,omitempty"`
	Spec                 DeletedDatabaseSpec   `json:"spec,omitempty"`
	Status               DeletedDatabaseStatus `json:"status,omitempty"`
}

type DeletedDatabaseSpec struct {
	// Database name
	DatabaseName string `json:"databaseName,omitempty"`
	// Database authentication secret
	// +optional
	AuthSecret *api.SecretVolumeSource `json:"authSecret,omitempty"`
	// If true, invoke destroy operation
	// +optional
	Destroy *bool `json:"destroy,omitempty"`
}

type DeletedDatabaseStatus struct {
	Message string    `json:"message,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Success time.Time `json:"success,omitempty"`
}

type DeletedDatabaseList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DatabaseSnapshot TPR objects
	Items []DeletedDatabase `json:"items,omitempty"`
}
