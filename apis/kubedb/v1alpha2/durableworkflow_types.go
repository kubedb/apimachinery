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

package v1alpha2

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"reflect"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeDurableWorkflow     = "dw"
	ResourceKindDurableWorkflow     = "DurableWorkflow"
	ResourceSingularDurableWorkflow = "durableworkflow"
	ResourcePluralDurableWorkflows  = "durableworkflows"
)

// DurableWorkflow is a durable workflow engine cluster. The engine (Temporal or Cadence) is selected by
// the DurableWorkflowVersion named in spec.version. The server roles are stateless; all state lives in
// the persistence stores referenced by spec.persistence.
//
// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=durableworkflows,singular=durableworkflow,shortName=dw,categories={datastore,kubedb,appscode,all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DurableWorkflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DurableWorkflowSpec   `json:"spec,omitempty"`
	Status DurableWorkflowStatus `json:"status,omitempty"`
}

// DurableWorkflowSpec defines the desired state of DurableWorkflow.
type DurableWorkflowSpec struct {
	// Version is the name of a DurableWorkflowVersion, for example "temporal-1.32.1" or "cadence-1.4.2".
	// The distribution (Temporal or Cadence) comes from that version.
	Version string `json:"version"`

	// NumHistoryShards is the number of history shards. It can't be changed after creation.
	// If omitted, the default of the DurableWorkflowVersion is used (16).
	// +kubebuilder:validation:Minimum=1
	// +optional
	NumHistoryShards *int32 `json:"numHistoryShards,omitempty"`

	// Topology configures the replicas and pod templates of each server role.
	// +optional
	Topology *DurableWorkflowTopology `json:"topology,omitempty"`

	// Persistence references the stores that hold all workflow state.
	Persistence DurableWorkflowPersistence `json:"persistence"`

	// Configuration provides custom static server configuration and dynamic configuration.
	// The key "dynamic_config.yaml" is the dynamic configuration, which the servers reload without a restart.
	// +optional
	Configuration *ConfigurationSpec `json:"configuration,omitempty"`

	// Archival of closed workflow histories and visibility records.
	// +optional
	Archival *DurableWorkflowArchival `json:"archival,omitempty"`

	// TLS configures certificates for internode, frontend and client traffic. The certificate aliases are
	// "internode", "frontend" and "client".
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// FrontendClientAuth controls whether the frontend requires client certificates.
	// Defaults to Required for Temporal and NotRequired for Cadence (the Cadence web UI can't present a
	// client certificate).
	// +optional
	FrontendClientAuth *DurableWorkflowClientAuth `json:"frontendClientAuth,omitempty"`

	// Authorization configures JWT authorization of API calls.
	// +optional
	Authorization *DurableWorkflowAuthorization `json:"authorization,omitempty"`

	// UI configures the web UI.
	// +optional
	UI *DurableWorkflowUI `json:"ui,omitempty"`

	// AdminTools configures a pod with the admin command line tools.
	// +optional
	AdminTools *DurableWorkflowAdminTools `json:"adminTools,omitempty"`

	// ServiceTemplates is an optional configuration for the services that expose the frontend.
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Monitor is used to monitor the cluster.
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`

	// DeletionPolicy controls the delete operation.
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// Halted indicates that all server pods are removed. The configuration, secrets and the persistence
	// stores are kept.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// HealthChecker defines attributes of the health checker.
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 3}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`
}

// DurableWorkflowTopology holds the server roles.
type DurableWorkflowTopology struct {
	// +optional
	Frontend *DurableWorkflowNode `json:"frontend,omitempty"`
	// +optional
	History *DurableWorkflowNode `json:"history,omitempty"`
	// +optional
	Matching *DurableWorkflowNode `json:"matching,omitempty"`
	// +optional
	Worker *DurableWorkflowNode `json:"worker,omitempty"`
	// InternalFrontend is only available for Temporal.
	// +optional
	InternalFrontend *DurableWorkflowNode `json:"internalFrontend,omitempty"`
}

// DurableWorkflowNode configures one server role.
type DurableWorkflowNode struct {
	// Replicas is the number of pods of this role.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// PodTemplate is an optional configuration for the pods of this role.
	// +optional
	PodTemplate *ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// DurableWorkflowPersistence references the persistence stores.
type DurableWorkflowPersistence struct {
	// Default is the store for workflow state: Postgres, MySQL or Cassandra. Its AppBinding reference is
	// required unless Managed is set.
	// +optional
	Default DurableWorkflowStore `json:"default,omitempty"`

	// Visibility is the store for visibility records. If omitted, the default store's server is used with
	// a separate database or keyspace. It can't be set together with Managed.
	// +optional
	Visibility *DurableWorkflowStore `json:"visibility,omitempty"`

	// Managed makes the operator create the backing database itself, a Postgres or MySQL named
	// "<name>-db" that is owned by this DurableWorkflow, instead of referencing an existing one. It is exclusive
	// with default.name and visibility, and immutable: the database is afterwards managed with its own
	// ops requests (an UpdateVersion of the DurableWorkflow never touches it).
	// +optional
	Managed *DurableWorkflowManagedStore `json:"managed,omitempty"`

	// KafkaRef is the AppBinding of a KubeDB Kafka. It is required by Cadence when the visibility store is
	// Elasticsearch or OpenSearch (the history service writes visibility records to a topic of it, and the worker
	// service indexes them), and is rejected otherwise. The operator creates the topics.
	// +optional
	KafkaRef *kmapi.ObjectReference `json:"kafkaRef,omitempty"`
}

// DurableWorkflowManagedEngine is the engine of a managed store.
// +kubebuilder:validation:Enum=Postgres;MySQL
type DurableWorkflowManagedEngine string

const (
	DurableWorkflowManagedPostgres DurableWorkflowManagedEngine = "Postgres"
	DurableWorkflowManagedMySQL    DurableWorkflowManagedEngine = "MySQL"
)

// DurableWorkflowManagedStore describes the database the operator creates for the default store (and the visibility
// store, which shares its server). The child inherits spec.deletionPolicy of the DurableWorkflow. Its connection
// limit is sized from the topology when it is created.
type DurableWorkflowManagedStore struct {
	// Engine is Postgres (default) or MySQL.
	// +kubebuilder:default=Postgres
	// +optional
	Engine DurableWorkflowManagedEngine `json:"engine,omitempty"`

	// Version is the name of a PostgresVersion or MySQLVersion. If omitted, the newest catalog entry of the
	// engine that is official, not deprecated, not end of life and needs no license is used. The choice is written
	// to the spec on creation.
	// +optional
	Version string `json:"version,omitempty"`

	// Replicas is the number of database pods: 1 (default) or 3 for a highly available database.
	// Postgres forms a streaming replication cluster, MySQL a group replication cluster.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Storage specifies the volume of each database pod.
	Storage core.PersistentVolumeClaimSpec `json:"storage"`

	// PodTemplate is an optional configuration for the database pods.
	// +optional
	PodTemplate *ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// DurableWorkflowStore references a persistence store.
type DurableWorkflowStore struct {
	// ObjectReference is the AppBinding of a KubeDB-managed or external database.
	*kmapi.ObjectReference `json:",omitempty"`

	// Type is required only when the AppBinding does not identify the engine.
	// +optional
	Type DurableWorkflowStoreType `json:"type,omitempty"`

	// DatabaseName is the database, keyspace or index name.
	// Defaults to "<name>" for the default store and "<name>_visibility" for the visibility store. A Cadence
	// Elasticsearch or OpenSearch index must start with "cadence-visibility-", and then defaults to
	// "cadence-visibility-<namespace>-<name>".
	// +optional
	DatabaseName string `json:"databaseName,omitempty"`

	// CreateDatabase lets the operator create the database or keyspace and a dedicated user, using the
	// admin credentials of the AppBinding. If false, both must exist already and AuthSecret is required.
	// Defaults to true.
	// +optional
	CreateDatabase *bool `json:"createDatabase,omitempty"`

	// AuthSecret holds the credentials of the user the servers connect with. The operator generates it
	// when CreateDatabase is true.
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// Cassandra holds the keyspace options. It is only used for Cassandra.
	// +optional
	Cassandra *DurableWorkflowCassandraOptions `json:"cassandra,omitempty"`
}

// DurableWorkflowCassandraOptions are used when the operator creates a keyspace.
type DurableWorkflowCassandraOptions struct {
	// +kubebuilder:default=1
	// +optional
	ReplicationFactor int32 `json:"replicationFactor,omitempty"`
	// +optional
	Datacenter string `json:"datacenter,omitempty"`
}

// +kubebuilder:validation:Enum=Postgres;MySQL;Cassandra;Elasticsearch;OpenSearch
type DurableWorkflowStoreType string

const (
	DurableWorkflowStorePostgres      DurableWorkflowStoreType = "Postgres"
	DurableWorkflowStoreMySQL         DurableWorkflowStoreType = "MySQL"
	DurableWorkflowStoreCassandra     DurableWorkflowStoreType = "Cassandra"
	DurableWorkflowStoreElasticsearch DurableWorkflowStoreType = "Elasticsearch"
	DurableWorkflowStoreOpenSearch    DurableWorkflowStoreType = "OpenSearch"
)

// +kubebuilder:validation:Enum=Required;NotRequired
type DurableWorkflowClientAuth string

const (
	DurableWorkflowClientAuthRequired    DurableWorkflowClientAuth = "Required"
	DurableWorkflowClientAuthNotRequired DurableWorkflowClientAuth = "NotRequired"
)

// DurableWorkflowArchival configures archival of closed workflow histories and visibility records. Each target is
// enabled by being set and renders the engine's static archival config (and the namespace/domain defaults), so a change
// rolls the pods. Namespaces that exist already keep their own archival state: enable it on them with the engine's CLI.
// Temporal builds without the "temporal_with_gstorage" tag refuse to start with a gcs target.
type DurableWorkflowArchival struct {
	// +optional
	History *DurableWorkflowArchivalTarget `json:"history,omitempty"`
	// +optional
	Visibility *DurableWorkflowArchivalTarget `json:"visibility,omitempty"`
}

const (
	DurableWorkflowArchivalProviderFilestore = "filestore"
	DurableWorkflowArchivalProviderS3        = "s3"
	DurableWorkflowArchivalProviderGCS       = "gcs"
)

// DurableWorkflowArchivalTarget is one archival target.
type DurableWorkflowArchivalTarget struct {
	// Provider is one of filestore, s3 and gcs.
	// +kubebuilder:validation:Enum=filestore;s3;gcs
	Provider string `json:"provider"`
	// URI of the archive. The scheme follows the provider: file:///absolute/path for filestore, s3://bucket/prefix for
	// s3 and gs://bucket/prefix for gcs.
	//
	// A filestore path is an emptyDir the operator mounts into every pod of the cluster. It is NOT durable and not
	// shared between pods: archives are lost with the pod and can't be read through another pod. Use it for tests only.
	URI string `json:"uri"`
	// ConfigSecret holds the credentials of the provider. Required for s3 and gcs, and not allowed for filestore.
	// A secret is read from the namespace of the DurableWorkflow.
	//
	// s3 keys: AWS_REGION (required), AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (both or none; both are passed to
	// the servers as env variables), AWS_ENDPOINT_URL (optional, an S3-compatible endpoint such as MinIO) and
	// AWS_S3_FORCE_PATH_STYLE (optional, "true" for path style addressing).
	//
	// gcs key: credentials.json, the service account key. It is mounted read-only and set as the credentialsPath.
	//
	// The history and visibility targets may share one s3 secret but can't use two different ones, as the
	// credentials are process wide env variables. A change of the secret data rolls the pods.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`
}

// ArchivalFilestoreDir returns the directory of a file:///path URI, or an error for any other form.
func ArchivalFilestoreDir(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if u.Scheme != "file" || u.Host != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("must look like file:///absolute/path")
	}
	p := path.Clean(u.Path)
	if !path.IsAbs(u.Path) || p == "/" {
		return "", fmt.Errorf("must be an absolute path below /")
	}
	return p, nil
}

// ArchivalURIScheme is the URI scheme a provider requires.
func ArchivalURIScheme(provider string) string {
	switch provider {
	case DurableWorkflowArchivalProviderFilestore:
		return "file"
	case DurableWorkflowArchivalProviderS3:
		return "s3"
	case DurableWorkflowArchivalProviderGCS:
		return "gs"
	}
	return ""
}

// DurableWorkflowAuthorization configures JWT authorization.
type DurableWorkflowAuthorization struct {
	// JWKSURL is the URL of the key set that verifies external tokens.
	// +optional
	JWKSURL string `json:"jwksURL,omitempty"`
	// PublicKeySecret holds a JWKS document under the key "jwks.json", which verifies external tokens. Temporal only:
	// Cadence uses its single static key for the tokens the cluster signs itself, so its external tokens come
	// through JWKSURL.
	// +optional
	PublicKeySecret *core.LocalObjectReference `json:"publicKeySecret,omitempty"`
	// MaxTokenTTL is the longest token lifetime accepted, in seconds. Cadence only; defaults to 3600.
	// +optional
	MaxTokenTTL *int64 `json:"maxTokenTTL,omitempty"`

	// PermissionsClaimName is the claim that lists the permissions of the caller. Temporal only; defaults to "permissions".
	// +optional
	PermissionsClaimName string `json:"permissionsClaimName,omitempty"`
	// PermissionsRegex parses a permission into "namespace" and "role" groups. Temporal only.
	// +optional
	PermissionsRegex string `json:"permissionsRegex,omitempty"`
	// Audience is the audience the tokens must carry. Temporal only.
	// +optional
	Audience string `json:"audience,omitempty"`

	// GroupsAttributePath is the claim that lists the groups of the caller. Cadence only; defaults to "groups".
	// +optional
	GroupsAttributePath string `json:"groupsAttributePath,omitempty"`
	// AdminAttributePath is the claim that marks an administrator. Cadence only; defaults to "admin".
	// +optional
	AdminAttributePath string `json:"adminAttributePath,omitempty"`
}

// DurableWorkflowUI configures the web UI.
type DurableWorkflowUI struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// Auth selects the UI authentication strategy. Cadence only.
	// +kubebuilder:default=Disabled
	// +optional
	Auth DurableWorkflowUIAuth `json:"auth,omitempty"`
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// +optional
	PodTemplate *ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// +kubebuilder:validation:Enum=Disabled;JWT
type DurableWorkflowUIAuth string

const (
	DurableWorkflowUIAuthDisabled DurableWorkflowUIAuth = "Disabled"
	DurableWorkflowUIAuthJWT      DurableWorkflowUIAuth = "JWT"
)

// DurableWorkflowAdminTools configures the admin tools pod.
type DurableWorkflowAdminTools struct {
	// +optional
	Enabled bool `json:"enabled,omitempty"`
	// +optional
	PodTemplate *ofstv2.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// DurableWorkflowStatus defines the observed state of DurableWorkflow.
type DurableWorkflowStatus struct {
	// Specifies the current phase of the cluster.
	// +optional
	Phase DatabasePhase `json:"phase,omitempty"`

	// observedGeneration is the most recent generation observed for this resource.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions applied to the cluster, such as Ready and HighlyAvailable.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`

	// Distribution is the resolved distribution of spec.version.
	// +optional
	Distribution string `json:"distribution,omitempty"`

	// Persistence reports the state of each persistence store.
	// +optional
	Persistence *DurableWorkflowPersistenceStatus `json:"persistence,omitempty"`
}

// DurableWorkflowPersistenceStatus reports the persistence stores.
type DurableWorkflowPersistenceStatus struct {
	// +optional
	Default DurableWorkflowStoreStatus `json:"default,omitempty"`
	// +optional
	Visibility DurableWorkflowStoreStatus `json:"visibility,omitempty"`
}

// DurableWorkflowStoreStatus reports one persistence store.
type DurableWorkflowStoreStatus struct {
	// Created is true once the database and user exist.
	// +optional
	Created bool `json:"created,omitempty"`
	// SchemaVersion is the schema version applied by the schema Job.
	// +optional
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// SchemaServerVersion is the server version whose schema was applied.
	// +optional
	SchemaServerVersion string `json:"schemaServerVersion,omitempty"`
	// Ref is the AppBinding of the server the store lives on, whether the user named it or the operator created it
	// (managed persistence). Backup and restore tools read it instead of resolving the store again.
	// +optional
	Ref *kmapi.ObjectReference `json:"ref,omitempty"`
	// Type is the engine of the store.
	// +optional
	Type DurableWorkflowStoreType `json:"type,omitempty"`
	// DatabaseName is the database, keyspace or index of the store on that server.
	// +optional
	DatabaseName string `json:"databaseName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DurableWorkflowList contains a list of DurableWorkflow.
type DurableWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DurableWorkflow `json:"items"`
}

var _ Accessor = &DurableWorkflow{}

func (m *DurableWorkflow) GetObjectMeta() metav1.ObjectMeta {
	return m.ObjectMeta
}

func (m *DurableWorkflow) GetConditions() []kmapi.Condition {
	return m.Status.Conditions
}

func (m *DurableWorkflow) SetCondition(cond kmapi.Condition) {
	m.Status.Conditions = setCondition(m.Status.Conditions, cond)
}

func (m *DurableWorkflow) RemoveCondition(typ string) {
	m.Status.Conditions = removeCondition(m.Status.Conditions, typ)
}

// MarshalJSON leaves out an empty default store. With managed persistence there is no reference, and the empty struct
// would be sent as {}, which the CRD refuses because name is required in a reference.
func (p DurableWorkflowPersistence) MarshalJSON() ([]byte, error) {
	type plain DurableWorkflowPersistence
	out := struct {
		plain
		Default *DurableWorkflowStore `json:"default,omitempty"`
	}{plain: plain(p)}
	if !reflect.DeepEqual(p.Default, DurableWorkflowStore{}) {
		out.Default = &p.Default
	}
	return json.Marshal(out)
}
