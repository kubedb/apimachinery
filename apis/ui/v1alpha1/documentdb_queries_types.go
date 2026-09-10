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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDocumentDBQueries = "DocumentDBQueries"
	ResourceDocumentDBQueries     = "documentdbqueries"
	// ResourceDocumentDBQuerieses mirrors the MongoDB/Postgres quirk: the Kind already
	// ends in "s", so the plural constant is the doubled-es name but the value is the
	// same string.
	ResourceDocumentDBQuerieses = "documentdbqueries"
)

// DocumentDBQueriesSpec defines the desired state of DocumentDBQueries.
//
// Rows come from pg_stat_statements on the PostgreSQL backend. Two caveats a reviewer
// should know: the statements are the rewritten SQL the gateway and the pg_documentdb
// extension emit (documentdb_api.*, documentdb_api_internal.*), not the Mongo commands the
// user typed; and the view is only present if the extension has been created, so an empty
// list is a legitimate result rather than an error.
type DocumentDBQueriesSpec struct {
	Queries []DocumentDBQuerySpec `json:"queries"`
}

type DocumentDBQuerySpec struct {
	// +optional
	DatabaseName string `json:"databaseName,omitempty"`
	Query        string `json:"query"`
	// +optional
	Calls *int64 `json:"calls,omitempty"`
	// +optional
	Rows *int64 `json:"rows,omitempty"`
	// +optional
	TotalTimeMilliSeconds *float64 `json:"totalTimeMilliSeconds,omitempty"`
	// +optional
	MeanTimeMilliSeconds *float64 `json:"meanTimeMilliSeconds,omitempty"`
	// +optional
	MinTimeMilliSeconds *float64 `json:"minTimeMilliSeconds,omitempty"`
	// +optional
	MaxTimeMilliSeconds *float64 `json:"maxTimeMilliSeconds,omitempty"`
	// +optional
	SharedBlksHit *int64 `json:"sharedBlksHit,omitempty"`
	// +optional
	SharedBlksRead *int64 `json:"sharedBlksRead,omitempty"`
}

// DocumentDBQueries is the Schema for the DocumentDBQueries API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DocumentDBQueries struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DocumentDBQueriesSpec `json:"spec,omitempty"`
}

// DocumentDBQueriesList contains a list of DocumentDBQueries

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DocumentDBQueriesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DocumentDBQueries `json:"items"`
}
