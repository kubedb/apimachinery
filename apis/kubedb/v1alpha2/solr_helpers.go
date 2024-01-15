package v1alpha2

import (
	"context"
	"fmt"
	"gomodules.xyz/pointer"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	coreutil "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/policy/secomp"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
	"kubedb.dev/apimachinery/apis"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	"reflect"
	"sort"
	"strings"
)

func (s *Solr) StatefulsetName(suffix string) string {
	sts := []string{s.Name}
	if suffix != "" {
		sts = append(sts, suffix)
	}
	return strings.Join(sts, "-")
}

// Owner returns owner reference to resources
func (s *Solr) Owner() *meta.OwnerReference {
	return meta.NewControllerRef(s, SchemeGroupVersion.WithKind(s.ResourceKind()))
}

func (s *Solr) ResourceKind() string {
	return ResourceKindSolr
}

func (s *Solr) GoverningServiceName() string {
	return meta_util.NameWithSuffix(s.ServiceName(), "pods")
}

func (s *Solr) OverseerDiscoveryServiceName() string {
	return meta_util.NameWithSuffix(s.ServiceName(), "overseer")
}

func (s *Solr) ServiceAccountName() string { return s.OffshootName() }

func (s *Solr) DefaultPodRoleName() string {
	return meta_util.NameWithSuffix(s.OffshootName(), "role")
}

func (s *Solr) DefaultPodRoleBindingName() string {
	return meta_util.NameWithSuffix(s.OffshootName(), "rolebinding")
}

func (s *Solr) ServiceName() string {
	return s.OffshootName()
}

func (s *Solr) SolrSecretName(suffix string) string {
	return strings.Join([]string{s.Name, suffix}, "-")
}

func (s *Solr) SolrSecretKey() string {
	return SolrSecretKey
}

func (s *Solr) Merge(opt map[string]string) map[string]string {
	if len(s.Spec.SolrOpts) == 0 {
		return opt
	}
	for _, y := range s.Spec.SolrOpts {
		sr := strings.Split(y, "=")
		_, ok := opt[sr[0]]
		if !ok || sr[0] != "-Dsolr.node.roles" {
			opt[sr[0]] = sr[1]
		}
	}
	return opt
}

func (s *Solr) Append(opt map[string]string) string {
	fl := 0
	as := ""
	for x, y := range opt {
		if fl == 1 {
			as += " "
		}
		as += fmt.Sprintf("%s=%s", x, y)
		fl = 1

	}
	return as
}

func GenerateAdditionalLibXMLPart(solrModules []string) string {
	libs := make(map[string]bool, 0)

	// Placeholder for users to specify libs via sysprop
	libs[SysPropLibPlaceholder] = true

	// Add all module library locations
	if len(solrModules) > 0 {
		libs[DistLibs] = true
	}
	for _, module := range solrModules {
		libs[fmt.Sprintf(ContribLibs, module)] = true
	}

	// Add all custom library locations
	//for _, libPath := range additionalLibs {
	//	libs[libPath] = true
	//}

	libList := make([]string, 0)
	for lib := range libs {
		libList = append(libList, lib)
	}
	sort.Strings(libList)
	return fmt.Sprintf("<str name=\"sharedLib\">%s</str>", strings.Join(libList, ","))
}

func getXMLConfigElement(name string, key string, value interface{}, kind string) string {
	if key == "" {
		key = name
	}
	pp := reflect.TypeOf(value).Kind()
	if pp == reflect.Int {
		return fmt.Sprintf("<%s name=\"%s\">${%s:%d}</%s>\n", kind, name, key, value, kind)
	} else if pp == reflect.String {
		return fmt.Sprintf("<%s name=\"%s\">${%s:%s}</%s>\n", kind, name, key, value, kind)
	} else if pp == reflect.Bool {
		return fmt.Sprintf("<%s name=\"%s\">${%s:%t}</%s>\n", kind, name, key, value, kind)
	}
	return ""
}

func intend(ss string, level int) string {
	level *= 2
	for level > 0 {
		level--
		ss += " "
	}
	return ss
}

func rec(mp map[string]interface{}, level int) string {
	ss := ""
	for x, y := range mp {
		kind := reflect.TypeOf(y).Kind()
		if kind == reflect.Int {
			val := y.(int)
			ss = intend(ss, level)
			ss = ss + getXMLConfigElement(x, Keys[x], val, "int")
		} else if kind == reflect.String {
			val := y.(string)
			ss = intend(ss, level)
			ss = ss + getXMLConfigElement(x, Keys[x], val, "str")
		} else if kind == reflect.Bool {
			val := y.(bool)
			ss = intend(ss, level)
			ss = ss + getXMLConfigElement(x, Keys[x], val, "bool")
		} else {
			ss = intend(ss, level)
			if x == "shardHandlerFactory" {
				ss = ss + fmt.Sprintf("<%s name=\"shardHandlerFactory\" class=\"HttpShardHandlerFactory\">\n", x)
			} else {
				ss = ss + fmt.Sprintf("<%s>\n", x)
			}
			//	fmt.Println(x, y)
			v, ok := y.(map[string]interface{})
			if !ok {
				fmt.Println("failed to decode")
			}
			ss += rec(v, level+1)
			ss = intend(ss, level)
			ss = ss + fmt.Sprintf("</%s>\n", x)
		}
	}
	return ss
}

func (s *Solr) SolrSecret() string {
	ss := "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n<solr>\n" + "  <str name=\"coreRootDirectory\">/var/solr/data</str>\n  %s\n" + rec(SolrConf, 1)
	ss += "  <metrics enabled=\"${metricsEnabled:true}\"/>\n"
	ss += "</solr>\n"
	return fmt.Sprintf(ss, GenerateAdditionalLibXMLPart(s.Spec.SolrModules))
}

func (s *Solr) OffshootName() string {
	return s.Name
}

func (s *Solr) PodControllerLabels(extraLabels ...map[string]string) map[string]string {
	return s.offshootLabels(meta_util.OverwriteKeys(s.OffshootSelectors(), extraLabels...), s.Spec.PodTemplate.Controller.Labels)
}

func (s *Solr) PodLabels(extraLabels ...map[string]string) map[string]string {
	return s.offshootLabels(meta_util.OverwriteKeys(s.OffshootSelectors(), extraLabels...), s.Spec.PodTemplate.Labels)
}

func (s *Solr) OffshootLabels() map[string]string {
	return s.offshootLabels(s.OffshootSelectors(), nil)
}

func (s *Solr) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, s.Labels, override))
}

func (s *Solr) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      s.ResourceFQN(),
		meta_util.InstanceLabelKey:  s.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (s *Solr) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", s.ResourcePlural(), kubedb.GroupName)
}

func (s *Solr) ResourcePlural() string {
	return ResourcePluralSolr
}

type SolrApp struct {
	*Solr
}

func (s SolrApp) Name() string {
	return s.Solr.Name
}

func (s SolrApp) Type() appcat.AppType {
	return appcat.AppType(fmt.Sprintf("%s/%s", kubedb.GroupName, ResourceSingularSolr))
}

func (s *Solr) AppBindingMeta() appcat.AppBindingMeta {
	return &SolrApp{s}
}

func (s *Solr) GetConnectionScheme() string {
	scheme := "http"
	if s.Spec.EnableSSL {
		scheme = "https"
	}
	return scheme
}

func (s *Solr) PVCName(alias string) string {
	return meta_util.NameWithSuffix(s.Name, alias)
}

func (s *Solr) SetDefaults() {
	if s.Spec.TerminationPolicy == "" {
		s.Spec.TerminationPolicy = TerminationPolicyDelete
	}

	if s.Spec.StorageType == "" {
		s.Spec.StorageType = StorageTypeDurable
	}

	if s.Spec.AuthSecret == nil {
		s.Spec.AuthSecret = &v1.LocalObjectReference{
			Name: s.SolrSecretName("admin-cred"),
		}
	}

	if s.Spec.AuthConfigSecret == nil {
		s.Spec.AuthConfigSecret = &v1.LocalObjectReference{
			Name: s.SolrSecretName("auth-config"),
		}
	}

	var slVersion catalog.SolrVersion
	err := DefaultClient.Get(context.TODO(), types.NamespacedName{
		Name: s.Spec.Version,
	}, &slVersion)

	if err != nil {
		klog.Errorf("can't get the solr version object %s for %s \n", err.Error(), s.Spec.Version)
		return
	}

	if s.Spec.Topology != nil {
		if s.Spec.Topology.Data != nil {
			if s.Spec.Topology.Data.Suffix == "" {
				s.Spec.Topology.Data.Suffix = string(SolrNodeRoleData)
			}
			if s.Spec.Topology.Data.Replicas == nil {
				s.Spec.Topology.Data.Replicas = pointer.Int32P(1)
			}

			if s.Spec.Topology.Data.PodTemplate.Spec.SecurityContext == nil {
				s.Spec.Topology.Data.PodTemplate.Spec.SecurityContext = &v1.PodSecurityContext{FSGroup: slVersion.Spec.SecurityContext.RunAsUser}
			}
			s.Spec.Topology.Data.PodTemplate.Spec.SecurityContext.FSGroup = slVersion.Spec.SecurityContext.RunAsUser
			s.setDefaultContainerSecurityContext(&slVersion, &s.Spec.Topology.Data.PodTemplate)
			s.setDefaultInitContainerSecurityContext(&slVersion, &s.Spec.Topology.Data.PodTemplate)
		}

		if s.Spec.Topology.Overseer != nil {
			if s.Spec.Topology.Overseer.Suffix == "" {
				s.Spec.Topology.Overseer.Suffix = string(SolrNodeRoleOverseer)
			}
			if s.Spec.Topology.Overseer.Replicas == nil {
				s.Spec.Topology.Overseer.Replicas = pointer.Int32P(1)
			}

			if s.Spec.Topology.Overseer.PodTemplate.Spec.SecurityContext == nil {
				s.Spec.Topology.Overseer.PodTemplate.Spec.SecurityContext = &v1.PodSecurityContext{FSGroup: slVersion.Spec.SecurityContext.RunAsUser}
			}
			s.Spec.Topology.Overseer.PodTemplate.Spec.SecurityContext.FSGroup = slVersion.Spec.SecurityContext.RunAsUser
			s.setDefaultContainerSecurityContext(&slVersion, &s.Spec.Topology.Overseer.PodTemplate)
			s.setDefaultInitContainerSecurityContext(&slVersion, &s.Spec.Topology.Overseer.PodTemplate)
		}

		if s.Spec.Topology.Coordinator != nil {
			if s.Spec.Topology.Coordinator.Suffix == "" {
				s.Spec.Topology.Coordinator.Suffix = string(SolrNodeRoleCoordinator)
			}
			if s.Spec.Topology.Coordinator.Replicas == nil {
				s.Spec.Topology.Coordinator.Replicas = pointer.Int32P(1)
			}

			if s.Spec.Topology.Coordinator.PodTemplate.Spec.SecurityContext == nil {
				s.Spec.Topology.Coordinator.PodTemplate.Spec.SecurityContext = &v1.PodSecurityContext{FSGroup: slVersion.Spec.SecurityContext.RunAsUser}
			}
			s.Spec.Topology.Coordinator.PodTemplate.Spec.SecurityContext.FSGroup = slVersion.Spec.SecurityContext.RunAsUser
			s.setDefaultContainerSecurityContext(&slVersion, &s.Spec.Topology.Coordinator.PodTemplate)
			s.setDefaultInitContainerSecurityContext(&slVersion, &s.Spec.Topology.Coordinator.PodTemplate)
		}
	} else {

		if s.Spec.Replicas == nil {
			s.Spec.Replicas = pointer.Int32P(1)
		}

		if s.Spec.PodTemplate.Spec.SecurityContext == nil {
			s.Spec.PodTemplate.Spec.SecurityContext = &v1.PodSecurityContext{FSGroup: slVersion.Spec.SecurityContext.RunAsUser}
		}

		s.Spec.PodTemplate.Spec.SecurityContext.FSGroup = slVersion.Spec.SecurityContext.RunAsUser
		s.setDefaultContainerSecurityContext(&slVersion, &s.Spec.PodTemplate)

		s.setDefaultInitContainerSecurityContext(&slVersion, &s.Spec.PodTemplate)
	}
}

func (s *Solr) setDefaultInitContainerSecurityContext(slVersion *catalog.SolrVersion, podTemplate *ofst.PodTemplateSpec) {

	initContainer := coreutil.GetContainerByName(podTemplate.Spec.InitContainers, SolrInitContainerName)
	if initContainer == nil {
		initContainer = &v1.Container{
			Name: SolrInitContainerName,
		}
	}
	if initContainer.SecurityContext == nil {
		initContainer.SecurityContext = &v1.SecurityContext{}
	}
	apis.SetDefaultResourceLimits(&initContainer.Resources, DefaultResources)
	s.assignDefaultContainerSecurityContext(slVersion, initContainer.SecurityContext)
	podTemplate.Spec.InitContainers = coreutil.UpsertContainer(podTemplate.Spec.InitContainers, *initContainer)
}

func (s *Solr) setDefaultContainerSecurityContext(slVersion *catalog.SolrVersion, podTemplate *ofst.PodTemplateSpec) {

	container := coreutil.GetContainerByName(podTemplate.Spec.Containers, SolrContainerName)
	if container == nil {
		container = &v1.Container{
			Name: SolrContainerName,
		}
	}
	if container.SecurityContext == nil {
		container.SecurityContext = &v1.SecurityContext{}
	}
	apis.SetDefaultResourceLimits(&container.Resources, DefaultResources)
	s.assignDefaultContainerSecurityContext(slVersion, container.SecurityContext)
	podTemplate.Spec.Containers = coreutil.UpsertContainer(podTemplate.Spec.Containers, *container)
}

func (s *Solr) assignDefaultContainerSecurityContext(slVersion *catalog.SolrVersion, sc *v1.SecurityContext) {
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
		sc.RunAsUser = slVersion.Spec.SecurityContext.RunAsUser
	}
	if sc.RunAsGroup == nil {
		sc.RunAsGroup = slVersion.Spec.SecurityContext.RunAsGroup
	}
	if sc.SeccompProfile == nil {
		sc.SeccompProfile = secomp.DefaultSeccompProfile()
	}
}

func (s *Solr) SetHealthCheckerDefaults() {
	if s.Spec.HealthChecker.PeriodSeconds == nil {
		s.Spec.HealthChecker.PeriodSeconds = pointer.Int32P(20)
	}
	if s.Spec.HealthChecker.TimeoutSeconds == nil {
		s.Spec.HealthChecker.TimeoutSeconds = pointer.Int32P(10)
	}
	if s.Spec.HealthChecker.FailureThreshold == nil {
		s.Spec.HealthChecker.FailureThreshold = pointer.Int32P(3)
	}
}

func (s *Solr) GetPersistentSecrets() []string {
	if s == nil {
		return nil
	}

	var secrets []string
	// Add Admin/Elastic user secret name
	if s.Spec.AuthSecret != nil {
		secrets = append(secrets, s.Spec.AuthSecret.Name)
	}

	if s.Spec.AuthConfigSecret != nil {
		secrets = append(secrets, s.Spec.AuthConfigSecret.Name)
	}

	return secrets
}
