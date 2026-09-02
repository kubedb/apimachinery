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
	ResourceKindNeo4jQueries = "Neo4jQueries"
	ResourceNeo4jQueries     = "neo4jqueries"
	ResourceNeo4jQuerieses   = "neo4jqueries"
)

type Neo4jQueriesSpec struct {
	CollectedAt        *metav1.Time     `json:"collectedAt,omitempty"`
	Queries            []Neo4jQuerySpec `json:"queries"`
	Incomplete         bool             `json:"incomplete,omitempty"`
	UnavailableServers []string         `json:"unavailableServers,omitempty"`
}

type Neo4jQuerySpec struct {
	ServerName        string       `json:"serverName"`
	Database          string       `json:"database"`
	TransactionID     string       `json:"transactionID,omitempty"`
	QueryID           string       `json:"queryID,omitempty"`
	Username          string       `json:"username,omitempty"`
	Query             string       `json:"query"`
	Status            string       `json:"status,omitempty"`
	StartTime         *metav1.Time `json:"startTime,omitempty"`
	ElapsedTimeMillis *int64       `json:"elapsedTimeMillis,omitempty"`
	CPUTimeMillis     *int64       `json:"cpuTimeMillis,omitempty"`
	WaitTimeMillis    *int64       `json:"waitTimeMillis,omitempty"`
	IdleTimeMillis    *int64       `json:"idleTimeMillis,omitempty"`
	AllocatedBytes    *int64       `json:"allocatedBytes,omitempty"`
	PageHits          *int64       `json:"pageHits,omitempty"`
	PageFaults        *int64       `json:"pageFaults,omitempty"`
	ActiveLockCount   *int64       `json:"activeLockCount,omitempty"`
	ConnectionID      string       `json:"connectionID,omitempty"`
	ClientAddress     string       `json:"clientAddress,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jQueries struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Neo4jQueriesSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jQueriesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neo4jQueries `json:"items"`
}
