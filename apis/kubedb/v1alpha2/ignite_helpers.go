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
	"sync"

	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"

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
	"kubedb.dev/apimachinery/crds"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	once          sync.Once
	DefaultClient client.Client
)

func SetDefaultClient(kc client.Client) {
	once.Do(func() {
		DefaultClient = kc
	})
}

func (i *Ignite) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralIgnite))
}

func (i *Ignite) AsOwner() *meta.OwnerReference {
	return meta.NewControllerRef(i, SchemeGroupVersion.WithKind(ResourceKindIgnite))
}

func (i *Ignite) ResourceKind() string {
	return ResourceKindIgnite
}

func (i *Ignite) ResourceSingular() string {
	return ResourceSingularIgnite
}

func (i *Ignite) ResourcePlural() string {
	return ResourcePluralIgnite
}

func (i *Ignite) Finalizer() string {
	return fmt.Sprintf("%s/%s", apis.Finalizer, i.ResourceSingular())
}

func (i *Ignite) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", i.ResourcePlural(), kubedb.GroupName)
}

func (i *Ignite) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      i.ResourceFQN(),
		meta_util.InstanceLabelKey:  i.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (i *Ignite) OffshootName() string {
	return i.Name
}

func (i *Ignite) GetAuthSecretName() string {
	if i.Spec.AuthSecret != nil && i.Spec.AuthSecret.Name != "" {
		return i.Spec.AuthSecret.Name
	}
	return i.DefaultAuthSecretName()
}

func (i *Ignite) GetPersistentSecrets() []string {
	var secrets []string
	secrets = append(secrets, i.GetAuthSecretName())
	return secrets
}

// Owner returns owner reference to resources
func (i *Ignite) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(i, SchemeGroupVersion.WithKind(i.ResourceKind()))
}

func (i *Ignite) SetDefaults() {
	if i.Spec.Replicas == nil {
		i.Spec.Replicas = pointer.Int32P(1)
	}

	if i.Spec.DeletionPolicy == "" {
		i.Spec.DeletionPolicy = DeletionPolicyDelete
	}

	if i.Spec.StorageType == "" {
		i.Spec.StorageType = StorageTypeDurable
	}

	var igVersion catalog.IgniteVersion
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: i.Spec.Version,
	}, &igVersion)
	if err != nil {
		klog.Errorf("can't get the ignite version object %s for %s \n", err.Error(), i.Spec.Version)
		return
	}

	i.setDefaultContainerSecurityContext(&igVersion, &i.Spec.PodTemplate)

	dbContainer := coreutil.GetContainerByName(i.Spec.PodTemplate.Spec.Containers, "ignite")
	if dbContainer != nil && (dbContainer.Resources.Requests == nil || dbContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&dbContainer.Resources, kubedb.DefaultResources)
	}

	i.SetHealthCheckerDefaults()

	/*	if i.Spec.Monitor != nil {
		if i.Spec.Monitor.Prometheus == nil {
			i.Spec.Monitor.Prometheus = &mona.PrometheusSpec{}
		}
		if i.Spec.Monitor.Prometheus != nil && i.Spec.Monitor.Prometheus.Exporter.Port == 0 {
			// i.Spec.Monitor.Prometheus.Exporter.Port =
		}
		i.Spec.Monitor.SetDefaults()
	}*/
}

func (i *Ignite) SetHealthCheckerDefaults() {
	if i.Spec.HealthChecker.PeriodSeconds == nil {
		i.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(10)
	}
	if i.Spec.HealthChecker.TimeoutSeconds == nil {
		i.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if i.Spec.HealthChecker.FailureThreshold == nil {
		i.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (i *Ignite) setDefaultContainerSecurityContext(igVersion *catalog.IgniteVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = igVersion.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, "ignite")
	if container == nil {
		container = &core.Container{
			Name: "ignite",
		}
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
	}

	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}
	i.assignDefaultContainerSecurityContext(igVersion, container.SecurityContext)

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, "ignite-init")
	if initContainer == nil {
		initContainer = &core.Container{
			Name: "ignite-init",
		}
		podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &core.SecurityContext{}
	}
	i.assignDefaultContainerSecurityContext(igVersion, initContainer.SecurityContext)
}

func (i *Ignite) assignDefaultContainerSecurityContext(igVersion *catalog.IgniteVersion, rc *core.SecurityContext) {
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
		rc.RunAsUser = igVersion.Spec.SecurityContext.RunAsUser
	}
	if rc.SeccompProfile == nil {
		rc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

func (i *Ignite) PetSetName() string {
	return i.OffshootName()
}

func (i *Ignite) ServiceName() string { return i.OffshootName() }

func (i *Ignite) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, i.Labels, override))
}

func (i *Ignite) OffshootLabels() map[string]string {
	return i.offshootLabels(i.OffshootSelectors(), nil)
}

func (i *Ignite) GoverningServiceName() string {
	return meta_util.NameWithSuffix(i.ServiceName(), "pods")
}

func (i *Ignite) DefaultAuthSecretName() string {
	return meta_util.NameWithSuffix(i.OffshootName(), "auth")
}

func (i *Ignite) ServiceAccountName() string {
	return i.OffshootName()
}

func (i *Ignite) DefaultPodRoleName() string {
	return meta_util.NameWithSuffix(i.OffshootName(), "role")
}

func (i *Ignite) DefaultPodRoleBindingName() string {
	return meta_util.NameWithSuffix(i.OffshootName(), "rolebinding")
}

type IgniteApp struct {
	*Ignite
}

func (i *IgniteApp) Name() string {
	return i.Ignite.Name
}

func (i IgniteApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularIgnite))
}

func (i *Ignite) AppBindingMeta() appcat.AppBindingMeta {
	return &IgniteApp{i}
}

func (i *Ignite) GetConnectionScheme() string {
	scheme := "http"
	return scheme
}

func (i *Ignite) PodLabels(extraLabels ...map[string]string) map[string]string {
	return i.offshootLabels(meta_util.OverwriteKeys(i.OffshootSelectors(), extraLabels...), i.Spec.PodTemplate.Labels)
}

func (i *Ignite) ConfigSecretName() string {
	return meta_util.NameWithSuffix(i.OffshootName(), "config")
}

func (i *Ignite) PVCName(alias string) string {
	return meta_util.NameWithSuffix(i.Name, alias)
}

func (i *Ignite) Address() string {
	klog.Infof("%v.%v.svc", i.Name, i.Namespace)
	return fmt.Sprintf("%v.%v.svc.cluster.local", i.Name, i.Namespace)
}
