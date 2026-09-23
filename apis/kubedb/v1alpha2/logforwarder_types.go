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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogForwarderSpec configures an operator-managed OpenTelemetry Collector sidecar that tails the
// database's log files and ships them to an observability backend.
//
// The operator owns the vendor-neutral half of the collector pipeline (file receiver, log parser,
// batching, and a persistent sending queue backed by a per-pod state PVC). The vendor-specific half
// — the exporter, any exporter-specific processors/extensions, and its credentials — is supplied as
// data: either a KubeDB-shipped Profile or a raw ExporterConfig. Because every exporter already
// ships in the collector image, adding a new backend never requires an operator code change or an
// image rebuild.
//
// This type is engine-neutral and intended to be embedded by any KubeDB database spec.
type LogForwarderSpec struct {
	// Enabled toggles log forwarding while retaining the generated configuration and the per-pod
	// state PVCs. It defaults to true when logForwarder is present.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Destination selects where logs are shipped and how the sidecar authenticates to the backend.
	Destination LogDestinationSpec `json:"destination"`

	// Sources selects which database log streams to forward. Each entry must name a log source
	// advertised by the database version's log capabilities.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=name
	Sources []LogSourceSpec `json:"sources,omitempty" patchStrategy:"merge" patchMergeKey:"name"`

	// StateStorage is the per-pod PVC template that persists receiver offsets and the exporter
	// sending queue. It must be a filesystem, ReadWriteOnce volume; never shared across pods.
	StateStorage core.PersistentVolumeClaimSpec `json:"stateStorage"`

	// Delivery tunes batching, the persistent sending queue, and retry/backpressure behavior.
	// +optional
	Delivery *LogDeliverySpec `json:"delivery,omitempty"`

	// Resources are the compute resources for the log-forwarder sidecar container.
	// +optional
	Resources core.ResourceRequirements `json:"resources,omitempty"`

	// SecurityContext is applied to the log-forwarder sidecar container on top of the operator's
	// safe defaults (non-root, dropped capabilities, read-only root filesystem).
	// +optional
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`

	// RolloutPolicy controls how pending log-forwarder changes are activated. Manual (default)
	// requires the normal restart OpsRequest; Automatic performs a serialized, health-aware restart.
	// +kubebuilder:validation:Enum=Manual;Automatic
	// +kubebuilder:default=Manual
	// +optional
	RolloutPolicy LogForwarderRolloutPolicy `json:"rolloutPolicy,omitempty"`
}

// LogDestinationSpec describes a single observability backend and how to reach it.
// Exactly one of Profile or ExporterConfig must be set.
// +kubebuilder:validation:XValidation:rule="(has(self.profile) && size(self.profile) > 0) != (has(self.exporterConfig) && size(self.exporterConfig) > 0)",message="exactly one of destination.profile or destination.exporterConfig must be set"
type LogDestinationSpec struct {
	// Name is a stable identifier for this destination. It names the exporter component and its
	// persistent queue, so it must not change while delivery state is pending.
	// +kubebuilder:default=primary
	// +optional
	Name string `json:"name,omitempty"`

	// Profile selects a KubeDB-shipped exporter profile (for example "splunk", "elastic",
	// "datadog", "loki", "otlphttp", or "syslog"). Mutually exclusive with ExporterConfig.
	// +optional
	Profile string `json:"profile,omitempty"`

	// ExporterConfig is a raw OpenTelemetry Collector exporter configuration block for backends
	// without a shipped profile. It is spliced into the generated pipeline verbatim. Reference
	// credentials with ${env:...} placeholders resolved from SecretRef; never inline secret values.
	// Mutually exclusive with Profile.
	// +optional
	ExporterConfig string `json:"exporterConfig,omitempty"`

	// ExtraProcessors is an optional raw collector "processors" fragment appended to the logs
	// pipeline (for example a transform processor required by the syslog exporter).
	// +optional
	ExtraProcessors string `json:"extraProcessors,omitempty"`

	// ExtraExtensions is an optional raw collector "extensions" fragment (for example an
	// oauth2client extension required by an OAuth2 backend).
	// +optional
	ExtraExtensions string `json:"extraExtensions,omitempty"`

	// Endpoint is the backend URL or host the exporter connects to. Its interpretation depends on
	// the profile/exporter (an OTLP base URL, a Splunk HEC URL, a syslog host, and so on).
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// TLS configures transport security for the connection to the backend.
	// +optional
	TLS *LogForwarderTLS `json:"tls,omitempty"`

	// SecretRef references a same-namespace Secret whose keys are projected as environment
	// variables into the sidecar (envFrom), resolving ${env:...} placeholders in the exporter
	// configuration. The operator consumes but never owns or deletes this Secret.
	// +optional
	SecretRef *core.LocalObjectReference `json:"secretRef,omitempty"`
}

// LogForwarderTLS configures transport security to the backend.
type LogForwarderTLS struct {
	// Mode selects TLS behavior: Disabled for plain internal HTTP, or Verify to validate the
	// backend's server certificate.
	// +kubebuilder:validation:Enum=Disabled;Verify
	// +kubebuilder:default=Verify
	// +optional
	Mode LogForwarderTLSMode `json:"mode,omitempty"`

	// CASecretRef optionally selects a Secret key holding a CA bundle used to verify the backend
	// certificate when a private CA is used. System trust is used when omitted.
	// +optional
	CASecretRef *core.SecretKeySelector `json:"caSecretRef,omitempty"`

	// ClientCertSecretRef and ClientKeySecretRef form an optional mTLS client-certificate pair.
	// Both must be set together.
	// +optional
	ClientCertSecretRef *core.SecretKeySelector `json:"clientCertSecretRef,omitempty"`
	// +optional
	ClientKeySecretRef *core.SecretKeySelector `json:"clientKeySecretRef,omitempty"`
}

// LogSourceSpec selects one database log stream to forward.
type LogSourceSpec struct {
	// Name identifies the log stream to forward (for example "query" or "security"). It must match
	// a source advertised by the database version's log capabilities.
	Name string `json:"name"`

	// InitialPosition sets where the receiver starts reading a file that has no saved checkpoint.
	// End (default) avoids replaying history; Beginning is an explicit catch-up. It is ignored once
	// a checkpoint exists for the file.
	// +kubebuilder:validation:Enum=End;Beginning
	// +kubebuilder:default=End
	// +optional
	InitialPosition LogInitialPosition `json:"initialPosition,omitempty"`
}

// LogDeliverySpec tunes batching, the persistent sending queue, and retry behavior.
type LogDeliverySpec struct {
	// QueueCapacityRequests is the maximum number of export requests buffered in the persistent
	// sending queue. This bounds requests, not bytes.
	// +kubebuilder:validation:Minimum=1
	// +optional
	QueueCapacityRequests *int32 `json:"queueCapacityRequests,omitempty"`

	// Workers is the number of concurrent senders draining the queue.
	// +kubebuilder:validation:Minimum=1
	// +optional
	Workers *int32 `json:"workers,omitempty"`

	// RetryInitialInterval is the delay before the first retry of a retryable failure.
	// +optional
	RetryInitialInterval *metav1.Duration `json:"retryInitialInterval,omitempty"`

	// RetryMaxInterval caps the exponential backoff between retries.
	// +optional
	RetryMaxInterval *metav1.Duration `json:"retryMaxInterval,omitempty"`

	// RetryMaxElapsedTime bounds total retry time; zero means retry retryable errors forever.
	// +optional
	RetryMaxElapsedTime *metav1.Duration `json:"retryMaxElapsedTime,omitempty"`

	// OnQueueFull selects behavior when the sending queue is saturated. Backpressure pauses log
	// tailing (never the database); Drop discards new records.
	// +kubebuilder:validation:Enum=Backpressure;Drop
	// +kubebuilder:default=Backpressure
	// +optional
	OnQueueFull LogQueueFullPolicy `json:"onQueueFull,omitempty"`
}

// +kubebuilder:validation:Enum=Manual;Automatic
type LogForwarderRolloutPolicy string

const (
	LogForwarderRolloutManual    LogForwarderRolloutPolicy = "Manual"
	LogForwarderRolloutAutomatic LogForwarderRolloutPolicy = "Automatic"
)

// +kubebuilder:validation:Enum=Disabled;Verify
type LogForwarderTLSMode string

const (
	LogForwarderTLSModeDisabled LogForwarderTLSMode = "Disabled"
	LogForwarderTLSModeVerify   LogForwarderTLSMode = "Verify"
)

// +kubebuilder:validation:Enum=End;Beginning
type LogInitialPosition string

const (
	LogInitialPositionEnd       LogInitialPosition = "End"
	LogInitialPositionBeginning LogInitialPosition = "Beginning"
)

// +kubebuilder:validation:Enum=Backpressure;Drop
type LogQueueFullPolicy string

const (
	LogQueueFullBackpressure LogQueueFullPolicy = "Backpressure"
	LogQueueFullDrop         LogQueueFullPolicy = "Drop"
)
