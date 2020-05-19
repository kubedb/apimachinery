/*
Copyright The KubeDB Authors.

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
	core "k8s.io/api/core/v1"
)

// List of possible condition types for a modification request
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
	Progressing               = "Progressing"
	ResumeDatabase            = "ResumeDatabase"
	ScalingDatabase           = "Scaling"
	ScalingDown               = "ScalingDown"
	ScalingUp                 = "ScalingUp"
	StartingBalancer          = "StartingBalancer"
	StoppingBalancer          = "StoppingBalancer"
	Successful                = "Successful"
	Updating                  = "Updating"
	UpgradedVersion           = "UpgradedVersion"
	UpgradingVersion          = "UpgradingVersion"
	VerticalScalingDatabase   = "VerticalScaling"
	VotingExclusionAdded      = "VotingExclusionAdded"
	VotingExclusionDeleted    = "VotingExclusionDeleted"
)

type ModificationRequestPhase string

const (
	// used for modification requests that are currently Progressing
	ModificationRequestPhaseProgressing ModificationRequestPhase = "Progressing"
	// used for modification requests that are executed successfully
	ModificationRequestPhaseSuccessful ModificationRequestPhase = "Successful"
	// used for modification requests that are waiting for approval
	ModificationRequestPhaseWaitingForApproval ModificationRequestPhase = "WaitingForApproval"
	// used for modification requests that are failed
	ModificationRequestPhaseFailed ModificationRequestPhase = "Failed"
	// used for modification requests that are approved
	ModificationRequestApproved ModificationRequestPhase = "Approved"
	// used for modification requests that are denied
	ModificationRequestDenied ModificationRequestPhase = "Denied"
)

// +kubebuilder:validation:Enum=Upgrade;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart;RotateCertificates
type ModificationRequestType string

const (
	// used for Upgrade operation
	ModificationRequestTypeUpgrade ModificationRequestType = "Upgrade"
	// used for HorizontalScaling operation
	ModificationRequestTypeHorizontalScaling ModificationRequestType = "HorizontalScaling"
	// used for VerticalScaling operation
	ModificationRequestTypeVerticalScaling ModificationRequestType = "VerticalScaling"
	// used for VolumeExpansion operation
	ModificationRequestTypeVolumeExpansion ModificationRequestType = "VolumeExpansion"
	// used for Restart operation
	ModificationRequestTypeRestart ModificationRequestType = "Restart"
	// used for RotateCertificates operation
	ModificationRequestTypeRotateCertificates ModificationRequestType = "RotateCertificates"
)

type UpgradeSpec struct {
	// Specifies the MySQLVersion object name
	TargetVersion string `json:"targetVersion,omitempty" protobuf:"bytes,1,opt,name=targetVersion"`
}

// Resources requested by a single application container
type ContainerResources struct {
	// Name of the container specified as a DNS_LABEL.
	// Each container in a pod must have a unique name (DNS_LABEL).
	// Cannot be updated.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources core.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,2,opt,name=resources"`
}
