package api

import (
	"fmt"
	"strings"
)

const (
	DatabaseNamePrefix = "kubedb"

	GenericKey = "kubedb.com"

	LabelDatabaseKind = GenericKey + "/kind"
	LabelDatabaseName = GenericKey + "/name"

	PostgresKey             = ResourceNamePostgres + "." + GenericKey
	PostgresDatabaseVersion = PostgresKey + "/version"

	ElasticKey             = ResourceNameElastic + ".kubedb.com"
	ElasticDatabaseVersion = ElasticKey + "/version"
)

func (p Postgres) OffshootName() string {
	return fmt.Sprintf("%v-%v", p.Name, ResourceCodePostgres)
}

func (p Postgres) ServiceName() string {
	return p.Name
}

func (p Postgres) OffshootLabels() map[string]string {
	return map[string]string{
		LabelDatabaseName: p.Name,
		LabelDatabaseKind: ResourceKindPostgres,
	}
}

func (p Postgres) StatefulSetLabels() map[string]string {
	labels := p.OffshootLabels()
	for key, val := range p.Labels {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, PostgresKey+"/") {
			labels[key] = val
		}
	}
	return labels
}

func (p Postgres) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range p.Annotations {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, PostgresKey+"/") {
			annotations[key] = val
		}
	}
	annotations[PostgresDatabaseVersion] = string(p.Spec.Version)
	return annotations
}

func (e Elastic) OffshootName() string {
	return fmt.Sprintf("%v-%v", e.Name, ResourceCodeElastic)
}

func (e Elastic) ServiceName() string {
	return e.Name
}

func (e Elastic) OffshootLabels() map[string]string {
	return map[string]string{
		LabelDatabaseKind: ResourceKindElastic,
		LabelDatabaseName: e.Name,
	}
}

func (e Elastic) StatefulSetLabels() map[string]string {
	labels := e.OffshootLabels()
	for key, val := range e.Labels {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, ElasticKey+"/") {
			labels[key] = val
		}
	}
	return labels
}

func (e Elastic) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range e.Annotations {
		if !strings.HasPrefix(key, GenericKey+"/") && !strings.HasPrefix(key, ElasticKey+"/") {
			annotations[key] = val
		}
	}
	annotations[ElasticDatabaseVersion] = string(e.Spec.Version)
	return annotations
}
