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

package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	// +optional
	// +kubebuilder:default=primary
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Enum=otlphttp;splunk_hec;elasticsearch;datadog;syslog
	Type string `json:"type"`
	// +optional
	OTLPHTTP *LogOTLPHTTPExporter `json:"otlpHttp,omitempty"`
	// +optional
	SplunkHEC *LogSplunkHECExporter `json:"splunkHec,omitempty"`
	// +optional
	Elasticsearch *LogElasticsearchExporter `json:"elasticsearch,omitempty"`
	// +optional
	Datadog *LogDatadogExporter `json:"datadog,omitempty"`
	// +optional
	Syslog *LogSyslogExporter `json:"syslog,omitempty"`
}

// LogExporterOptions contains shared queue and transport intent.
// The runtime must validate support against the pinned exporter build.
type LogExporterOptions struct {
	// +optional
	TLS *LogForwarderTLS `json:"tls,omitempty"`
	// +optional
	SendingQueue *LogSendingQueue `json:"sendingQueue,omitempty"`
}

// LogOTLPHTTPExporter configures native OTLP/HTTP export.
// +kubebuilder:validation:XValidation:rule="has(self.endpoint) != has(self.logsEndpoint)",message="exactly one of endpoint or logsEndpoint is required"
type LogOTLPHTTPExporter struct {
	LogExporterOptions `json:",inline"`
	// Endpoint is a base URL; LogsEndpoint is a complete logs URL.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint,omitempty"`
	// +optional
	// +kubebuilder:validation:MinLength=1
	LogsEndpoint string `json:"logsEndpoint,omitempty"`
	// +optional
	Auth *LogExporterAuth `json:"auth,omitempty"`
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogSplunkHECExporter configures the HEC endpoint and selected token key.
type LogSplunkHECExporter struct {
	LogExporterOptions `json:",inline"`
	// +kubebuilder:validation:MinLength=1
	Endpoint       string                 `json:"endpoint"`
	TokenSecretRef core.SecretKeySelector `json:"tokenSecretRef"`
	// +optional
	Index string `json:"index,omitempty"`
	// +optional
	Source string `json:"source,omitempty"`
	// +optional
	Sourcetype string `json:"sourcetype,omitempty"`
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogElasticsearchExporter preserves Elasticsearch's distinct retry schema.
type LogElasticsearchExporter struct {
	LogExporterOptions `json:",inline"`
	// +kubebuilder:validation:MinItems=1
	Endpoints []string `json:"endpoints"`
	// +optional
	APIKeySecretRef *core.SecretKeySelector `json:"apiKeySecretRef,omitempty"`
	// +optional
	LogsIndex string `json:"logsIndex,omitempty"`
	// +optional
	Mapping *LogElasticsearchMapping `json:"mapping,omitempty"`
	// +optional
	Retry *LogElasticsearchRetry `json:"retry,omitempty"`
}

type LogElasticsearchMapping struct {
	// Mode must be supported by the pinned exporter and destination version.
	// +optional
	Mode string `json:"mode,omitempty"`
}

type LogElasticsearchRetry struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	InitialInterval *metav1.Duration `json:"initialInterval,omitempty"`
	// +optional
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
	// +optional
	// +kubebuilder:validation:Minimum=0
	MaxRetries *int32 `json:"maxRetries,omitempty"`
}

type LogDatadogExporter struct {
	// Datadog exporter support for queue and TLS controls is build-dependent.
	LogExporterOptions `json:",inline"`
	API                LogDatadogAPI `json:"api"`
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

type LogDatadogAPI struct {
	// +kubebuilder:validation:MinLength=1
	Site         string                 `json:"site"`
	KeySecretRef core.SecretKeySelector `json:"keySecretRef"`
}

type LogSyslogExporter struct {
	LogExporterOptions `json:",inline"`
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
	// +kubebuilder:validation:Enum=tcp
	Network string `json:"network"`
	// +kubebuilder:validation:Enum=rfc5424;rfc3164
	Protocol string `json:"protocol"`
	// +optional
	EnableOctetCounting *bool `json:"enableOctetCounting,omitempty"`
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=23
	Facility *int32 `json:"facility,omitempty"`
	// +optional
	RetryOnFailure *LogRetryOnFailure `json:"retryOnFailure,omitempty"`
}

// LogSendingQueue bounds export requests, not bytes. KubeDB wires persistence.
// Blocking overflow does not guarantee backend delivery or block DB writes.
type LogSendingQueue struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	// +kubebuilder:validation:Minimum=1
	QueueSize *int32 `json:"queueSize,omitempty"`
	// +optional
	// +kubebuilder:validation:Minimum=1
	NumConsumers *int32 `json:"numConsumers,omitempty"`
	// +optional
	BlockOnOverflow *bool `json:"blockOnOverflow,omitempty"`
}

type LogRetryOnFailure struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	InitialInterval *metav1.Duration `json:"initialInterval,omitempty"`
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
	// +optional
	PollInterval *metav1.Duration `json:"pollInterval,omitempty"`
	// MaxLogSize uses Collector sizes, for example 1MiB.
	// +optional
	MaxLogSize string `json:"maxLogSize,omitempty"`
}

type LogProcessors struct {
	// +optional
	Batch *LogBatchProcessor `json:"batch,omitempty"`
}

// LogBatchProcessor configures only the managed batch processor.
type LogBatchProcessor struct {
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
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
	// +optional
	EmptyDir *core.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// +optional
	VolumeClaimTemplate *core.PersistentVolumeClaimSpec `json:"volumeClaimTemplate,omitempty"`
}

// LogExporterAuth selects non-optional keys from same-namespace Secrets.
// +kubebuilder:validation:XValidation:rule="self.type in ['Token', 'Bearer'] ? has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) && !has(self.valueSecretRef) && !has(self.headerName) : self.type == 'Basic' ? has(self.usernameSecretRef) && has(self.passwordSecretRef) && !has(self.tokenSecretRef) && !has(self.valueSecretRef) && !has(self.headerName) : self.type == 'Header' ? has(self.headerName) && has(self.valueSecretRef) && !has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) : !has(self.tokenSecretRef) && !has(self.usernameSecretRef) && !has(self.passwordSecretRef) && !has(self.valueSecretRef) && !has(self.headerName)",message="auth fields must match the selected type"
type LogExporterAuth struct {
	// +kubebuilder:validation:Enum=None;Token;Bearer;Basic;Header
	Type string `json:"type"`
	// +optional
	TokenSecretRef *core.SecretKeySelector `json:"tokenSecretRef,omitempty"`
	// +optional
	UsernameSecretRef *core.SecretKeySelector `json:"usernameSecretRef,omitempty"`
	// +optional
	PasswordSecretRef *core.SecretKeySelector `json:"passwordSecretRef,omitempty"`
	// +optional
	// +kubebuilder:validation:MinLength=1
	HeaderName string `json:"headerName,omitempty"`
	// +optional
	ValueSecretRef *core.SecretKeySelector `json:"valueSecretRef,omitempty"`
}
