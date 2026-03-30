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

package qdrant

// CollectionInfo represents basic information about a collection
type CollectionInfo struct {
	Name string `json:"name"`
}

// CollectionsResult contains the list of collections
type CollectionsResult struct {
	Collections []CollectionInfo `json:"collections"`
}

// UsageHardware represents hardware usage information
type UsageHardware struct {
	CPU                 int `json:"cpu"`
	PayloadIORead       int `json:"payload_io_read"`
	PayloadIOWrite      int `json:"payload_io_write"`
	PayloadIndexIORead  int `json:"payload_index_io_read"`
	PayloadIndexIOWrite int `json:"payload_index_io_write"`
	VectorIORead        int `json:"vector_io_read"`
	VectorIOWrite       int `json:"vector_io_write"`
}

// UsageInference represents inference usage information
type UsageInference struct {
	Models map[string]interface{} `json:"models"`
}

// Usage represents the usage information in the response
type Usage struct {
	Hardware  UsageHardware  `json:"hardware"`
	Inference UsageInference `json:"inference"`
}

// ListCollectionsResponse represents the response from listing all collections
type ListCollectionsResponse struct {
	Usage  *Usage             `json:"usage"`
	Time   float64            `json:"time"`
	Status string             `json:"status"`
	Result *CollectionsResult `json:"result"`
}

// CollectionInfoResponse represents the response from getting collection info
type CollectionInfoResponse struct {
	Usage  *Usage             `json:"usage"`
	Time   float64            `json:"time"`
	Status string             `json:"status"`
	Result *CollectionDetails `json:"result"`
}

// CollectionDetails contains detailed information about a collection
type CollectionDetails struct {
	Name           string                 `json:"name"`
	VectorsCount   uint64                 `json:"vectors_count"`
	PointsCount    uint64                 `json:"points_count"`
	PayloadSchema  map[string]interface{} `json:"payload_schema"`
	Status         string                 `json:"status"`
	Conditions     string                 `json:"conditions,omitempty"`
	OptimizeHidden *bool                  `json:"optimize_hidden,omitempty"`
	AutoMigrate    *bool                  `json:"auto_migrate,omitempty"`
	RAMUsage       uint64                 `json:"ram_usage,omitempty"`
	DiskUsage      uint64                 `json:"disk_usage,omitempty"`
}

// VectorParams defines the configuration for a single vector field
type VectorParams struct {
	Size     uint64 `json:"size"`
	Distance string `json:"distance"`
}

// VectorParamsMap maps vector names to their configurations
type VectorParamsMap map[string]*VectorParams

// CreateCollectionRequest represents the request body for creating a collection
type CreateCollectionRequest struct {
	CollectionName string      `json:"-"`
	VectorsConfig  interface{} `json:"vectors_config"`
}

// CreateCollectionResponse represents the response from creating a collection
type CreateCollectionResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
}
