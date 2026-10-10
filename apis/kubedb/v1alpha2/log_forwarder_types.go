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
	core "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogForwarderSpec configures Sidecar collection or platform-managed NodeAgent collection.
// +kubebuilder:validation:XValidation:rule="!has(self.collectionMode) || self.collectionMode != 'NodeAgent' || (!has(self.exporter) && !has(self.processors) && !has(self.extensions) && !has(self.sourceStorage) && !has(self.securityContext) && !(has(self.resources) && (has(self.resources.requests) || has(self.resources.limits) || has(self.resources.claims))) && !(has(self.stateStorage) && (has(self.stateStorage.accessModes) || has(self.stateStorage.volumeMode) || has(self.stateStorage.selector) || has(self.stateStorage.storageClassName) || has(self.stateStorage.volumeName) || has(self.stateStorage.dataSource) || has(self.stateStorage.dataSourceRef) || has(self.stateStorage.volumeAttributesClassName) || (has(self.stateStorage.resources) && (has(self.stateStorage.resources.requests) || has(self.stateStorage.resources.limits))))))",message="NodeAgent Collector configuration is platform-owned; Sidecar fields are forbidden"
// +kubebuilder:validation:XValidation:rule="(has(self.enabled) && !self.enabled) || (has(self.collectionMode) && self.collectionMode == 'NodeAgent') || has(self.exporter)",message="enabled Sidecar requires exporter"
// +kubebuilder:validation:XValidation:rule="!has(self.collectionMode) || self.collectionMode != 'NodeAgent' || !has(self.sources) || self.sources.all(s, !has(s.fileLog))",message="NodeAgent cannot tune the shared file reader per database"
type LogForwarderSpec struct {
	// Enabled toggles KubeDB forwarding without deleting retained logs or state.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// CollectionMode selects the source topology; defaults to Sidecar.
	// +optional
	// +kubebuilder:default=Sidecar
	CollectionMode LogCollectionMode `json:"collectionMode,omitempty"`
	// Sources selects the log streams advertised by the database version.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	Sources []LogSourceSpec `json:"sources,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
	// Exporter selects the Sidecar destination and its transport settings.
	// +optional
	Exporter *LogExporterSpec `json:"exporter,omitempty"`
	// Processors configures Sidecar processing without exposing arbitrary pipelines.
	// +optional
	Processors *LogProcessors `json:"processors,omitempty"`
	// Extensions configures additional Sidecar services and selected credential keys.
	// +optional
	Extensions *LogExtensions `json:"extensions,omitempty"`
	// StateStorage is the per-pod Filesystem/RWO PVC for checkpoints and queued records.
	// +optional
	StateStorage core.PersistentVolumeClaimSpec `json:"stateStorage,omitempty"`
	// SourceStorage overrides per-pod source storage where the engine adapter supports it.
	// +optional
	SourceStorage *LogSourceStorage `json:"sourceStorage,omitempty"`
	// Resources sets the Sidecar compute budget, including its explicit memory limit.
	// +optional
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// SecurityContext customizes the Sidecar's operator-managed security defaults.
	// +optional
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`
	// RolloutPolicy requires an OpsRequest to activate pending configuration changes.
	// +optional
	// +kubebuilder:validation:Enum=Manual
	// +kubebuilder:default=Manual
	RolloutPolicy LogForwarderRolloutPolicy `json:"rolloutPolicy,omitempty"`
}

// LogForwarderRolloutPolicy selects controlled OpsRequest activation.
// +kubebuilder:validation:Enum=Manual
type LogForwarderRolloutPolicy string

const LogForwarderRolloutManual LogForwarderRolloutPolicy = "Manual"

// LogForwarderTLSMode selects verified TLS or explicitly requested plaintext.
// +kubebuilder:validation:Enum=Disabled;Verify
type LogForwarderTLSMode string

const (
	LogForwarderTLSModeDisabled LogForwarderTLSMode = "Disabled"
	LogForwarderTLSModeVerify   LogForwarderTLSMode = "Verify"
)

// LogForwarderTLS configures trusted transport and optional client authentication.
// +kubebuilder:validation:XValidation:rule="has(self.clientCertSecretRef) == has(self.clientKeySecretRef)",message="clientCertSecretRef and clientKeySecretRef must be supplied together"
type LogForwarderTLS struct {
	// Mode defaults to Verify; Disabled permits plaintext, not unverified HTTPS.
	// +optional
	// +kubebuilder:validation:Enum=Disabled;Verify
	// +kubebuilder:default=Verify
	Mode LogForwarderTLSMode `json:"mode,omitempty"`
	// CASecretRef selects a private server-trust CA bundle when system trust is insufficient.
	// +optional
	CASecretRef *core.SecretKeySelector `json:"caSecretRef,omitempty"`
	// ClientCertSecretRef selects a client certificate and pairs with ClientKeySecretRef.
	// +optional
	ClientCertSecretRef *core.SecretKeySelector `json:"clientCertSecretRef,omitempty"`
	// ClientKeySecretRef selects a client private key and pairs with ClientCertSecretRef.
	// +optional
	ClientKeySecretRef *core.SecretKeySelector `json:"clientKeySecretRef,omitempty"`
}

// LogSourceSpec selects a source supported by the engine/version adapter.
type LogSourceSpec struct {
	// Name identifies the selected database log stream.
	Name string `json:"name"`
	// FileLog tunes a Sidecar reader; saved checkpoints take precedence over startAt.
	// +optional
	FileLog *LogFileLogReceiver `json:"fileLog,omitempty"`
}

// LogCollectionMode selects the source collection topology.
// +kubebuilder:validation:Enum=Sidecar;NodeAgent
type LogCollectionMode string

const (
	LogCollectionModeSidecar   LogCollectionMode = "Sidecar"
	LogCollectionModeNodeAgent LogCollectionMode = "NodeAgent"
)

// LogExporterSpec selects exactly one compiled Collector exporter.
// Name is part of persistent delivery identity, not a license to rename queues.
// +kubebuilder:validation:XValidation:rule="(self.type == 'otlphttp') == has(self.otlpHttp)",message="otlphttp requires only the otlpHttp block"
// +kubebuilder:validation:XValidation:rule="(self.type == 'splunk_hec') == has(self.splunkHec)",message="splunk_hec requires only the splunkHec block"
// +kubebuilder:validation:XValidation:rule="(self.type == 'elasticsearch') == has(self.elasticsearch)",message="elasticsearch requires only the elasticsearch block"
// +kubebuilder:validation:XValidation:rule="(self.type == 'datadog') == has(self.datadog)",message="datadog requires only the datadog block"
// +kubebuilder:validation:XValidation:rule="(self.type == 'syslog') == has(self.syslog)",message="syslog requires only the syslog block"
type LogExporterSpec struct {
	// Name identifies the stable Collector component or database source.
	// +optional
	// +kubebuilder:default=primary
	Name string `json:"name,omitempty"`
	// Type selects the native Collector component or authentication type.
	// +kubebuilder:validation:Enum=otlphttp;splunk_hec;elasticsearch;datadog;syslog
	Type string `json:"type"`
	// OTLPHTTP configures the OTLP/HTTP exporter.
	// +optional
	OTLPHTTP *LogOTLPHTTPExporter `json:"otlpHttp,omitempty"`
	// SplunkHEC configures the Splunk HEC exporter.
	// +optional
	SplunkHEC *LogSplunkHECExporter `json:"splunkHec,omitempty"`
	// Elasticsearch configures direct Elasticsearch export.
	// +optional
	Elasticsearch *LogElasticsearchExporter `json:"elasticsearch,omitempty"`
	// Datadog configures the Datadog exporter.
	// +optional
	Datadog *LogDatadogExporter `json:"datadog,omitempty"`
	// Syslog configures syslog export.
	// +optional
	Syslog *LogSyslogExporter `json:"syslog,omitempty"`
}

// LogExporterOptions contains shared queue and transport intent.
// The runtime must validate support against the pinned exporter build.
type LogExporterOptions struct {
	// TLS configures verified transport and optional client authentication.
	// +optional
	TLS *LogForwarderTLS `json:"tls,omitempty"`
	// SendingQueue configures bounded persistent delivery buffering.
	// +optional
	SendingQueue *LogSendingQueue `json:"sendingQueue,omitempty"`
}

// LogOTLPHTTPExporter configures native OTLP/HTTP export.
// +kubebuilder:validation:XValidation:rule="has(self.endpoint) != has(self.logsEndpoint)",message="exactly one of endpoint or logsEndpoint is required"
type LogOTLPHTTPExporter struct {
	// LogExporterOptions embeds shared transport and queue options.
	LogExporterOptions `json:",inline"`
	// Endpoint is a base URL; LogsEndpoint is a complete logs URL.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint,omitempty"`
	// LogsEndpoint sets the complete OTLP logs URL without automatic path suffixes.
	// +optional
	// +kubebuilder:validation:MinLength=1
	LogsEndpoint string `json:"logsEndpoint,omitempty"`
	// Auth selects destination authentication.
	// +optional
	Auth *LogExporterAuth `json:"auth,omitempty"`
	// RetryOnFailure configures retryable export failures.
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogSplunkHECExporter configures the HEC endpoint and selected token key.
type LogSplunkHECExporter struct {
	// LogExporterOptions embeds shared transport and queue options.
	LogExporterOptions `json:",inline"`
	// Endpoint sets the destination address.
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`
	// TokenSecretRef selects the token from a same-namespace Secret.
	TokenSecretRef core.SecretKeySelector `json:"tokenSecretRef"`
	// Index sets the Splunk destination index.
	// +optional
	Index string `json:"index,omitempty"`
	// Source sets the Splunk event source.
	// +optional
	Source string `json:"source,omitempty"`
	// Sourcetype sets the Splunk event sourcetype.
	// +optional
	Sourcetype string `json:"sourcetype,omitempty"`
	// RetryOnFailure configures retryable export failures.
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogElasticsearchExporter preserves Elasticsearch's distinct retry schema.
type LogElasticsearchExporter struct {
	// LogExporterOptions embeds shared transport and queue options.
	LogExporterOptions `json:",inline"`
	// Endpoints lists Elasticsearch destination URLs.
	// +kubebuilder:validation:MinItems=1
	Endpoints []string `json:"endpoints"`
	// APIKeySecretRef selects an Elasticsearch API key.
	// +optional
	APIKeySecretRef *core.SecretKeySelector `json:"apiKeySecretRef,omitempty"`
	// LogsIndex sets the destination logs index.
	// +optional
	LogsIndex string `json:"logsIndex,omitempty"`
	// Mapping selects native Elasticsearch document mapping.
	// +optional
	Mapping *LogElasticsearchMapping `json:"mapping,omitempty"`
	// Retry configures Elasticsearch-specific retry behavior.
	// +optional
	Retry *LogElasticsearchRetry `json:"retry,omitempty"`
}

type LogElasticsearchMapping struct {
	// Mode must be supported by the pinned exporter and destination version.
	// +optional
	Mode string `json:"mode,omitempty"`
}

type LogElasticsearchRetry struct {
	// Enabled toggles the native feature.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// InitialInterval sets the initial retry delay.
	// +optional
	InitialInterval *metav1.Duration `json:"initialInterval,omitempty"`
	// MaxInterval caps the retry delay.
	// +optional
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
	// MaxRetries caps the retry count.
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxRetries *int32 `json:"maxRetries,omitempty"`
}

type LogDatadogExporter struct {
	// Datadog exporter support for queue and TLS controls is build-dependent.
	LogExporterOptions `json:",inline"`
	// API configures the Datadog intake credentials.
	API LogDatadogAPI `json:"api"`
	// RetryOnFailure configures retryable export failures.
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

type LogDatadogAPI struct {
	// Site sets the Datadog destination site.
	// +kubebuilder:validation:MinLength=1
	Site string `json:"site"`
	// KeySecretRef selects the Datadog API key.
	KeySecretRef core.SecretKeySelector `json:"keySecretRef"`
}

type LogSyslogExporter struct {
	// LogExporterOptions embeds shared transport and queue options.
	LogExporterOptions `json:",inline"`
	// Endpoint sets the destination address.
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`
	// Port sets the syslog destination port.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
	// Network selects the supported TCP transport.
	// +kubebuilder:validation:Enum=tcp
	Network string `json:"network"`
	// Protocol selects the syslog message format.
	// +kubebuilder:validation:Enum=rfc5424;rfc3164
	Protocol string `json:"protocol"`
	// EnableOctetCounting enables TCP octet-counted framing.
	// +optional
	EnableOctetCounting *bool `json:"enableOctetCounting,omitempty"`
	// Facility sets the syslog facility number.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	Facility *int32 `json:"facility,omitempty"`
	// RetryOnFailure configures retryable export failures.
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogSendingQueue bounds export requests, not bytes. KubeDB wires persistence.
// Blocking overflow does not guarantee backend delivery or block DB writes.
type LogSendingQueue struct {
	// Enabled toggles the native feature.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// QueueSize bounds queued requests, not bytes.
	// +optional
	// +kubebuilder:validation:Minimum=1
	QueueSize *int32 `json:"queueSize,omitempty"`
	// NumConsumers sets concurrent queue consumers.
	// +optional
	// +kubebuilder:validation:Minimum=1
	NumConsumers *int32 `json:"numConsumers,omitempty"`
	// BlockOnOverflow blocks the collector producer when the queue fills.
	// +optional
	BlockOnOverflow *bool `json:"blockOnOverflow,omitempty"`
}

type LogRetryOnFailure struct {
	// Enabled toggles the native feature.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// InitialInterval sets the initial retry delay.
	// +optional
	InitialInterval *metav1.Duration `json:"initialInterval,omitempty"`
	// MaxInterval caps the retry delay.
	// +optional
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
	// MaxElapsedTime of zero means unlimited retry of retryable failures.
	// +optional
	MaxElapsedTime *metav1.Duration `json:"maxElapsedTime,omitempty"`
}

// LogFileLogReceiver tunes a managed reader. Paths and parsers are operator-owned.
type LogFileLogReceiver struct {
	// StartAt applies only to files without saved checkpoints.
	// +optional
	// +kubebuilder:validation:Enum=end;beginning
	StartAt string `json:"startAt,omitempty"`
	// PollInterval sets the file discovery interval.
	// +optional
	PollInterval *metav1.Duration `json:"pollInterval,omitempty"`
	// MaxLogSize uses Collector sizes, for example 1MiB.
	// +optional
	MaxLogSize string `json:"maxLogSize,omitempty"`
}

// LogProcessors configures managed Sidecar processing stages.
// +kubebuilder:validation:XValidation:rule="!has(self.extraConfig) || size(self.extraConfig) == 0 || (has(self.order) && size(self.order) > 0)",message="extraConfig requires explicit processor order"
type LogProcessors struct {
	// MemoryLimiter tunes the always-enabled first processor and its paired memory thresholds.
	// +optional
	MemoryLimiter *LogMemoryLimiterProcessor `json:"memoryLimiter,omitempty"`
	// Batch tunes managed batching, which always runs last.
	// +optional
	Batch *LogBatchProcessor `json:"batch,omitempty"`
	// Attributes applies ordered actions to log-record attributes.
	// +optional
	Attributes *LogAttributesProcessor `json:"attributes,omitempty"`
	// Filter drops log records matching log-only OTTL conditions.
	// +optional
	Filter *LogFilterProcessor `json:"filter,omitempty"`
	// Transform applies log-only OTTL normalization and redaction statements.
	// +optional
	Transform *LogTransformProcessor `json:"transform,omitempty"`
	// ExtraConfig defines native processors without replacing managed safety components.
	// +optional
	ExtraConfig string `json:"extraConfig,omitempty"`
	// Order lists each user stage exactly once; excludes memoryLimiter and batch.
	// +optional
	// +listType=atomic
	Order []string `json:"order,omitempty"`
}

// LogMemoryLimiterProcessor selects a complete percentage or absolute-MiB pair.
// +kubebuilder:validation:XValidation:rule="has(self.limitPercentage) == has(self.spikeLimitPercentage)",message="limitPercentage and spikeLimitPercentage are a required pair"
// +kubebuilder:validation:XValidation:rule="has(self.limitMiB) == has(self.spikeLimitMiB)",message="limitMiB and spikeLimitMiB are a required pair"
// +kubebuilder:validation:XValidation:rule="!has(self.limitPercentage) || !has(self.limitMiB)",message="percentage and MiB pairs cannot be mixed"
// +kubebuilder:validation:XValidation:rule="!has(self.limitPercentage) || !has(self.spikeLimitPercentage) || self.spikeLimitPercentage < self.limitPercentage",message="spikeLimitPercentage must be below limitPercentage"
// +kubebuilder:validation:XValidation:rule="!has(self.limitMiB) || !has(self.spikeLimitMiB) || self.spikeLimitMiB < self.limitMiB",message="spikeLimitMiB must be below limitMiB"
type LogMemoryLimiterProcessor struct {
	// CheckInterval sets the memory-check interval; defaults to 1s.
	// +optional
	CheckInterval *metav1.Duration `json:"checkInterval,omitempty"`
	// LimitPercentage is the hard percentage threshold; pairs with SpikeLimitPercentage.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=99
	LimitPercentage *int32 `json:"limitPercentage,omitempty"`
	// SpikeLimitPercentage is the percentage spike allowance; pairs with LimitPercentage.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=98
	SpikeLimitPercentage *int32 `json:"spikeLimitPercentage,omitempty"`
	// LimitMiB is the absolute hard threshold; pairs with SpikeLimitMiB instead of percentages.
	// +optional
	// +kubebuilder:validation:Minimum=1
	LimitMiB *int32 `json:"limitMiB,omitempty"`
	// SpikeLimitMiB is the absolute spike allowance; pairs with LimitMiB instead of percentages.
	// +optional
	// +kubebuilder:validation:Minimum=1
	SpikeLimitMiB *int32 `json:"spikeLimitMiB,omitempty"`
}

// LogAttributesProcessor configures native attribute actions.
type LogAttributesProcessor struct {
	// Actions lists the attribute modifications in execution order.
	// +kubebuilder:validation:MinItems=1
	// +listType=atomic
	Actions []LogAttributeAction `json:"actions"`
}

// LogAttributeAction modifies one attribute using native Collector semantics.
type LogAttributeAction struct {
	// Key identifies the attribute to modify.
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
	// Action selects the supported native operation.
	// +kubebuilder:validation:Enum=insert;update;upsert;delete;hash;extract;convert
	Action string `json:"action"`
	// Value supplies a scalar value for insert/update/upsert instead of a source reference.
	// +optional
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	Value *apiextensionsv1.JSON `json:"value,omitempty"`
	// FromAttribute copies an existing attribute instead of using Value or FromContext.
	// +optional
	FromAttribute string `json:"fromAttribute,omitempty"`
	// FromContext copies a native context value instead of Value or FromAttribute.
	// +optional
	FromContext string `json:"fromContext,omitempty"`
	// Pattern is the extraction expression used by the extract action.
	// +optional
	Pattern string `json:"pattern,omitempty"`
	// ConvertedType selects the resulting scalar type for the convert action.
	// +optional
	// +kubebuilder:validation:Enum=int;double;string
	ConvertedType string `json:"convertedType,omitempty"`
}

// LogProcessorErrorMode selects explicit OTTL error handling.
// +kubebuilder:validation:Enum=propagate;ignore
type LogProcessorErrorMode string

// LogFilterProcessor configures native OTTL log-record filtering.
type LogFilterProcessor struct {
	// ErrorMode defaults to propagate so failed filtering is not silently bypassed.
	// +optional
	ErrorMode LogProcessorErrorMode `json:"errorMode,omitempty"`
	// Logs selects conditions evaluated against incoming log records.
	Logs LogFilterLogs `json:"logs"`
}

// LogFilterLogs contains log-only filter conditions.
type LogFilterLogs struct {
	// LogRecord drops a record when any condition evaluates to true.
	// +kubebuilder:validation:MinItems=1
	// +listType=atomic
	LogRecord []string `json:"logRecord"`
}

// LogTransformProcessor configures native OTTL transformations of logs.
type LogTransformProcessor struct {
	// ErrorMode defaults to propagate so failed redaction is not silently bypassed.
	// +optional
	ErrorMode LogProcessorErrorMode `json:"errorMode,omitempty"`
	// LogStatements groups ordered statements by their OTTL context.
	// +kubebuilder:validation:MinItems=1
	// +listType=atomic
	LogStatements []LogTransformStatements `json:"logStatements"`
}

// LogTransformStatements contains one native OTTL statement group.
type LogTransformStatements struct {
	// Context selects the native log, resource, or scope evaluation context.
	// +kubebuilder:validation:Enum=log;resource;scope
	Context string `json:"context"`
	// Statements lists transformations in execution order.
	// +kubebuilder:validation:MinItems=1
	// +listType=atomic
	Statements []string `json:"statements"`
	// Conditions optionally restrict when this group executes.
	// +optional
	// +listType=atomic
	Conditions []string `json:"conditions,omitempty"`
}

// LogExtensions configures additional services without exposing pipeline wiring.
type LogExtensions struct {
	// ExtraConfig defines native extensions activated through service.extensions.
	// +optional
	ExtraConfig string `json:"extraConfig,omitempty"`
	// SecretEnv binds extension environment names to selected same-namespace Secret keys.
	// +optional
	SecretEnv map[string]core.SecretKeySelector `json:"secretEnv,omitempty"`
}

// LogBatchProcessor configures only the managed batch processor.
type LogBatchProcessor struct {
	// Timeout sets the batch flush interval.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// SendBatchSize sets the batching trigger in records.
	// +optional
	// +kubebuilder:validation:Minimum=1
	SendBatchSize *int32 `json:"sendBatchSize,omitempty"`
	// Zero means no maximum, as in the native Collector.
	// +optional
	// +kubebuilder:validation:Minimum=0
	SendBatchMaxSize *int32 `json:"sendBatchMaxSize,omitempty"`
}

// LogSourceStorage is a per-emitting-pod log volume override.
// +kubebuilder:validation:XValidation:rule="has(self.emptyDir) != has(self.volumeClaimTemplate)",message="exactly one source storage type is required"
type LogSourceStorage struct {
	// EmptyDir selects bounded ephemeral source storage.
	// +optional
	EmptyDir *core.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// VolumeClaimTemplate selects durable per-pod source storage.
	// +optional
	VolumeClaimTemplate *core.PersistentVolumeClaimSpec `json:"volumeClaimTemplate,omitempty"`
}

// LogExporterAuth selects non-optional keys from same-namespace Secrets.
// +kubebuilder:validation:XValidation:rule="self.type in ['Token', 'Bearer'] ? has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) && !has(self.valueSecretRef) && !has(self.headerName) : self.type == 'Basic' ? has(self.usernameSecretRef) && has(self.passwordSecretRef) && !has(self.tokenSecretRef) && !has(self.valueSecretRef) && !has(self.headerName) : self.type == 'Header' ? has(self.headerName) && has(self.valueSecretRef) && !has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) : !has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) && !has(self.valueSecretRef) && !has(self.headerName)",message="auth fields must match the selected type"
// +kubebuilder:validation:XValidation:rule="(self.type == 'Extension') == has(self.extensionRef)",message="Extension authentication requires only extensionRef"
type LogExporterAuth struct {
	// Type selects the native Collector component or authentication type.
	// +kubebuilder:validation:Enum=None;Token;Bearer;Basic;Header;Extension
	Type string `json:"type"`
	// ExtensionRef names an authentication extension defined in extensions.extraConfig.
	// +optional
	ExtensionRef string `json:"extensionRef,omitempty"`
	// TokenSecretRef selects the token from a same-namespace Secret.
	// +optional
	TokenSecretRef *core.SecretKeySelector `json:"tokenSecretRef,omitempty"`
	// UsernameSecretRef selects the authentication username.
	// +optional
	UsernameSecretRef *core.SecretKeySelector `json:"usernameSecretRef,omitempty"`
	// PasswordSecretRef selects the authentication password.
	// +optional
	PasswordSecretRef *core.SecretKeySelector `json:"passwordSecretRef,omitempty"`
	// HeaderName sets the authentication HTTP header name.
	// +optional
	// +kubebuilder:validation:MinLength=1
	HeaderName string `json:"headerName,omitempty"`
	// ValueSecretRef selects the authentication header value.
	// +optional
	ValueSecretRef *core.SecretKeySelector `json:"valueSecretRef,omitempty"`
}
