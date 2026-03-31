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

import "encoding/json"

// Distance represents the distance function used to compare vectors.
type Distance string

const (
	DistanceCosine    Distance = "Cosine"
	DistanceEuclid    Distance = "Euclid"
	DistanceDot       Distance = "Dot"
	DistanceManhattan Distance = "Manhattan"
)

// CollectionInfo represents basic information about a collection
type CollectionInfo struct {
	Name string `json:"name"`
}

// CollectionsResult contains the list of collections
type CollectionsResult struct {
	Collections []CollectionInfo `json:"collections"`
}

// VectorParams represents parameters for a single vector data storage
type VectorParams struct {
	Size              uint64   `json:"size"`
	Distance          Distance `json:"distance"`
	OnDisk            *bool    `json:"on_disk,omitempty"`
	Datatype          *string  `json:"datatype,omitempty"`
	MultivectorConfig any      `json:"multivector_config,omitempty"`
}

// VectorsConfig represents vector configuration for single or multiple vector modes
type VectorsConfig struct {
	Single *VectorParams            `json:"-"`
	Named  map[string]*VectorParams `json:"-"`
}

func (v VectorsConfig) MarshalJSON() ([]byte, error) {
	if v.Single != nil {
		return json.Marshal(v.Single)
	}
	if v.Named != nil {
		return json.Marshal(v.Named)
	}
	return json.Marshal(nil)
}

// CreateCollectionRequest represents the request body for creating a collection.
type CreateCollectionRequest struct {
	VectorsConfig          *VectorsConfig `json:"vectors,omitempty"`
	ShardNumber            *uint          `json:"shard_number,omitempty"`
	ReplicationFactor      *uint          `json:"replication_factor,omitempty"`
	WriteConsistencyFactor *uint          `json:"write_consistency_factor,omitempty"`
	OnDiskPayload          *bool          `json:"on_disk_payload,omitempty"`
	Metadata               map[string]any `json:"metadata,omitempty"`
}

// NewVectorsConfig creates a VectorsConfig from a single vector params.
func NewVectorsConfig(params *VectorParams) *VectorsConfig {
	return &VectorsConfig{Single: params}
}

// NewNamedVectorsConfig creates a VectorsConfig from named vector params.
func NewNamedVectorsConfig(named map[string]*VectorParams) *VectorsConfig {
	return &VectorsConfig{Named: named}
}

// CreateCollectionResponse represents the response from creating a collection
type CreateCollectionResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
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
	Models map[string]any `json:"models"`
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
	Name           string         `json:"name"`
	VectorsCount   uint64         `json:"vectors_count"`
	PointsCount    uint64         `json:"points_count"`
	PayloadSchema  map[string]any `json:"payload_schema"`
	Status         string         `json:"status"`
	Conditions     string         `json:"conditions,omitempty"`
	OptimizeHidden *bool          `json:"optimize_hidden,omitempty"`
	AutoMigrate    *bool          `json:"auto_migrate,omitempty"`
	RAMUsage       uint64         `json:"ram_usage,omitempty"`
	DiskUsage      uint64         `json:"disk_usage,omitempty"`
}
