package v1alpha1

import (
	"fmt"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (r Memcached) OffshootName() string {
	return r.Name
}

func (r Memcached) OffshootSelectors() map[string]string {
	return map[string]string{
		LabelDatabaseKind: ResourceKindMemcached,
		LabelDatabaseName: r.Name,
	}
}

func (r Memcached) OffshootLabels() map[string]string {
	return filterTags(r.OffshootSelectors(), r.Labels)
}

func (r Memcached) ResourceShortCode() string {
	return ResourceCodeMemcached
}

func (r Memcached) ResourceKind() string {
	return ResourceKindMemcached
}

func (r Memcached) ResourceSingular() string {
	return ResourceSingularMemcached
}

func (r Memcached) ResourcePlural() string {
	return ResourcePluralMemcached
}

func (m Memcached) ServiceName() string {
	return m.OffshootName()
}

func (m Memcached) StatsServiceName() string {
	return m.OffshootName() + "-stats"
}

type MemcachedStatsService struct {
	memcached Memcached
}

func (m MemcachedStatsService) GetNamespace() string {
	return m.memcached.GetNamespace()
}

func (m MemcachedStatsService) ServiceName() string {
	return m.memcached.StatsServiceName()
}

func (m MemcachedStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", m.memcached.Namespace, m.memcached.Name)
}

func (m MemcachedStatsService) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", m.memcached.Namespace, m.memcached.ResourcePlural(), m.memcached.Name)
}

func (m MemcachedStatsService) Scheme() string {
	return ""
}

func (m Memcached) StatsAccessor() mona.StatsAccessor {
	return MemcachedStatsService{memcached: m}
}

func (m *Memcached) GetMonitoringVendor() string {
	if m.Spec.Monitor != nil {
		return m.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (m Memcached) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralMemcached,
		Singular:      ResourceSingularMemcached,
		Kind:          ResourceKindMemcached,
		ShortNames:    []string{ResourceCodeMemcached},
		ResourceScope: string(apiextensions.NamespaceScoped),
		Versions: []apiextensions.CustomResourceDefinitionVersion{
			{
				Name:    SchemeGroupVersion.Version,
				Served:  true,
				Storage: true,
			},
		},
		Labels: crdutils.Labels{
			LabelsMap: map[string]string{"app": "kubedb"},
		},
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.Memcached",
		EnableValidation:        true,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
			{
				Name:     "Version",
				Type:     "string",
				JSONPath: ".spec.version",
			},
			{
				Name:     "Status",
				Type:     "string",
				JSONPath: ".status.phase",
			},
			{
				Name:     "Age",
				Type:     "date",
				JSONPath: ".metadata.creationTimestamp",
			},
		},
	}, setNameSchema)
}
