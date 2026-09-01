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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	ResourceKindNeo4jNodesStats = "Neo4jNodesStats"
	ResourceNeo4jNodesStats     = "neo4jnodesstats"
)

type Neo4jNodesStatsSpec struct {
	CollectedAt *metav1.Time        `json:"collectedAt,omitempty"`
	Servers     []Neo4jServerStat   `json:"servers"`
	Databases   []Neo4jDatabaseStat `json:"databases"`
}

type Neo4jServerStat struct {
	ServerID         string   `json:"serverID,omitempty"`
	Name             string   `json:"name"`
	Address          string   `json:"address,omitempty"`
	State            string   `json:"state"`
	Health           string   `json:"health"`
	Version          string   `json:"version,omitempty"`
	ModeConstraint   string   `json:"modeConstraint,omitempty"`
	Hosting          []string `json:"hosting,omitempty"`
	RequestedHosting []string `json:"requestedHosting,omitempty"`
}

type Neo4jDatabaseStat struct {
	Name            string `json:"name"`
	Type            string `json:"type,omitempty"`
	Access          string `json:"access,omitempty"`
	DatabaseID      string `json:"databaseID,omitempty"`
	ServerID        string `json:"serverID,omitempty"`
	Address         string `json:"address,omitempty"`
	Role            string `json:"role,omitempty"`
	Writer          *bool  `json:"writer,omitempty"`
	RequestedStatus string `json:"requestedStatus,omitempty"`
	CurrentStatus   string `json:"currentStatus,omitempty"`
	StatusMessage   string `json:"statusMessage,omitempty"`
	Default         bool   `json:"default,omitempty"`
	Home            bool   `json:"home,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jNodesStats struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Neo4jNodesStatsSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jNodesStatsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neo4jNodesStats `json:"items"`
}
