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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:resource:shortName=vpacheckpoint

// VerticalPodAutopilotCheckpoint is the checkpoint of the internal state of VPA that
// is used for recovery after recommender's restart.
type VerticalPodAutopilotCheckpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the checkpoint.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// +optional
	Spec VerticalPodAutopilotCheckpointSpec `json:"spec,omitempty"`

	// Data of the checkpoint.
	// +optional
	Status VerticalPodAutopilotCheckpointStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutopilotCheckpointList is a list of VerticalPodAutopilotCheckpoint objects.
type VerticalPodAutopilotCheckpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []VerticalPodAutopilotCheckpoint `json:"items"`
}

// VerticalPodAutopilotCheckpointSpec is the specification of the checkpoint object.
type VerticalPodAutopilotCheckpointSpec struct {
	// Name of the VPA object that stored VerticalPodAutopilotCheckpoint object.
	VPAObjectName string `json:"vpaObjectName,omitempty"`

	// Name of the checkpointed container.
	ContainerName string `json:"containerName,omitempty"`
}

// VerticalPodAutopilotCheckpointStatus contains data of the checkpoint.
type VerticalPodAutopilotCheckpointStatus struct {
	// The time when the status was last refreshed.
	// +nullable
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Version of the format of the stored data.
	Version string `json:"version,omitempty"`

	// Checkpoint of histogram for consumption of CPU.
	CPUHistogram HistogramCheckpoint `json:"cpuHistogram,omitempty"`

	// Checkpoint of histogram for consumption of memory.
	MemoryHistogram HistogramCheckpoint `json:"memoryHistogram,omitempty"`

	// Timestamp of the fist sample from the histograms.
	// +nullable
	FirstSampleStart metav1.Time `json:"firstSampleStart,omitempty"`

	// Timestamp of the last sample from the histograms.
	// +nullable
	LastSampleStart metav1.Time `json:"lastSampleStart,omitempty"`

	// Total number of samples in the histograms.
	TotalSamplesCount int `json:"totalSamplesCount,omitempty"`
}

type BucketWeight struct {
	Index  int    `json:"index"`
	Weight uint32 `json:"weight"`
}

// HistogramCheckpoint contains data needed to reconstruct the histogram.
type HistogramCheckpoint struct {
	// Reference timestamp for samples collected within this histogram.
	// +nullable
	ReferenceTimestamp metav1.Time `json:"referenceTimestamp,omitempty"`

	// Map from bucket index to bucket weight.
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:XPreserveUnknownFields
	BucketWeights []BucketWeight `json:"bucketWeights,omitempty"`

	// Sum of samples to be used as denominator for weights from BucketWeights.
	TotalWeight float64 `json:"totalWeight,omitempty"`
}
