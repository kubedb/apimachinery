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

	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	core "k8s.io/api/core/v1"
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

// DurableWorkflowCertificateType is a certificate alias.
// +kubebuilder:validation:Enum=ca;internode;frontend;client
type DurableWorkflowCertificateType string

const (
	DurableWorkflowCertificateTypeCA        DurableWorkflowCertificateType = "ca"
	DurableWorkflowCertificateTypeInternode DurableWorkflowCertificateType = "internode"
	DurableWorkflowCertificateTypeFrontend  DurableWorkflowCertificateType = "frontend"
	DurableWorkflowCertificateTypeClient    DurableWorkflowCertificateType = "client"
)

func (*DurableWorkflow) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralDurableWorkflows))
}

var _ apis.ResourceInfo = &DurableWorkflow{}

func (r *DurableWorkflow) ResourceShortCode() string {
	return ResourceCodeDurableWorkflow
}

func (r *DurableWorkflow) ResourceKind() string {
	return ResourceKindDurableWorkflow
}

func (r *DurableWorkflow) ResourceSingular() string {
	return ResourceSingularDurableWorkflow
}

func (r *DurableWorkflow) ResourcePlural() string {
	return ResourcePluralDurableWorkflows
}

func (r *DurableWorkflow) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", r.ResourcePlural(), kubedb.GroupName)
}

func (r *DurableWorkflow) OffshootName() string {
	return r.Name
}

func (r *DurableWorkflow) ServiceAccountName() string {
	return r.OffshootName()
}

func (r *DurableWorkflow) AsOwner() *meta.OwnerReference {
	return meta.NewControllerRef(r, SchemeGroupVersion.WithKind(r.ResourceKind()))
}

func (r *DurableWorkflow) GetDeletionPolicy() string {
	return string(r.Spec.DeletionPolicy)
}

func (r *DurableWorkflow) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      r.ResourceFQN(),
		meta_util.InstanceLabelKey:  r.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (r *DurableWorkflow) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDurableWorkflow
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, r.Labels, override))
}

func (r *DurableWorkflow) OffshootLabels() map[string]string {
	return r.offshootLabels(r.OffshootSelectors(), nil)
}

// RoleSelectors selects the pods of one role.
func (r *DurableWorkflow) RoleSelectors(role string) map[string]string {
	return r.OffshootSelectors(map[string]string{kubedb.DurableWorkflowRoleLabelKey: role})
}

// RoleLabels are the labels of the resources of one role.
func (r *DurableWorkflow) RoleLabels(role string, override map[string]string) map[string]string {
	return r.offshootLabels(r.RoleSelectors(role), override)
}

// PetSetName is the name of the PetSet of a role.
func (r *DurableWorkflow) PetSetName(role string) string {
	return meta_util.NameWithSuffix(r.OffshootName(), role)
}

// GoverningServiceName is the headless Service of a role.
func (r *DurableWorkflow) GoverningServiceName(role string) string {
	return meta_util.NameWithSuffix(r.PetSetName(role), "pods")
}

// ServiceName is the Service that exposes the frontend.
func (r *DurableWorkflow) ServiceName() string {
	return r.OffshootName()
}

// InternalFrontendServiceName is the Service that exposes the Temporal internal-frontend.
func (r *DurableWorkflow) InternalFrontendServiceName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), kubedb.DurableWorkflowRoleInternalFrontend)
}

func (r *DurableWorkflow) UIServiceName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), kubedb.DurableWorkflowRoleUI)
}

func (r *DurableWorkflow) ConfigSecretName() string {
	uid := string(r.UID)
	return meta_util.NameWithSuffix(r.OffshootName(), uid[len(uid)-6:])
}

func (r *DurableWorkflow) DynamicConfigName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "dynamic-config")
}

func (r *DurableWorkflow) DefaultPodRoleName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "role")
}

func (r *DurableWorkflow) DefaultPodRoleBindingName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "rolebinding")
}

// StoreAuthSecretName is the auth secret of the default or visibility store.
func (r *DurableWorkflow) StoreAuthSecretName(visibility bool) string {
	store := r.Spec.Persistence.Default
	suffix := "default-auth"
	if visibility {
		suffix = "visibility-auth"
		if r.Spec.Persistence.Visibility != nil {
			store = *r.Spec.Persistence.Visibility
		}
	}
	if store.AuthSecret != nil && store.AuthSecret.Name != "" {
		return store.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(r.OffshootName(), suffix)
}

// StoreDatabaseName is the database, keyspace or index of the default or visibility store. The visibility store
// shares the server of the default store when none is declared, and always gets a database of its own.
func (r *DurableWorkflow) StoreDatabaseName(visibility bool) string {
	if visibility {
		if v := r.Spec.Persistence.Visibility; v != nil && v.DatabaseName != "" {
			return v.DatabaseName
		}
		return r.Name + "_visibility"
	}
	if d := r.Spec.Persistence.Default.DatabaseName; d != "" {
		return d
	}
	return r.Name
}

// ShouldCreateDatabase tells if the operator creates the database and user of a store.
func (s DurableWorkflowStore) ShouldCreateDatabase() bool {
	return s.CreateDatabase == nil || *s.CreateDatabase
}

// VisibilityStore returns the visibility store, which is the default store's server if none is set.
func (r *DurableWorkflow) VisibilityStore() DurableWorkflowStore {
	if r.Spec.Persistence.Visibility != nil {
		return *r.Spec.Persistence.Visibility
	}
	return r.DefaultStore()
}

// Replicas returns the replicas of a role.
func (r *DurableWorkflow) Replicas(role string) int32 {
	switch role {
	case kubedb.DurableWorkflowRoleUI:
		if r.Spec.UI != nil && r.Spec.UI.Replicas != nil {
			return *r.Spec.UI.Replicas
		}
		return 1
	case kubedb.DurableWorkflowRoleAdminTools:
		return 1
	}
	if n := r.node(role); n != nil && n.Replicas != nil {
		return *n.Replicas
	}
	return 1
}

// PodTemplate returns the pod template of a role.
func (r *DurableWorkflow) PodTemplate(role string) *ofst.PodTemplateSpec {
	switch role {
	case kubedb.DurableWorkflowRoleUI:
		if r.Spec.UI != nil {
			return r.Spec.UI.PodTemplate
		}
		return nil
	case kubedb.DurableWorkflowRoleAdminTools:
		if r.Spec.AdminTools != nil {
			return r.Spec.AdminTools.PodTemplate
		}
		return nil
	}
	if n := r.node(role); n != nil {
		return n.PodTemplate
	}
	return nil
}

// HasUI and HasAdminTools tell if the optional pods are wanted.
func (r *DurableWorkflow) HasUI() bool { return r.Spec.UI != nil && r.Spec.UI.Enabled }

func (r *DurableWorkflow) HasAdminTools() bool {
	return r.Spec.AdminTools != nil && r.Spec.AdminTools.Enabled
}

// PodRoles are the roles that run pods: the server roles in rollout order, then the web UI and the admin tools if enabled.
func (r *DurableWorkflow) PodRoles(v *catalog.DurableWorkflowVersion) []string {
	roles := r.Roles(v)
	if r.HasUI() && v.Spec.UI != nil {
		roles = append(roles, kubedb.DurableWorkflowRoleUI)
	}
	if r.HasAdminTools() {
		roles = append(roles, kubedb.DurableWorkflowRoleAdminTools)
	}
	return roles
}

func (r *DurableWorkflow) node(role string) *DurableWorkflowNode {
	t := r.Spec.Topology
	if t == nil {
		return nil
	}
	switch role {
	case kubedb.DurableWorkflowRoleFrontend:
		return t.Frontend
	case kubedb.DurableWorkflowRoleHistory:
		return t.History
	case kubedb.DurableWorkflowRoleMatching:
		return t.Matching
	case kubedb.DurableWorkflowRoleWorker:
		return t.Worker
	case kubedb.DurableWorkflowRoleInternalFrontend:
		return t.InternalFrontend
	}
	return nil
}

// Roles returns the enabled server roles, in rollout order.
func (r *DurableWorkflow) Roles(v *catalog.DurableWorkflowVersion) []string {
	roles := []string{
		kubedb.DurableWorkflowRoleHistory,
		kubedb.DurableWorkflowRoleMatching,
	}
	if v.IsTemporal() && r.Spec.Topology != nil && r.Spec.Topology.InternalFrontend != nil {
		roles = append(roles, kubedb.DurableWorkflowRoleInternalFrontend)
	}
	return append(roles, kubedb.DurableWorkflowRoleFrontend, kubedb.DurableWorkflowRoleWorker)
}

// IsHighlyAvailable is true if every enabled role (and the UI) has at least two replicas.
func (r *DurableWorkflow) IsHighlyAvailable(v *catalog.DurableWorkflowVersion) bool {
	for _, role := range r.Roles(v) {
		if r.Replicas(role) < 2 {
			return false
		}
	}
	if r.HasUI() && r.Replicas(kubedb.DurableWorkflowRoleUI) < 2 {
		return false
	}
	return true
}

// GetPersistentSecrets returns the secrets that outlive the cluster under the Halt policy: the credentials of the
// application users the operator created in the persistence stores.
func (r *DurableWorkflow) GetPersistentSecrets() []string {
	var secrets []string
	if r.Spec.Persistence.Default.ShouldCreateDatabase() {
		secrets = append(secrets, r.StoreAuthSecretName(false))
	}
	if v := r.Spec.Persistence.Visibility; v != nil && v.ShouldCreateDatabase() {
		secrets = append(secrets, r.StoreAuthSecretName(true))
	}
	return secrets
}

// IsAuthorizationEnabled reports whether JWT authorization is on.
func (r *DurableWorkflow) IsAuthorizationEnabled() bool {
	return r.Spec.Authorization != nil
}

// AuthKeysSecretName is the secret with the RSA key pair a Cadence cluster signs its internal tokens with.
func (r *DurableWorkflow) AuthKeysSecretName() string {
	return meta_util.NameWithSuffix(r.Name, "auth-keys")
}

func (r *DurableWorkflow) IsTLSEnabled() bool {
	return r.Spec.TLS != nil && r.Spec.TLS.IssuerRef != nil
}

// IsFrontendClientAuthRequired reports whether the frontend requires client certificates.
func (r *DurableWorkflow) IsFrontendClientAuthRequired(v *catalog.DurableWorkflowVersion) bool {
	if r.Spec.FrontendClientAuth != nil {
		return *r.Spec.FrontendClientAuth == DurableWorkflowClientAuthRequired
	}
	return v.IsTemporal()
}

// GetCertSecretName returns the secret name for a certificate alias if any,
// otherwise returns default certificate secret name for the given alias.
func (r *DurableWorkflow) GetCertSecretName(alias DurableWorkflowCertificateType) string {
	if r.Spec.TLS != nil {
		name, ok := kmapi.GetCertificateSecretName(r.Spec.TLS.Certificates, string(alias))
		if ok {
			return name
		}
	}
	return r.CertificateName(alias)
}

func (r *DurableWorkflow) CertificateName(alias DurableWorkflowCertificateType) string {
	return meta_util.NameWithSuffix(r.Name, fmt.Sprintf("%s-cert", string(alias)))
}

func (r *DurableWorkflow) SetTLSDefaults() {
	if !r.IsTLSEnabled() {
		return
	}
	for _, alias := range []DurableWorkflowCertificateType{
		DurableWorkflowCertificateTypeInternode,
		DurableWorkflowCertificateTypeFrontend,
		DurableWorkflowCertificateTypeClient,
	} {
		r.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(r.Spec.TLS.Certificates, string(alias), r.CertificateName(alias))
	}
}

func (r *DurableWorkflow) SetHealthCheckerDefaults() {
	if r.Spec.HealthChecker.PeriodSeconds == nil {
		r.Spec.HealthChecker.PeriodSeconds = ptr.To(int32(10))
	}
	if r.Spec.HealthChecker.TimeoutSeconds == nil {
		r.Spec.HealthChecker.TimeoutSeconds = ptr.To(int32(10))
	}
	if r.Spec.HealthChecker.FailureThreshold == nil {
		r.Spec.HealthChecker.FailureThreshold = ptr.To(int32(3))
	}
}

func (r *DurableWorkflow) SetDefaults(kc client.Client) {
	if r.Spec.DeletionPolicy == "" {
		r.Spec.DeletionPolicy = DeletionPolicyDelete
	}
	if r.Spec.Topology == nil {
		r.Spec.Topology = &DurableWorkflowTopology{}
	}
	t := r.Spec.Topology
	for _, n := range []**DurableWorkflowNode{&t.Frontend, &t.History, &t.Matching, &t.Worker} {
		if *n == nil {
			*n = &DurableWorkflowNode{}
		}
		if (*n).Replicas == nil {
			(*n).Replicas = ptr.To(int32(1))
		}
	}
	if t.InternalFrontend != nil && t.InternalFrontend.Replicas == nil {
		t.InternalFrontend.Replicas = ptr.To(int32(1))
	}

	r.setDefaultStore(&r.Spec.Persistence.Default, false)
	if m := r.Spec.Persistence.Managed; m != nil {
		m.Engine = m.ManagedEngine()
		m.Replicas = ptr.To(m.ManagedReplicas())
		if m.Version == "" {
			// Written to the spec once, so that a newer catalog entry never changes an existing database.
			if v, err := ResolveManagedVersion(context.TODO(), kc, m.Engine); err != nil {
				klog.Errorf("can't pick a version for the managed %s of %s/%s: %s", m.Engine, r.Namespace, r.Name, err.Error())
			} else {
				m.Version = v
			}
		}
	}
	if r.Spec.Persistence.Visibility != nil {
		r.setDefaultStore(r.Spec.Persistence.Visibility, true)
	}

	var version catalog.DurableWorkflowVersion
	if err := kc.Get(context.TODO(), types.NamespacedName{Name: r.Spec.Version}, &version); err != nil {
		klog.Errorf("can't get the DurableWorkflowVersion %s: %s", r.Spec.Version, err.Error())
		return
	}

	if r.Spec.NumHistoryShards == nil {
		shards := version.Spec.DefaultHistoryShards
		if shards == 0 {
			shards = 16
		}
		r.Spec.NumHistoryShards = ptr.To(shards)
	}

	for _, role := range r.Roles(&version) {
		if n := r.node(role); n != nil {
			r.setDefaultContainerSecurityContext(&version, n)
		}
	}
	r.SetHealthCheckerDefaults()

	if r.Spec.Monitor != nil {
		if r.Spec.Monitor.Prometheus == nil {
			r.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if r.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			r.Spec.Monitor.Prometheus.Exporter.Port = version.Spec.Ports.Metrics
		}
		r.Spec.Monitor.SetDefaults()
	}

	r.SetTLSDefaults()
}

func (r *DurableWorkflow) setDefaultStore(s *DurableWorkflowStore, visibility bool) {
	if s.DatabaseName == "" {
		s.DatabaseName = r.StoreDatabaseName(visibility)
	}
	if s.CreateDatabase == nil {
		s.CreateDatabase = ptr.To(true)
	}
	if s.AuthSecret == nil {
		s.AuthSecret = &SecretReference{}
	}
	if s.AuthSecret.Kind == "" {
		s.AuthSecret.Kind = kubedb.ResourceKindSecret
	}
}

func (r *DurableWorkflow) setDefaultContainerSecurityContext(v *catalog.DurableWorkflowVersion, n *DurableWorkflowNode) {
	if n.PodTemplate == nil {
		n.PodTemplate = &ofst.PodTemplateSpec{}
	}
	podTemplate := n.PodTemplate
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = v.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.DurableWorkflowContainerName)
	if container == nil {
		container = &core.Container{Name: kubedb.DurableWorkflowContainerName}
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
		container = coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.DurableWorkflowContainerName)
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}
	sc := container.SecurityContext
	if sc.AllowPrivilegeEscalation == nil {
		sc.AllowPrivilegeEscalation = ptr.To(false)
	}
	if sc.Capabilities == nil {
		sc.Capabilities = &core.Capabilities{Drop: []core.Capability{"ALL"}}
	}
	if sc.RunAsNonRoot == nil {
		sc.RunAsNonRoot = ptr.To(true)
	}
	if sc.RunAsUser == nil {
		sc.RunAsUser = v.Spec.SecurityContext.RunAsUser
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
	apis.SetDefaultResourceLimits(&container.Resources, kubedb.DefaultResourcesDurableWorkflow)
	apis.SetDefaultResizePolicy(podTemplate.Spec.Containers, podTemplate.Spec.InitContainers)
}

func (r *DurableWorkflow) PodLabels(role string, extraLabels ...map[string]string) map[string]string {
	var override map[string]string
	if pt := r.PodTemplate(role); pt != nil {
		override = pt.Labels
	}
	return r.offshootLabels(meta_util.OverwriteKeys(r.RoleSelectors(role), extraLabels...), override)
}

func (r *DurableWorkflow) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(r.Spec.ServiceTemplates, alias)
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

type DurableWorkflowApp struct {
	*DurableWorkflow
}

func (r *DurableWorkflowApp) Name() string {
	return r.DurableWorkflow.Name
}

func (r *DurableWorkflowApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceKindDurableWorkflow))
}

func (r *DurableWorkflow) AppBindingMeta() appcat.AppBindingMeta {
	return &DurableWorkflowApp{r}
}

type durableWorkflowStatsService struct {
	*DurableWorkflow
}

func (r durableWorkflowStatsService) GetNamespace() string {
	return r.DurableWorkflow.GetNamespace()
}

func (r durableWorkflowStatsService) ServiceName() string {
	return r.OffshootName() + "-stats"
}

func (r durableWorkflowStatsService) ServiceMonitorName() string {
	return r.ServiceName()
}

func (r durableWorkflowStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return r.OffshootLabels()
}

func (r durableWorkflowStatsService) Path() string {
	return kubedb.DefaultStatsPath
}

func (r durableWorkflowStatsService) Scheme() string {
	return "http"
}

func (r durableWorkflowStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (r *DurableWorkflow) StatsService() mona.StatsAccessor {
	return &durableWorkflowStatsService{r}
}

func (r *DurableWorkflow) StatsServiceLabels() map[string]string {
	return r.ServiceLabels(StatsServiceAlias, map[string]string{kubedb.LabelRole: kubedb.RoleStats})
}

type DurableWorkflowBind struct {
	*DurableWorkflow
}

var _ DBBindInterface = &DurableWorkflowBind{}

// ServiceNames returns the frontend Service and, if enabled, the UI Service.
func (r *DurableWorkflowBind) ServiceNames() (string, string) {
	ui := ""
	if r.Spec.UI != nil && r.Spec.UI.Enabled {
		ui = r.UIServiceName()
	}
	return r.ServiceName(), ui
}

// Ports cannot be resolved without the catalog; the controller fills the AppBinding ports itself.
func (r *DurableWorkflowBind) Ports() (int, int) {
	return 0, 0
}

func (r *DurableWorkflowBind) SecretName() string {
	return r.GetCertSecretName(DurableWorkflowCertificateTypeClient)
}

func (r *DurableWorkflowBind) CertSecretName() string {
	return r.GetCertSecretName(DurableWorkflowCertificateTypeClient)
}
