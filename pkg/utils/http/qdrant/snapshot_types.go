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

// Snapshot represents information about a snapshot.
type Snapshot struct {
	Name         string `json:"name"`
	Size         uint64 `json:"size"`
	CreationTime string `json:"creation_time"`
	Checksum     string `json:"checksum"`
}

// CreateSnapshotResponse represents the response from creating a snapshot.
type CreateSnapshotResponse struct {
	Time   float64  `json:"time"`
	Status string   `json:"status"`
	Result Snapshot `json:"result"`
}

// ListSnapshotsResponse represents the response from listing snapshots.
type ListSnapshotsResponse struct {
	Time   float64    `json:"time"`
	Status string     `json:"status"`
	Result []Snapshot `json:"result"`
}

// DeleteSnapshotResponse represents the response from deleting a snapshot.
type DeleteSnapshotResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
}

// RecoverSnapshotResponse represents the response from recovering a snapshot.
type RecoverSnapshotResponse struct {
	Time   float64 `json:"time"`
	Status string  `json:"status"`
	Result bool    `json:"result"`
}
