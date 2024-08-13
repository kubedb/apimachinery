package v1alpha2

import (
	"context"
	"fmt"
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
	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/crds"
	catalog "kubedb.dev/cassandra/api/catalog/v1alpha1"
	kubedb_local "kubedb.dev/cassandra/api/kubedb"
	"strconv"
	"strings"
)

type CassandraApp struct {
	*Cassandra
}

func (r *Cassandra) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(v1alpha2.SchemeGroupVersion.WithResource(ResourcePluralCassandra))
}

func (r *Cassandra) AppBindingMeta() appcat.AppBindingMeta {
	return &CassandraApp{r}
}

func (r CassandraApp) Name() string {
	return r.Cassandra.Name
}

func (r CassandraApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularCassandra))
}

// Owner returns owner reference to resources
func (r *Cassandra) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(r, v1alpha2.SchemeGroupVersion.WithKind(r.ResourceKind()))
}

func (r *Cassandra) ResourceKind() string {
	return ResourceKindCassandra
}

func (r *Cassandra) OffshootName() string {
	return r.Name
}

func (r *Cassandra) OffshootRackName(value string) string {
	return meta_util.NameWithSuffix(r.OffshootName(), value)
}

func (r *Cassandra) OffshootRackPetSetName(rackName string) string {
	rack := meta_util.NameWithSuffix("rack", rackName)
	return meta_util.NameWithSuffix(r.OffshootName(), rack)
}

func (r *Cassandra) OffshootLabels() map[string]string {
	return r.offshootLabels(r.OffshootSelectors(), nil)
}

func (r *Cassandra) OffshootRackLabels(petSetName string) map[string]string {
	return r.offshootLabels(r.OffshootRackSelectors(petSetName), nil)
}

func (r *Cassandra) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, r.Labels, override))
}

func (r *Cassandra) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      r.ResourceFQN(),
		meta_util.InstanceLabelKey:  r.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (r *Cassandra) OffshootRackSelectors(petSetName string, extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      r.ResourceFQN(),
		meta_util.InstanceLabelKey:  r.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
		meta_util.PartOfLabelKey:    petSetName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (r *Cassandra) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", r.ResourcePlural(), kubedb.GroupName)
}

func (r *Cassandra) ResourcePlural() string {
	return ResourcePluralCassandra
}

func (r *Cassandra) ServiceName() string {
	return r.OffshootName()
}

func (r *Cassandra) PrimaryServiceDNS() string {
	return fmt.Sprintf("%s.%s.svc", r.ServiceName(), r.Namespace)
}

func (r *Cassandra) GoverningServiceName() string {
	return meta_util.NameWithSuffix(r.ServiceName(), "pods")
}

func (r *Cassandra) RackGoverningServiceName(name string) string {
	return meta_util.NameWithSuffix(name, "pods")
}

func (r *Cassandra) RackGoverningServiceDNS(petSetName string, replicaNo int) string {
	return fmt.Sprintf("%s-%d.%s.%s.svc", petSetName, replicaNo, r.RackGoverningServiceName(petSetName), r.GetNamespace())
}

func (r *Cassandra) GetAuthSecretName() string {
	if r.Spec.AuthSecret != nil && r.Spec.AuthSecret.Name != "" {
		return r.Spec.AuthSecret.Name
	}
	return r.DefaultUserCredSecretName("admin")
}

func (r *Cassandra) ConfigSecretName() string {
	return meta_util.NameWithSuffix(r.OffshootName(), "config")
}

func (r *Cassandra) DefaultUserCredSecretName(username string) string {
	return meta_util.NameWithSuffix(r.Name, strings.ReplaceAll(fmt.Sprintf("%s-cred", username), "_", "-"))
}

func (r *Cassandra) PVCName(alias string) string {
	return meta_util.NameWithSuffix(r.Name, alias)
}

func (r *Cassandra) PetSetName() string {
	return r.OffshootName()
}

func (r *Cassandra) PodLabels(extraLabels ...map[string]string) map[string]string {
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootSelectors(), extraLabels...), r.Spec.PodTemplate.Labels)
}

func (r *Cassandra) RackPodLabels(petSetName string, labels map[string]string, extraLabels ...map[string]string) map[string]string {
	return r.offshootLabels(meta_util.OverwriteKeys(r.OffshootRackSelectors(petSetName), extraLabels...), labels)
}

func (r *Cassandra) GetConnectionScheme() string {
	scheme := "http"
	return scheme
}

func (r *Cassandra) SetHealthCheckerDefaults() {
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

func (r *Cassandra) Finalizer() string {
	return fmt.Sprintf("%s/%s", apis.Finalizer, r.ResourceSingular())
}

func (r *Cassandra) ResourceSingular() string {
	return ResourceSingularCassandra
}

func (r *Cassandra) SetDefaults() {
	var chVersion catalog.CassandraVersion
	err := v1alpha2.DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: r.Spec.Version,
	}, &chVersion)
	if err != nil {
		klog.Errorf("can't get the cassandra version object %s for %s \n", err.Error(), r.Spec.Version)
		return
	}
	if r.Spec.Topology != nil {
		rackName := map[string]bool{}
		racks := r.Spec.Topology.Rack
		for index, rack := range racks {
			if rack.Replicas == nil {
				rack.Replicas = pointer.Int32P(1)
			}
			if rack.Name == "" {
				for i := 1; ; i += 1 {
					rack.Name = r.OffshootRackName(strconv.Itoa(i))
					if !rackName[rack.Name] {
						rackName[rack.Name] = true
						break
					}
				}
			} else {
				rackName[rack.Name] = true
			}
			if rack.StorageType == "" {
				rack.StorageType = v1alpha2.StorageTypeDurable
			}

			if rack.PodTemplate == nil {
				rack.PodTemplate = &ofst.PodTemplateSpec{}
			}

			dbContainer := coreutil.GetContainerByName(rack.PodTemplate.Spec.Containers, kubedb_local.CassandraContainerName)
			if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
				apis.SetDefaultResourceLimits(&dbContainer.Resources, kubedb.DefaultResources)
			}
			r.setDefaultContainerSecurityContext(&chVersion, rack.PodTemplate)
			racks[index] = rack
		}
		r.Spec.Topology.Rack = racks
	} else {
		if r.Spec.Replicas == nil {
			r.Spec.Replicas = pointer.Int32P(1)
		}
		if r.Spec.DeletionPolicy == "" {
			r.Spec.DeletionPolicy = v1alpha2.TerminationPolicyDelete
		}
		if r.Spec.StorageType == "" {
			r.Spec.StorageType = v1alpha2.StorageTypeDurable
		}

		if r.Spec.PodTemplate == nil {
			r.Spec.PodTemplate = &ofst.PodTemplateSpec{}
		}
		r.setDefaultContainerSecurityContext(&chVersion, r.Spec.PodTemplate)
		dbContainer := coreutil.GetContainerByName(r.Spec.PodTemplate.Spec.Containers, kubedb_local.CassandraContainerName)
		if dbContainer != nil && (dbContainer.Resources.Requests == nil && dbContainer.Resources.Limits == nil) {
			apis.SetDefaultResourceLimits(&dbContainer.Resources, kubedb.DefaultResources)
		}
		r.SetHealthCheckerDefaults()
	}
}

func (r *Cassandra) setDefaultContainerSecurityContext(chVersion *catalog.CassandraVersion, podTemplate *ofst.PodTemplateSpec) {
	if podTemplate == nil {
		return
	}
	if podTemplate.Spec.SecurityContext == nil {
		podTemplate.Spec.SecurityContext = &core.PodSecurityContext{}
	}
	if podTemplate.Spec.SecurityContext.FSGroup == nil {
		podTemplate.Spec.SecurityContext.FSGroup = chVersion.Spec.SecurityContext.RunAsUser
	}

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, kubedb_local.CassandraContainerName)
	if container == nil {
		container = &core.Container{
			Name: kubedb_local.CassandraContainerName,
		}
		podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &core.SecurityContext{}
	}
	r.assignDefaultContainerSecurityContext(chVersion, container.SecurityContext)

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, kubedb_local.CassandraInitContainerName)
	if initContainer == nil {
		initContainer = &core.Container{
			Name: kubedb_local.CassandraInitContainerName,
		}
		podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &core.SecurityContext{}
	}
	r.assignDefaultContainerSecurityContext(chVersion, initContainer.SecurityContext)
}

func (r *Cassandra) assignDefaultContainerSecurityContext(chVersion *catalog.CassandraVersion, rc *core.SecurityContext) {
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
