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
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindPostgresInsight = "PostgresInsight"
	ResourcePostgresInsight     = "postgresinsight"
	ResourcePostgresInsights    = "postgresinsights"
)

// PostgresInsightSpec defines the desired state of PostgresInsight
type PostgresInsightSpec struct {
	Version           string                      `json:"version" protobuf:"bytes,1,opt,name=version"`
	ConnectionURL     string                      `json:"connectionURL" protobuf:"bytes,2,opt,name=connectionURL"`
	Status            string                      `json:"status" protobuf:"bytes,3,opt,name=status"`
	Mode              string                      `json:"mode" protobuf:"bytes,4,opt,name=mode"`
	ReplicationStatus []PostgresReplicationStatus `json:"replicationStatus,omitempty" protobuf:"bytes,5,rep,name=replicationStatus"`
	ConnectionInfo    PostgresConnectionInfo      `json:"connectionInfo,omitempty" protobuf:"bytes,6,opt,name=connectionInfo"`
	BackupInfo        PostgresBackupInfo          `json:"backupInfo,omitempty" protobuf:"bytes,7,opt,name=backupInfo"`
	VacuumInfo        PostgresVacuumInfo          `json:"vacuumInfo,omitempty" protobuf:"bytes,8,opt,name=vacuumInfo"`
}

type PostgresVacuumInfo struct {
	AutoVacuum          string `json:"autoVacuum,omitempty" protobuf:"bytes,1,opt,name=autoVacuum"`
	ActiveVacuumProcess int64  `json:"activeVacuumProcess,omitempty" protobuf:"varint,2,opt,name=activeVacuumProcess"`
}

type PostgresBackupInfo struct {
}

type PostgresConnectionInfo struct {
	MaxConnections    int64 `json:"maxConnections,omitempty" protobuf:"varint,1,opt,name=maxConnections"`
	ActiveConnections int64 `json:"activeConnections,omitempty" protobuf:"varint,2,opt,name=activeConnections"`
}

// Ref: https://www.postgresql.org/docs/10/monitoring-stats.html#PG-STAT-REPLICATION-VIEW

type PostgresReplicationStatus struct {
	ApplicationName string `json:"applicationName,omitempty" protobuf:"bytes,1,opt,name=applicationName"`
	State           string `json:"state,omitempty" protobuf:"bytes,2,opt,name=state"`
	WriteLag        int64  `json:"writeLag,omitempty" protobuf:"varint,3,opt,name=writeLag"`
	FlushLag        int64  `json:"flushLag,omitempty" protobuf:"varint,4,opt,name=flushLag"`
	ReplayLag       int64  `json:"replayLag,omitempty" protobuf:"varint,5,opt,name=replayLag"`
}

// PostgresInsight is the Schema for the postgresinsights API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PostgresInsight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   PostgresInsightSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status api.PostgresStatus  `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PostgresInsightList contains a list of PostgresInsight

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PostgresInsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []PostgresInsight `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&PostgresInsight{}, &PostgresInsightList{})
}
