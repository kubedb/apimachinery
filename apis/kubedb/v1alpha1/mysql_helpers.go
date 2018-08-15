package v1alpha1

import (
	"fmt"
	"reflect"

	"github.com/appscode/go/log"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	meta_util "github.com/appscode/kutil/meta"
	"github.com/golang/glog"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (m MySQL) OffshootName() string {
	return m.Name
}

func (m MySQL) OffshootSelectors() map[string]string {
	return map[string]string{
		LabelDatabaseName: m.Name,
		LabelDatabaseKind: ResourceKindMySQL,
	}
}

func (m MySQL) OffshootLabels() map[string]string {
	return filterTags(m.OffshootSelectors(), m.Labels)
}

var _ ResourceInfo = &MySQL{}

func (m MySQL) ResourceShortCode() string {
	return ResourceCodeMySQL
}

func (m MySQL) ResourceKind() string {
	return ResourceKindMySQL
}

func (m MySQL) ResourceSingular() string {
	return ResourceSingularMySQL
}

func (m MySQL) ResourcePlural() string {
	return ResourcePluralMySQL
}

func (m MySQL) ServiceName() string {
	return m.OffshootName()
}

type mySQLStatsService struct {
	*MySQL
}

func (m mySQLStatsService) GetNamespace() string {
	return m.MySQL.GetNamespace()
}

func (m mySQLStatsService) ServiceName() string {
	return m.OffshootName() + "-stats"
}

func (m mySQLStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", m.Namespace, m.Name)
}

func (m mySQLStatsService) Path() string {
	return fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", m.Namespace, m.ResourcePlural(), m.Name)
}

func (m mySQLStatsService) Scheme() string {
	return ""
}

func (m MySQL) StatsService() mona.StatsAccessor {
	return &mySQLStatsService{&m}
}

func (m *MySQL) GetMonitoringVendor() string {
	if m.Spec.Monitor != nil {
		return m.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (m MySQL) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralMySQL,
		Singular:      ResourceSingularMySQL,
		Kind:          ResourceKindMySQL,
		ShortNames:    []string{ResourceCodeMySQL},
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
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.MySQL",
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

func (m *MySQL) Migrate() {
	if m == nil {
		return
	}
	m.Spec.Migrate()
}

func (m *MySQLSpec) Migrate() {
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

func (m *MySQL) Equal(other *MySQL) bool {
	if EnableStatusSubresource {
		// At this moment, metadata.Generation is incremented only by `spec`.
		// issue tracked: https://github.com/kubernetes/kubernetes/issues/67428
		// So look for changes in metadata.labels as well.
		if m.Generation <= m.Status.ObservedGeneration && reflect.DeepEqual(other.Labels, m.Labels) {
			return true
		}
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, m)
			glog.InfoDepth(1, "meta.Generation [%d] is higher than status.observedGeneration [%d] in MySQL %s/%s with Diff: %s",
				m.Generation, m.Status.ObservedGeneration, m.Namespace, m.Name, diff)
		}
		return false
	}

	if !meta_util.Equal(other.Spec, m.Spec) || !reflect.DeepEqual(other.Labels, m.Labels) {
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, m)
			glog.InfoDepth(1, "MySQL %s/%s has changed. Diff: %s", m.Namespace, m.Name, diff)
		}
		return false
	}
	return true
}
