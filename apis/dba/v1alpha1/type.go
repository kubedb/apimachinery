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

type RequestConditionType string

// These are the possible conditions for a certificate request.
const (
	AccessApproved            RequestConditionType = "Approved"
	AccessDenied              RequestConditionType = "Denied"
	DisableSharding           RequestConditionType = "DisableSharding"
	EnableSharding            RequestConditionType = "EnableSharding"
	Failed                    RequestConditionType = "Failed"
	HorizontalScalingDatabase RequestConditionType = "HorizontalScaling"
	MigratingData             RequestConditionType = "MigratingData"
	NodeCreated               RequestConditionType = "NodeCreated"
	NodeDeleted               RequestConditionType = "NodeDeleted"
	NodeRestarted             RequestConditionType = "NodeRestarted"
	PauseDatabase             RequestConditionType = "PauseDatabase"
	Processing                RequestConditionType = "Processing"
	ResumeDatabase            RequestConditionType = "ResumeDatabase"
	ScalingDatabase           RequestConditionType = "Scaling"
	ScalingDown               RequestConditionType = "ScalingDown"
	ScalingUp                 RequestConditionType = "ScalingUp"
	StartingBalancer          RequestConditionType = "StartingBalancer"
	StoppingBalancer          RequestConditionType = "StoppingBalancer"
	Successful                RequestConditionType = "Successful"
	Updating                  RequestConditionType = "Updating"
	UpgradedDatabaseVersion   RequestConditionType = "UpgradedDatabaseVersion"
	UpgradingDatabaseVersion  RequestConditionType = "UpgradingDatabaseVersion"
	VerticalScalingDatabase   RequestConditionType = "VerticalScaling"
	VotingExclusionAdded      RequestConditionType = "VotingExclusionAdded"
	VotingExclusionDeleted    RequestConditionType = "VotingExclusionDeleted"
)

type ModificationRequestPhase string

const (
	// used for modification requests that are currently processing
	ModificationRequestPhaseProcessing ModificationRequestPhase = "Processing"
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

// +kubebuilder:validation:Enum=Upgrade;HorizontalScaling;VerticalScaling;VolumeExpansion;Restart
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
)

type UpdateSpec struct {
	// Specifies the ElasticsearchVersion object name
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
