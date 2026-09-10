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
	ResourceKindDocumentDBSchemaOverview = "DocumentDBSchemaOverview"
	ResourceDocumentDBSchemaOverview     = "documentdbschemaoverview"
	ResourceDocumentDBSchemaOverviews    = "documentdbschemaoverviews"
)

// DocumentDBSchemaOverviewSpec defines the desired state of DocumentDBSchemaOverview.
//
// DocumentDB presents collections over the MongoDB wire protocol, so the unit here is a
// collection, mirroring MongoDBSchemaOverview rather than the generic table shape. Unlike
// MongoDBSchemaOverview, sizes are scalars: DocumentDB is not sharded, so there is no
// per-shard array to report.
type DocumentDBSchemaOverviewSpec struct {
	Collections []DocumentDBCollectionSpec `json:"collections"`
}

// DocumentDBCollectionSpec describes one collection.
type DocumentDBCollectionSpec struct {
	DatabaseName string `json:"databaseName"`
	Name         string `json:"name"`

	// +optional
	DocumentCount *int64 `json:"documentCount,omitempty"`
	// SizeBytes is the size of the collection's documents.
	// +optional
	SizeBytes *int64 `json:"sizeBytes,omitempty"`
	// StorageSizeBytes is the on-disk size including padding and free space.
	// +optional
	StorageSizeBytes *int64 `json:"storageSizeBytes,omitempty"`
	// +optional
	TotalIndexSizeBytes *int64 `json:"totalIndexSizeBytes,omitempty"`
	// +optional
	IndexCount *int32 `json:"indexCount,omitempty"`
}

// DocumentDBSchemaOverview is the Schema for the DocumentDBSchemaOverviews API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DocumentDBSchemaOverview struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DocumentDBSchemaOverviewSpec `json:"spec,omitempty"`
}

// DocumentDBSchemaOverviewList contains a list of DocumentDBSchemaOverview

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DocumentDBSchemaOverviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DocumentDBSchemaOverview `json:"items"`
}
