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

// PointId is either a uint64 or a UUID string.
type PointId = any

func PointIdFromNum(n uint64) PointId  { return n }
func PointIdFromUUID(u string) PointId { return u }

// PointStruct represents a point to be upserted into a collection.
type PointStruct struct {
	Id      PointId        `json:"id"`
	Vector  any            `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

// UpsertPointsRequest represents a request to upsert points into a collection.
type UpsertPointsRequest struct {
	Points []PointStruct `json:"points"`
}

// UpsertPointsResponse represents the response from an upsert operation.
type UpsertPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result Result  `json:"result"`
}

// GetPointsRequest represents a request to retrieve points by their IDs.
type GetPointsRequest struct {
	Ids []PointId `json:"ids"`
}

// GetPointsResponse represents the response from a get points operation.
type GetPointsResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result []Point `json:"result"`
}

// Point represents a point retrieved from a collection.
type Point struct {
	Id      PointId        `json:"id"`
	Vector  any            `json:"vector,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}
