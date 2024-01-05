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
	"path/filepath"
	"strings"

	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gomodules.xyz/pointer"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (r *Rabbitmq) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralRabbitmq))
}

func (r *Rabbitmq) AsOwner() *meta.OwnerReference {
	return meta.NewControllerRef(r, SchemeGroupVersion.WithKind(ResourceKindRabbitmq))
}

func (r *Rabbitmq) ResourceShortCode() string {
	return ResourceCodeRabbitmq
}

func (r *Rabbitmq) ResourceKind() string {
	return ResourceKindRabbitmq
}

func (r *Rabbitmq) ResourceSingular() string {
	return ResourceSingularRabbitmq
}

func (r *Rabbitmq) ResourcePlural() string {
	return ResourcePluralRabbitmq
}

func (r *Rabbitmq) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", r.ResourcePlural(), kubedb.GroupName)
}

// Owner returns owner reference to resources
func (r *Rabbitmq) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(r, SchemeGroupVersion.WithKind(r.ResourceKind()))
}

func (r *Rabbitmq) OffshootName() string {
	return r.Name
}

func (r *Rabbitmq) ServiceName() string {
	return r.OffshootName()
}

func (r *Rabbitmq) GoverningServiceName() string {
	return meta_util.NameWithSuffix(r.ServiceName(), "pods")
}

func (r *Rabbitmq) StandbyServiceName() string {
	return meta_util.NameWithPrefix(r.ServiceName(), KafkaStandbyServiceSuffix)
}

func (r *Rabbitmq) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, r.Labels, override))
}

func (r *Rabbitmq) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      r.ResourceFQN(),
		meta_util.InstanceLabelKey:  r.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (r *Rabbitmq) OffshootLabels() map[string]string {
	return r.offshootLabels(r.OffshootSelectors(), nil)
}

func (r *Rabbitmq) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(r.Spec.ServiceTemplates, alias)
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

type RabbitmqStatsService struct {
	*Rabbitmq
}

func (ks RabbitmqStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (ks RabbitmqStatsService) GetNamespace() string {
	return ks.Rabbitmq.GetNamespace()
}

func (ks RabbitmqStatsService) ServiceName() string {
	return ks.OffshootName() + "-stats"
}

func (ks RabbitmqStatsService) ServiceMonitorName() string {
	return ks.ServiceName()
}

func (ks RabbitmqStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return ks.OffshootLabels()
}

func (ks RabbitmqStatsService) Path() string {
	return DefaultStatsPath
}

func (ks RabbitmqStatsService) Scheme() string {
	return ""
}

func (r *Rabbitmq) StatsService() mona.StatsAccessor {
	return &RabbitmqStatsService{r}
}

func (r *Rabbitmq) StatsServiceLabels() map[string]string {
	return r.ServiceLabels(StatsServiceAlias, map[string]string{LabelRole: RoleStats})
}

func (r *Rabbitmq) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), r.Spec.PodTemplate.Controller.Labels)
}

func (r *Rabbitmq) PodLabels(extraLabels ...map[string]string) map[string]string {
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), r.Spec.PodTemplate.Labels)
}

func (r *Rabbitmq) StatefulSetName() string {
	return r.OffshootName()
}

func (r *Rabbitmq) ServiceAccountName() string {
	return r.OffshootName()
}

func (r *Rabbitmq) DefaultPodRoleName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "role")
}

func (r *Rabbitmq) DefaultPodRoleBindingName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "rolebinding")
}

func (r *Rabbitmq) ConfigSecretName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "config")
}

func (r *Rabbitmq) DefaultUserCredSecretName(username string) string {
	return meta_util.NameWithSuffix(r.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (r *Rabbitmq) DefaultErlangCookieSecretName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "erlang-cookie")
}

// CertificateName returns the default certificate name and/or certificate secret name for a certificate alias
func (r *Rabbitmq) CertificateName(alias RabbitmqCertificateAlias) string {
	return meta_util.NameWithSuffix(r.Name, fmt.Sprintf("%s-cert", string(alias)))
}

// ClientCertificateCN returns the CN for a client certificate
func (r *Rabbitmq) ClientCertificateCN(alias RabbitmqCertificateAlias) string {
	return fmt.Sprintf("%s-%s", r.Name, string(alias))
}

// GetCertSecretName returns the secret name for a certificate alias if any,
// otherwise returns default certificate secret name for the given alias.
func (r *Rabbitmq) GetCertSecretName(alias RabbitmqCertificateAlias) string {
	if r.Spec.TLS != nil {
		name, ok := kmapi.GetCertificateSecretName(r.Spec.TLS.Certificates, string(alias))
		if ok {
			return name
		}
	}
	return r.CertificateName(alias)
}

// returns the CertSecretVolumeName
// Values will be like: client-certs, server-certs etc.
func (r *Rabbitmq) CertSecretVolumeName(alias RabbitmqCertificateAlias) string {
	return string(alias) + "-certs"
}

// returns CertSecretVolumeMountPath
// if configDir is "/opt/kafka/config",
// mountPath will be, "/opt/kafka/config/<alias>".
func (r *Rabbitmq) CertSecretVolumeMountPath(configDir string, cert string) string {
	return filepath.Join(configDir, cert)
}

func (r *Rabbitmq) PVCName(alias string) string {
	return meta_util.NameWithSuffix(r.Name, alias)
}

func (r *Rabbitmq) SetHealthCheckerDefaults() {
	if r.Spec.HealthChecker.PeriodSeconds == nil {
		r.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if r.Spec.HealthChecker.TimeoutSeconds == nil {
		r.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if r.Spec.HealthChecker.FailureThreshold == nil {
		r.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (r *Rabbitmq) SetDefaults() {
	if r.Spec.TerminationPolicy == "" {
		r.Spec.TerminationPolicy = TerminationPolicyDelete
	}

	if r.Spec.StorageType == "" {
		r.Spec.StorageType = StorageTypeDurable
	}

	r.SetHealthCheckerDefaults()
}

func (r *Rabbitmq) SetTLSDefaults() {
	if r.Spec.TLS == nil || r.Spec.TLS.IssuerRef == nil {
		return
	}
	r.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(r.Spec.TLS.Certificates, string(RabbitmqServerCert), r.CertificateName(RabbitmqServerCert))
	r.Spec.TLS.Certificates = kmapi.SetMissingSecretNameForCertificate(r.Spec.TLS.Certificates, string(RabbitmqClientCert), r.CertificateName(RabbitmqClientCert))
}

type RabbitmqApp struct {
	*Rabbitmq
}

func (r RabbitmqApp) Name() string {
	return r.Rabbitmq.Name
}

func (r RabbitmqApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularRabbitmq))
}

func (r *Rabbitmq) AppBindingMeta() appcat.AppBindingMeta {
	return &RabbitmqApp{r}
}

func (r *Rabbitmq) GetConnectionScheme() string {
	scheme := "http"
	if r.Spec.EnableSSL {
		scheme = "https"
	}
	return scheme
}
