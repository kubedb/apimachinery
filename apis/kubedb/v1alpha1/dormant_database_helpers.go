package v1alpha1

import (
	meta_util "github.com/appscode/kutil/meta"
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/golang/glog"
	"github.com/appscode/go/log"
)

func (d DormantDatabase) OffshootName() string {
	return d.Name
}

func (d DormantDatabase) ResourceShortCode() string {
	return ResourceCodeDormantDatabase
}

func (d DormantDatabase) ResourceKind() string {
	return ResourceKindDormantDatabase
}

func (d DormantDatabase) ResourceSingular() string {
	return ResourceSingularDormantDatabase
}

func (d DormantDatabase) ResourcePlural() string {
	return ResourcePluralDormantDatabase
}

func (d DormantDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crdutils.NewCustomResourceDefinition(crdutils.Config{
		Group:         SchemeGroupVersion.Group,
		Plural:        ResourcePluralDormantDatabase,
		Singular:      ResourceSingularDormantDatabase,
		Kind:          ResourceKindDormantDatabase,
		ShortNames:    []string{ResourceCodeDormantDatabase},
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
		SpecDefinitionName:      "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1.DormantDatabase",
		EnableValidation:        false,
		GetOpenAPIDefinitions:   GetOpenAPIDefinitions,
		EnableStatusSubresource: EnableStatusSubresource,
		AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
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

func (d *DormantDatabase) Migrate() {
	if d == nil {
		return
	}
	d.Spec.Origin.Spec.Elasticsearch.Migrate()
	d.Spec.Origin.Spec.Postgres.Migrate()
	d.Spec.Origin.Spec.MySQL.Migrate()
	d.Spec.Origin.Spec.MongoDB.Migrate()
	d.Spec.Origin.Spec.Redis.Migrate()
	d.Spec.Origin.Spec.Memcached.Migrate()
	d.Spec.Origin.Spec.Etcd.Migrate()
}

func (d *DormantDatabase) Equal(other *DormantDatabase) bool {
	if EnableStatusSubresource {
		if d.Status.ObservedGeneration >= d.Generation {
			return true
		}
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, d)
			glog.Infof("meta.Generation [%d] is higher than status.observedGeneration [%d] in DormantDatabase %s/%s with Diff: %s",
				d.Generation, d.Status.ObservedGeneration, d.Namespace, d.Name, diff)
		}
		return false
	}
	if !meta_util.Equal(d.Spec,other.Spec) {
		if glog.V(log.LevelDebug) {
			diff := meta_util.Diff(other, d)
			glog.Infof("DormantDatabase %s/%s has changed. Diff: %s", d.Namespace, d.Name, diff)
		}
		return false
	}
	return true
}
