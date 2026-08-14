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

//go:generate go-enum --mustparse --names --values
package v1alpha1

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeEtcdOpsRequest     = "etcdops"
	ResourceKindEtcdOpsRequest     = "EtcdOpsRequest"
	ResourceSingularEtcdOpsRequest = "etcdopsrequest"
	ResourcePluralEtcdOpsRequest   = "etcdopsrequests"
)

// EtcdOpsRequest defines an Etcd DBA operation.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=etcdopsrequests,singular=etcdopsrequest,shortName=etcdops,categories={ops,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type EtcdOpsRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EtcdOpsRequestSpec `json:"spec,omitempty"`
	Status OpsRequestStatus   `json:"status,omitempty"`
}

// EtcdOpsRequestSpec is the spec for EtcdOpsRequest
type EtcdOpsRequestSpec struct {
	// Specifies the Etcd reference
	DatabaseRef core.LocalObjectReference `json:"databaseRef"`
	// Specifies the ops request type: UpdateVersion, HorizontalScaling, VerticalScaling etc.
	Type EtcdOpsRequestType `json:"type"`
	// Specifies information necessary for updating Etcd version
	UpdateVersion *EtcdUpdateVersionSpec `json:"updateVersion,omitempty"`
	// Specifies information necessary for horizontal scaling
	HorizontalScaling *EtcdHorizontalScalingSpec `json:"horizontalScaling,omitempty"`
	// Specifies information necessary for vertical scaling
	VerticalScaling *EtcdVerticalScalingSpec `json:"verticalScaling,omitempty"`
	// Specifies information necessary for volume expansion
	VolumeExpansion *EtcdVolumeExpansionSpec `json:"volumeExpansion,omitempty"`
	// Specifies information necessary for custom configuration of Etcd
	Configuration *EtcdCustomConfigurationSpec `json:"configuration,omitempty"`
	// Specifies information necessary for configuring TLS
	TLS *EtcdTLSSpec `json:"tls,omitempty"`
	// Specifies information necessary for configuring authSecret of the database
	Authentication *AuthSpec `json:"authentication,omitempty"`
	// Specifies information necessary for restarting database
	Restart *RestartSpec `json:"restart,omitempty"`
	// Specifies information necessary for migrating storageClass of the PVCs
	Migration *StorageMigrationSpec `json:"migration,omitempty"`
	// Specifies information necessary for moving the raft leadership to another member
	MoveLeader *EtcdMoveLeaderSpec `json:"moveLeader,omitempty"`
	// Specifies information necessary for defragmenting the backend of the etcd members
	Defragment *EtcdDefragmentSpec `json:"defragment,omitempty"`
	// Specifies information necessary for compacting the etcd keyspace history
	Compact *EtcdCompactSpec `json:"compact,omitempty"`
	// Timeout for each step of the ops request in second. If a step doesn't finish within the specified timeout, the ops request will result in failure.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// ApplyOption is to control the execution of OpsRequest depending on the database state.
	// +kubebuilder:default="IfReady"
	Apply ApplyOption `json:"apply,omitempty"`
	// +kubebuilder:default=1
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

// +kubebuilder:validation:Enum=UpdateVersion;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;Reconfigure;ReconfigureTLS;RotateAuth;StorageMigration;MoveLeader;Defragment;Compact
// ENUM(UpdateVersion, HorizontalScaling, VerticalScaling, VolumeExpansion, Restart, Reconfigure, ReconfigureTLS, RotateAuth, StorageMigration, MoveLeader, Defragment, Compact)
type EtcdOpsRequestType string

// EtcdUpdateVersionSpec contains the update version information of an Etcd cluster
type EtcdUpdateVersionSpec struct {
	// Specifies the target version name from catalog
	TargetVersion string `json:"targetVersion,omitempty"`
}

// EtcdHorizontalScalingSpec contains the horizontal scaling information of an Etcd cluster.
// Members are added as learners and promoted once they have caught up, and removed
// through the etcd membership API before the pod is deleted.
type EtcdHorizontalScalingSpec struct {
	// Number of etcd members in the cluster
	Replicas *int32 `json:"replicas,omitempty"`
}

// EtcdVerticalScalingSpec is the spec for Etcd vertical scaling
type EtcdVerticalScalingSpec struct {
	// Resource spec for the etcd container
	Etcd *ContainerResources `json:"etcd,omitempty"`
	// Resource spec for the exporter sidecar container
	Exporter *ContainerResources `json:"exporter,omitempty"`
	// Mode selects how the vertical scaling is actuated. Defaults to Restart.
	// +optional
	// +kubebuilder:default=Restart
	Mode VerticalScalingMode `json:"mode,omitempty"`
}

// EtcdVolumeExpansionSpec is the spec for Etcd volume expansion
type EtcdVolumeExpansionSpec struct {
	// volume specification for the etcd data directory
	Etcd *resource.Quantity  `json:"etcd,omitempty"`
	Mode VolumeExpansionMode `json:"mode"`
}

// EtcdCustomConfigurationSpec is the Etcd-specific reconfiguration spec.
// It embeds the generic ReconfigurationSpec and adds the KubeDB-managed etcd
// tuning knobs which end up on the etcd command line.
type EtcdCustomConfigurationSpec struct {
	ReconfigurationSpec `json:",inline,omitempty"`

	// Tuning holds the KubeDB-managed etcd tuning knobs to apply.
	// +optional
	Tuning *dbapi.EtcdTuningConfig `json:"tuning,omitempty"`
}

// EtcdTLSSpec is the spec for reconfiguring the Etcd TLS configuration.
// The embedded TLSSpec already carries `rotateCertificates` and `remove`, which
// cover certificate rotation and TLS removal for the client, peer and metrics
// certificates alike.
type EtcdTLSSpec struct {
	TLSSpec `json:",inline,omitempty"`
}

// EtcdMoveLeaderSpec is the spec for transferring the raft leadership of an Etcd cluster.
type EtcdMoveLeaderSpec struct {
	// NewLeader is the name of the member (pod) that the raft leadership should be
	// transferred to. If empty, the operator picks a healthy, up-to-date member.
	// +optional
	NewLeader string `json:"newLeader,omitempty"`
}

// EtcdDefragmentSpec is the spec for defragmenting the backend of the Etcd members.
// Defragmentation is always applied to every member, one at a time, so that the
// cluster never loses quorum.
type EtcdDefragmentSpec struct{}

// EtcdCompactSpec is the spec for compacting the Etcd keyspace history.
type EtcdCompactSpec struct {
	// Revision is the revision to compact the keyspace history up to. If it is not
	// given, the operator compacts up to the current revision at execution time.
	// +optional
	Revision *int64 `json:"revision,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EtcdOpsRequestList is a list of EtcdOpsRequests
type EtcdOpsRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of EtcdOpsRequest CRD objects
	Items []EtcdOpsRequest `json:"items,omitempty"`
}
