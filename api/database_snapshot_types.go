package api

import (
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
	// Backup Spec
	Backup BackupSpec `json:"backup,omitempty"`
}

type DatabaseSnapshotStatus struct {
	StartTime      *unversioned.Time `json:"startTime,omitempty"`
	CompletionTime *unversioned.Time `json:"completionTime,omitempty"`
	Active         *bool             `json:"active,omitempty"`
	Succeeded      *bool             `json:"succeeded,omitempty"`
	Failed         *bool             `json:"failed,omitempty"`
	Message        string            `json:"message,omitempty"`
}

type DatabaseSnapshotList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DatabaseSnapshot TPR objects
	Items []DatabaseSnapshot `json:"items,omitempty"`
}
