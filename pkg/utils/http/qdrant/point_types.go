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

// PointId represents a point identifier in Qdrant
type PointId struct {
	Num  uint64 `json:"num,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

// Value represents a payload value in Qdrant
type Value struct{}

// PointStruct represents a point with vectors and payload for upsert operations
type PointStruct struct {
	Id      PointId          `json:"id"`
	Vectors interface{}      `json:"vectors"`
	Payload map[string]Value `json:"payload,omitempty"`
}

// UpsertPointsRequest represents the request body for upserting points
type UpsertPointsRequest struct {
	CollectionName string        `json:"-"`
	Points         []PointStruct `json:"points"`
}

// UpsertPointsResponse represents the response from upserting points
type UpsertPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result Result  `json:"result"`
}

// GetPointsRequest represents the request body for retrieving points
type GetPointsRequest struct {
	CollectionName string    `json:"-"`
	Ids            []PointId `json:"ids"`
}

// GetPointsResponse represents the response from retrieving points
type GetPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result []Point `json:"result"`
}

// Point represents a retrieved point from Qdrant
type Point struct {
	Id      PointId          `json:"id"`
	Vectors interface{}      `json:"vectors,omitempty"`
	Payload map[string]Value `json:"payload,omitempty"`
}
