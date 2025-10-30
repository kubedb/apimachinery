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
	pslister "kubeops.dev/petset/client/listers/apps/v1"

	promapi "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	metautil "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	_ "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
)

func (_ HanaDB) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralHanaDB))
}

func (h *HanaDB) ResourceKind() string {
	return ResourceKindHanaDB
}

func (h *HanaDB) ResourcePlural() string {
	return ResourcePluralHanaDB
}

func (h *HanaDB) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", h.ResourcePlural(), SchemeGroupVersion.Group)
}

func (h *HanaDB) ResourceShortCode() string {
	return ResourceCodeHanaDB
}

func (h *HanaDB) OffshootName() string {
	return h.Name
}

func (h *HanaDB) ServiceName() string {
	return h.OffshootName()
}

func (h *HanaDB) GoverningServiceName() string {
	return metautil.NameWithSuffix(h.ServiceName(), "pods")
}

func (h *HanaDB) offshootLabels(selector, override map[string]string) map[string]string {
	selector[metautil.ComponentLabelKey] = kubedb.ComponentDatabase
	return metautil.FilterKeys(SchemeGroupVersion.Group, selector, metautil.OverwriteKeys(nil, h.Labels, override))
}

func (h *HanaDB) ServiceLabels(alias ServiceAlias, extraLabels ...map[string]string) map[string]string {
	svcTemplate := GetServiceTemplate(h.Spec.ServiceTemplates, alias)
	return h.offshootLabels(metautil.OverwriteKeys(h.OffshootSelectors(), extraLabels...), svcTemplate.Labels)
}

func (h *HanaDB) OffshootLabels() map[string]string {
	return h.offshootLabels(h.OffshootSelectors(), nil)
}

func (h *HanaDB) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      h.ResourceFQN(),
		metautil.InstanceLabelKey:  h.Name,
		metautil.ManagedByLabelKey: SchemeGroupVersion.Group,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (h *HanaDB) OffshootPodSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		metautil.NameLabelKey:      h.ResourceFQN(),
		metautil.InstanceLabelKey:  h.Name,
		metautil.ManagedByLabelKey: SchemeGroupVersion.Group,
	}
	return metautil.OverwriteKeys(selector, extraSelectors...)
}

func (h *HanaDB) PodControllerLabels(podTemplate *ofst.PodTemplateSpec, extraLabels ...map[string]string) map[string]string {
	if podTemplate != nil && podTemplate.Controller.Labels != nil {
		return h.offshootLabels(metautil.OverwriteKeys(h.OffshootSelectors(), extraLabels...), podTemplate.Controller.Labels)
	}
	return h.offshootLabels(metautil.OverwriteKeys(h.OffshootSelectors(), extraLabels...), nil)
}

func (h *HanaDB) PodLabels(podTemplate *ofst.PodTemplateSpec, extraLabels ...map[string]string) map[string]string {
	if podTemplate != nil && podTemplate.Labels != nil {
		return h.offshootLabels(metautil.OverwriteKeys(h.OffshootSelectors(), extraLabels...), podTemplate.Labels)
	}
	return h.offshootLabels(metautil.OverwriteKeys(h.OffshootSelectors(), extraLabels...), nil)
}

func (h *HanaDB) ServiceAccountName() string {
	return h.OffshootName()
}

// Owner returns owner reference to resources
func (h *HanaDB) Owner() *metav1.OwnerReference {
	return metav1.NewControllerRef(h, SchemeGroupVersion.WithKind(h.ResourceKind()))
}

func (h *HanaDB) GetAuthSecretName() string {
	if h.Spec.AuthSecret != nil && h.Spec.AuthSecret.Name != "" {
		return h.Spec.AuthSecret.Name
	}
	return metautil.NameWithSuffix(h.OffshootName(), "auth")
}

func (h *HanaDB) GetPersistentSecrets() []string {
	var secrets []string
	secrets = append(secrets, h.GetAuthSecretName())
	return secrets
}

func (m *HanaDB) GetNameSpacedName() string {
	return m.Namespace + "/" + m.Name
}

func (h *HanaDB) DefaultPodRoleName() string {
	return metautil.NameWithSuffix(h.OffshootName(), "role")
}

func (h *HanaDB) DefaultPodRoleBindingName() string {
	return metautil.NameWithSuffix(h.OffshootName(), "rolebinding")
}

func (r *HanaDB) Finalizer() string {
	return fmt.Sprintf("%s/%s", apis.Finalizer, r.ResourceSingular())
}

func (r *HanaDB) ResourceSingular() string {
	return ResourceSingularHanaDB
}

type hanadbStatsService struct {
	*HanaDB
}

func (os hanadbStatsService) TLSConfig() *promapi.TLSConfig {
	return nil
}

func (os hanadbStatsService) GetNamespace() string {
	return os.HanaDB.GetNamespace()
}

func (os hanadbStatsService) ServiceName() string {
	return os.OffshootName() + "-stats"
}

func (os hanadbStatsService) ServiceMonitorName() string {
	return os.ServiceName()
}

func (os hanadbStatsService) ServiceMonitorAdditionalLabels() map[string]string {
	return os.OffshootLabels()
}

func (os hanadbStatsService) Path() string {
	return kubedb.DefaultStatsPath
}

func (os hanadbStatsService) Scheme() string {
	return ""
}

func (h *HanaDB) StatsService() mona.StatsAccessor {
	return &hanadbStatsService{h}
}

type hanadbApp struct {
	*HanaDB
}

func (r hanadbApp) Name() string {
	return r.HanaDB.Name
}

func (r hanadbApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", SchemeGroupVersion.Group, ResourceSingularHanaDB))
}

func (h HanaDB) AppBindingMeta() appcat.AppBindingMeta {
	return &hanadbApp{&h}
}

func (h *HanaDB) StatsServiceLabels() map[string]string {
	return h.ServiceLabels(StatsServiceAlias, map[string]string{kubedb.LabelRole: kubedb.RoleStats})
}

func (h *HanaDB) PetSetName() string {
	return h.OffshootName()
}

func (h *HanaDB) ObserverPetSetName() string {
	return fmt.Sprintf("%s-observer", h.PetSetName())
}

func (h *HanaDB) ConfigSecretName() string {
	return metautil.NameWithSuffix(h.OffshootName(), "config")
}

func (m *HanaDB) IsStandalone() bool {
	return m.Spec.Topology == nil
}

func (h *HanaDB) SetHealthCheckerDefaults() {
	if h.Spec.HealthChecker.PeriodSeconds == nil {
		h.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if h.Spec.HealthChecker.TimeoutSeconds == nil {
		h.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if h.Spec.HealthChecker.FailureThreshold == nil {
		h.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (h *HanaDB) SetDefaults(kc client.Client) {
	if h == nil {
		return
	}
	if h.Spec.StorageType == "" {
		h.Spec.StorageType = StorageTypeDurable
	}
	if h.Spec.DeletionPolicy == "" {
		h.Spec.DeletionPolicy = DeletionPolicyDelete
	}

	if h.Spec.PodTemplate == nil {
		h.Spec.PodTemplate = &ofst.PodTemplateSpec{}
	}

	var hanadbVersion catalog.HanaDBVersion
	err := kc.Get(context.TODO(), types.NamespacedName{
		Name: h.Spec.Version,
	}, &hanadbVersion)
	if err != nil {
		klog.Errorf("can't get the HanaDB version object %s for %s \n", h.Spec.Version, err.Error())
		return
	}

	if h.IsStandalone() {
		if h.Spec.Replicas == nil {
			h.Spec.Replicas = pointer.Int32P(1)
		} else if ptr.Deref(h.Spec.Replicas, 1) != 1 {
			klog.Errorf("")
			return
		}
	}

	h.setDefaultContainerSecurityContext(&hanadbVersion, h.Spec.PodTemplate)
	h.SetHealthCheckerDefaults()
}

func (m *HanaDB) setDefaultContainerSecurityContext(hanadbVersion *catalog.HanaDBVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = hanadbVersion.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.MSSQLContainerName)
	if container == nil {
		container = &core.Container{
			Name: kubedb.MSSQLContainerName,
		}
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}

	m.assignDefaultContainerSecurityContext(hanadbVersion, container.SecurityContext, true)

	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
}

func (m *HanaDB) assignDefaultContainerSecurityContext(hanadbVersion *catalog.HanaDBVersion, sc *core.SecurityContext, isMainContainer bool) {
	if sc.AllowPrivilegeEscalation == nil {
		sc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if sc.Capabilities == nil {
		if isMainContainer {
			sc.Capabilities = &core.Capabilities{
				Drop: []core.Capability{"ALL"},
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
		sc.RunAsUser = hanadbVersion.Spec.SecurityContext.RunAsUser
	}
	if sc.RunAsGroup == nil {
		sc.RunAsGroup = hanadbVersion.Spec.SecurityContext.RunAsUser
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

func (m *HanaDB) setDefaultContainerResourceLimits(podTemplate *ofst.PodTemplateSpec) {
	dbContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb.MSSQLContainerName)
	if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&dbContainer.Resources, kubedb.DefaultResourcesHanaDB)
	}
}

func (m *HanaDB) ReplicasAreReady(lister pslister.PetSetLister) (bool, string, error) {
	// Desire number of petSets
	expectedItems := 1
	return checkReplicasOfPetSet(lister.PetSets(m.Namespace), labels.SelectorFromSet(m.OffshootLabels()), expectedItems)
}
