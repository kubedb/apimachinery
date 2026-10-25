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
	"context"
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultYugabyteDBReplicationFactor is the replication factor of a YugabyteDB that does not set one.
const DefaultYugabyteDBReplicationFactor int32 = 3

func (*YugabyteDB) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralYugabyteDB))
}

func (r *YugabyteDB) ResourceSingular() string {
	return ResourceSingularYugabyteDB
}

func (r *YugabyteDB) ResourceKind() string {
	return ResourceKindYugabyteDB
}

func (r *YugabyteDB) ResourcePlural() string {
	return ResourcePluralYugabyteDB
}

func (r *YugabyteDB) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", r.ResourcePlural(), kubedb.GroupName)
}

func (r *YugabyteDB) OffshootName() string {
	return r.Name
}

func (r *YugabyteDB) ServiceAccountName() string {
	return r.OffshootName()
}

func (r *YugabyteDB) GetAuthSecretName() string {
	if r.Spec.AuthSecret != nil && r.Spec.AuthSecret.Name != "" {
		return r.Spec.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(r.OffshootName(), "auth")
}

// ConfigSecretName is the Secret that holds the flag templates rendered by the operator.
func (r *YugabyteDB) ConfigSecretName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "config")
}

func (r *YugabyteDB) GetPersistentSecrets() []string {
	var secrets []string
	if !r.Spec.DisableSecurity && !IsVirtualAuthSecretReferred(r.Spec.AuthSecret) && r.Spec.AuthSecret != nil && r.Spec.AuthSecret.Name != "" {
		secrets = append(secrets, r.GetAuthSecretName())
	}
	return secrets
}

func (r *YugabyteDB) GetDeletionPolicy() string {
	return string(r.Spec.DeletionPolicy)
}

func (r *YugabyteDB) AsOwner() *meta.OwnerReference {
	return meta.NewControllerRef(r, SchemeGroupVersion.WithKind(r.ResourceKind()))
}

func (r *YugabyteDB) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      r.ResourceFQN(),
		meta_util.InstanceLabelKey:  r.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

// RoleSelectors selects the pods of one role.
func (r *YugabyteDB) RoleSelectors(role YugabyteDBRole) map[string]string {
	return r.OffshootSelectors(map[string]string{kubedb.LabelRole: string(role)})
}

func (r *YugabyteDB) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, r.Labels, override))
}

func (r *YugabyteDB) OffshootLabels() map[string]string {
	return r.offshootLabels(r.OffshootSelectors(), nil)
}

// PodLabels returns the labels of the pods of a role.
func (r *YugabyteDB) PodLabels(role YugabyteDBRole, extraLabels ...map[string]string) map[string]string {
	return r.offshootLabels(meta_util.OverwriteKeys(r.RoleSelectors(role), extraLabels...), r.Node(role).PodTemplate.Labels)
}

func (r *YugabyteDB) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(r.Spec.ServiceTemplates, alias)
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

// Node returns the configuration of a role. It never returns nil.
func (r *YugabyteDB) Node(role YugabyteDBRole) *YugabyteDBNode {
	if r.Spec.Topology != nil {
		switch role {
		case YugabyteDBRoleMaster:
			if r.Spec.Topology.Master != nil {
				return r.Spec.Topology.Master
			}
		case YugabyteDBRoleTServer:
			if r.Spec.Topology.TServer != nil {
				return r.Spec.Topology.TServer
			}
		case YugabyteDBRoleReadReplica:
			if r.Spec.Topology.ReadReplica != nil {
				return &r.Spec.Topology.ReadReplica.YugabyteDBNode
			}
		}
	}
	return &YugabyteDBNode{}
}

// Roles returns the roles that have pods: the masters, the tservers and, when it is configured, the read replicas.
func (r *YugabyteDB) Roles() []YugabyteDBRole {
	roles := []YugabyteDBRole{YugabyteDBRoleMaster, YugabyteDBRoleTServer}
	if r.IsReadReplicaEnabled() {
		roles = append(roles, YugabyteDBRoleReadReplica)
	}
	return roles
}

// IsReadReplicaEnabled reports whether spec.topology.readReplica is set.
func (r *YugabyteDB) IsReadReplicaEnabled() bool {
	return r.Spec.Topology != nil && r.Spec.Topology.ReadReplica != nil
}

// ReadReplicaPlacementUUID is the placement_uuid of the read replica tservers.
func (r *YugabyteDB) ReadReplicaPlacementUUID() string {
	if r.IsReadReplicaEnabled() && r.Spec.Topology.ReadReplica.PlacementUUID != "" {
		return r.Spec.Topology.ReadReplica.PlacementUUID
	}
	return kubedb.YugabyteDBDefaultReadReplicaPlacementUUID
}

// ReadReplicaReplicationFactor is the number of read replica copies of every tablet. It is 0 without read replicas.
func (r *YugabyteDB) ReadReplicaReplicationFactor() int32 {
	if !r.IsReadReplicaEnabled() {
		return 0
	}
	if rf := r.Spec.Topology.ReadReplica.ReplicationFactor; rf != nil {
		return *rf
	}
	return 1
}

// IsReadReplica reports whether a role is the read replica role.
func (role YugabyteDBRole) IsReadReplica() bool {
	return role == YugabyteDBRoleReadReplica
}

// IsTServer reports whether the pods of a role run yb-tserver: the tservers and the read replicas.
func (role YugabyteDBRole) IsTServer() bool {
	return role == YugabyteDBRoleTServer || role == YugabyteDBRoleReadReplica
}

func (r *YugabyteDB) ReplicationFactor() int32 {
	if r.Spec.ReplicationFactor != nil {
		return *r.Spec.ReplicationFactor
	}
	return DefaultYugabyteDBReplicationFactor
}

// Replicas is the number of pods of a role. The masters always equal the replication factor.
func (r *YugabyteDB) Replicas(role YugabyteDBRole) int32 {
	if role == YugabyteDBRoleMaster {
		return r.ReplicationFactor()
	}
	if n := r.Node(role); n.Replicas != nil {
		return *n.Replicas
	}
	if role == YugabyteDBRoleReadReplica {
		return r.ReadReplicaReplicationFactor()
	}
	return r.ReplicationFactor()
}

// DataVolumes is the number of data volumes of each pod of a role.
func (r *YugabyteDB) DataVolumes(role YugabyteDBRole) int32 {
	if n := r.Node(role); n.DataVolumes != nil {
		return *n.DataVolumes
	}
	return 1
}

func (r *YugabyteDB) ContainerName(role YugabyteDBRole) string {
	if role == YugabyteDBRoleMaster {
		return kubedb.YugabyteDBMasterContainerName
	}
	return kubedb.YugabyteDBTServerContainerName
}

func (r *YugabyteDB) PetSetName(role YugabyteDBRole) string {
	return meta_util.NameWithSuffix(r.OffshootName(), string(role))
}

// ServiceName is the client Service. It selects the tservers.
func (r *YugabyteDB) ServiceName() string {
	return r.OffshootName()
}

// GoverningServiceName is the headless Service of both PetSets. Its pods resolve before they are Ready.
func (r *YugabyteDB) GoverningServiceName() string {
	return meta_util.NameWithSuffix(r.ServiceName(), "pods")
}

// PodName returns the name of the pod of a role with the given ordinal.
func (r *YugabyteDB) PodName(role YugabyteDBRole, ordinal int32) string {
	return fmt.Sprintf("%s-%d", r.PetSetName(role), ordinal)
}

// PodFQDN is the stable address a pod registers with YugabyteDB. Blacklists must use exactly this string.
func (r *YugabyteDB) PodFQDN(role YugabyteDBRole, ordinal int32) string {
	return fmt.Sprintf("%s.%s.%s.svc", r.PodName(role, ordinal), r.GoverningServiceName(), r.Namespace)
}

// RPCPort is the port of the RPC server of a role.
func (r *YugabyteDB) RPCPort(role YugabyteDBRole) int32 {
	if role == YugabyteDBRoleMaster {
		return kubedb.YugabyteDBMasterRPCPort
	}
	return kubedb.YugabyteDBTServerRPCPort
}

// UIPort is the port of the web server of a role.
func (r *YugabyteDB) UIPort(role YugabyteDBRole) int32 {
	if role == YugabyteDBRoleMaster {
		return kubedb.YugabyteDBMasterUIPort
	}
	return kubedb.YugabyteDBTServerUIPort
}

// MasterAddresses returns the master RPC addresses, comma separated.
func (r *YugabyteDB) MasterAddresses() string {
	addrs := make([]string, 0, r.Replicas(YugabyteDBRoleMaster))
	for i := int32(0); i < r.Replicas(YugabyteDBRoleMaster); i++ {
		addrs = append(addrs, fmt.Sprintf("%s:%d", r.PodFQDN(YugabyteDBRoleMaster, i), kubedb.YugabyteDBMasterRPCPort))
	}
	return strings.Join(addrs, ",")
}

// MasterAddressList returns the master RPC addresses.
func (r *YugabyteDB) MasterAddressList() []string {
	return strings.Split(r.MasterAddresses(), ",")
}

func (r *YugabyteDB) IsYSQLEnabled() bool {
	return r.Spec.YSQL == nil || r.Spec.YSQL.Enabled
}

func (r *YugabyteDB) IsYCQLEnabled() bool {
	return r.Spec.YCQL != nil && r.Spec.YCQL.Enabled
}

// IsConnectionManagerEnabled reports whether the YSQL Connection Manager is on.
func (r *YugabyteDB) IsConnectionManagerEnabled() bool {
	return r.IsYSQLEnabled() && r.Spec.YSQL != nil && r.Spec.YSQL.ConnectionManager != nil && r.Spec.YSQL.ConnectionManager.Enabled
}

// ConnectionManagerMaxClientConnections is --ysql_conn_mgr_max_client_connections.
func (r *YugabyteDB) ConnectionManagerMaxClientConnections() int32 {
	if r.Spec.YSQL != nil && r.Spec.YSQL.ConnectionManager != nil && r.Spec.YSQL.ConnectionManager.MaxClientConnections != nil {
		return *r.Spec.YSQL.ConnectionManager.MaxClientConnections
	}
	return kubedb.YugabyteDBDefaultConnMgrMaxClientConnection
}

// YSQLInternalPort is the port PostgreSQL listens on inside the pod: 6433 behind the Connection Manager, else 5433.
func (r *YugabyteDB) YSQLInternalPort() int32 {
	if r.IsConnectionManagerEnabled() {
		return kubedb.YugabyteDBYSQLInternalPort
	}
	return kubedb.YugabyteDBYSQLPort
}

// IsWebserverTLS reports whether the web UIs and the metrics endpoints of the masters and tservers serve HTTPS.
func (r *YugabyteDB) IsWebserverTLS() bool {
	return r.Spec.TLS != nil && r.Spec.TLS.Webserver != nil && *r.Spec.TLS.Webserver
}

// GetWebserverScheme is the scheme of the web UIs (7000 and 9000) and of /prometheus-metrics.
func (r *YugabyteDB) GetWebserverScheme() string {
	if r.IsWebserverTLS() {
		return "https"
	}
	return "http"
}

// PITRFor returns the PITR schedule of a database, or nil.
func (r *YugabyteDB) PITRFor(database string) *YugabyteDBPITR {
	for i := range r.Spec.PITR {
		if r.Spec.PITR[i].Database == database {
			return &r.Spec.PITR[i]
		}
	}
	return nil
}

// PITRStatusFor returns the observed state of the PITR schedule of a database, or nil.
func (r *YugabyteDB) PITRStatusFor(database string) *YugabyteDBPITRStatus {
	for i := range r.Status.PITR {
		if r.Status.PITR[i].Database == database {
			return &r.Status.PITR[i]
		}
	}
	return nil
}

// GetKind returns the kind of the schedule, YSQL when unset.
func (p YugabyteDBPITR) GetKind() YugabyteDBPITRKind {
	if p.Kind == "" {
		return YugabyteDBPITRKindYSQL
	}
	return p.Kind
}

func (r *YugabyteDB) IsNodeToNodeTLS() bool {
	return r.Spec.TLS != nil && (r.Spec.TLS.NodeToNode == nil || *r.Spec.TLS.NodeToNode)
}

func (r *YugabyteDB) IsClientToNodeTLS() bool {
	return r.Spec.TLS != nil && (r.Spec.TLS.ClientToNode == nil || *r.Spec.TLS.ClientToNode)
}

func (r *YugabyteDB) IsTLSEnabled() bool {
	return r.IsNodeToNodeTLS() || r.IsClientToNodeTLS()
}

func (r *YugabyteDB) GetConnectionScheme() string {
	if r.IsClientToNodeTLS() {
		return "https"
	}
	return "http"
}

func (r *YugabyteDB) DefaultPodRoleName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "role")
}

func (r *YugabyteDB) DefaultPodRoleBindingName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "rolebinding")
}

// CertificateName is the name of the Secret of a certificate alias, unless the user named it.
func (r *YugabyteDB) CertificateName(alias YugabyteDBCertificateType) string {
	return meta_util.NameWithSuffix(r.Name, fmt.Sprintf("%s-cert", string(alias)))
}

// GetCertSecretName returns the Secret of a certificate alias if any, otherwise the default name.
func (r *YugabyteDB) GetCertSecretName(alias YugabyteDBCertificateType) string {
	if r.Spec.TLS != nil {
		if name, ok := kmapi.GetCertificateSecretName(r.Spec.TLS.Certificates, string(alias)); ok {
			return name
		}
	}
	return r.CertificateName(alias)
}

func (r *YugabyteDB) GetStorageClassName(role YugabyteDBRole) string {
	if s := r.Node(role).Storage; s != nil && s.StorageClassName != nil {
		return *s.StorageClassName
	}
	return ""
}

func (r *YugabyteDB) SetDefaults(kc client.Client) {
	if r.Spec.DeletionPolicy == "" {
		r.Spec.DeletionPolicy = DeletionPolicyDelete
	}
	if r.Spec.ReplicationFactor == nil {
		r.Spec.ReplicationFactor = ptr.To(DefaultYugabyteDBReplicationFactor)
	}
	if r.Spec.YSQL == nil {
		r.Spec.YSQL = &YugabyteDBYSQLSpec{Enabled: true}
	}
	if r.Spec.YCQL == nil {
		r.Spec.YCQL = &YugabyteDBYCQLSpec{}
	}
	if r.Spec.Topology == nil {
		r.Spec.Topology = &YugabyteDBTopology{}
	}
	if r.Spec.Topology.Master == nil {
		r.Spec.Topology.Master = &YugabyteDBNode{}
	}
	if r.Spec.Topology.TServer == nil {
		r.Spec.Topology.TServer = &YugabyteDBNode{}
	}
	if rr := r.Spec.Topology.ReadReplica; rr != nil {
		if rr.PlacementUUID == "" {
			rr.PlacementUUID = kubedb.YugabyteDBDefaultReadReplicaPlacementUUID
		}
		if rr.Replicas == nil {
			rr.Replicas = ptr.To(r.ReadReplicaReplicationFactor())
		}
	}
	r.Spec.Topology.Master.Replicas = ptr.To(r.ReplicationFactor())
	if r.Spec.Topology.TServer.Replicas == nil {
		r.Spec.Topology.TServer.Replicas = ptr.To(r.ReplicationFactor())
	}
	for _, role := range r.Roles() {
		r.setNodeDefaults(r.Node(role))
	}

	if !r.Spec.DisableSecurity {
		if r.Spec.AuthSecret == nil {
			r.Spec.AuthSecret = &SecretReference{}
		}
		if r.Spec.AuthSecret.Kind == "" {
			r.Spec.AuthSecret.Kind = kubedb.ResourceKindSecret
		}
	}

	for i := range r.Spec.PITR {
		if r.Spec.PITR[i].Kind == "" {
			r.Spec.PITR[i].Kind = YugabyteDBPITRKindYSQL
		}
	}

	r.SetHealthCheckerDefaults()
	r.SetTLSDefaults()

	var ybVersion catalog.YugabyteDBVersion
	err := kc.Get(context.TODO(), types.NamespacedName{Name: r.Spec.Version}, &ybVersion)
	if err != nil {
		klog.Errorf("can't get the YugabyteDB version object %s for %s \n", err.Error(), r.Spec.Version)
		return
	}

	for _, role := range r.Roles() {
		node := r.Node(role)
		r.setDefaultContainerSecurityContext(&ybVersion, role, &node.PodTemplate)
		if c := coreutil.GetContainerByName(node.PodTemplate.Spec.Containers, r.ContainerName(role)); c != nil {
			if role == YugabyteDBRoleMaster {
				apis.SetDefaultResourceLimits(&c.Resources, kubedb.DefaultResourcesYugabyteDBMaster)
			} else {
				apis.SetDefaultResourceLimits(&c.Resources, kubedb.DefaultResourcesYugabyteDBTServer)
			}
		}
		apis.SetDefaultResizePolicy(node.PodTemplate.Spec.Containers, node.PodTemplate.Spec.InitContainers)
	}

	if r.Spec.Monitor != nil {
		if r.Spec.Monitor.Prometheus == nil {
			r.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if r.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			r.Spec.Monitor.Prometheus.Exporter.Port = kubedb.YugabyteDBTServerUIPort
		}
		r.Spec.Monitor.SetDefaults()
		if r.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsUser == nil {
			r.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsUser = ybVersion.Spec.SecurityContext.RunAsUser
		}
		if r.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsGroup == nil {
			r.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsGroup = ybVersion.Spec.SecurityContext.RunAsUser
		}
	}
}

func (r *YugabyteDB) setNodeDefaults(n *YugabyteDBNode) {
	if n.StorageType == "" {
		n.StorageType = StorageTypeDurable
	}
	if n.DataVolumes == nil {
		n.DataVolumes = ptr.To(int32(1))
	}
	if n.StorageType == StorageTypeDurable && n.Storage == nil {
		n.Storage = &core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce},
			Resources: core.VolumeResourceRequirements{
				Requests: core.ResourceList{core.ResourceStorage: resource.MustParse("10Gi")},
			},
		}
	}
}

func (r *YugabyteDB) SetTLSDefaults() {
	if r.Spec.TLS == nil || r.Spec.TLS.IssuerRef == nil {
		return
	}
	if r.Spec.TLS.NodeToNode == nil {
		r.Spec.TLS.NodeToNode = ptr.To(true)
	}
	if r.Spec.TLS.ClientToNode == nil {
		r.Spec.TLS.ClientToNode = ptr.To(true)
	}
	r.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(r.Spec.TLS.Certificates, string(YugabyteDBCertificateTypeServer), r.CertificateName(YugabyteDBCertificateTypeServer))
	r.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(r.Spec.TLS.Certificates, string(YugabyteDBCertificateTypeClient), r.CertificateName(YugabyteDBCertificateTypeClient))
}

func (r *YugabyteDB) SetHealthCheckerDefaults() {
	if r.Spec.HealthChecker.PeriodSeconds == nil {
		r.Spec.HealthChecker.PeriodSeconds = ptr.To(int32(20))
	}
	if r.Spec.HealthChecker.TimeoutSeconds == nil {
		r.Spec.HealthChecker.TimeoutSeconds = ptr.To(int32(10))
	}
	if r.Spec.HealthChecker.FailureThreshold == nil {
		r.Spec.HealthChecker.FailureThreshold = ptr.To(int32(3))
	}
}

func (r *YugabyteDB) setDefaultContainerSecurityContext(ybVersion *catalog.YugabyteDBVersion, role YugabyteDBRole, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = ybVersion.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, r.ContainerName(role))
	if container == nil {
		container = &core.Container{Name: r.ContainerName(role)}
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
		container = coreutil.GetContainerByName(podTemplate.Spec.Containers, r.ContainerName(role))
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}
	r.assignDefaultContainerSecurityContext(ybVersion, container.SecurityContext)
}

func (r *YugabyteDB) assignDefaultContainerSecurityContext(ybVersion *catalog.YugabyteDBVersion, rc *core.SecurityContext) {
	if rc.AllowPrivilegeEscalation == nil {
		rc.AllowPrivilegeEscalation = ptr.To(false)
	}
	if rc.Capabilities == nil {
		rc.Capabilities = &core.Capabilities{Drop: []core.Capability{"ALL"}}
	}
	if rc.RunAsNonRoot == nil {
		rc.RunAsNonRoot = ptr.To(true)
	}
	if rc.RunAsUser == nil {
		rc.RunAsUser = ybVersion.Spec.SecurityContext.RunAsUser
	}
	if rc.SeccompProfile == nil {
		rc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

type YugabyteDBApp struct {
	*YugabyteDB
}

func (r *YugabyteDBApp) Name() string {
	return r.YugabyteDB.Name
}

func (r *YugabyteDBApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceKindYugabyteDB))
}

func (r *YugabyteDB) AppBindingMeta() appcat.AppBindingMeta {
	return &YugabyteDBApp{r}
}

type yugabyteDBStatsService struct {
	*YugabyteDB
}

func (r yugabyteDBStatsService) GetNamespace() string {
	return r.YugabyteDB.GetNamespace()
}

func (r yugabyteDBStatsService) ServiceName() string {
	return r.OffshootName() + "-stats"
}

func (r yugabyteDBStatsService) ServiceMonitorName() string {
	return r.ServiceName()
}

func (r yugabyteDBStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return r.OffshootLabels()
}

func (r yugabyteDBStatsService) Path() string {
	return "/prometheus-metrics"
}

func (r yugabyteDBStatsService) Scheme() string {
	return r.GetWebserverScheme()
}

func (r yugabyteDBStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (r *YugabyteDB) StatsService() mona.StatsAccessor {
	return &yugabyteDBStatsService{r}
}

func (r *YugabyteDB) StatsServiceLabels() map[string]string {
	return r.ServiceLabels(StatsServiceAlias, map[string]string{kubedb.LabelRole: kubedb.RoleStats})
}

func (r *YugabyteDB) SidekickLabels(skName string) map[string]string {
	return meta_util.OverwriteKeys(nil, kubedb.CommonSidekickLabels(), map[string]string{
		meta_util.InstanceLabelKey: skName,
		kubedb.SidekickOwnerName:   r.Name,
		kubedb.SidekickOwnerKind:   r.ResourceFQN(),
	})
}

type YugabyteDBBind struct {
	*YugabyteDB
}

var _ DBBindInterface = &YugabyteDBBind{}

// ServiceNames returns the client Service. YugabyteDB has no separate UI Service.
func (r *YugabyteDBBind) ServiceNames() (string, string) {
	return r.ServiceName(), ""
}

// Ports returns the YSQL port, or the YCQL port when YSQL is disabled.
func (r *YugabyteDBBind) Ports() (int, int) {
	if r.IsYSQLEnabled() {
		return kubedb.YugabyteDBYSQLPort, 0
	}
	return kubedb.YugabyteDBYCQLPort, 0
}

func (r *YugabyteDBBind) SecretName() string {
	return r.GetAuthSecretName()
}

func (r *YugabyteDBBind) CertSecretName() string {
	return r.GetCertSecretName(YugabyteDBCertificateTypeClient)
}

var _ Accessor = &YugabyteDB{}

func (r *YugabyteDB) GetObjectMeta() meta.ObjectMeta {
	return r.ObjectMeta
}

func (r *YugabyteDB) GetConditions() []kmapi.Condition {
	return r.Status.Conditions
}

func (r *YugabyteDB) SetCondition(cond kmapi.Condition) {
	r.Status.Conditions = setCondition(r.Status.Conditions, cond)
}

func (r *YugabyteDB) RemoveCondition(typ string) {
	r.Status.Conditions = removeCondition(r.Status.Conditions, typ)
}
