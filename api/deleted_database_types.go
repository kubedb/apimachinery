package api

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	ResourceKindDeletedDatabase = "DeletedDatabase"
	ResourceNameDeletedDatabase = "deleted-database"
	ResourceTypeDeletedDatabase = "deleteddatabases"
)

type DeletedDatabase struct {
	unversioned.TypeMeta `json:",inline,omitempty"`
	api.ObjectMeta       `json:"metadata,omitempty"`
	Spec                 DeletedDatabaseSpec   `json:"spec,omitempty"`
	Status               DeletedDatabaseStatus `json:"status,omitempty"`
}

type DeletedDatabaseSpec struct {
	// Database authentication secret
	// +optional
	AuthSecret *api.SecretVolumeSource `json:"authSecret,omitempty"`
	// If true, invoke destroy operation
	// +optional
	Destroy bool `json:"destroy,omitempty"`
	// If true, invoke recover operation
	// +optional
	Recover bool `json:"recover,omitempty"`
	// Origin Spec to store original database information
	Origin OriginSpec `json:"origin,omitempty"`
}

type OriginSpec struct {
	api.ObjectMeta `json:"metadata,omitempty"`
	// OriginDatabase Spec to store original database Spec
	OriginDatabase OriginDatabaseSpec `json:"originDatabase,omitempty"`
}

type OriginDatabaseSpec struct {
	// Elastic Spec
	// +optional
	Elastic *ElasticSpec `json:"elastic,omitempty"`
	// Postgres Spec
	// +optional
	Postgres *PostgresSpec `json:"postgres,omitempty"`
}

type DeletedDatabasePhase string

const (
	// used for Databases that are deleted
	PhaseDatabaseDeleted DeletedDatabasePhase = "Deleted"
	// used for Databases that are currently deleting
	PhaseDatabaseDeleting DeletedDatabasePhase = "Deleting"
	// used for Databases that are destroyed
	PhaseDatabaseDestroyed DeletedDatabasePhase = "Destroyed"
	// used for Databases that are currently destroying
	PhaseDatabaseDestroying DeletedDatabasePhase = "Destroying"
	// used for Databases that are currently recovering
	PhaseDatabaseRecovering DeletedDatabasePhase = "Recovering"
)

type DeletedDatabaseStatus struct {
	CreationTime *unversioned.Time    `json:"creationTime,omitempty"`
	DeletionTime *unversioned.Time    `json:"deletionTime,omitempty"`
	DestroyTime  *unversioned.Time    `json:"destroyTime,omitempty"`
	Phase        DeletedDatabasePhase `json:"phase,omitempty"`
	Reason       string               `json:"reason,omitempty"`
}

type DeletedDatabaseList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DeletedDatabase TPR objects
	Items []DeletedDatabase `json:"items,omitempty"`
}
