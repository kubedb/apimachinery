package v1alpha1

import (
	"fmt"

	meta_util "github.com/appscode/kutil/meta"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	"github.com/golang/glog"
	"github.com/appscode/go/log"
	"reflect"
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

type mongoDBStatsService struct {
	*MongoDB
}

func (m mongoDBStatsService) GetNamespace() string {
	return m.MongoDB.GetNamespace()
}

func (m mongoDBStatsService) ServiceName() string {
	return m.OffshootName() + "-stats"
}

func (m mongoDBStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", m.Namespace, m.Name)
}

func (m mongoDBStatsService) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", m.Namespace, m.ResourcePlural(), m.Name)
}

func (m mongoDBStatsService) Scheme() string {
	return ""
}

func (m MongoDB) StatsService() mona.StatsAccessor {
	return &mongoDBStatsService{&m}
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
		Categories:    []string{"datastore", "kubedb", "appscode"},
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

func (m *MongoDB) Migrate() {
	if m == nil {
		return
	}
	m.Spec.Migrate()
}

func (m *MongoDBSpec) Migrate() {
	if m == nil {
		return
	}
	m.BackupSchedule.Migrate()
	if len(m.NodeSelector) > 0 {
		m.PodTemplate.Spec.NodeSelector = m.NodeSelector
		m.NodeSelector = nil
	}
	if m.Resources != nil {
		m.PodTemplate.Spec.Resources = *m.Resources
		m.Resources = nil
	}
	if m.Affinity != nil {
		m.PodTemplate.Spec.Affinity = m.Affinity
		m.Affinity = nil
	}
	if len(m.SchedulerName) > 0 {
		m.PodTemplate.Spec.SchedulerName = m.SchedulerName
		m.SchedulerName = ""
	}
	if len(m.Tolerations) > 0 {
		m.PodTemplate.Spec.Tolerations = m.Tolerations
		m.Tolerations = nil
	}
	if len(m.ImagePullSecrets) > 0 {
		m.PodTemplate.Spec.ImagePullSecrets = m.ImagePullSecrets
		m.ImagePullSecrets = nil
	}
}

func (m *MongoDB) Equal(other *MongoDB) bool {
	if EnableStatusSubresource {
		// At this moment, metadata.Generation is incremented only by `spec`.
		// issue tracked: https://github.com/kubernetes/kubernetes/issues/67428
		// So look for changes in metadata.labels as well.
		if m.Generation <= m.Status.ObservedGeneration && reflect.DeepEqual(other.Labels, m.Labels) {
			return true
		}
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, m)
			glog.Infof("meta.Generation [%m] is higher than status.observedGeneration [%m] in MongoDB %s/%s with Diff: %s",
				m.Generation, m.Status.ObservedGeneration, m.Namespace, m.Name, diff)
		}
		return false
	}

	if !meta_util.Equal(other.Spec, m.Spec) || !reflect.DeepEqual(other.Labels, m.Labels) {
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, m)
			glog.Infof("MongoDB %s/%s has changed. Diff: %s", m.Namespace, m.Name, diff)
		}
		return false
	}
	return true
}
