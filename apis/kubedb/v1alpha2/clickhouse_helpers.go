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

	"gomodules.xyz/pointer"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

type ClickhouseApp struct {
	*ClickHouse
}

func (r *ClickHouse) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralClickhouse))
}

func (c *ClickHouse) AppBindingMeta() appcat.AppBindingMeta {
	return &ClickhouseApp{c}
}

func (r ClickhouseApp) Name() string {
	return r.ClickHouse.Name
}

func (r ClickhouseApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularClickhouse))
}

// Owner returns owner reference to resources
func (c *ClickHouse) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(c, SchemeGroupVersion.WithKind(c.ResourceKind()))
}

func (c *ClickHouse) ResourceKind() string {
	return ResourceKindClickhouse
}

func (c *ClickHouse) ServiceName() string {
	return c.OffshootName()
}

func (c *ClickHouse) OffshootName() string {
	return c.Name
}

func (c *ClickHouse) OffshootLabels() map[string]string {
	return c.offshootLabels(c.OffshootSelectors(), nil)
}

func (c *ClickHouse) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, c.Labels, override))
}

func (c *ClickHouse) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      c.ResourceFQN(),
		meta_util.InstanceLabelKey:  c.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (c *ClickHouse) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", c.ResourcePlural(), kubedb.GroupName)
}

func (c *ClickHouse) ResourcePlural() string {
	return ResourcePluralClickhouse
}

func (c *ClickHouse) GoverningServiceName() string {
	return meta_util.NameWithSuffix(c.ServiceName(), "pods")
}

func (c *ClickHouse) GetAuthSecretName() string {
	if c.Spec.AuthSecret != nil && c.Spec.AuthSecret.Name != "" {
		return c.Spec.AuthSecret.Name
	}
	return c.DefaultUserCredSecretName("admin")
}

func (c *ClickHouse) DefaultUserCredSecretName(username string) string {
	return meta_util.NameWithSuffix(c.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (c *ClickHouse) PVCName(alias string) string {
	return meta_util.NameWithSuffix(c.Name, alias)
}

func (c *ClickHouse) PetSetName() string {
	return c.OffshootName()
}

func (c *ClickHouse) PodLabels(extraLabels ...map[string]string) map[string]string {
	return c.offshootLabels(meta_util.OverwriteKeys(c.OffshootSelectors(), extraLabels...), c.Spec.PodTemplate.Labels)
}

func (c *ClickHouse) GetConnectionScheme() string {
	scheme := "http"
	if c.Spec.EnableSSL {
		scheme = "https"
	}
	return scheme
}

func (c *ClickHouse) SetHealthCheckerDefaults() {
	if c.Spec.HealthChecker.PeriodSeconds == nil {
		c.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if c.Spec.HealthChecker.TimeoutSeconds == nil {
		c.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if c.Spec.HealthChecker.FailureThreshold == nil {
		c.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (c *ClickHouse) PrimaryServiceDNS() string {
	return fmt.Sprintf("%s.%s.svc", c.ServiceName(), c.Namespace)
}

func (c *ClickHouse) Finalizer() string {
	return fmt.Sprintf("%s/%s", apis.Finalizer, c.ResourceSingular())
}

func (c *ClickHouse) ResourceSingular() string {
	return ResourceSingularClickhouse
}

func (c *ClickHouse) GetPersistentSecrets() []string {
	var secrets []string
	secrets = append(secrets, c.GetAuthSecretName())
	return secrets
}

func (c *ClickHouse) SetDefaults() {
	if c.Spec.Replicas == nil {
		c.Spec.Replicas = pointer.Int32P(1)
	}
	if c.Spec.TerminationPolicy == "" {
		c.Spec.TerminationPolicy = TerminationPolicyDelete
	}
	if c.Spec.StorageType == "" {
		c.Spec.StorageType = StorageTypeDurable
	}

	var chVersion catalog.ClickHouseVersion
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: c.Spec.Version,
	}, &chVersion)
	if err != nil {
		klog.Errorf("can't get the clickhouse version object %s for %s \n", err.Error(), c.Spec.Version)
		return
	}
	c.setDefaultContainerSecurityContext(&chVersion, &c.Spec.PodTemplate)

	dbContainer := coreutil.GetContainerByName(c.Spec.PodTemplate.Spec.Containers, ClickHouseContainerName)
	if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&dbContainer.Resources, DefaultResources)
	}
	c.SetHealthCheckerDefaults()
}

func (r *ClickHouse) setDefaultContainerSecurityContext(chVersion *catalog.ClickHouseVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = chVersion.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, ClickHouseContainerName)
	if container == nil {
		container = &core.Container{
			Name: ClickHouseContainerName,
		}
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}
	r.assignDefaultContainerSecurityContext(chVersion, container.SecurityContext)

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, ClickHouseInitContainerName)
	if initContainer == nil {
		initContainer = &core.Container{
			Name: ClickHouseInitContainerName,
		}
		podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &core.SecurityContext{}
	}
	r.assignDefaultInitContainerSecurityContext(chVersion, initContainer.SecurityContext)
}

func (r *ClickHouse) assignDefaultContainerSecurityContext(chVersion *catalog.ClickHouseVersion, rc *core.SecurityContext) {
	if rc.AllowPrivilegeEscalation == nil {
		rc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if rc.Capabilities == nil {
		rc.Capabilities = &core.Capabilities{
			Drop: []core.Capability{"ALL"},
		}
	}
	if rc.RunAsNonRoot == nil {
		rc.RunAsNonRoot = pointer.BoolP(true)
	}
	if rc.RunAsUser == nil {
		rc.RunAsUser = chVersion.Spec.SecurityContext.RunAsUser
	}
	if rc.SeccompProfile == nil {
		rc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

func (r *ClickHouse) assignDefaultInitContainerSecurityContext(chVersion *catalog.ClickHouseVersion, rc *core.SecurityContext) {
	if rc.AllowPrivilegeEscalation == nil {
		rc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if rc.Capabilities == nil {
		rc.Capabilities = &core.Capabilities{
			Drop: []core.Capability{"ALL"},
		}
	}
	if rc.RunAsNonRoot == nil {
		rc.RunAsNonRoot = pointer.BoolP(true)
	}
	if rc.RunAsUser == nil {
		rc.RunAsUser = chVersion.Spec.SecurityContext.RunAsUser
	}
	if rc.SeccompProfile == nil {
		rc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

// returns the CertSecretVolumeName
// Values will be like: client-certs, server-certs etc.
//func (r *ClickHouse) CertSecretVolumeName(alias ClickHouseCertificateAlias) string {
//	return string(alias) + "-certs"
//}
