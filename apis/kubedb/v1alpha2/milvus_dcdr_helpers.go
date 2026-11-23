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
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis/kubedb"

	meta_util "kmodules.xyz/client-go/meta"
)

const (
	// MilvusDCDRClusterNameLabel is the OCM spoke cluster (data center) label on the pods of a
	// per data center PetSet.
	MilvusDCDRClusterNameLabel = "open-cluster-management.io/cluster-name"
	// MilvusDCDRActiveLabel marks the proxy (or standalone) pods of the active data center. The
	// spoke agent sets it once its data center is the writable primary and removes it when it
	// fences, so the client Service <db> resolves only to the active data center.
	MilvusDCDRActiveLabel = "dr.kubedb.com/active"

	MilvusDCDRSwitchoverToAnnotation           = "dr.kubedb.com/switchover-to"
	MilvusDCDRSwitchoverAbortAnnotation        = "dr.kubedb.com/switchover-abort"
	MilvusDCDRSwitchoverTimeoutAnnotation      = "dr.kubedb.com/switchover-timeout"
	MilvusDCDRSwitchoverStartedAnnotation      = "dr.kubedb.com/switchover-started"
	MilvusDCDRSwitchoverDrainedAnnotation      = "dr.kubedb.com/switchover-drained"
	MilvusDCDRAcceptFailoverDataLossAnnotation = "dr.kubedb.com/accept-failover-data-loss"
	MilvusDCDRAcceptRebuildDataLossAnnotation  = "dr.kubedb.com/accept-rebuild-data-loss"
	MilvusDCDRRebuildHoldAnnotation            = "dr.kubedb.com/rebuild-hold"
	// MilvusDCDRStoppedDCAnnotation lists data centers whose Milvus PetSets the operator keeps
	// at zero replicas (comma separated). Set by the orchestrator during a rebuild.
	MilvusDCDRStoppedDCAnnotation = "dr.kubedb.com/stopped-dc"
	// MilvusDCDRStoppedCDCAnnotation lists data centers whose CDC the operator keeps at zero
	// replicas, so a seed runs before replication starts.
	MilvusDCDRStoppedCDCAnnotation = "dr.kubedb.com/stopped-cdc"

	// MilvusDCDRDMLChannelNum is rootCoord.dmlChannelNum of every data center. CDC needs the same
	// physical channel count on every cluster.
	MilvusDCDRDMLChannelNum = 16
	// MilvusDCDRReplicationUser is the Milvus user CDC connects to the peer data centers with.
	// A replication connection cannot be changed once it is set, so it uses its own user whose
	// password never rotates, instead of the root user.
	MilvusDCDRReplicationUser = "kubedb_cdc"
	// MilvusDCDRDefaultMaxFailoverLagSeconds is the default of MaxFailoverLagSeconds.
	MilvusDCDRDefaultMaxFailoverLagSeconds = int64(30)

	// MilvusDCDRCDCContainerName is the CDC container of the per data center CDC pod.
	MilvusDCDRCDCContainerName = "cdc"
	// MilvusDCDRAgentContainerName is the fence agent sidecar of the per data center CDC pod.
	MilvusDCDRAgentContainerName = "dcdr-agent"
)

// IsDCDR reports whether this Milvus is deployed for cross data center disaster recovery.
// It is unrelated to IsDistributed, which reports the deployment mode inside a data center.
func (m *Milvus) IsDCDR() bool {
	return m.Spec.Distributed
}

// DCDRDataCenter returns the storage spec of one data center, or nil.
func (m *Milvus) DCDRDataCenter(dc string) *MilvusDataCenterSpec {
	if m.Spec.DisasterRecovery == nil {
		return nil
	}
	for i := range m.Spec.DisasterRecovery.DataCenters {
		if m.Spec.DisasterRecovery.DataCenters[i].Name == dc {
			return &m.Spec.DisasterRecovery.DataCenters[i]
		}
	}
	return nil
}

// DCDRDataCenterNames returns the data centers in spec order.
func (m *Milvus) DCDRDataCenterNames() []string {
	if m.Spec.DisasterRecovery == nil {
		return nil
	}
	names := make([]string, 0, len(m.Spec.DisasterRecovery.DataCenters))
	for _, dc := range m.Spec.DisasterRecovery.DataCenters {
		names = append(names, dc.Name)
	}
	return names
}

// ForDataCenter returns a copy of this Milvus whose meta and object storage are those of one
// data center. Everything else, names included, is unchanged.
func (m *Milvus) ForDataCenter(dc string) (*Milvus, error) {
	spec := m.DCDRDataCenter(dc)
	if spec == nil {
		return nil, fmt.Errorf("data center %q is not listed in spec.disasterRecovery.dataCenters of Milvus %s/%s", dc, m.Namespace, m.Name)
	}
	out := m.DeepCopy()
	out.Spec.MetaStorage = spec.MetaStorage.DeepCopy()
	out.Spec.ObjectStorage = spec.ObjectStorage.DeepCopy()
	return out, nil
}

// DCDRClusterID is the Milvus cluster id (msgChannel.chanNamePrefix.cluster) of one data
// center. It also prefixes the data center's physical channel names, so it must never change.
func (m *Milvus) DCDRClusterID(dc string) string {
	return sanitizeClusterID(m.Name + "-" + dc)
}

func sanitizeClusterID(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

// DCDRPchannels returns the physical channel names of one data center.
func (m *Milvus) DCDRPchannels(dc string) []string {
	out := make([]string, MilvusDCDRDMLChannelNum)
	for i := range out {
		out[i] = fmt.Sprintf("%s-rootcoord-dml_%d", m.DCDRClusterID(dc), i)
	}
	return out
}

// DCDRPetSetName is the PetSet of one role in one data center: <db>-<dc>-<role>[-<group>], or
// <db>-<dc> for a Standalone Milvus.
func (m *Milvus) DCDRPetSetName(dc string, nodeRole MilvusNodeRoleType, groupName string) string {
	name := meta_util.NameWithSuffix(m.OffshootName(), dc)
	if nodeRole != "" {
		name = meta_util.NameWithSuffix(name, string(nodeRole))
	}
	if groupName != "" {
		name = meta_util.NameWithSuffix(name, groupName)
	}
	return name
}

// DCDRCDCPetSetName is the CDC PetSet of one data center.
func (m *Milvus) DCDRCDCPetSetName(dc string) string {
	return m.DCDRPetSetName(dc, MilvusNodeRoleCDC, "")
}

// DCDRServiceName is the Service that reaches one data center's proxies, whatever its role.
func (m *Milvus) DCDRServiceName(dc string) string {
	return meta_util.NameWithSuffix(m.OffshootName(), dc)
}

// DCDRServiceAddress is the cross cluster address of one data center's proxies.
func (m *Milvus) DCDRServiceAddress(dc string) string {
	return fmt.Sprintf("%s.%s.svc.%s:%d", m.DCDRServiceName(dc), m.Namespace, kubedb.KubeSliceDomainSuffix, kubedb.MilvusGrpcPort)
}

// DCDRClientAddress is the cross cluster address clients use; it resolves to the active data center.
func (m *Milvus) DCDRClientAddress() string {
	return fmt.Sprintf("%s.%s.svc.%s:%d", m.ServiceName(), m.Namespace, kubedb.KubeSliceDomainSuffix, kubedb.MilvusGrpcPort)
}

// DCDRConfigSecretName is the rendered milvus.yaml of one data center.
func (m *Milvus) DCDRConfigSecretName(dc string) string {
	return meta_util.NameWithSuffix(m.ConfigSecretName(), dc)
}

// DCDRReplicationAuthSecretName holds the password of MilvusDCDRReplicationUser.
func (m *Milvus) DCDRReplicationAuthSecretName() string {
	return meta_util.NameWithSuffix(m.OffshootName(), "replication-auth")
}

// DCDRAgentConfigSecretName holds the peer list the fence agent of one data center demotes with.
func (m *Milvus) DCDRAgentConfigSecretName(dc string) string {
	return meta_util.NameWithSuffix(m.DCDRCDCPetSetName(dc), "agent")
}

// DCDRMaxFailoverLagSeconds returns MaxFailoverLagSeconds or its default.
func (m *Milvus) DCDRMaxFailoverLagSeconds() int64 {
	if m.Spec.DisasterRecovery != nil && m.Spec.DisasterRecovery.MaxFailoverLagSeconds != nil {
		return *m.Spec.DisasterRecovery.MaxFailoverLagSeconds
	}
	return MilvusDCDRDefaultMaxFailoverLagSeconds
}

// DCDRMaxRebuildLossSeconds returns MaxRebuildLossSeconds or its default, 0.
func (m *Milvus) DCDRMaxRebuildLossSeconds() int64 {
	if m.Spec.DisasterRecovery != nil && m.Spec.DisasterRecovery.MaxRebuildLossSeconds != nil {
		return *m.Spec.DisasterRecovery.MaxRebuildLossSeconds
	}
	return 0
}

// DCDRStopped reports whether the annotation (MilvusDCDRStoppedDCAnnotation or
// MilvusDCDRStoppedCDCAnnotation) lists the data center.
func (m *Milvus) DCDRStopped(annotation, dc string) bool {
	for _, v := range strings.Split(m.GetAnnotations()[annotation], ",") {
		if strings.TrimSpace(v) == dc {
			return true
		}
	}
	return false
}

// DCDRStatusFor returns the status entry of one data center, or nil.
func (s *MilvusDisasterRecoveryStatus) DCDRStatusFor(dc string) *MilvusDCStatus {
	if s == nil {
		return nil
	}
	for i := range s.DataCenters {
		if s.DataCenters[i].ClusterName == dc {
			return &s.DataCenters[i]
		}
	}
	return nil
}
