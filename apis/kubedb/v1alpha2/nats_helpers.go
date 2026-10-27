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
	"time"

	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"
	apiutils "kubedb.dev/apimachinery/pkg/utils"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	metautil "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	pslister "kubeops.dev/petset/client/listers/apps/v1"
)

func (NATS) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralNATS))
}

var _ apis.ResourceInfo = &NATS{}

func (n *NATS) ResourceKind() string {
	return ResourceKindNATS
}

func (n *NATS) ResourceSingular() string {
	return ResourceSingularNATS
}

func (n *NATS) ResourcePlural() string {
	return ResourcePluralNATS
}

func (n *NATS) ResourceShortCode() string {
	return ResourceCodeNATS
}

func (n *NATS) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralNATS, kubedb.GroupName)
}

func (n *NATS) AsOwner() *metav1.OwnerReference {
	return metav1.NewControllerRef(n, SchemeGroupVersion.WithKind(ResourceKindNATS))
}

func (n *NATS) GetNameSpacedName() string {
	return n.Namespace + "/" + n.Name
}

// ---------------------------------------------------------------------------
// Naming
// ---------------------------------------------------------------------------

func (n *NATS) OffshootName() string {
	return n.Name
}

func (n *NATS) PetSetName() string {
	return n.OffshootName()
}

func (n *NATS) ServiceAccountName() string {
	return n.OffshootName()
}

// ServiceName is the client (load balanced) service used by applications.
func (n *NATS) ServiceName() string {
	return n.OffshootName()
}

// GoverningServiceName is the headless service shared by all pods (publishNotReadyAddresses: true),
// because routes must resolve before the peers are Ready.
func (n *NATS) GoverningServiceName() string {
	return metautil.NameWithSuffix(n.ServiceName(), kubedb.NATSServiceSuffixPods)
}

func (n *NATS) StatsServiceName() string {
	return metautil.NameWithSuffix(n.ServiceName(), kubedb.NATSServiceSuffixStats)
}

// ConfigSecretName is the Secret holding nats.conf, auth.conf, user.conf and accounts.conf.
func (n *NATS) ConfigSecretName() string {
	return metautil.NameWithSuffix(n.OffshootName(), kubedb.NATSSecretSuffixConfig)
}

// PodName returns the name of the pod at the given PetSet ordinal. It is also the server_name.
func (n *NATS) PodName(ordinal int) string {
	return fmt.Sprintf("%s-%d", n.PetSetName(), ordinal)
}

// PrimaryServiceDNS is the in-cluster DNS name clients should use.
func (n *NATS) PrimaryServiceDNS() string {
	return fmt.Sprintf("%s.%s.svc", n.ServiceName(), n.Namespace)
}

// ClientURL is the URL of the client Service.
func (n *NATS) ClientURL() string {
	return fmt.Sprintf("nats://%s:%d", n.PrimaryServiceDNS(), kubedb.NATSClientPort)
}

// GoverningServiceDNS returns the fully qualified DNS name of a single pod.
func (n *NATS) GoverningServiceDNS(podName string) string {
	return fmt.Sprintf("%s.%s.%s.svc.%s", podName, n.GoverningServiceName(), n.Namespace, apiutils.FindDomain())
}

// RouteURL is the route endpoint of a single pod.
func (n *NATS) RouteURL(podName string) string {
	return fmt.Sprintf("nats://%s:%d", n.GoverningServiceDNS(podName), kubedb.NATSClusterPort)
}

// RouteURLs lists the route of every ordinal < replicas. It is rendered into cluster.routes, so a
// scale operation changes the config Secret and the reloader applies the new list live.
func (n *NATS) RouteURLs(replicas int32) []string {
	urls := make([]string, 0, replicas)
	for i := int32(0); i < replicas; i++ {
		urls = append(urls, n.RouteURL(n.PodName(int(i))))
	}
	return urls
}

// MonitorScheme is https when TLS is on for the monitor port.
func (n *NATS) MonitorScheme() string {
	if n.IsMonitorTLS() {
		return "https"
	}
	return "http"
}

// ---------------------------------------------------------------------------
// Feature switches
// ---------------------------------------------------------------------------

// IsJetStreamEnabled is true unless spec.jetStream.enabled is explicitly false.
func (n *NATS) IsJetStreamEnabled() bool {
	return n.Spec.JetStream == nil || n.Spec.JetStream.Enabled == nil || *n.Spec.JetStream.Enabled
}

// IsEphemeral reports a JetStream file store on an emptyDir.
func (n *NATS) IsEphemeral() bool {
	return n.IsJetStreamEnabled() && n.Spec.JetStream != nil && n.Spec.JetStream.StorageType == StorageTypeEphemeral
}

// AuthMode returns the effective auth mode.
func (n *NATS) AuthMode() NATSAuthMode {
	if n.Spec.Auth == nil || n.Spec.Auth.Mode == "" {
		return NATSAuthModeConfig
	}
	return n.Spec.Auth.Mode
}

func (n *NATS) IsOperatorMode() bool {
	return n.AuthMode() == NATSAuthModeOperator
}

// HasResolverPVC: Operator mode always gets the resolver PVC, whatever the JetStream storage type.
func (n *NATS) HasResolverPVC() bool {
	return n.IsOperatorMode()
}

func (n *NATS) IsTLSEnabled() bool {
	return n.Spec.TLS != nil
}

func tlsChannel(tlsOn bool, sw *bool) bool {
	return tlsOn && (sw == nil || *sw)
}

func (n *NATS) IsClientTLS() bool {
	return tlsChannel(n.IsTLSEnabled(), n.Spec.ClientTLS)
}

func (n *NATS) IsClusterTLS() bool {
	return tlsChannel(n.IsTLSEnabled(), n.Spec.ClusterTLS)
}

func (n *NATS) IsMonitorTLS() bool {
	return tlsChannel(n.IsTLSEnabled(), n.Spec.MonitorTLS)
}

// IsGatewayTLS reports TLS on the gateway port. The gateway uses the server certificate.
func (n *NATS) IsGatewayTLS() bool {
	return tlsChannel(n.IsTLSEnabled(), n.Spec.GatewayTLS)
}

// IsLeafNodeTLS reports TLS on the leafnodes port.
func (n *NATS) IsLeafNodeTLS() bool {
	return tlsChannel(n.IsTLSEnabled(), n.Spec.LeafNodeTLS)
}

// IsGatewayEnabled reports whether the gateway block is rendered: an explicit spec.gateway, or a
// distributed NATS, whose per-DC clusters are always joined by gateways.
func (n *NATS) IsGatewayEnabled() bool {
	return n.Spec.Gateway != nil || n.Spec.Distributed
}

// IsLeafNodeAccepting reports whether the leafnodes listener (hub side) is open.
func (n *NATS) IsLeafNodeAccepting() bool {
	return n.Spec.LeafNodes != nil && n.Spec.LeafNodes.Accept
}

// GatewayName is the gateway name of a non-distributed NATS; it must equal the cluster name. For a
// distributed NATS the name is the DC's cluster name and is derived by the operator.
func (n *NATS) GatewayName() string {
	if n.Spec.Gateway != nil && n.Spec.Gateway.Name != "" {
		return n.Spec.Gateway.Name
	}
	return n.Name
}

// ArbiterReplicas returns the replica count of the arbiter PetSet, defaulting to 1.
func (n *NATS) ArbiterReplicas() int32 {
	if n.Spec.Arbiter == nil || n.Spec.Arbiter.Replicas == nil {
		return kubedb.NATSDefaultArbiterReplicas
	}
	return *n.Spec.Arbiter.Replicas
}

// ---------------------------------------------------------------------------
// Sizing
// ---------------------------------------------------------------------------

// NATSLameDuckDefaults are used when neither spec.lameDuck nor a NATSVersion supplies a value.
const (
	NATSDefaultLameDuckGracePeriod = 10 * time.Second
	NATSDefaultLameDuckDuration    = 30 * time.Second
)

// NATSTerminationGracePeriodSeconds is lameDuck.duration + 30s, with a floor of 60s. The grace period
// sits inside the duration (the server closes clients over duration - gracePeriod), and the extra 30s
// covers the Raft leader hand-off, the JetStream flush and the final shutdown.
func NATSTerminationGracePeriodSeconds(lameDuckDuration time.Duration) int64 {
	secs := int64((lameDuckDuration + kubedb.NATSTerminationGraceExtra) / time.Second)
	if secs < kubedb.NATSMinTerminationGracePeriodSeconds {
		return kubedb.NATSMinTerminationGracePeriodSeconds
	}
	return secs
}

// LameDuckTimes returns the effective lame-duck grace period and duration.
func (n *NATS) LameDuckTimes() (grace, duration time.Duration) {
	grace, duration = NATSDefaultLameDuckGracePeriod, NATSDefaultLameDuckDuration
	if n.Spec.LameDuck != nil {
		if n.Spec.LameDuck.GracePeriod.Duration > 0 {
			grace = n.Spec.LameDuck.GracePeriod.Duration
		}
		if n.Spec.LameDuck.Duration.Duration > 0 {
			duration = n.Spec.LameDuck.Duration.Duration
		}
	}
	return grace, duration
}

// MinTerminationGracePeriodSeconds is the smallest terminationGracePeriodSeconds the pod may have.
func (n *NATS) MinTerminationGracePeriodSeconds() int64 {
	_, duration := n.LameDuckTimes()
	return NATSTerminationGracePeriodSeconds(duration)
}

// TerminationGracePeriodSeconds is the value to put on the pod: the user's, if set, else the minimum.
func (n *NATS) TerminationGracePeriodSeconds() int64 {
	if t := n.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds; t != nil {
		return *t
	}
	return n.MinTerminationGracePeriodSeconds()
}

// JetStreamStorageRequest is the PVC request of the file store, or nil.
func (n *NATS) JetStreamStorageRequest() *resource.Quantity {
	if n.Spec.JetStream == nil || n.Spec.JetStream.Storage == nil {
		return nil
	}
	if q, ok := n.Spec.JetStream.Storage.Resources.Requests[core.ResourceStorage]; ok {
		return &q
	}
	return nil
}

// MaxFileStoreBytes is max_file_store: 90% of the PVC request. The 10% headroom is for the Raft WAL,
// meta snapshots and filesystem overhead; a full volume leaves the server unable to write its own
// Raft state. It is 0 (unset) when JetStream is off or there is no storage request.
func (n *NATS) MaxFileStoreBytes() int64 {
	if !n.IsJetStreamEnabled() {
		return 0
	}
	q := n.JetStreamStorageRequest()
	if q == nil {
		return 0
	}
	return NATSMaxFileStoreBytes(q.Value())
}

// NATSMaxFileStoreBytes applies the 90% rule to a volume size in bytes.
func NATSMaxFileStoreBytes(pvcBytes int64) int64 {
	return pvcBytes / 100 * kubedb.NATSMaxFileStorePercent
}

// MaxMemoryStoreBytes is max_memory_store: memoryStoreSize, or 0 (memory streams disabled).
func (n *NATS) MaxMemoryStoreBytes() int64 {
	if !n.IsJetStreamEnabled() || n.Spec.JetStream == nil || n.Spec.JetStream.MemoryStoreSize == nil {
		return 0
	}
	return n.Spec.JetStream.MemoryStoreSize.Value()
}

// MemoryLimit returns limits.memory of the nats container.
func (n *NATS) MemoryLimit() *resource.Quantity {
	c := coreutil.GetContainerByName(n.Spec.PodTemplate.Spec.Containers, kubedb.NATSContainerName)
	if c == nil {
		return nil
	}
	if q, ok := c.Resources.Limits[core.ResourceMemory]; ok {
		return &q
	}
	return nil
}

// GoMemLimitBytes is GOMEMLIMIT: 90% of limits.memory, or 0 when there is no limit.
func (n *NATS) GoMemLimitBytes() int64 {
	q := n.MemoryLimit()
	if q == nil {
		return 0
	}
	return q.Value() / 100 * kubedb.NATSGoMemLimitPercent
}

// ---------------------------------------------------------------------------
// Labels & selectors
// ---------------------------------------------------------------------------

func (n *NATS) offshootLabels(selector, override map[string]string) map[string]string {
	selector[metautil.ComponentLabelKey] = kubedb.ComponentDatabase
	return metautil.FilterKeys(SchemeGroupVersion.Group, selector, metautil.OverwriteKeys(nil, n.Labels, override))
}

func (n *NATS) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      n.ResourceFQN(),
		metautil.InstanceLabelKey:  n.Name,
		metautil.ManagedByLabelKey: SchemeGroupVersion.Group,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (n *NATS) OffshootLabels() map[string]string {
	return n.offshootLabels(n.OffshootSelectors(), nil)
}

func (n *NATS) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(n.Spec.ServiceTemplates, alias)
	return n.offshootLabels(metautil.OverwriteKeys(n.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (n *NATS) PodLabels(extraLabels ...map[string]string) map[string]string {
	return n.offshootLabels(metautil.OverwriteKeys(n.OffshootSelectors(), extraLabels...), n.Spec.PodTemplate.Labels)
}

func (n *NATS) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return n.offshootLabels(metautil.OverwriteKeys(n.OffshootSelectors(), extraLabels...), n.Spec.PodTemplate.Controller.Labels)
}

func (n *NATS) StatsServiceLabels() map[string]string {
	return n.ServiceLabels(StatsServiceAlias, map[string]string{kubedb.LabelRole: kubedb.RoleStats})
}

// ---------------------------------------------------------------------------
// Secrets
// ---------------------------------------------------------------------------

// GetAuthSecretName is the "admin" credential Secret, <name>-auth by default.
func (n *NATS) GetAuthSecretName() string {
	if n.Spec.AuthSecret != nil && n.Spec.AuthSecret.Name != "" {
		return n.Spec.AuthSecret.Name
	}
	return metautil.NameWithSuffix(n.OffshootName(), "auth")
}

// SysAuthSecretName holds the operator-only kubedb-sys credential. It is not user-configurable.
func (n *NATS) SysAuthSecretName() string {
	return metautil.NameWithSuffix(n.OffshootName(), kubedb.NATSSecretSuffixSysAuth)
}

// ClusterAuthSecretName holds the route credential.
func (n *NATS) ClusterAuthSecretName() string {
	return metautil.NameWithSuffix(n.OffshootName(), kubedb.NATSSecretSuffixClusterAuth)
}

// OperatorSecretName holds the operator seed, signing-key seed and operator.jwt (Operator mode).
// It is never mounted into a pod.
func (n *NATS) OperatorSecretName() string {
	return metautil.NameWithSuffix(n.OffshootName(), kubedb.NATSSecretSuffixOperator)
}

// SysCredsSecretName holds sys.creds (Operator mode).
func (n *NATS) SysCredsSecretName() string {
	return metautil.NameWithSuffix(n.OffshootName(), kubedb.NATSSecretSuffixSysCreds)
}

// AccountSecretName holds the seed, signing key and account.jwt of a managed account (Operator mode).
func (n *NATS) AccountSecretName(account string) string {
	return metautil.NameWithSuffix(n.OffshootName(), fmt.Sprintf(kubedb.NATSSecretSuffixAccountFmt, strings.ToLower(account)))
}

func (n *NATS) GetPersistentSecrets() []string {
	var secrets []string
	if !IsVirtualAuthSecretReferred(n.Spec.AuthSecret) && n.Spec.AuthSecret != nil && n.Spec.AuthSecret.Name != "" {
		secrets = append(secrets, n.GetAuthSecretName())
	}
	return secrets
}

func (n *NATS) GetDeletionPolicy() string {
	return string(n.Spec.DeletionPolicy)
}

// ---------------------------------------------------------------------------
// Certificates
// ---------------------------------------------------------------------------

// CertificateName returns the default secret name holding the certificate of the given alias.
func (n *NATS) CertificateName(alias NATSCertificateAlias) string {
	return metautil.NameWithSuffix(n.Name, fmt.Sprintf("%s-cert", string(alias)))
}

// GetCertSecretName returns the referenced secret name for the certificate alias.
func (n *NATS) GetCertSecretName(alias NATSCertificateAlias) string {
	if n.Spec.TLS == nil {
		return ""
	}
	if _, cert := kmapi.GetCertificate(n.Spec.TLS.Certificates, string(alias)); cert != nil {
		if cert.SecretName != "" {
			return cert.SecretName
		}
		return n.CertificateName(alias)
	}
	return ""
}

// CertificateDNSNames are the SANs of the server certificate: the client Service and every pod through
// the governing Service.
func (n *NATS) CertificateDNSNames() []string {
	domain := apiutils.FindDomain()
	names := []string{
		n.ServiceName(),
		fmt.Sprintf("%s.%s", n.ServiceName(), n.Namespace),
		fmt.Sprintf("%s.%s.svc", n.ServiceName(), n.Namespace),
		fmt.Sprintf("%s.%s.svc.%s", n.ServiceName(), n.Namespace, domain),
		fmt.Sprintf("%s.%s", n.GoverningServiceName(), n.Namespace),
		fmt.Sprintf("%s.%s.svc", n.GoverningServiceName(), n.Namespace),
		fmt.Sprintf("%s.%s.svc.%s", n.GoverningServiceName(), n.Namespace, domain),
		fmt.Sprintf("*.%s.%s.svc", n.GoverningServiceName(), n.Namespace),
		fmt.Sprintf("*.%s.%s.svc.%s", n.GoverningServiceName(), n.Namespace, domain),
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// ---------------------------------------------------------------------------
// AppBinding & monitoring
// ---------------------------------------------------------------------------

type natsApp struct {
	*NATS
}

func (r natsApp) Name() string {
	return r.NATS.Name
}

func (r natsApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", SchemeGroupVersion.Group, ResourceSingularNATS))
}

func (n NATS) AppBindingMeta() appcat.AppBindingMeta {
	return &natsApp{&n}
}

type natsStatsService struct {
	*NATS
}

func (s natsStatsService) GetNamespace() string {
	return s.NATS.GetNamespace()
}

func (s natsStatsService) ServiceName() string {
	return s.NATS.StatsServiceName()
}

func (s natsStatsService) ServiceMonitorName() string {
	return s.ServiceName()
}

func (s natsStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return s.OffshootLabels()
}

func (s natsStatsService) Path() string {
	return kubedb.DefaultStatsPath
}

func (s natsStatsService) Scheme() string {
	sc := promapi.SchemeHTTP
	return sc.String()
}

func (s natsStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (n *NATS) StatsService() mona.StatsAccessor {
	return &natsStatsService{n}
}

func (n *NATS) ReplicasAreReady(lister pslister.PetSetLister) (bool, string, error) {
	// Desire number of petSets
	expectedItems := 1
	return checkReplicasOfPetSet(lister.PetSets(n.Namespace), labels.SelectorFromSet(n.OffshootLabels()), expectedItems)
}

// ---------------------------------------------------------------------------
// Defaulting
// ---------------------------------------------------------------------------

func (n *NATS) SetHealthCheckerDefaults() {
	if n.Spec.HealthChecker.PeriodSeconds == nil {
		n.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(20)
	}
	if n.Spec.HealthChecker.TimeoutSeconds == nil {
		n.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if n.Spec.HealthChecker.FailureThreshold == nil {
		n.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func natsStorageSpec(size resource.Quantity) *core.PersistentVolumeClaimSpec {
	return &core.PersistentVolumeClaimSpec{
		AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce},
		Resources: core.VolumeResourceRequirements{
			Requests: core.ResourceList{core.ResourceStorage: size},
		},
	}
}

func (n *NATS) SetDefaults(natsVersion *catalog.NATSVersion) {
	if n == nil {
		return
	}

	if n.Spec.Replicas == nil {
		n.Spec.Replicas = pointer.Int32P(kubedb.NATSDefaultReplicas)
	}
	if n.Spec.DeletionPolicy == "" {
		n.Spec.DeletionPolicy = DeletionPolicyDelete
	}

	if n.Spec.JetStream == nil {
		n.Spec.JetStream = &NATSJetStreamSpec{}
	}
	if n.Spec.JetStream.Enabled == nil {
		n.Spec.JetStream.Enabled = pointer.BoolP(true)
	}
	if *n.Spec.JetStream.Enabled {
		if n.Spec.JetStream.StorageType == "" {
			n.Spec.JetStream.StorageType = StorageTypeDurable
		}
		if n.Spec.JetStream.Storage == nil {
			n.Spec.JetStream.Storage = natsStorageSpec(kubedb.NATSDefaultJetStreamStorage)
		}
	}

	if n.Spec.Auth == nil {
		n.Spec.Auth = &NATSAuthSpec{}
	}
	if n.Spec.Auth.Mode == "" {
		n.Spec.Auth.Mode = NATSAuthModeConfig
	}
	if n.Spec.Auth.Mode == NATSAuthModeOperator {
		if n.Spec.Auth.Operator == nil {
			n.Spec.Auth.Operator = &NATSOperatorSpec{}
		}
		if n.Spec.Auth.Operator.ResolverStorage == nil {
			n.Spec.Auth.Operator.ResolverStorage = natsStorageSpec(kubedb.NATSDefaultResolverStorage)
		}
	}

	if n.Spec.Gateway != nil && !n.Spec.Distributed && n.Spec.Gateway.Name == "" {
		n.Spec.Gateway.Name = n.Name
	}
	if n.Spec.Distributed {
		if n.Spec.Arbiter == nil {
			n.Spec.Arbiter = &NATSArbiterSpec{}
		}
		if n.Spec.Arbiter.Replicas == nil {
			n.Spec.Arbiter.Replicas = pointer.Int32P(kubedb.NATSDefaultArbiterReplicas)
		}
		if n.Spec.Arbiter.Storage == nil {
			n.Spec.Arbiter.Storage = natsStorageSpec(kubedb.NATSDefaultArbiterStorage)
		}
	}

	if n.Spec.AuthSecret == nil {
		n.Spec.AuthSecret = &SecretReference{}
	}
	if n.Spec.AuthSecret.Kind == "" {
		n.Spec.AuthSecret.Kind = kubedb.ResourceKindSecret
	}

	// Lame-duck timing comes from the catalog; the pod needs duration + 30s to drain.
	if n.Spec.LameDuck == nil {
		n.Spec.LameDuck = &NATSLameDuckSpec{}
	}
	if n.Spec.LameDuck.Duration.Duration == 0 {
		n.Spec.LameDuck.Duration.Duration = NATSDefaultLameDuckDuration
		if natsVersion != nil && natsVersion.Spec.LameDuck.Duration.Duration > 0 {
			n.Spec.LameDuck.Duration.Duration = natsVersion.Spec.LameDuck.Duration.Duration
		}
	}
	if n.Spec.LameDuck.GracePeriod.Duration == 0 {
		n.Spec.LameDuck.GracePeriod.Duration = NATSDefaultLameDuckGracePeriod
		if natsVersion != nil && natsVersion.Spec.LameDuck.GracePeriod.Duration > 0 {
			n.Spec.LameDuck.GracePeriod.Duration = natsVersion.Spec.LameDuck.GracePeriod.Duration
		}
	}
	if n.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds == nil {
		n.Spec.PodTemplate.Spec.TerminationGracePeriodSeconds = pointer.Int64P(n.MinTerminationGracePeriodSeconds())
	}

	if n.Spec.PodTemplate.Spec.ServiceAccountName == "" {
		n.Spec.PodTemplate.Spec.ServiceAccountName = n.OffshootName()
	}

	n.SetTLSDefaults()
	n.SetHealthCheckerDefaults()
	n.setDefaultContainerSecurityContext(natsVersion, &n.Spec.PodTemplate)
	n.setDefaultContainerResourceLimits(&n.Spec.PodTemplate)

	if n.Spec.Monitor != nil {
		if n.Spec.Monitor.Prometheus == nil {
			n.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if n.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			n.Spec.Monitor.Prometheus.Exporter.Port = kubedb.NATSExporterPort
		}
		n.Spec.Monitor.SetDefaults()
		if natsVersion != nil {
			if n.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsUser == nil {
				n.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsUser = natsVersion.Spec.SecurityContext.RunAsUser
			}
			if n.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsGroup == nil {
				n.Spec.Monitor.Prometheus.Exporter.SecurityContext.RunAsGroup = natsVersion.Spec.SecurityContext.RunAsUser
			}
		}
	}
}

func (n *NATS) SetTLSDefaults() {
	if n.Spec.TLS == nil || n.Spec.TLS.IssuerRef == nil {
		return
	}

	for _, alias := range []NATSCertificateAlias{NATSServerCert, NATSClientCert} {
		defaultOrg := []string{kubedb.KubeDBOrganization}
		defaultOrgUnit := []string{string(alias)}
		if _, cert := kmapi.GetCertificate(n.Spec.TLS.Certificates, string(alias)); cert != nil && cert.Subject != nil {
			if cert.Subject.Organizations != nil {
				defaultOrg = cert.Subject.Organizations
			}
			if cert.Subject.OrganizationalUnits != nil {
				defaultOrgUnit = cert.Subject.OrganizationalUnits
			}
		}

		spec := kmapi.CertificateSpec{
			Alias:      string(alias),
			SecretName: n.CertificateName(alias),
			Subject: &kmapi.X509Subject{
				Organizations:       defaultOrg,
				OrganizationalUnits: defaultOrgUnit,
			},
		}
		// Only the server certificate is presented on a listening socket.
		if alias == NATSServerCert {
			spec.DNSNames = n.CertificateDNSNames()
		}
		n.Spec.TLS.Certificates = kmapi.SetMissingSpecForCertificate(n.Spec.TLS.Certificates, spec)
	}

	// All channels use TLS by default.
	if n.Spec.ClientTLS == nil {
		n.Spec.ClientTLS = pointer.BoolP(true)
	}
	if n.Spec.ClusterTLS == nil {
		n.Spec.ClusterTLS = pointer.BoolP(true)
	}
	if n.Spec.MonitorTLS == nil {
		n.Spec.MonitorTLS = pointer.BoolP(true)
	}
	if n.Spec.GatewayTLS == nil {
		n.Spec.GatewayTLS = pointer.BoolP(true)
	}
	if n.Spec.LeafNodeTLS == nil {
		n.Spec.LeafNodeTLS = pointer.BoolP(true)
	}
}

func (n *NATS) setDefaultContainerSecurityContext(natsVersion *catalog.NATSVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = natsRunAsUser(natsVersion)
	}

	for _, name := range []string{kubedb.NATSContainerName, kubedb.NATSReloaderContainerName} {
		container := coreutil.GetContainerByName(podTemplate.Spec.Containers, name)
		if container == nil {
			container = &core.Container{Name: name}
		}
		if container.SecurityContext == nil {
			container.SecurityContext = &core.SecurityContext{}
		}
		n.assignDefaultContainerSecurityContext(natsVersion, container.SecurityContext)
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
	}
}

func (n *NATS) assignDefaultContainerSecurityContext(natsVersion *catalog.NATSVersion, sc *core.SecurityContext) {
	if sc.AllowPrivilegeEscalation == nil {
		sc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if sc.Capabilities == nil {
		sc.Capabilities = &core.Capabilities{
			Drop: []core.Capability{"ALL"},
		}
	}
	if sc.RunAsNonRoot == nil {
		sc.RunAsNonRoot = pointer.BoolP(true)
	}
	if sc.RunAsUser == nil {
		sc.RunAsUser = natsRunAsUser(natsVersion)
	}
	if sc.RunAsGroup == nil {
		sc.RunAsGroup = natsRunAsUser(natsVersion)
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

// natsRunAsUser falls back to the uid of the upstream nats image when the catalog does not pin one.
func natsRunAsUser(natsVersion *catalog.NATSVersion) *int64 {
	if natsVersion != nil && natsVersion.Spec.SecurityContext.RunAsUser != nil {
		return natsVersion.Spec.SecurityContext.RunAsUser
	}
	return pointer.Int64P(kubedb.NATSUserID)
}

func (n *NATS) setDefaultContainerResourceLimits(podTemplate *ofst.PodTemplateSpec) {
	dbContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.NATSContainerName)
	if dbContainer == nil {
		dbContainer = &core.Container{Name: kubedb.NATSContainerName}
	}
	apis.SetDefaultResourceLimits(&dbContainer.Resources, kubedb.NATSDefaultResources)
	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *dbContainer)

	reloader := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.NATSReloaderContainerName)
	if reloader == nil {
		reloader = &core.Container{Name: kubedb.NATSReloaderContainerName}
	}
	apis.SetDefaultResourceLimits(&reloader.Resources, kubedb.NATSSidecarDefaultResources)
	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *reloader)

	apis.SetDefaultResizePolicy(podTemplate.Spec.Containers, podTemplate.Spec.InitContainers)
}

// ---------------------------------------------------------------------------
// User configuration: operator-owned keys
// ---------------------------------------------------------------------------

// The only keys spec.configuration may carry (in inline, or in the referenced Secret).
// nats.conf is rendered as extra top-level keys in user.conf; accounts.conf holds extra accounts and is
// included inside the operator's accounts { } block, so it exists in Config auth mode only.
var NATSAllowedConfigurationKeys = []string{kubedb.NATSConfKey, kubedb.NATSAccountsConfKey}

// NATSReservedAccounts cannot be defined in accounts.conf.
var NATSReservedAccounts = []string{kubedb.NATSAccountSys, kubedb.NATSAccountApp}

var natsOwnedKeys = map[string]struct{}{
	// design Section 4: the operator-owned list
	"server_name":    {},
	"listen":         {},
	"port":           {},
	"host":           {},
	"pid_file":       {},
	"cluster":        {},
	"gateway":        {},
	"leafnodes":      {},
	"jetstream":      {},
	"store_dir":      {},
	"system_account": {},
	"accounts":       {},
	"authorization":  {},
	"tls":            {},
	"include":        {},
	// Not in the design list, but written by the operator in some mode, and a later duplicate wins
	// silently: pod.conf (zone tags), disableSecurity, Operator mode.
	"server_tags":      {},
	"no_auth_user":     {},
	"operator":         {},
	"resolver":         {},
	"resolver_preload": {},
}

// natsOwnedKeyPrefixes: http*, https*, lame_duck_*.
var natsOwnedKeyPrefixes = []string{"http", "lame_duck_"}

// IsNATSOperatorOwnedKey reports whether a top-level nats.conf key is written by the operator.
// NATS lowercases keys, so the comparison is case-insensitive.
func IsNATSOperatorOwnedKey(key string) bool {
	k := strings.ToLower(key)
	if _, ok := natsOwnedKeys[k]; ok {
		return true
	}
	for _, p := range natsOwnedKeyPrefixes {
		if strings.HasPrefix(k, p) {
			return true
		}
	}
	return false
}

// NATSTopLevelKeys returns the top-level keys of a NATS configuration text, in order. It understands
// comments (# and //), quoted strings, and nested {} and [] blocks, with "key: value", "key = value"
// and "key value" forms. It is a safety scan for the webhook and the renderer, not a full parser.
func NATSTopLevelKeys(conf string) ([]string, error) {
	var keys []string
	i, n := 0, len(conf)

	skipComment := func() bool {
		if i < n && conf[i] == '#' || (i+1 < n && conf[i] == '/' && conf[i+1] == '/') {
			for i < n && conf[i] != '\n' {
				i++
			}
			return true
		}
		return false
	}
	skipQuoted := func() error {
		q := conf[i]
		i++
		for i < n && conf[i] != q {
			i++
		}
		if i >= n {
			return fmt.Errorf("unterminated string in configuration")
		}
		i++
		return nil
	}
	skipBlock := func() error {
		depth := 0
		for i < n {
			c := conf[i]
			switch {
			case c == '#' || (c == '/' && i+1 < n && conf[i+1] == '/'):
				skipComment()
				continue
			case c == '"' || c == '\'':
				if err := skipQuoted(); err != nil {
					return err
				}
				continue
			case c == '{' || c == '[' || c == '(':
				depth++
			case c == '}' || c == ']' || c == ')':
				depth--
				if depth == 0 {
					i++
					return nil
				}
			}
			i++
		}
		return fmt.Errorf("unbalanced braces in configuration")
	}

	for i < n {
		c := conf[i]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == ',' || c == ';' {
			i++
			continue
		}
		if skipComment() {
			continue
		}

		// key
		var key string
		if c == '"' || c == '\'' {
			start := i + 1
			if err := skipQuoted(); err != nil {
				return nil, err
			}
			key = conf[start : i-1]
		} else {
			start := i
			for i < n && !strings.ContainsRune(" \t\r\n:={[,;", rune(conf[i])) {
				i++
			}
			key = conf[start:i]
			if key == "" {
				return nil, fmt.Errorf("unexpected character %q in configuration", string(c))
			}
		}
		keys = append(keys, key)

		// separator
		for i < n && (conf[i] == ' ' || conf[i] == '\t') {
			i++
		}
		if i < n && (conf[i] == ':' || conf[i] == '=') {
			i++
			for i < n && (conf[i] == ' ' || conf[i] == '\t') {
				i++
			}
		}

		// value
		switch {
		case i >= n || conf[i] == '\n' || conf[i] == '\r':
		case conf[i] == '{' || conf[i] == '[':
			if err := skipBlock(); err != nil {
				return nil, err
			}
		case conf[i] == '"' || conf[i] == '\'':
			if err := skipQuoted(); err != nil {
				return nil, err
			}
		default:
			for i < n && conf[i] != '\n' && conf[i] != '\r' && conf[i] != ';' {
				if skipComment() {
					break
				}
				i++
			}
		}
	}
	return keys, nil
}

// ValidateNATSUserConfig checks one key of spec.configuration. It is shared by the webhook and the
// renderer: duplicate keys in a NATS config are not an error, the last one silently wins, and
// user.conf is included last, so an operator-owned key in user content would override the operator's.
func ValidateNATSUserConfig(fileKey, content string, authMode NATSAuthMode) error {
	switch fileKey {
	case kubedb.NATSConfKey:
		keys, err := NATSTopLevelKeys(content)
		if err != nil {
			return fmt.Errorf("%s: %w", fileKey, err)
		}
		var bad []string
		for _, k := range keys {
			if IsNATSOperatorOwnedKey(k) {
				bad = append(bad, k)
			}
		}
		if len(bad) > 0 {
			return fmt.Errorf("%s sets operator-owned keys: %s", fileKey, strings.Join(bad, ", "))
		}
	case kubedb.NATSAccountsConfKey:
		if authMode == NATSAuthModeOperator {
			return fmt.Errorf("%s is not allowed in Operator auth mode, accounts live in the account JWTs", fileKey)
		}
		keys, err := NATSTopLevelKeys(content)
		if err != nil {
			return fmt.Errorf("%s: %w", fileKey, err)
		}
		for _, k := range keys {
			for _, r := range NATSReservedAccounts {
				if k == r {
					return fmt.Errorf("%s: account name %q is reserved", fileKey, k)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported configuration key %q, supported keys are %s", fileKey, strings.Join(NATSAllowedConfigurationKeys, ", "))
	}
	return nil
}

// ValidateNATSServerConfig checks the typed server limits and timeouts.
func ValidateNATSServerConfig(sc *NATSServerConfig, path *field.Path) field.ErrorList {
	var allErr field.ErrorList
	if sc == nil {
		return nil
	}
	positive32 := func(name string, v *int32) {
		if v != nil && *v <= 0 {
			allErr = append(allErr, field.Invalid(path.Child(name), *v, fmt.Sprintf("%s must be positive", name)))
		}
	}
	positive32("maxConnections", sc.MaxConnections)
	positive32("maxPayload", sc.MaxPayload)
	positive32("maxSubscriptions", sc.MaxSubscriptions)
	positive32("maxControlLine", sc.MaxControlLine)
	if sc.MaxPending != nil && *sc.MaxPending <= 0 {
		allErr = append(allErr, field.Invalid(path.Child("maxPending"), *sc.MaxPending, "maxPending must be positive"))
	}
	if sc.MaxPayload != nil && sc.MaxPending != nil && int64(*sc.MaxPayload) > *sc.MaxPending {
		allErr = append(allErr, field.Invalid(path.Child("maxPayload"), *sc.MaxPayload, "maxPayload must be <= maxPending"))
	}
	if sc.WriteDeadline != nil && sc.WriteDeadline.Duration <= 0 {
		allErr = append(allErr, field.Invalid(path.Child("writeDeadline"), sc.WriteDeadline.Duration.String(), "writeDeadline must be positive"))
	}
	if sc.AuthTimeout != nil && sc.AuthTimeout.Duration <= 0 {
		allErr = append(allErr, field.Invalid(path.Child("authTimeout"), sc.AuthTimeout.Duration.String(), "authTimeout must be positive"))
	}
	return allErr
}

// ValidateNATSLameDuck checks gracePeriod < duration and duration >= 30s, the server minimum.
func ValidateNATSLameDuck(ld *NATSLameDuckSpec, path *field.Path) field.ErrorList {
	var allErr field.ErrorList
	if ld == nil {
		return nil
	}
	if ld.Duration.Duration != 0 && ld.Duration.Duration < kubedb.NATSMinLameDuckDuration {
		allErr = append(allErr, field.Invalid(path.Child("duration"), ld.Duration.Duration.String(),
			"lameDuck.duration must be at least 30s, the server minimum"))
	}
	if ld.GracePeriod.Duration < 0 {
		allErr = append(allErr, field.Invalid(path.Child("gracePeriod"), ld.GracePeriod.Duration.String(),
			"lameDuck.gracePeriod can not be negative"))
	}
	if ld.Duration.Duration != 0 && ld.GracePeriod.Duration >= ld.Duration.Duration {
		allErr = append(allErr, field.Invalid(path.Child("gracePeriod"), ld.GracePeriod.Duration.String(),
			"lameDuck.gracePeriod must be less than lameDuck.duration"))
	}
	return allErr
}
