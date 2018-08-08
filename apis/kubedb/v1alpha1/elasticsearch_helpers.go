package v1alpha1

import (
	"fmt"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/meta"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (e Elasticsearch) OffshootName() string {
	return e.Name
}

func (e Elasticsearch) OffshootSelectors() map[string]string {
	return map[string]string{
		LabelDatabaseKind: ResourceKindElasticsearch,
		LabelDatabaseName: e.Name,
	}
}

func (e Elasticsearch) OffshootLabels() map[string]string {
	return filterTags(e.OffshootSelectors(), e.Labels)
}

var _ ResourceInfo = &Elasticsearch{}

func (e Elasticsearch) ResourceShortCode() string {
	return ResourceCodeElasticsearch
}

func (e Elasticsearch) ResourceKind() string {
	return ResourceKindElasticsearch
}

func (e Elasticsearch) ResourceSingular() string {
	return ResourceSingularElasticsearch
}

func (e Elasticsearch) ResourcePlural() string {
	return ResourcePluralElasticsearch
}

func (e Elasticsearch) ServiceName() string {
	return e.OffshootName()
}

func (e *Elasticsearch) MasterServiceName() string {
	return fmt.Sprintf("%v-master", e.ServiceName())
}

func (e Elasticsearch) StatsServiceName() string {
	return e.OffshootName() + "-stats"
}

type ElasticsearchStatsService struct {
	mongodb Elasticsearch
}

func (e ElasticsearchStatsService) GetNamespace() string {
	return e.mongodb.GetNamespace()
}

func (e ElasticsearchStatsService) ServiceName() string {
	return e.mongodb.StatsServiceName()
}

func (e ElasticsearchStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", e.mongodb.Namespace, e.mongodb.Name)
}

func (e ElasticsearchStatsService) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", e.mongodb.Namespace, e.mongodb.ResourcePlural(), e.mongodb.Name)
}

func (e ElasticsearchStatsService) Scheme() string {
	return ""
}

func (e Elasticsearch) StatsAccessor() mona.StatsAccessor {
	return &ElasticsearchStatsService{mongodb: e}
}

func (e *Elasticsearch) GetMonitoringVendor() string {
	if e.Spec.Monitor != nil {
		return e.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (e Elasticsearch) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralElasticsearch,
		Singular:      ResourceSingularElasticsearch,
		Kind:          ResourceKindElasticsearch,
		ShortNames:    []string{ResourceCodeElasticsearch},
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
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.Elasticsearch",
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

const (
	ESSearchGuardDisabled = ElasticsearchKey + "/searchguard-disabled"
)

func (e Elasticsearch) SearchGuardDisabled() bool {
	v, _ := meta.GetBoolValue(e.Annotations, ESSearchGuardDisabled)
	return v
}
