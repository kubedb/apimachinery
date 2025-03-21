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
	kube "kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	"gomodules.xyz/pointer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
	coreutil "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HazelcastApp struct {
	*Hazelcast
}

func (h *Hazelcast) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralHazelcast))
}

func (h *Hazelcast) OffshootName() string {
	return h.Name
}

func (h *Hazelcast) HazelcastSecretName(suffix string) string {
	return strings.Join([]string{h.Name, suffix}, "-")
}

func (h *Hazelcast) StatefulSetName() string {
	return h.Name
}

func (h *Hazelcast) GetAuthSecretName() string {
	if h.Spec.AuthSecret != nil && h.Spec.AuthSecret.Name != "" {
		return h.Spec.AuthSecret.Name
	}
	return meta_util.NameWithSuffix(h.OffshootName(), "auth")
}

// Owner returns owner reference to resources
func (h *Hazelcast) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(h, SchemeGroupVersion.WithKind(h.ResourceKind()))
}

func (h *Hazelcast) ServiceAccountName() string {
	return h.OffshootName()
}

func (h *Hazelcast) ResourceKind() string {
	return ResourceKindHazelcast
}

func (h *Hazelcast) ResourcePlural() string {
	return ResourcePluralHazelcast
}

func (h *Hazelcast) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", h.ResourcePlural(), kube.GroupName)
}

func (h *Hazelcast) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return h.offshootLabels(meta_util.OverwriteKeys(h.OffshootSelectors(), extraLabels...), h.Spec.PodTemplate.Controller.Labels)
}

func (h *Hazelcast) PVCName(alias string) string {
	return meta_util.NameWithSuffix(h.Name, alias)
}

func (h *Hazelcast) PodLabels(extraLabels ...map[string]string) map[string]string {
	return h.offshootLabels(meta_util.OverwriteKeys(h.OffshootSelectors(), extraLabels...), h.Spec.PodTemplate.Labels)
}

func (h *Hazelcast) OffshootLabels() map[string]string {
	return h.offshootLabels(h.OffshootSelectors(), nil)
}

func (h *Hazelcast) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kube.ComponentDatabase
	return meta_util.FilterKeys(kube.GroupName, selector, meta_util.OverwriteKeys(nil, h.Labels, override))
}

func (h *Hazelcast) GoverningServiceName() string {
	return meta_util.NameWithSuffix(h.ServiceName(), "pods")
}

func (h *Hazelcast) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      h.ResourceFQN(),
		meta_util.InstanceLabelKey:  h.Name,
		meta_util.ManagedByLabelKey: kube.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (h *Hazelcast) ServiceName() string {
	return h.OffshootName()
}

func (h *Hazelcast) SetHealthCheckerDefaults() {
	if h.Spec.HealthChecker.PeriodSeconds == nil {
		h.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(20)
	}
	if h.Spec.HealthChecker.TimeoutSeconds == nil {
		h.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if h.Spec.HealthChecker.FailureThreshold == nil {
		h.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (h *Hazelcast) SetDefaults(kc client.Client) {
	if h.Spec.DeletionPolicy == "" {
		h.Spec.DeletionPolicy = DeletionPolicyDelete
	}

	if h.Spec.StorageType == "" {
		h.Spec.StorageType = StorageTypeDurable
	}

	var hzVersion catalog.HazelcastVersion
	err := kc.Get(context.TODO(), types.NamespacedName{
		Name: h.Spec.Version,
	}, &hzVersion)
	if err != nil {
		klog.Errorf("can't get the solr version object %s for %s \n", err.Error(), h.Spec.Version)
		return
	}
	if h.Spec.Replicas == nil {
		h.Spec.Replicas = pointer.Int32P(1)
	}
	if h.Spec.PodTemplate.Spec.SecurityContext == nil {
		h.Spec.PodTemplate.Spec.SecurityContext = &v1.PodSecurityContext{}
	}
	h.Spec.PodTemplate.Spec.SecurityContext.FSGroup = hzVersion.Spec.SecurityContext.RunAsUser
	h.setDefaultContainerSecurityContext(&hzVersion, &h.Spec.PodTemplate)
	h.setDefaultContainerResourceLimits(&h.Spec.PodTemplate)
}

func (h *Hazelcast) setDefaultContainerSecurityContext(hzVersion *catalog.HazelcastVersion, podTemplate *ofst.PodTemplateSpec) {
	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, "hazelcast-init")
	if initContainer == nil {
		initContainer = &v1.Container{
			Name: "hazelcast-init",
		}
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &v1.SecurityContext{}
	}
	h.assignDefaultContainerSecurityContext(hzVersion, initContainer.SecurityContext)
	podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)
	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, "hazelcast")
	if container == nil {
		container = &v1.Container{
			Name: "hazelcast",
		}
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &v1.SecurityContext{}
	}
	h.assignDefaultContainerSecurityContext(hzVersion, container.SecurityContext)
	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
}

func (h *Hazelcast) assignDefaultContainerSecurityContext(hzVersion *catalog.HazelcastVersion, sc *v1.SecurityContext) {
	if sc.AllowPrivilegeEscalation == nil {
		sc.AllowPrivilegeEscalation = pointer.BoolP(false)
	}
	if sc.Capabilities == nil {
		sc.Capabilities = &v1.Capabilities{
			Drop: []v1.Capability{"ALL"},
		}
	}
	if sc.RunAsNonRoot == nil {
		sc.RunAsNonRoot = pointer.BoolP(true)
	}
	if sc.RunAsUser == nil {
		sc.RunAsUser = hzVersion.Spec.SecurityContext.RunAsUser
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

var (
	DefaultResourcesMemoryIntensive = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(".500"),
			v1.ResourceMemory: resource.MustParse("1.5Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("1.5Gi"),
		},
	}
	DefaultInitContainerResource = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(".200"),
			v1.ResourceMemory: resource.MustParse("256Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
)

func (h *Hazelcast) setDefaultContainerResourceLimits(podTemplate *ofst.PodTemplateSpec) {
	dbContainer := coreutil.GetContainerByName(podTemplate.Spec.Containers, "hazelcast")
	if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&dbContainer.Resources, DefaultResourcesMemoryIntensive)
	}

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, "hazelcast-init")
	if initContainer != nil && (initContainer.Resources.Requests == nil && initContainer.Resources.Limits == nil) {
		apis.SetDefaultResourceLimits(&initContainer.Resources, DefaultInitContainerResource)
	}
}

func (h *Hazelcast) GetPersistentSecrets() []string {
	if h == nil {
		return nil
	}

	var secrets []string
	secrets = append(secrets, h.GetAuthSecretName())

	return secrets
}

func (h HazelcastApp) Name() string {
	return h.Hazelcast.Name
}

func (h HazelcastApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kube.GroupName, ResourceSingularHazelcast))
}

func (h *Hazelcast) AppBindingMeta() appcat.AppBindingMeta {
	return &HazelcastApp{h}
}

func (h *Hazelcast) GetConnectionScheme() string {
	scheme := "http"
	return scheme
}
