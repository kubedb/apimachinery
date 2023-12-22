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
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis/kafka"
	"kubedb.dev/apimachinery/crds"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	"kmodules.xyz/client-go/apiextensions"
	meta_util "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (k *ConnectCluster) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralConnectCluster))
}

func (k *ConnectCluster) AsOwner() *meta.OwnerReference {
	return meta.NewControllerRef(k, SchemeGroupVersion.WithKind(ResourceKindConnectCluster))
}

func (k *ConnectCluster) ResourceShortCode() string {
	return ResourceCodeConnectCluster
}

func (k *ConnectCluster) ResourceKind() string {
	return ResourceKindConnectCluster
}

func (k *ConnectCluster) ResourceSingular() string {
	return ResourceSingularConnectCluster
}

func (k *ConnectCluster) ResourcePlural() string {
	return ResourcePluralConnectCluster
}

func (k *ConnectCluster) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", k.ResourcePlural(), kafka.GroupName)
}

// Owner returns owner reference to resources
func (k *ConnectCluster) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(k, SchemeGroupVersion.WithKind(k.ResourceKind()))
}

func (k *ConnectCluster) OffshootName() string {
	return k.Name
}

func (k *ConnectCluster) ServiceName() string {
	return k.OffshootName()
}

func (k *ConnectCluster) GoverningServiceName() string {
	return meta_util.NameWithSuffix(k.ServiceName(), "pods")
}

func (k *ConnectCluster) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentKafka
	return meta_util.FilterKeys(kafka.GroupName, selector, meta_util.OverwriteKeys(nil, k.Labels, override))
}

func (k *ConnectCluster) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      k.ResourceFQN(),
		meta_util.InstanceLabelKey:  k.Name,
		meta_util.ManagedByLabelKey: kafka.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (k *ConnectCluster) OffshootLabels() map[string]string {
	return k.offshootLabels(k.OffshootSelectors(), nil)
}

// GetServiceTemplate returns a pointer to the desired serviceTemplate referred by "aliaS". Otherwise, it returns nil.
func (k *ConnectCluster) GetServiceTemplate(templates []NamedServiceTemplateSpec, alias ServiceAlias) ofst.ServiceTemplateSpec {
	for i := range templates {
		c := templates[i]
		if c.Alias == alias {
			return c.ServiceTemplateSpec
		}
	}
	return ofst.ServiceTemplateSpec{}
}

func (k *ConnectCluster) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := k.GetServiceTemplate(k.Spec.ServiceTemplates, alias)
	return k.offshootLabels(meta_util.OverwriteKeys(k.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (k *ConnectCluster) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return k.offshootLabels(meta_util.OverwriteKeys(k.OffshootSelectors(), extraLabels...), k.Spec.PodTemplate.Controller.Labels)
}

type connectClusterStatsService struct {
	*ConnectCluster
}

func (ks connectClusterStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (ks connectClusterStatsService) GetNamespace() string {
	return ks.ConnectCluster.GetNamespace()
}

func (ks connectClusterStatsService) ServiceName() string {
	return ks.OffshootName() + "-stats"
}

func (ks connectClusterStatsService) ServiceMonitorName() string {
	return ks.ServiceName()
}

func (ks connectClusterStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return ks.OffshootLabels()
}

func (ks connectClusterStatsService) Path() string {
	return DefaultStatsPath
}

func (ks connectClusterStatsService) Scheme() string {
	return ""
}

func (k *ConnectCluster) StatsService() mona.StatsAccessor {
	return &connectClusterStatsService{k}
}

func (k *ConnectCluster) StatsServiceLabels() map[string]string {
	return k.ServiceLabels(StatsServiceAlias, map[string]string{LabelRole: RoleStats})
}

func (k *ConnectCluster) PodLabels(extraLabels ...map[string]string) map[string]string {
	return k.offshootLabels(meta_util.OverwriteKeys(k.OffshootSelectors(), extraLabels...), k.Spec.PodTemplate.Labels)
}

func (k *ConnectCluster) StatefulSetName() string {
	return k.OffshootName()
}

func (k *ConnectCluster) ConfigSecretName() string {
	return meta_util.NameWithSuffix(k.OffshootName(), "config")
}

func (k *ConnectCluster) KafkaClientCredentialsSecretName() string {
	return meta_util.NameWithSuffix(k.Name, "kafka-client-cred")
}

func (k *ConnectCluster) DefaultUserCredSecretName(username string) string {
	return meta_util.NameWithSuffix(k.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (k *ConnectCluster) DefaultKeystoreCredSecretName() string {
	return meta_util.NameWithSuffix(k.Name, strings.ReplaceAll("connect-keystore-cred", "_", "-"))
}

// CertificateName returns the default certificate name and/or certificate secret name for a certificate alias
func (k *ConnectCluster) CertificateName(alias ConnectClusterCertificateAlias) string {
	return meta_util.NameWithSuffix(k.Name, fmt.Sprintf("%s-connect-cert", string(alias)))
}

// GetCertSecretName returns the secret name for a certificate alias if any,
// otherwise returns default certificate secret name for the given alias.
func (k *ConnectCluster) GetCertSecretName(alias ConnectClusterCertificateAlias) string {
	if k.Spec.TLS != nil {
		name, ok := kmapi.GetCertificateSecretName(k.Spec.TLS.Certificates, string(alias))
		if ok {
			return name
		}
	}
	return k.CertificateName(alias)
}

func (k *ConnectCluster) PVCName(alias string) string {
	return meta_util.NameWithSuffix(k.Name, alias)
}

func (k *ConnectCluster) SetDefaults() {
	if k.Spec.TerminationPolicy == "" {
		k.Spec.TerminationPolicy = TerminationPolicyDelete
	}
}

type ConnectClusterApp struct {
	*ConnectCluster
}

func (r ConnectClusterApp) Name() string {
	return r.ConnectCluster.Name
}

func (r ConnectClusterApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kafka.GroupName, ResourceSingularConnectCluster))
}

func (k *ConnectCluster) AppBindingMeta() appcat.AppBindingMeta {
	return &ConnectClusterApp{k}
}

func (k *ConnectCluster) GetConnectionScheme() string {
	scheme := "http"
	if k.Spec.EnableSSL {
		scheme = "https"
	}
	return scheme
}
