package v1alpha2

import (
	"fmt"
	"gomodules.xyz/pointer"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	metautil "kmodules.xyz/client-go/meta"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	"kubedb.dev/apimachinery/apis/kubedb"
	"strings"
)

type MsSQLApp struct {
	*MsSQL
}

func (m MsSQLApp) Name() string {
	return m.MsSQL.Name
}

func (m MsSQLApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularMsSQL))
}

func (m *MsSQL) ResourceKind() string {
	return ResourceKindMsSQL
}

func (m *MsSQL) ResourcePlural() string {
	return ResourcePluralMsSQL
}

func (m *MsSQL) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", m.ResourcePlural(), kubedb.GroupName)
}

// Owner returns owner reference to resources
func (m *MsSQL) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(m, SchemeGroupVersion.WithKind(m.ResourceKind()))
}

func (m *MsSQL) OffshootName() string {
	return m.Name
}

func (m *MsSQL) ServiceName() string {
	return m.OffshootName()
}

func (m *MsSQL) SecondaryServiceName() string {
	return metautil.NameWithPrefix(m.ServiceName(), "secondary")
}

func (m *MsSQL) GoverningServiceName() string {
	return metautil.NameWithSuffix(m.ServiceName(), "pods")
}

func (m *MsSQL) DefaultUserCredSecretName(username string) string {
	return metautil.NameWithSuffix(m.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (m *MsSQL) offshootLabels(selector, override map[string]string) map[string]string {
	selector[metautil.ComponentLabelKey] = ComponentDatabase
	return metautil.FilterKeys(kubedb.GroupName, selector, metautil.OverwriteKeys(nil, m.Labels, override))
}

func (m *MsSQL) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(m.Spec.ServiceTemplates, alias)
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (m *MsSQL) OffshootLabels() map[string]string {
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MsSQL) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      m.ResourceFQN(),
		metautil.InstanceLabelKey:  m.Name,
		metautil.ManagedByLabelKey: kubedb.GroupName,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (m *MsSQL) IsClustering() bool {
	return m.Spec.Topology != nil && m.Spec.Topology.Mode != nil && *m.Spec.Topology.Mode == MsSQLModeAvailabilityGroup
}

func (m *MsSQL) IsStandalone() bool {
	return m.Spec.Topology == nil || (m.Spec.Topology.Mode != nil && *m.Spec.Topology.Mode == MsSQLModeStandalone)
}

func (m *MsSQL) PVCName(alias string) string {
	return metautil.NameWithSuffix(m.Name, alias)
}

func (m *MsSQL) PodLabels(extraLabels ...map[string]string) map[string]string {
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), m.Spec.PodTemplate.Labels)
}
func (m *MsSQL) PodLabel(podTemplate *ofst.PodTemplateSpec) map[string]string {
	if podTemplate != nil && podTemplate.Labels != nil {
		return m.offshootLabels(m.OffshootSelectors(), m.Spec.PodTemplate.Labels)
	}
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MsSQL) ConfigSecretName() string {
	return metautil.NameWithSuffix(m.OffshootName(), "config")
}

func (m *MsSQL) StatefulSetName() string {
	return m.OffshootName()
}

func (m *MsSQL) ServiceAccountName() string {
	return m.OffshootName()
}

func (m *MsSQL) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), m.Spec.PodTemplate.Controller.Labels)
}

func (m *MsSQL) PodControllerLabel(podTemplate *ofst.PodTemplateSpec) map[string]string {
	if podTemplate != nil && podTemplate.Controller.Labels != nil {
		return m.offshootLabels(m.OffshootSelectors(), podTemplate.Controller.Labels)
	}
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MsSQL) GetPersistentSecrets() []string {
	var secrets []string
	if m.Spec.AuthSecret != nil {
		secrets = append(secrets, m.Spec.AuthSecret.Name)
	}
	return secrets
}

func (m *MsSQL) AppBindingMeta() appcat.AppBindingMeta {
	return &MsSQLApp{m}
}

func (m MsSQL) SetHealthCheckerDefaults() {
	if m.Spec.HealthChecker.PeriodSeconds == nil {
		m.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if m.Spec.HealthChecker.TimeoutSeconds == nil {
		m.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if m.Spec.HealthChecker.FailureThreshold == nil {
		m.Spec.HealthChecker.FailureThreshold = pointer.Int32P(1)
	}
}

func (m MsSQL) GetAuthSecretName() string {
	if m.Spec.AuthSecret != nil && m.Spec.AuthSecret.Name != "" {
		return m.Spec.AuthSecret.Name
	}
	return m.DefaultUserCredSecretName(MsSQLSAUser)
}

func (m *MsSQL) GetNameSpacedName() string {
	return m.Namespace + "/" + m.Name
}

func (m *MsSQL) PrimaryServiceDNS() string {
	return fmt.Sprintf("%s.%s.svc", m.ServiceName(), m.Namespace)
}
