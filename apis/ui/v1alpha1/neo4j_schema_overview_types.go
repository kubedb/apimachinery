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
	ResourceKindNeo4jSchemaOverview = "Neo4jSchemaOverview"
	ResourceNeo4jSchemaOverview     = "neo4jschemaoverview"
	ResourceNeo4jSchemaOverviews    = "neo4jschemaoverviews"
)

type Neo4jSchemaOverviewSpec struct {
	Databases []Neo4jDatabaseSchema `json:"databases"`
}

type Neo4jDatabaseSchema struct {
	DatabaseName      string                    `json:"databaseName"`
	NodeTypes         []Neo4jNodeTypeSchema     `json:"nodeTypes,omitempty"`
	RelationshipTypes []Neo4jRelationshipSchema `json:"relationshipTypes,omitempty"`
	Indexes           []Neo4jIndexSchema        `json:"indexes,omitempty"`
	Constraints       []Neo4jConstraintSchema   `json:"constraints,omitempty"`
}

type Neo4jNodeTypeSchema struct {
	NodeType   string                `json:"nodeType"`
	Labels     []string              `json:"labels,omitempty"`
	Properties []Neo4jPropertySchema `json:"properties,omitempty"`
}

type Neo4jRelationshipSchema struct {
	RelationshipType string                `json:"relationshipType"`
	Properties       []Neo4jPropertySchema `json:"properties,omitempty"`
}

type Neo4jPropertySchema struct {
	Name      string   `json:"name"`
	Types     []string `json:"types,omitempty"`
	Mandatory bool     `json:"mandatory"`
}

type Neo4jIndexSchema struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	State             string   `json:"state"`
	EntityType        string   `json:"entityType"`
	LabelsOrTypes     []string `json:"labelsOrTypes,omitempty"`
	Properties        []string `json:"properties,omitempty"`
	PopulationPercent *float64 `json:"populationPercent,omitempty"`
	IndexProvider     string   `json:"indexProvider,omitempty"`
	OwningConstraint  string   `json:"owningConstraint,omitempty"`
}

type Neo4jConstraintSchema struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	EntityType    string   `json:"entityType"`
	LabelsOrTypes []string `json:"labelsOrTypes,omitempty"`
	Properties    []string `json:"properties,omitempty"`
	PropertyType  string   `json:"propertyType,omitempty"`
	OwnedIndex    string   `json:"ownedIndex,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jSchemaOverview struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Neo4jSchemaOverviewSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Neo4jSchemaOverviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Neo4jSchemaOverview `json:"items"`
}
