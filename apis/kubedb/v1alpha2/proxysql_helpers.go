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

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	"gomodules.xyz/pointer"
	"k8s.io/apimachinery/pkg/labels"
	appslister "k8s.io/client-go/listers/apps/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (_ ProxySQL) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralProxySQL))
}

var _ apis.ResourceInfo = &ProxySQL{}

func (p ProxySQL) OffshootName() string {
	return p.Name
}

func (p ProxySQL) OffshootSelectors() map[string]string {
	return map[string]string{
		meta_util.NameLabelKey:      p.ResourceFQN(),
		meta_util.InstanceLabelKey:  p.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
		LabelProxySQLLoadBalance:    string(*p.Spec.Mode),
	}
}

func (p ProxySQL) OffshootLabels() map[string]string {
	return p.offshootLabels(p.OffshootSelectors(), nil)
}

func (p ProxySQL) PodLabels() map[string]string {
	return p.offshootLabels(p.OffshootSelectors(), p.Spec.PodTemplate.Labels)
}

func (p ProxySQL) PodControllerLabels() map[string]string {
	return p.offshootLabels(p.OffshootSelectors(), p.Spec.PodTemplate.Controller.Labels)
}

func (p ProxySQL) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(p.Spec.ServiceTemplates, alias)
	return p.offshootLabels(meta_util.OverwriteKeys(p.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (p ProxySQL) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, p.Labels, override))
}

func (p ProxySQL) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralProxySQL, kubedb.GroupName)
}

func (p ProxySQL) ResourceShortCode() string {
	return ResourceCodeProxySQL
}

func (p ProxySQL) ResourceKind() string {
	return ResourceKindProxySQL
}

func (p ProxySQL) ResourceSingular() string {
	return ResourceSingularProxySQL
}

func (p ProxySQL) ResourcePlural() string {
	return ResourcePluralProxySQL
}

func (p ProxySQL) GetAuthSecretName() string {
	if p.Spec.AuthSecret != nil && p.Spec.AuthSecret.Name != "" {
		return p.Spec.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(p.OffshootName(), "auth")
}

func (p ProxySQL) ServiceName() string {
	return p.OffshootName()
}

func (p ProxySQL) GoverningServiceName() string {
	return meta_util.NameWithSuffix(p.ServiceName(), "pods")
}

type proxysqlApp struct {
	*ProxySQL
}

func (p proxysqlApp) Name() string {
	return p.ProxySQL.Name
}

func (p proxysqlApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularProxySQL))
}

func (p ProxySQL) AppBindingMeta() appcat.AppBindingMeta {
	return &proxysqlApp{&p}
}

type proxysqlStatsService struct {
	*ProxySQL
}

func (p proxysqlStatsService) GetNamespace() string {
	return p.ProxySQL.GetNamespace()
}

func (p proxysqlStatsService) ServiceName() string {
	return p.OffshootName() + "-stats"
}

func (p proxysqlStatsService) ServiceMonitorName() string {
	return p.ServiceName()
}

func (p proxysqlStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return p.OffshootLabels()
}

func (p proxysqlStatsService) Path() string {
	return DefaultStatsPath
}

func (p proxysqlStatsService) Scheme() string {
	return ""
}

func (p ProxySQL) StatsService() mona.StatsAccessor {
	return &proxysqlStatsService{&p}
}

func (p ProxySQL) StatsServiceLabels() map[string]string {
	return p.ServiceLabels(StatsServiceAlias, map[string]string{LabelRole: RoleStats})
}

func (p *ProxySQL) SetDefaults() {
	if p == nil {
		return
	}

	if p == nil || p.Spec.Mode == nil || p.Spec.Backend == nil {
		return
	}

	if p.Spec.Replicas == nil {
		p.Spec.Replicas = pointer.Int32P(1)
	}

	p.Spec.Monitor.SetDefaults()
	p.SetHealthCheckerDefaults()
	apis.SetDefaultResourceLimits(&p.Spec.PodTemplate.Spec.Resources, DefaultResources)
}

func (p *ProxySQL) SetHealthCheckerDefaults() {
	if p.Spec.HealthChecker.PeriodSeconds == nil {
		p.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if p.Spec.HealthChecker.TimeoutSeconds == nil {
		p.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if p.Spec.HealthChecker.FailureThreshold == nil {
		p.Spec.HealthChecker.FailureThreshold = pointer.Int32P(1)
	}
}

func (m *ProxySQL) SetTLSDefaults(issuerType IssuerType) {
	if m.Spec.TLS == nil || m.Spec.TLS.IssuerRef == nil {
		return
	}
	if issuerType == IssuerTypeACME {
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLServerCert), m.CertificateName(ProxySQLServerCert))
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLClientCert), m.CertificateName(ProxySQLServerCert))
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLMetricsExporterCert), m.CertificateName(ProxySQLServerCert))

	} else if issuerType == IssuerTypeSelfSigned {
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLServerCert), m.CertificateName(ProxySQLServerCert))
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLClientCert), m.CertificateName(ProxySQLClientCert))
		m.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(m.Spec.TLS.Certificates, string(ProxySQLMetricsExporterCert), m.CertificateName(ProxySQLMetricsExporterCert))
	}
}

func (p *ProxySQLSpec) GetPersistentSecrets() []string {
	if p == nil {
		return nil
	}

	var secrets []string
	if p.AuthSecret != nil {
		secrets = append(secrets, p.AuthSecret.Name)
	}
	return secrets
}

func (p *ProxySQL) ReplicasAreReady(lister appslister.StatefulSetLister) (bool, string, error) {
	// Desire number of statefulSets
	expectedItems := 1
	return checkReplicas(lister.StatefulSets(p.Namespace), labels.SelectorFromSet(p.OffshootLabels()), expectedItems)
}

// GetCertSecretName returns the secret name for a certificate alias if any,
// otherwise returns default certificate secret name for the given alias.
func (m *ProxySQL) GetCertSecretName(alias ProxySQLCertificateAlias) string {
	if m.Spec.TLS != nil {
		name, ok := kmapi.GetCertificateSecretName(m.Spec.TLS.Certificates, string(alias))
		if ok {
			return name
		}
	}
	return m.CertificateName(alias)
}

// CertificateName returns the default certificate name and/or certificate secret name for a certificate alias
func (m *ProxySQL) CertificateName(alias ProxySQLCertificateAlias) string {
	return meta_util.NameWithSuffix(m.Name, fmt.Sprintf("%s-cert", string(alias)))
}

// IsCluster returns boolean true if the proxysql is in cluster mode, otherwise false
func (m *ProxySQL) IsCluster() bool {
	r := m.Spec.Replicas
	return *r > 1
}
