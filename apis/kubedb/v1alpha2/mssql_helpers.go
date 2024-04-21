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
	"time"

	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	metautil "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

type MSSQLApp struct {
	*MSSQL
}

func (m *MSSQL) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMSSQL))
}

func (m MSSQLApp) Name() string {
	return m.MSSQL.Name
}

func (m MSSQLApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularMSSQL))
}

func (m *MSSQL) ResourceKind() string {
	return ResourceKindMSSQL
}

func (m *MSSQL) ResourcePlural() string {
	return ResourcePluralMSSQL
}

func (m *MSSQL) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", m.ResourcePlural(), kubedb.GroupName)
}

// Owner returns owner reference to resources
func (m *MSSQL) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(m, SchemeGroupVersion.WithKind(m.ResourceKind()))
}

func (m *MSSQL) OffshootName() string {
	return m.Name
}

func (m *MSSQL) ServiceName() string {
	return m.OffshootName()
}

func (m *MSSQL) SecondaryServiceName() string {
	return metautil.NameWithPrefix(m.ServiceName(), "secondary")
}

func (m *MSSQL) GoverningServiceName() string {
	return metautil.NameWithSuffix(m.ServiceName(), "pods")
}

func (m *MSSQL) DefaultUserCredSecretName(username string) string {
	return metautil.NameWithSuffix(m.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (m *MSSQL) offshootLabels(selector, override map[string]string) map[string]string {
	selector[metautil.ComponentLabelKey] = ComponentDatabase
	return metautil.FilterKeys(kubedb.GroupName, selector, metautil.OverwriteKeys(nil, m.Labels, override))
}

func (m *MSSQL) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(m.Spec.ServiceTemplates, alias)
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (m *MSSQL) OffshootLabels() map[string]string {
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MSSQL) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      m.ResourceFQN(),
		metautil.InstanceLabelKey:  m.Name,
		metautil.ManagedByLabelKey: kubedb.GroupName,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (m *MSSQL) IsAvailabilityGroup() bool {
	return m.Spec.Topology != nil &&
		m.Spec.Topology.Mode != nil &&
		*m.Spec.Topology.Mode == MSSQLModeAvailabilityGroup
}

func (m *MSSQL) IsStandalone() bool {
	return m.Spec.Topology == nil
}

func (m *MSSQL) PVCName(alias string) string {
	return metautil.NameWithSuffix(m.Name, alias)
}

func (m *MSSQL) PodLabels(extraLabels ...map[string]string) map[string]string {
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), m.Spec.PodTemplate.Labels)
}

func (m *MSSQL) PodLabel(podTemplate *ofst.PodTemplateSpec) map[string]string {
	if podTemplate != nil && podTemplate.Labels != nil {
		return m.offshootLabels(m.OffshootSelectors(), m.Spec.PodTemplate.Labels)
	}
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MSSQL) ConfigSecretName() string {
	return metautil.NameWithSuffix(m.OffshootName(), "config")
}

func (m *MSSQL) PetSetName() string {
	return m.OffshootName()
}

func (m *MSSQL) ServiceAccountName() string {
	return m.OffshootName()
}

func (m *MSSQL) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return m.offshootLabels(metautil.OverwriteKeys(m.OffshootSelectors(), extraLabels...), m.Spec.PodTemplate.Controller.Labels)
}

func (m *MSSQL) PodControllerLabel(podTemplate *ofst.PodTemplateSpec) map[string]string {
	if podTemplate != nil && podTemplate.Controller.Labels != nil {
		return m.offshootLabels(m.OffshootSelectors(), podTemplate.Controller.Labels)
	}
	return m.offshootLabels(m.OffshootSelectors(), nil)
}

func (m *MSSQL) GetPersistentSecrets() []string {
	var secrets []string
	if m.Spec.AuthSecret != nil {
		secrets = append(secrets, m.Spec.AuthSecret.Name)
	}

	secrets = append(secrets, MSSQLEndpointCertsSecretName)
	secrets = append(secrets, MSSQLDbmLoginSecretName)
	secrets = append(secrets, MSSQLMasterKeySecretName)

	return secrets
}

func (m *MSSQL) AppBindingMeta() appcat.AppBindingMeta {
	return &MSSQLApp{m}
}

func (m MSSQL) SetHealthCheckerDefaults() {
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

func (m MSSQL) GetAuthSecretName() string {
	if m.Spec.AuthSecret != nil && m.Spec.AuthSecret.Name != "" {
		return m.Spec.AuthSecret.Name
	}
	return metautil.NameWithSuffix(m.OffshootName(), "auth")
}

func (m *MSSQL) GetNameSpacedName() string {
	return m.Namespace + "/" + m.Name
}

func (m *MSSQL) PrimaryServiceDNS() string {
	return fmt.Sprintf("%s.%s.svc", m.ServiceName(), m.Namespace)
}

func (m *MSSQL) SetDefaults() {
	if m == nil {
		return
	}
	if m.Spec.StorageType == "" {
		m.Spec.StorageType = StorageTypeDurable
	}
	if m.Spec.TerminationPolicy == "" {
		m.Spec.TerminationPolicy = TerminationPolicyDelete
	}

	if m.IsStandalone() {
		if m.Spec.Replicas == nil {
			m.Spec.Replicas = pointer.Int32P(1)
		}
	} else {
		if m.Spec.LeaderElection == nil {
			m.Spec.LeaderElection = &MSSQLLeaderElectionConfig{
				// The upper limit of election timeout is 50000ms (50s), which should only be used when deploying a
				// globally-distributed etcd cluster. A reasonable round-trip time for the continental United States is around 130-150ms,
				// and the time between US and Japan is around 350-400ms. If the network has uneven performance or regular packet
				// delays/loss then it is possible that a couple of retries may be necessary to successfully send a packet.
				// So 5s is a safe upper limit of global round-trip time. As the election timeout should be an order of magnitude
				// bigger than broadcast time, in the case of ~5s for a globally distributed cluster, then 50 seconds becomes
				// a reasonable maximum.
				Period: meta.Duration{Duration: 300 * time.Millisecond},
				// the amount of HeartbeatTick can be missed before the failOver
				ElectionTick: 10,
				// this value should be one.
				HeartbeatTick: 1,
			}
		}
		if m.Spec.LeaderElection.TransferLeadershipInterval == nil {
			m.Spec.LeaderElection.TransferLeadershipInterval = &meta.Duration{Duration: 1 * time.Second}
		}
		if m.Spec.LeaderElection.TransferLeadershipTimeout == nil {
			m.Spec.LeaderElection.TransferLeadershipTimeout = &meta.Duration{Duration: 60 * time.Second}
		}
	}

	if m.Spec.PodTemplate == nil {
		m.Spec.PodTemplate = &ofst.PodTemplateSpec{}
	}

	var mssqlVersion catalog.MSSQLVersion
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: m.Spec.Version,
	}, &mssqlVersion)
	if err != nil {
		klog.Errorf("can't get the MSSQL version object %s for %s \n", m.Spec.Version, err.Error())
		return
	}

	m.setDefaultContainerSecurityContext(&mssqlVersion, m.Spec.PodTemplate)

	m.SetHealthCheckerDefaults()

	m.setDefaultContainerResourceLimits(m.Spec.PodTemplate)
}

func (m *MSSQL) setDefaultContainerSecurityContext(mssqlVersion *catalog.MSSQLVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = mssqlVersion.Spec.SecurityContext.RunAsGroup
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, MSSQLContainerName)
	if container == nil {
		container = &core.Container{
			Name: MSSQLContainerName,
		}
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}

	m.assignDefaultContainerSecurityContext(mssqlVersion, container.SecurityContext, true)

	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, MSSQLInitContainerName)
	if initContainer == nil {
		initContainer = &core.Container{
			Name: MSSQLInitContainerName,
		}
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &core.SecurityContext{}
	}
	m.assignDefaultContainerSecurityContext(mssqlVersion, initContainer.SecurityContext, false)
	podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)

	if m.IsAvailabilityGroup() {
		coordinatorContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, MSSQLCoordinatorContainerName)
		if coordinatorContainer == nil {
			coordinatorContainer = &core.Container{
				Name: MSSQLCoordinatorContainerName,
			}
		}
		if coordinatorContainer.SecurityContext == nil {
			coordinatorContainer.SecurityContext = &core.SecurityContext{}
		}
		m.assignDefaultContainerSecurityContext(mssqlVersion, coordinatorContainer.SecurityContext, false)
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *coordinatorContainer)
	}
}

func (m *MSSQL) assignDefaultContainerSecurityContext(mssqlVersion *catalog.MSSQLVersion, sc *core.SecurityContext, isMainContainer bool) {
	if sc.AllowPrivilegeEscalation == nil {
		sc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if sc.Capabilities == nil {
		if isMainContainer {
			sc.Capabilities = &core.Capabilities{
				Drop: []core.Capability{"ALL"},
				Add:  []core.Capability{"NET_BIND_SERVICE"},
			}
		} else {
			sc.Capabilities = &core.Capabilities{
				Drop: []core.Capability{"ALL"},
			}
		}
	}
	if sc.RunAsNonRoot == nil {
		sc.RunAsNonRoot = pointer.BoolP(true)
	}
	if sc.RunAsUser == nil {
		sc.RunAsUser = mssqlVersion.Spec.SecurityContext.RunAsUser
	}
	if sc.RunAsGroup == nil {
		sc.RunAsGroup = mssqlVersion.Spec.SecurityContext.RunAsGroup
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

func (m *MSSQL) setDefaultContainerResourceLimits(podTemplate *ofst.PodTemplateSpec) {
	dbContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, MSSQLContainerName)
	if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&dbContainer.Resources, DefaultResourcesMemoryIntensive)
	}

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, MSSQLInitContainerName)
	if initContainer != nil && (initContainer.Resources.Requests == nil && initContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&initContainer.Resources, DefaultInitContainerResource)
	}

	if m.IsAvailabilityGroup() {
		coordinatorContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, MSSQLCoordinatorContainerName)
		if coordinatorContainer != nil && (coordinatorContainer.Resources.Requests == nil && coordinatorContainer.Resources.Limits == nil) {
			apis.SetDefaultResourceLimits(&coordinatorContainer.Resources, CoordinatorDefaultResources)
		}
	}
}
