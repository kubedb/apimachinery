package v1alpha1

import (
	"fmt"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (p MongoDB) OffshootName() string {
	return p.Name
}

func (p MongoDB) OffshootSelectors() map[string]string {
	return map[string]string{
		LabelDatabaseName: p.Name,
		LabelDatabaseKind: ResourceKindMongoDB,
	}
}

func (p MongoDB) OffshootLabels() map[string]string {
	return filterTags(p.OffshootSelectors(), p.Labels)
}

func (p MongoDB) ResourceShortCode() string {
	return ResourceCodeMongoDB
}

func (p MongoDB) ResourceKind() string {
	return ResourceKindMongoDB
}

func (p MongoDB) ResourceSingular() string {
	return ResourceSingularMongoDB
}

func (p MongoDB) ResourcePlural() string {
	return ResourcePluralMongoDB
}

func (p MongoDB) ServiceName() string {
	return p.OffshootName()
}

func (m MongoDB) StatsServiceName() string {
	return m.OffshootName() + "-stats"
}

type MongoDBStatsService struct {
	mongodb MongoDB
}

func (m MongoDBStatsService) GetNamespace() string {
	return m.mongodb.GetNamespace()
}

func (m MongoDBStatsService) ServiceName() string {
	return m.mongodb.StatsServiceName()
}

func (m MongoDBStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", m.mongodb.Namespace, m.mongodb.Name)
}

func (m MongoDBStatsService) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", m.mongodb.Namespace, m.mongodb.ResourcePlural(), m.mongodb.Name)
}

func (m MongoDBStatsService) Scheme() string {
	return ""
}

func (m MongoDB) StatsAccessor() mona.StatsAccessor {
	return &MongoDBStatsService{mongodb: m}
}

func (m *MongoDB) GetMonitoringVendor() string {
	if m.Spec.Monitor != nil {
		return m.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (p MongoDB) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralMongoDB,
		Singular:      ResourceSingularMongoDB,
		Kind:          ResourceKindMongoDB,
		ShortNames:    []string{ResourceCodeMongoDB},
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
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.MongoDB",
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
