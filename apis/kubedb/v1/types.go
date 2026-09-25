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
	kmapi "kmodules.xyz/client-go/api/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InitSpec struct {
	// Initialized indicates that this database has been initialized.
	// This will be set by the operator when status.conditions["Provisioned"] is set to ensure
	// that database is not mistakenly reset when recovered using disaster recovery tools.
	Initialized bool `json:"initialized,omitempty"`
	// Wait for initial DataRestore condition
	WaitForInitialRestore bool              `json:"waitForInitialRestore,omitempty"`
	Script                *ScriptSourceSpec `json:"script,omitempty"`

	Archiver *ArchiverRecovery `json:"archiver,omitempty"`
}

type ScriptSourceSpec struct {
	ScriptPath        string `json:"scriptPath,omitempty"`
	core.VolumeSource `json:",inline,omitempty"`
	Git               *GitRepo `json:"git,omitempty"`
}

type GitRepo struct {
	// https://github.com/kubernetes/git-sync/tree/master
	Args []string `json:"args"`
	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	Env []core.EnvVar `json:"env,omitempty"`
	// Security options the pod should run with.
	// More info: https://kubernetes.io/docs/concepts/policy/security-context/
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
	// +optional
	SecurityContext *core.SecurityContext `json:"securityContext,omitempty"`
	// Compute Resources required by the sidecar container.
	// +optional
	Resources core.ResourceRequirements `json:"resources,omitempty"`
	// Authentication secret for git repository
	// +optional
	AuthSecret *core.LocalObjectReference `json:"authSecret,omitempty"`
}

type RemoteReplicaSpec struct {
	// SourceRef specifies the  source object
	SourceRef core.ObjectReference `json:"sourceRef" protobuf:"bytes,1,opt,name=sourceRef"`
}

// +kubebuilder:validation:Enum=Provisioning;DataRestoring;Ready;Critical;NotReady;Halted;Unknown
type DatabasePhase string

const (
	// used for Databases that are currently provisioning
	DatabasePhaseProvisioning DatabasePhase = "Provisioning"
	// used for Databases for which data is currently restoring
	DatabasePhaseDataRestoring DatabasePhase = "DataRestoring"
	// used for Databases that are currently ReplicaReady, AcceptingConnection and Ready
	DatabasePhaseReady DatabasePhase = "Ready"
	// used for Databases that can connect, ReplicaReady == false || Ready == false (eg, ES yellow)
	DatabasePhaseCritical DatabasePhase = "Critical"
	// used for Databases that can't connect
	DatabasePhaseNotReady DatabasePhase = "NotReady"
	// used for Databases that are halted
	DatabasePhaseHalted DatabasePhase = "Halted"
	// used for Databases for which Phase can't be calculated
	DatabasePhaseUnknown DatabasePhase = "Unknown"
)

// +kubebuilder:validation:Enum=Durable;Ephemeral
type StorageType string

const (
	// default storage type and requires spec.storage to be configured
	StorageTypeDurable StorageType = "Durable"
	// Uses emptyDir as storage
	StorageTypeEphemeral StorageType = "Ephemeral"
)

// +kubebuilder:validation:Enum=Halt;Delete;WipeOut;DoNotTerminate
type DeletionPolicy string

const (
	// Deletes database pods, service but leave the PVCs and stash backup data intact.
	DeletionPolicyHalt DeletionPolicy = "Halt"
	// Deletes database pods, service, pvcs but leave the stash backup data intact.
	DeletionPolicyDelete DeletionPolicy = "Delete"
	// Deletes database pods, service, pvcs and stash backup data.
	DeletionPolicyWipeOut DeletionPolicy = "WipeOut"
	// Rejects attempt to delete database using ValidationWebhook.
	DeletionPolicyDoNotTerminate DeletionPolicy = "DoNotTerminate"
)

// +kubebuilder:validation:Enum=primary;standby;stats;dashboard;readreplica
type ServiceAlias string

const (
	PrimaryServiceAlias     ServiceAlias = "primary"
	StandbyServiceAlias     ServiceAlias = "standby"
	StatsServiceAlias       ServiceAlias = "stats"
	DashboardServiceAlias   ServiceAlias = "dashboard"
	ReadReplicaServiceAlias ServiceAlias = "readreplica"
)

// +kubebuilder:validation:Enum=fscopy;clone;sync;none
type PITRReplicationStrategy string

const (
	// ReplicationStrategySync means data will be synced from primary to secondary
	ReplicationStrategySync PITRReplicationStrategy = "sync"
	// ReplicationStrategyFSCopy means data will be copied from filesystem
	ReplicationStrategyFSCopy PITRReplicationStrategy = "fscopy"
	// ReplicationStrategyClone means volumeSnapshot will be used to create pvc's
	ReplicationStrategyClone PITRReplicationStrategy = "clone"
	// ReplicationStrategyNone means no replication will be used
	// data will be fully restored in every replicas instead of replication
	ReplicationStrategyNone PITRReplicationStrategy = "none"
)

// +kubebuilder:validation:Enum=DNS;IP;IPv4;IPv6
type AddressType string

const (
	AddressTypeDNS AddressType = "DNS"
	// Uses spec.podIP as address for db pods.
	AddressTypeIP AddressType = "IP"
	// Uses first IPv4 address from spec.podIP, spec.podIPs fields as address for db pods.
	AddressTypeIPv4 AddressType = "IPv4"
	// Uses first IPv6 address from spec.podIP, spec.podIPs fields as address for db pods.
	AddressTypeIPv6 AddressType = "IPv6"
)

func (a AddressType) IsIP() bool {
	return a == AddressTypeIP || a == AddressTypeIPv4 || a == AddressTypeIPv6
}

type NamedServiceTemplateSpec struct {
	// Alias represents the identifier of the service.
	Alias ServiceAlias `json:"alias"`

	// ServiceTemplate is an optional configuration for a service used to expose database
	// +optional
	ofstv1.ServiceTemplateSpec `json:",inline,omitempty"`
}

type KernelSettings struct {
	// DisableDefaults can be set to false to avoid defaulting via mutator
	DisableDefaults bool `json:"disableDefaults,omitempty"`
	// Privileged specifies the status whether the init container
	// requires privileged access to perform the following commands.
	// +optional
	Privileged bool `json:"privileged,omitempty"`
	// Sysctls hold a list of sysctls commands needs to apply to kernel.
	// +optional
	Sysctls []core.Sysctl `json:"sysctls,omitempty"`
}

// AutoOpsSpec defines the specifications of automatic ops-request recommendation generation
type AutoOpsSpec struct {
	// Disabled specifies whether the ops-request recommendation generation will be disabled or not.
	// +optional
	Disabled bool `json:"disabled,omitempty"`
}

type SystemUserSecretsSpec struct {
	// ReplicationUserSecret contains replication system user credentials
	// +optional
	ReplicationUserSecret *SecretReference `json:"replicationUserSecret,omitempty"`

	// MonitorUserSecret contains monitor system user credentials
	// +optional
	MonitorUserSecret *SecretReference `json:"monitorUserSecret,omitempty"`
}

type SecretReference struct {
	// +optional
	// SecretSource references the secret manager used for virtual secret
	SecretStoreName string `json:"secretStoreName,omitempty"`

	appcat.TypedLocalObjectReference `json:",inline,omitempty"`
	// Recommendation engine will generate RotateAuth opsReq using this field
	// +optional
	RotateAfter *metav1.Duration `json:"rotateAfter,omitempty"`
	// ActiveFrom holds the RFC3339 time. The referred authSecret is in-use from this timestamp.
	// +optional
	ActiveFrom        *metav1.Time `json:"activeFrom,omitempty"`
	ExternallyManaged bool         `json:"externallyManaged,omitempty"`
}

type Age struct {
	// Populated by Provisioner when authSecret is created or Ops Manager when authSecret is updated.
	LastUpdateTimestamp metav1.Time `json:"lastUpdateTimestamp,omitempty"`
}

type ConfigurationSpec struct {
	// SecretName is an optional field to provide custom configuration file for the database (i.e. elasticsearch.yml, mongod.conf ..).
	// If specified, these configurations will be used with default configurations (if any) and applyConfig configurations (if any).
	// configurations from this secret will override default configurations.
	// This secret must be created by user.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// Inline contains key-value pairs of configurations to be applied to the database.
	// These configurations will override both default configurations and configurations from the config secret (if any).
	// +optional
	Inline map[string]string `json:"inline,omitempty"`
}

type Archiver struct {
	// Pause is used to stop the archiver backup for the database
	// +optional
	Pause bool `json:"pause,omitempty"`
	// Ref is the name and namespace reference to the Archiver CR
	// Ref is the name and namespace reference to the Archiver CR
	Ref kmapi.ObjectReference `json:"ref"`
}

type ArchiverRecovery struct {
	RecoveryTimestamp metav1.Time `json:"recoveryTimestamp"`
	// +optional
	EncryptionSecret *kmapi.ObjectReference `json:"encryptionSecret,omitempty"`
	// +optional
	ManifestRepository *kmapi.ObjectReference `json:"manifestRepository,omitempty"`

	// FullDBRepository means db restore + manifest restore
	FullDBRepository    *kmapi.ObjectReference   `json:"fullDBRepository,omitempty"`
	ReplicationStrategy *PITRReplicationStrategy `json:"replicationStrategy,omitempty"`

	// ManifestOptions provide options to select particular manifest object to restore
	// +optional
	ManifestOptions *ManifestOptions `json:"manifestOptions,omitempty"`
}

type ManifestOptions struct {
	// Archiver specifies whether to restore the Archiver manifest or not
	// +kubebuilder:default=false
	// +optional
	Archiver *bool `json:"archiver,omitempty"`

	// ArchiverRef specifies the new name and namespace of the Archiver yaml after restore
	// +optional
	ArchiverRef *kmapi.ObjectReference `json:"archiverRef,omitempty"`

	// InitScript specifies whether to restore the InitScript or not
	// +kubebuilder:default=false
	// +optional
	InitScript *bool `json:"initScript,omitempty"`
}

type Accessor interface {
	GetObjectMeta() metav1.ObjectMeta
	GetConditions() []kmapi.Condition
	SetCondition(cond kmapi.Condition)
	RemoveCondition(typ string)
	client.Object
}

func setCondition(conditions []kmapi.Condition, cond kmapi.Condition) []kmapi.Condition {
	for i, c := range conditions {
		if c.Type == cond.Type {
			conditions[i] = cond
			return conditions
		}
	}
	conditions = append(conditions, cond)
	return conditions
}

func removeCondition(conditions []kmapi.Condition, typ string) []kmapi.Condition {
	for i, c := range conditions {
		if string(c.Type) == typ {
			conditions = append(conditions[:i], conditions[i+1:]...)
			break
		}
	}
	return conditions
}

// --- Log forwarding (shared, engine-neutral) ---

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
