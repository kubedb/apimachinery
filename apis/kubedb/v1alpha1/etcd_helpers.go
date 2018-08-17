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

func (e Etcd) OffshootName() string {
	return e.Name
}

func (e Etcd) OffshootSelectors() map[string]string {
	return map[string]string{
		LabelDatabaseName: e.Name,
		LabelDatabaseKind: ResourceKindEtcd,
	}
}

func (e Etcd) OffshootLabels() map[string]string {
	return meta_util.FilterKeys(GenericKey, e.OffshootSelectors(), e.Labels)
}

func (e Etcd) ResourceShortCode() string {
	return ResourceCodeEtcd
}

func (e Etcd) ResourceKind() string {
	return ResourceKindEtcd
}

func (e Etcd) ResourceSingular() string {
	return ResourceSingularEtcd
}

func (e Etcd) ResourcePlural() string {
	return ResourcePluralEtcd
}

func (e Etcd) ServiceName() string {
	return e.OffshootName()
}

type etcdStatsService struct {
	*Etcd
}

func (e etcdStatsService) GetNamespace() string {
	return e.Etcd.GetNamespace()
}

func (e etcdStatsService) ServiceName() string {
	return e.OffshootName() + "-stats"
}

func (e etcdStatsService) ServiceMonitorName() string {
	return fmt.Sprintf("kubedb-%s-%s", e.Namespace, e.Name)
}

func (e etcdStatsService) Path() string {
	return fmt.Sprintf("/metrics")
}

func (e etcdStatsService) Scheme() string {
	return ""
}

func (e Etcd) StatsService() mona.StatsAccessor {
	return &etcdStatsService{&e}
}

func (e *Etcd) GetMonitoringVendor() string {
	if e.Spec.Monitor != nil {
		return e.Spec.Monitor.Agent.Vendor()
	}
	return ""
}

func (e Etcd) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralEtcd,
		Singular:      ResourceSingularEtcd,
		Kind:          ResourceKindEtcd,
		ShortNames:    []string{ResourceCodeEtcd},
		Categories:    []string{"datastore", "kubedb", "appscode", "all"},
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
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.Etcd",
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

func (e *Etcd) Migrate() {
	if e == nil {
		return
	}
	e.Spec.Migrate()
}

func (e *EtcdSpec) Migrate() {
	if e == nil {
		return
	}
}

func (e *Etcd) AlreadyObserved(other *Etcd) bool {
	if e == nil {
		return other == nil
	}
	if other == nil { // && d != nil
		return false
	}
	if e == other {
		return true
	}

	var match bool

	if EnableStatusSubresource {
		match = e.Status.ObservedGeneration >= e.Generation
	} else {
		match = meta_util.Equal(e.Spec, other.Spec)
	}
	if match {
		match = reflect.DeepEqual(e.Labels, other.Labels)
	}
	if match {
		match = meta_util.EqualAnnotation(e.Annotations, other.Annotations)
	}

	if !match && bool(glog.V(log.LevelDebug)) {
		diff := meta_util.Diff(other, e)
		glog.V(log.LevelDebug).Infof("%s %s/%s has changed. Diff: %s", meta_util.GetKind(e), e.Namespace, e.Name, diff)
	}
	return match
}
