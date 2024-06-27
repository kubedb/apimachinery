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
	"strings"
	"unsafe"

	"kubedb.dev/apimachinery/apis/kubedb"
	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientgoapiv1 "kmodules.xyz/client-go/api/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	monitoringagentapiapiv1 "kmodules.xyz/monitoring-agent-api/api/v1"
	ofstv1 "kmodules.xyz/offshoot-api/api/v1"
	ofstconv "kmodules.xyz/offshoot-api/api/v1/conversion"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
	rtconv "sigs.k8s.io/controller-runtime/pkg/conversion"
)

// Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec is an autogenerated conversion function.
func Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(in *ofstv1.PodTemplateSpec, out *ofstv2.PodTemplateSpec, s conversion.Scope) error {
	return ofstconv.Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(in, out, s)
}

// Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec is an autogenerated conversion function.
func Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(in *ofstv2.PodTemplateSpec, out *ofstv1.PodTemplateSpec, s conversion.Scope) error {
	return ofstconv.Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(in, out, s)
}

// Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec is an autogenerated conversion function.
func Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(in *[]corev1.Container, out *CoordinatorSpec, s conversion.Scope) error {
	var container *corev1.Container
	for i := range *in {
		if strings.HasSuffix((*in)[i].Name, "coordinator") {
			container = &(*in)[i]
		}
	}
	if container == nil || !(container.Resources.Requests != nil || container.Resources.Limits != nil) {
		return nil
	}
	out.Resources = *container.Resources.DeepCopy()
	out.SecurityContext = container.SecurityContext
	return nil
}

// Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container is an autogenerated conversion function.
func Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(in *CoordinatorSpec, out *[]corev1.Container, s conversion.Scope) error {
	if in == nil {
		return nil
	}

	var container corev1.Container
	container.SecurityContext = (*corev1.SecurityContext)(unsafe.Pointer(in.SecurityContext))
	container.Resources = *in.Resources.DeepCopy()
	found := false

	for i := range *out {
		if strings.HasSuffix((*out)[i].Name, "coordinator") || (*out)[i].Name == kubedb.ReplicationModeDetectorContainerName {
			if container.SecurityContext != nil {
				(*out)[i].SecurityContext = container.SecurityContext
			}
			if container.Resources.Limits != nil || container.Resources.Requests != nil {
				(*out)[i].Resources = *container.Resources.DeepCopy()
			}
			found = true
		}
	}
	if !found && (container.SecurityContext != nil || container.Resources.Limits != nil || container.Resources.Requests != nil) {
		*out = append(*out, container)
	}

	return nil
}

func Convert_v1_ElasticsearchNode_To_v1alpha2_ElasticsearchNode(in *v1.ElasticsearchNode, out *ElasticsearchNode, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Suffix = in.Suffix
	out.HeapSizePercentage = (*int32)(unsafe.Pointer(in.HeapSizePercentage))
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.MaxUnavailable = (*intstr.IntOrString)(unsafe.Pointer(in.MaxUnavailable))
	out.NodeSelector = *(*map[string]string)(unsafe.Pointer(&in.PodTemplate.Spec.NodeSelector))
	out.Tolerations = *(*[]corev1.Toleration)(unsafe.Pointer(&in.PodTemplate.Spec.Tolerations))
	dbContainer := core_util.GetContainerByName(in.PodTemplate.Spec.Containers, kubedb.ElasticsearchContainerName)
	if dbContainer != nil {
		out.Resources = *(*corev1.ResourceRequirements)(unsafe.Pointer(&dbContainer.Resources))
	}
	return nil
}

func Convert_v1alpha2_ElasticsearchNode_To_v1_ElasticsearchNode(in *ElasticsearchNode, out *v1.ElasticsearchNode, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Suffix = in.Suffix
	out.HeapSizePercentage = (*int32)(unsafe.Pointer(in.HeapSizePercentage))
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.MaxUnavailable = (*intstr.IntOrString)(unsafe.Pointer(in.MaxUnavailable))
	out.PodTemplate.Spec.NodeSelector = *(*map[string]string)(unsafe.Pointer(&in.NodeSelector))
	out.PodTemplate.Spec.Tolerations = *(*[]corev1.Toleration)(unsafe.Pointer(&in.Tolerations))
	return nil
}

func Convert_v1alpha2_MariaDBSpec_To_v1_MariaDBSpec(in *MariaDBSpec, out *v1.MariaDBSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.WsrepSSTMethod = v1.GaleraWsrepSSTMethod(in.WsrepSSTMethod)
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	// WARNING: in.Coordinator requires manual conversion: does not exist in peer-type
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.MariaDBCoordinatorContainerName
		}
	}
	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*v1.Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1_MariaDBSpec_To_v1alpha2_MariaDBSpec(in *v1.MariaDBSpec, out *MariaDBSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.WsrepSSTMethod = GaleraWsrepSSTMethod(in.WsrepSSTMethod)
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1alpha2_PostgresSpec_To_v1_PostgresSpec(in *PostgresSpec, out *v1.PostgresSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StandbyMode = (*v1.PostgresStandbyMode)(unsafe.Pointer(in.StandbyMode))
	out.StreamingMode = (*v1.PostgresStreamingMode)(unsafe.Pointer(in.StreamingMode))
	out.Mode = (*v1.PostgreSQLMode)(unsafe.Pointer(in.Mode))
	out.RemoteReplica = (*v1.RemoteReplicaSpec)(unsafe.Pointer(in.RemoteReplica))
	out.LeaderElection = (*v1.PostgreLeaderElectionConfig)(unsafe.Pointer(in.LeaderElection))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.ClientAuthMode = v1.PostgresClientAuthMode(in.ClientAuthMode)
	out.SSLMode = v1.PostgresSSLMode(in.SSLMode)
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.PostgresCoordinatorContainerName
		}
	}
	out.EnforceFsGroup = in.EnforceFsGroup
	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*v1.Archiver)(unsafe.Pointer(in.Archiver))
	out.Arbiter = (*v1.ArbiterSpec)(unsafe.Pointer(in.Arbiter))
	out.Replication = (*v1.PostgresReplication)(unsafe.Pointer(in.Replication))
	return nil
}

func Convert_v1_PostgresSpec_To_v1alpha2_PostgresSpec(in *v1.PostgresSpec, out *PostgresSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StandbyMode = (*PostgresStandbyMode)(unsafe.Pointer(in.StandbyMode))
	out.StreamingMode = (*PostgresStreamingMode)(unsafe.Pointer(in.StreamingMode))
	out.Mode = (*PostgreSQLMode)(unsafe.Pointer(in.Mode))
	out.RemoteReplica = (*RemoteReplicaSpec)(unsafe.Pointer(in.RemoteReplica))
	out.LeaderElection = (*PostgreLeaderElectionConfig)(unsafe.Pointer(in.LeaderElection))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.ClientAuthMode = PostgresClientAuthMode(in.ClientAuthMode)
	out.SSLMode = PostgresSSLMode(in.SSLMode)
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.EnforceFsGroup = in.EnforceFsGroup
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*Archiver)(unsafe.Pointer(in.Archiver))
	out.Arbiter = (*ArbiterSpec)(unsafe.Pointer(in.Arbiter))
	out.Replication = (*PostgresReplication)(unsafe.Pointer(in.Replication))
	return nil
}

func Convert_v1alpha2_MySQLSpec_To_v1_MySQLSpec(in *MySQLSpec, out *v1.MySQLSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	if in.Topology != nil {
		in, out := &in.Topology, &out.Topology
		*out = new(v1.MySQLTopology)
		if err := Convert_v1alpha2_MySQLTopology_To_v1_MySQLTopology(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Topology = nil
	}
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	out.UseAddressType = v1.AddressType(in.UseAddressType)
	// WARNING: in.Coordinator requires manual conversion: does not exist in peer-type
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.MySQLCoordinatorContainerName
		}
	}
	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.AllowedReadReplicas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedReadReplicas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*v1.Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1_MySQLSpec_To_v1alpha2_MySQLSpec(in *v1.MySQLSpec, out *MySQLSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	if in.Topology != nil {
		in, out := &in.Topology, &out.Topology
		*out = new(MySQLTopology)
		if err := Convert_v1_MySQLTopology_To_v1alpha2_MySQLTopology(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Topology = nil
	}
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.UseAddressType = AddressType(in.UseAddressType)
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.AllowedReadReplicas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedReadReplicas))
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1alpha2_MongoDBSpec_To_v1_MongoDBSpec(in *MongoDBSpec, out *v1.MongoDBSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.ReplicaSet = (*v1.MongoDBReplicaSet)(unsafe.Pointer(in.ReplicaSet))
	if in.ShardTopology != nil {
		in, out := &in.ShardTopology, &out.ShardTopology
		*out = new(v1.MongoDBShardingTopology)
		if err := Convert_v1alpha2_MongoDBShardingTopology_To_v1_MongoDBShardingTopology(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.ShardTopology = nil
	}
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.EphemeralStorage = (*corev1.EmptyDirVolumeSource)(unsafe.Pointer(in.EphemeralStorage))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.ClusterAuthMode = v1.ClusterAuthMode(in.ClusterAuthMode)
	out.SSLMode = v1.SSLMode(in.SSLMode)
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(ofstv2.PodTemplateSpec)
		if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.PodTemplate = nil
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.KeyFileSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.KeyFileSecret))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	out.StorageEngine = v1.StorageEngine(in.StorageEngine)
	// WARNING: in.Coordinator requires manual conversion: does not exist in peer-type
	if out.PodTemplate == nil {
		out.PodTemplate = &ofstv2.PodTemplateSpec{}
	}
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.ReplicationModeDetectorContainerName
		}
	}
	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	if in.Arbiter != nil {
		in, out := &in.Arbiter, &out.Arbiter
		*out = new(v1.MongoArbiterNode)
		if err := Convert_v1alpha2_MongoArbiterNode_To_v1_MongoArbiterNode(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Arbiter = nil
	}
	if in.Hidden != nil {
		in, out := &in.Hidden, &out.Hidden
		*out = new(v1.MongoHiddenNode)
		if err := Convert_v1alpha2_MongoHiddenNode_To_v1_MongoHiddenNode(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Hidden = nil
	}
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*v1.Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1_MongoDBSpec_To_v1alpha2_MongoDBSpec(in *v1.MongoDBSpec, out *MongoDBSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.ReplicaSet = (*MongoDBReplicaSet)(unsafe.Pointer(in.ReplicaSet))
	if in.ShardTopology != nil {
		in, out := &in.ShardTopology, &out.ShardTopology
		*out = new(MongoDBShardingTopology)
		if err := Convert_v1_MongoDBShardingTopology_To_v1alpha2_MongoDBShardingTopology(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.ShardTopology = nil
	}
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.EphemeralStorage = (*corev1.EmptyDirVolumeSource)(unsafe.Pointer(in.EphemeralStorage))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.ClusterAuthMode = ClusterAuthMode(in.ClusterAuthMode)
	out.SSLMode = SSLMode(in.SSLMode)
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if in.PodTemplate != nil {
		in, out := &in.PodTemplate, &out.PodTemplate
		*out = new(ofstv1.PodTemplateSpec)
		if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.PodTemplate = nil
	}
	if in.PodTemplate != nil {
		if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
			return err
		}
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.KeyFileSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.KeyFileSecret))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.StorageEngine = StorageEngine(in.StorageEngine)
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	if in.Arbiter != nil {
		in, out := &in.Arbiter, &out.Arbiter
		*out = new(MongoArbiterNode)
		if err := Convert_v1_MongoArbiterNode_To_v1alpha2_MongoArbiterNode(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Arbiter = nil
	}
	if in.Hidden != nil {
		in, out := &in.Hidden, &out.Hidden
		*out = new(MongoHiddenNode)
		if err := Convert_v1_MongoHiddenNode_To_v1alpha2_MongoHiddenNode(*in, *out, s); err != nil {
			return err
		}
	} else {
		out.Hidden = nil
	}
	out.HealthChecker = in.HealthChecker
	out.Archiver = (*Archiver)(unsafe.Pointer(in.Archiver))
	return nil
}

func Convert_v1alpha2_RedisSpec_To_v1_RedisSpec(in *RedisSpec, out *v1.RedisSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Mode = v1.RedisMode(in.Mode)
	out.SentinelRef = (*v1.RedisSentinelRef)(unsafe.Pointer(in.SentinelRef))
	out.Cluster = (*v1.RedisClusterSpec)(unsafe.Pointer(in.Cluster))
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.DisableAuth = in.DisableAuth
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	// WARNING: in.Coordinator requires manual conversion: does not exist in peer-type
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.RedisCoordinatorContainerName
		}
	}
	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	return nil
}

func Convert_v1_RedisSpec_To_v1alpha2_RedisSpec(in *v1.RedisSpec, out *RedisSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Mode = RedisMode(in.Mode)
	out.SentinelRef = (*RedisSentinelRef)(unsafe.Pointer(in.SentinelRef))
	out.Cluster = (*RedisClusterSpec)(unsafe.Pointer(in.Cluster))
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.DisableAuth = in.DisableAuth
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	return nil
}

func Convert_v1alpha2_PerconaXtraDBSpec_To_v1_PerconaXtraDBSpec(in *PerconaXtraDBSpec, out *v1.PerconaXtraDBSpec, s conversion.Scope) error {
	if err := Convert_v1alpha2_AutoOpsSpec_To_v1_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StorageType = v1.StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*v1.SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.Init = (*v1.InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v1_PodTemplateSpec_To_v2_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]v1.NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = v1.TerminationPolicy(in.TerminationPolicy)
	// WARNING: in.Coordinator requires manual conversion: does not exist in peer-type
	if err := Convert_v1alpha2_CoordinatorSpec_To_Slice_v1_Container(&in.Coordinator, &out.PodTemplate.Spec.Containers, s); err != nil {
		return err
	}
	for i := range out.PodTemplate.Spec.Containers {
		if out.PodTemplate.Spec.Containers[i].Name == "" {
			out.PodTemplate.Spec.Containers[i].Name = kubedb.PerconaXtraDBCoordinatorContainerName
		}
	}

	out.AllowedSchemas = (*v1.AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.SystemUserSecrets = (*v1.SystemUserSecretsSpec)(unsafe.Pointer(in.SystemUserSecrets))
	return nil
}

func Convert_v1_PerconaXtraDBSpec_To_v1alpha2_PerconaXtraDBSpec(in *v1.PerconaXtraDBSpec, out *PerconaXtraDBSpec, s conversion.Scope) error {
	if err := Convert_v1_AutoOpsSpec_To_v1alpha2_AutoOpsSpec(&in.AutoOps, &out.AutoOps, s); err != nil {
		return err
	}
	out.Version = in.Version
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.StorageType = StorageType(in.StorageType)
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.AuthSecret = (*SecretReference)(unsafe.Pointer(in.AuthSecret))
	out.Init = (*InitSpec)(unsafe.Pointer(in.Init))
	out.Monitor = (*monitoringagentapiapiv1.AgentSpec)(unsafe.Pointer(in.Monitor))
	out.ConfigSecret = (*corev1.LocalObjectReference)(unsafe.Pointer(in.ConfigSecret))
	if err := Convert_v2_PodTemplateSpec_To_v1_PodTemplateSpec(&in.PodTemplate, &out.PodTemplate, s); err != nil {
		return err
	}

	if err := Convert_Slice_v1_Container_To_v1alpha2_CoordinatorSpec(&in.PodTemplate.Spec.Containers, &out.Coordinator, s); err != nil {
		return err
	}
	out.ServiceTemplates = *(*[]NamedServiceTemplateSpec)(unsafe.Pointer(&in.ServiceTemplates))
	out.RequireSSL = in.RequireSSL
	out.TLS = (*clientgoapiv1.TLSConfig)(unsafe.Pointer(in.TLS))
	out.Halted = in.Halted
	out.TerminationPolicy = TerminationPolicy(in.TerminationPolicy)
	out.AllowedSchemas = (*AllowedConsumers)(unsafe.Pointer(in.AllowedSchemas))
	out.HealthChecker = in.HealthChecker
	out.SystemUserSecrets = (*SystemUserSecretsSpec)(unsafe.Pointer(in.SystemUserSecrets))
	return nil
}

func Convert_v1alpha2_KafkaNode_To_v1_KafkaNode(in *KafkaNode, out *v1.KafkaNode, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Suffix = in.Suffix
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.PodTemplate.Spec.NodeSelector = *(*map[string]string)(unsafe.Pointer(&in.NodeSelector))
	out.PodTemplate.Spec.Tolerations = *(*[]corev1.Toleration)(unsafe.Pointer(&in.Tolerations))
	out.PodTemplate.Spec.Containers = core_util.UpsertContainer(out.PodTemplate.Spec.Containers, corev1.Container{
		Name:      kubedb.KafkaContainerName,
		Resources: in.Resources,
	})
	return nil
}

func Convert_v1_KafkaNode_To_v1alpha2_KafkaNode(in *v1.KafkaNode, out *KafkaNode, s conversion.Scope) error {
	out.Replicas = (*int32)(unsafe.Pointer(in.Replicas))
	out.Suffix = in.Suffix
	out.Storage = (*corev1.PersistentVolumeClaimSpec)(unsafe.Pointer(in.Storage))
	out.NodeSelector = *(*map[string]string)(unsafe.Pointer(&in.PodTemplate.Spec.NodeSelector))
	out.Tolerations = *(*[]corev1.Toleration)(unsafe.Pointer(&in.PodTemplate.Spec.Tolerations))
	dbContainer := core_util.GetContainerByName(in.PodTemplate.Spec.Containers, kubedb.KafkaContainerName)
	if dbContainer != nil {
		out.Resources = *(*corev1.ResourceRequirements)(unsafe.Pointer(&dbContainer.Resources))
	}
	return nil
}

func (src *Elasticsearch) ConvertTo(dstRaw rtconv.Hub) error {
	dst := dstRaw.(*v1.Elasticsearch)
	err := Convert_v1alpha2_Elasticsearch_To_v1_Elasticsearch(src, dst, nil)
	if err != nil {
		return err
	}
	if len(dst.Spec.PodTemplate.Spec.Containers) > 0 {
		dst.Spec.PodTemplate.Spec.Containers[0].Name = "elasticsearch" // db container name used in sts
	}
	return nil
}

func (dst *Elasticsearch) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Elasticsearch_To_v1alpha2_Elasticsearch(srcRaw.(*v1.Elasticsearch), dst, nil)
}

func (src *MariaDB) ConvertTo(dstRaw rtconv.Hub) error {
	dst := dstRaw.(*v1.MariaDB)
	err := Convert_v1alpha2_MariaDB_To_v1_MariaDB(src, dst, nil)
	if err != nil {
		return err
	}
	if len(dst.Spec.PodTemplate.Spec.Containers) > 0 {
		dst.Spec.PodTemplate.Spec.Containers[0].Name = "mariadb" // db container name used in sts
	}
	return nil
}

func (dst *MariaDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MariaDB_To_v1alpha2_MariaDB(srcRaw.(*v1.MariaDB), dst, nil)
}

func (src *MongoDB) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_MongoDB_To_v1_MongoDB(src, dstRaw.(*v1.MongoDB), nil)
}

func (dst *MongoDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MongoDB_To_v1alpha2_MongoDB(srcRaw.(*v1.MongoDB), dst, nil)
}

func (src *Memcached) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_Memcached_To_v1_Memcached(src, dstRaw.(*v1.Memcached), nil)
}

func (dst *Memcached) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Memcached_To_v1alpha2_Memcached(srcRaw.(*v1.Memcached), dst, nil)
}

func (src *MySQL) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_MySQL_To_v1_MySQL(src, dstRaw.(*v1.MySQL), nil)
}

func (dst *MySQL) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_MySQL_To_v1alpha2_MySQL(srcRaw.(*v1.MySQL), dst, nil)
}

func (src *PerconaXtraDB) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_PerconaXtraDB_To_v1_PerconaXtraDB(src, dstRaw.(*v1.PerconaXtraDB), nil)
}

func (dst *PerconaXtraDB) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_PerconaXtraDB_To_v1alpha2_PerconaXtraDB(srcRaw.(*v1.PerconaXtraDB), dst, nil)
}

func (src *PgBouncer) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_PgBouncer_To_v1_PgBouncer(src, dstRaw.(*v1.PgBouncer), nil)
}

func (dst *PgBouncer) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_PgBouncer_To_v1alpha2_PgBouncer(srcRaw.(*v1.PgBouncer), dst, nil)
}

func (src *Postgres) ConvertTo(dstRaw rtconv.Hub) error {
	// return Convert_v1alpha2_Postgres_To_v1_Postgres(src, dstRaw.(*v1.Postgres), nil)
	dst := dstRaw.(*v1.Postgres)
	err := Convert_v1alpha2_Postgres_To_v1_Postgres(src, dst, nil)
	if err != nil {
		return err
	}
	if len(dst.Spec.PodTemplate.Spec.Containers) > 0 {
		dst.Spec.PodTemplate.Spec.Containers[0].Name = "postgres" // db container name used in sts
	}
	return nil
}

func (dst *Postgres) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Postgres_To_v1alpha2_Postgres(srcRaw.(*v1.Postgres), dst, nil)
}

func (src *ProxySQL) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_ProxySQL_To_v1_ProxySQL(src, dstRaw.(*v1.ProxySQL), nil)
}

func (dst *ProxySQL) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_ProxySQL_To_v1alpha2_ProxySQL(srcRaw.(*v1.ProxySQL), dst, nil)
}

func (src *RedisSentinel) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_RedisSentinel_To_v1_RedisSentinel(src, dstRaw.(*v1.RedisSentinel), nil)
}

func (dst *RedisSentinel) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_RedisSentinel_To_v1alpha2_RedisSentinel(srcRaw.(*v1.RedisSentinel), dst, nil)
}

func (src *Redis) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_Redis_To_v1_Redis(src, dstRaw.(*v1.Redis), nil)
}

func (dst *Redis) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Redis_To_v1alpha2_Redis(srcRaw.(*v1.Redis), dst, nil)
}

func (src *Kafka) ConvertTo(dstRaw rtconv.Hub) error {
	return Convert_v1alpha2_Kafka_To_v1_Kafka(src, dstRaw.(*v1.Kafka), nil)
}

func (dst *Kafka) ConvertFrom(srcRaw rtconv.Hub) error {
	return Convert_v1_Kafka_To_v1alpha2_Kafka(srcRaw.(*v1.Kafka), dst, nil)
}
