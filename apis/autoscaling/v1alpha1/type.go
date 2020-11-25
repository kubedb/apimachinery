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
	"time"

	core "k8s.io/api/core/v1"
)

// List of possible condition types for a ops request
const (
	AccessApproved            = "Approved"
	AccessDenied              = "Denied"
	DisableSharding           = "DisableSharding"
	EnableSharding            = "EnableSharding"
	Failure                   = "Failure"
	HorizontalScalingDatabase = "HorizontalScaling"
	MigratingData             = "MigratingData"
	NodeCreated               = "NodeCreated"
	NodeDeleted               = "NodeDeleted"
	NodeRestarted             = "NodeRestarted"
	PauseDatabase             = "PauseDatabase"
	PausingDatabase           = "PausingDatabase"
	PausedDatabase            = "PausedDatabase"
	Progressing               = "Progressing"
	ResumeDatabase            = "ResumeDatabase"
	ResumingDatabase          = "ResumingDatabase"
	ResumedDatabase           = "ResumedDatabase"
	ScalingDatabase           = "Scaling"
	ScalingDown               = "ScalingDown"
	ScalingUp                 = "ScalingUp"
	Successful                = "Successful"
	Updating                  = "Updating"
	UpgradedVersion           = "UpgradedVersion"
	UpgradingVersion          = "UpgradingVersion"
	VerticalScalingDatabase   = "VerticalScaling"
	VotingExclusionAdded      = "VotingExclusionAdded"
	VotingExclusionDeleted    = "VotingExclusionDeleted"
	UpdateStatefulSets        = "UpdateStatefulSets"

	// MongoDB Constants
	StartingBalancer            = "StartingBalancer"
	StoppingBalancer            = "StoppingBalancer"
	UpdateShardImage            = "UpdateShardImage"
	UpdateStatefulSetResources  = "UpdateStatefulSetResources"
	UpdateShardResources        = "UpdateShardResources"
	ScaleDownShard              = "ScaleDownShard"
	ScaleUpShard                = "ScaleUpShard"
	UpdateReplicaSetImage       = "UpdateReplicaSetImage"
	UpdateConfigServerImage     = "UpdateConfigServerImage"
	UpdateMongosImage           = "UpdateMongosImage"
	UpdateReplicaSetResources   = "UpdateReplicaSetResources"
	UpdateConfigServerResources = "UpdateConfigServerResources"
	UpdateMongosResources       = "UpdateMongosResources"
	FlushRouterConfig           = "FlushRouterConfig"
	ScaleDownReplicaSet         = "ScaleDownReplicaSet"
	ScaleUpReplicaSet           = "ScaleUpReplicaSet"
	ScaleUpShardReplicas        = "ScaleUpShardReplicas"
	ScaleDownShardReplicas      = "ScaleDownShardReplicas"
	ScaleDownConfigServer       = "ScaleDownConfigServer "
	ScaleUpConfigServer         = "ScaleUpConfigServer "
	ScaleMongos                 = "ScaleMongos"
	VolumeExpansion             = "VolumeExpansion"
)

type AutoscalerPhase string

const (
	// used for ops requests that are currently Progressing
	AutoscalerPhaseProgressing AutoscalerPhase = "Progressing"
	// used for ops requests that are executed successfully
	AutoscalerPhaseSuccessful AutoscalerPhase = "Successful"
	// used for ops requests that are waiting for approval
	AutoscalerPhaseWaitingForApproval AutoscalerPhase = "WaitingForApproval"
	// used for ops requests that are failed
	AutoscalerPhaseFailed AutoscalerPhase = "Failed"
	// used for ops requests that are approved
	AutoscalerApproved AutoscalerPhase = "Approved"
	// used for ops requests that are denied
	AutoscalerDenied AutoscalerPhase = "Denied"
)

// +kubebuilder:validation:Enum=Upgrade;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;RotateCertificates
type AutoscalerType string

const (
	// used for Upgrade operation
	AutoscalerTypeUpgrade AutoscalerType = "Upgrade"
	// used for HorizontalScaling operation
	AutoscalerTypeHorizontalScaling AutoscalerType = "HorizontalScaling"
	// used for VerticalScaling operation
	AutoscalerTypeVerticalScaling AutoscalerType = "VerticalScaling"
	// used for VolumeExpansion operation
	AutoscalerTypeVolumeExpansion AutoscalerType = "VolumeExpansion"
	// used for Restart operation
	AutoscalerTypeRestart AutoscalerType = "Restart"
	// used for RotateCertificates operation
	AutoscalerTypeRotateCertificates AutoscalerType = "RotateCertificates"
)

// ContainerControlledValues controls which resource value should be autoscaled.
type ContainerControlledValues string

const (
	// ContainerControlledValuesRequestsAndLimits means resource request and limits
	// are scaled automatically. The limit is scaled proportionally to the request.
	ContainerControlledValuesRequestsAndLimits ContainerControlledValues = "RequestsAndLimits"
	// ContainerControlledValuesRequestsOnly means only requested resource is autoscaled.
	ContainerControlledValuesRequestsOnly ContainerControlledValues = "RequestsOnly"
)

// AutoscalerTrigger controls if autoscaler is enabled.
type AutoscalerTrigger string

const (
	// AutoscalerTriggerOn means the autoscaler is enabled.
	AutoscalerTriggerOn AutoscalerTrigger = "On"
	// AutoscalerTriggerOff means the autoscaler is disabled.
	AutoscalerTriggerOff AutoscalerTrigger = "Off"
)

type ComputeAutoscalerSpec struct {
	// Whether compute autoscaler is enabled. The default is Off".
	Trigger AutoscalerTrigger `json:"trigger,omitempty" protobuf:"bytes,9,opt,name=trigger,casttype=AutoscalerTrigger"`
	// Specifies the minimal amount of resources that will be recommended.
	// The default is no minimum.
	// +optional
	MinAllowed core.ResourceList `json:"minAllowed,omitempty" protobuf:"bytes,2,rep,name=minAllowed,casttype=k8s.io/api/core/v1.ResourceList,castkey=k8s.io/api/core/v1.ResourceName"`
	// Specifies the maximum amount of resources that will be recommended.
	// The default is no maximum.
	// +optional
	MaxAllowed core.ResourceList `json:"maxAllowed,omitempty" protobuf:"bytes,3,rep,name=maxAllowed,casttype=k8s.io/api/core/v1.ResourceList,castkey=k8s.io/api/core/v1.ResourceName"`

	// Specifies the type of recommendations that will be computed
	// (and possibly applied) by VPA.
	// If not specified, the default of [ResourceCPU, ResourceMemory] will be used.
	// +optional
	// +patchStrategy=merge
	ControlledResources []core.ResourceName `json:"controlledResources,omitempty" patchStrategy:"merge" protobuf:"bytes,5,rep,name=controlledResources,casttype=k8s.io/api/core/v1.ResourceName"`

	// Specifies which resource values should be controlled.
	// The default is "RequestsAndLimits".
	// +optional
	ContainerControlledValues *ContainerControlledValues `json:"containerControlledValues,omitempty" protobuf:"bytes,6,opt,name=containerControlledValues,casttype=ContainerControlledValues"`

	// Specifies the minimum resource difference in percentage
	// The default is 10%.
	// +optional
	ResourceDiffPercentage int32 `json:"resourceDiffPercentage,omitempty" protobuf:"varint,7,opt,name=resourceDiffPercentage"`

	// Specifies the minimum pod life time
	// The default is 12h.
	// +optional
	PodLifeTimeThreshold time.Duration `json:"podLifeTimeThreshold,omitempty" protobuf:"varint,8,opt,name=podLifeTimeThreshold,casttype=time.Duration"`
}

type StorageAutoscalerSpec struct {
	// Whether compute autoscaler is enabled. The default is Off".
	Trigger          AutoscalerTrigger `json:"trigger,omitempty" protobuf:"bytes,1,opt,name=trigger,casttype=AutoscalerTrigger"`
	UsageThreshold   int32             `json:"usageThreshold,omitempty" protobuf:"varint,2,opt,name=usageThreshold"`
	ScalingThreshold int32             `json:"scalingThreshold,omitempty" protobuf:"varint,3,opt,name=scalingThreshold"`
}
