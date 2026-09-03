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
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindDruidInsight = "DruidInsight"
	ResourceDruidInsight     = "druidinsight"
	ResourceDruidInsights    = "druidinsights"
)

type DruidInsightSpec struct {
	Version     string                  `json:"version"`
	Summary     DruidClusterSummary     `json:"summary"`
	Services    []DruidServiceStatus    `json:"services,omitempty"`
	Datasources []DruidDatasourceStatus `json:"datasources,omitempty"`
}

type DruidClusterSummary struct {
	TotalServices       *int32 `json:"totalServices,omitempty"`
	TotalDatasources    *int32 `json:"totalDatasources,omitempty"`
	TotalSegments       *int64 `json:"totalSegments,omitempty"`
	AvailableSegments   *int64 `json:"availableSegments,omitempty"`
	UnavailableSegments *int64 `json:"unavailableSegments,omitempty"`
	RealtimeSegments    *int64 `json:"realtimeSegments,omitempty"`
	TotalRows           *int64 `json:"totalRows,omitempty"`
	TotalSizeBytes      *int64 `json:"totalSizeBytes,omitempty"`
}

type DruidServiceStatus struct {
	Server              string       `json:"server"`
	Host                string       `json:"host,omitempty"`
	Type                string       `json:"type"`
	Tier                string       `json:"tier,omitempty"`
	PlaintextPort       *int32       `json:"plaintextPort,omitempty"`
	TLSPort             *int32       `json:"tlsPort,omitempty"`
	Leader              *bool        `json:"leader,omitempty"`
	StartTime           *metav1.Time `json:"startTime,omitempty"`
	Version             string       `json:"version,omitempty"`
	AvailableProcessors *int32       `json:"availableProcessors,omitempty"`
	TotalMemoryBytes    *int64       `json:"totalMemoryBytes,omitempty"`
	CurrentSizeBytes    *int64       `json:"currentSizeBytes,omitempty"`
	MaxSizeBytes        *int64       `json:"maxSizeBytes,omitempty"`
}

type DruidDatasourceStatus struct {
	Name                string       `json:"name"`
	SegmentCount        *int64       `json:"segmentCount,omitempty"`
	AvailableSegments   *int64       `json:"availableSegments,omitempty"`
	UnavailableSegments *int64       `json:"unavailableSegments,omitempty"`
	RealtimeSegments    *int64       `json:"realtimeSegments,omitempty"`
	TotalRows           *int64       `json:"totalRows,omitempty"`
	TotalSizeBytes      *int64       `json:"totalSizeBytes,omitempty"`
	MinTime             *metav1.Time `json:"minTime,omitempty"`
	MaxTime             *metav1.Time `json:"maxTime,omitempty"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DruidInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DruidInsightSpec  `json:"spec,omitempty"`
	Status dbapi.DruidStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DruidInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DruidInsight `json:"items"`
}
