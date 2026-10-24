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
	ResourceCodeDurableWorkflowVersion     = "dwversion"
	ResourceKindDurableWorkflowVersion     = "DurableWorkflowVersion"
	ResourceSingularDurableWorkflowVersion = "durableworkflowversion"
	ResourcePluralDurableWorkflowVersion   = "durableworkflowversions"
)

// DurableWorkflowVersion defines a Temporal or Cadence version.

// +genclient
// +genclient:nonNamespaced
// +genclient:skipVerbs=updateStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=durableworkflowversions,singular=durableworkflowversion,scope=Cluster,shortName=dwversion,categories={catalog,kubedb,appscode}
// +kubebuilder:printcolumn:name="Distribution",type="string",JSONPath=".spec.distribution"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="SERVER_IMAGE",type="string",JSONPath=".spec.server.image"
// +kubebuilder:printcolumn:name="Deprecated",type="boolean",JSONPath=".spec.deprecated"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DurableWorkflowVersion struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DurableWorkflowVersionSpec `json:"spec,omitempty"`
}

// DurableWorkflowVersionSpec is the spec for a DurableWorkflowVersion.
type DurableWorkflowVersionSpec struct {
	// Version of the engine, for example 1.32.1 or 1.4.2.
	Version string `json:"version"`

	// Distribution is the engine: Temporal or Cadence.
	Distribution DurableWorkflowDistro `json:"distribution"`

	// EndOfLife refers if this version reached into its end of the life or not, based on https://endoflife.date/
	// +optional
	EndOfLife bool `json:"endOfLife"`

	// Deprecated versions usable but regarded as obsolete and best avoided, typically due to having been superseded.
	// +optional
	Deprecated bool `json:"deprecated,omitempty"`

	// Server is the image that runs every server role.
	Server DurableWorkflowVersionImage `json:"server"`

	// AdminTools is the image of the schema Job and the admin tools pod.
	// Cadence uses the server image.
	AdminTools DurableWorkflowVersionImage `json:"adminTools"`

	// UI is the web UI image.
	// +optional
	UI *DurableWorkflowVersionImage `json:"ui,omitempty"`

	// Ports of each server role.
	Ports DurableWorkflowVersionPorts `json:"ports"`

	// Schema describes the persistence schema tooling.
	Schema DurableWorkflowVersionSchema `json:"schema"`

	// DefaultHistoryShards is used when spec.numHistoryShards is omitted.
	// +kubebuilder:default=16
	// +optional
	DefaultHistoryShards int32 `json:"defaultHistoryShards,omitempty"`

	// Rollout holds what zero-downtime rolling needs.
	// +optional
	Rollout DurableWorkflowVersionRollout `json:"rollout,omitempty"`

	// SecurityContext is for the additional config for the server containers.
	// +optional
	SecurityContext SecurityContext `json:"securityContext"`

	// update constraints
	// +optional
	UpdateConstraints UpdateConstraints `json:"updateConstraints,omitempty"`
}

// +kubebuilder:validation:Enum=Temporal;Cadence
type DurableWorkflowDistro string

const (
	DurableWorkflowDistroTemporal DurableWorkflowDistro = "Temporal"
	DurableWorkflowDistroCadence  DurableWorkflowDistro = "Cadence"
)

// DurableWorkflowVersionImage is a container image.
type DurableWorkflowVersionImage struct {
	Image string `json:"image"`
}

// DurableWorkflowVersionPorts lists the ports of every server role. A zero port means the role doesn't
// have that port.
type DurableWorkflowVersionPorts struct {
	Frontend         DurableWorkflowRolePorts `json:"frontend"`
	History          DurableWorkflowRolePorts `json:"history"`
	Matching         DurableWorkflowRolePorts `json:"matching"`
	Worker           DurableWorkflowRolePorts `json:"worker"`
	InternalFrontend DurableWorkflowRolePorts `json:"internalFrontend,omitempty"`
	// Metrics is the Prometheus port of every role.
	Metrics int32 `json:"metrics"`
	// UI is the port of the web UI.
	UI int32 `json:"ui,omitempty"`
}

// DurableWorkflowRolePorts are the ports of one role.
type DurableWorkflowRolePorts struct {
	// GRPC is the gRPC port.
	// +optional
	GRPC int32 `json:"grpc,omitempty"`
	// Membership is the membership (ringpop) port. Temporal only; Cadence's ringpop uses TChannel.
	// +optional
	Membership int32 `json:"membership,omitempty"`
	// HTTP is the HTTP API port. Temporal frontend only.
	// +optional
	HTTP int32 `json:"http,omitempty"`
	// TChannel is the TChannel port. Cadence only, and also its membership port.
	// +optional
	TChannel int32 `json:"tchannel,omitempty"`
}

// DurableWorkflowVersionSchema describes the schema tooling of a version.
type DurableWorkflowVersionSchema struct {
	// Tool is how the schema is applied: "temporal-tools" (temporal-sql-tool, temporal-cassandra-tool)
	// or "cadence-server" (cadence-server update-schema).
	// +kubebuilder:validation:Enum=temporal-tools;cadence-server
	Tool string `json:"tool"`
	// Dir is the directory in the admin tools image that holds the versioned schema, for example
	// /etc/temporal/schema.
	// +optional
	Dir string `json:"dir,omitempty"`
	// Stores lists the supported default and visibility stores.
	// +optional
	Stores []DurableWorkflowVersionStore `json:"stores,omitempty"`
}

// DurableWorkflowVersionStore is a supported persistence engine of a version.
type DurableWorkflowVersionStore struct {
	// Type is Postgres, MySQL, Cassandra, Elasticsearch or OpenSearch.
	Type string `json:"type"`
	// Default is true if the engine can be the default store.
	Default bool `json:"default,omitempty"`
	// Visibility is true if the engine can be the visibility store.
	Visibility bool `json:"visibility,omitempty"`
}

// DurableWorkflowVersionRollout holds what zero-downtime rolling needs.
type DurableWorkflowVersionRollout struct {
	// RoleOrder is the order in which the roles are rolled.
	// +optional
	RoleOrder []string `json:"roleOrder,omitempty"`

	// FrontendShutdownFailHealthCheckDuration is how long a frontend reports unhealthy before it stops
	// accepting requests. Temporal only.
	// +optional
	FrontendShutdownFailHealthCheckDuration *metav1.Duration `json:"frontendShutdownFailHealthCheckDuration,omitempty"`
	// FrontendShutdownDrainDuration is the drain duration of the frontend.
	// +optional
	FrontendShutdownDrainDuration *metav1.Duration `json:"frontendShutdownDrainDuration,omitempty"`
	// MatchingShutdownDrainDuration is the drain duration of matching.
	// +optional
	MatchingShutdownDrainDuration *metav1.Duration `json:"matchingShutdownDrainDuration,omitempty"`
	// HistoryShutdownDrainDuration is the drain duration of history.
	// +optional
	HistoryShutdownDrainDuration *metav1.Duration `json:"historyShutdownDrainDuration,omitempty"`

	// PreStopSleep is the sleep before the container receives SIGTERM.
	// +optional
	PreStopSleep *metav1.Duration `json:"preStopSleep,omitempty"`

	// HistoryMinInterval is the minimum interval between two history pod replacements.
	// +optional
	HistoryMinInterval *metav1.Duration `json:"historyMinInterval,omitempty"`

	// PostHopSettle is how long to wait after a version hop before starting the next one.
	// +optional
	PostHopSettle *metav1.Duration `json:"postHopSettle,omitempty"`

	// OnlineSchemaUpdate is false if the schema update from the previous hop locks heavily and must not run
	// while the old servers are serving.
	// +kubebuilder:default=true
	// +optional
	OnlineSchemaUpdate *bool `json:"onlineSchemaUpdate,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DurableWorkflowVersionList is a list of DurableWorkflowVersions
type DurableWorkflowVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of DurableWorkflowVersion CRD objects
	Items []DurableWorkflowVersion `json:"items,omitempty"`
}
