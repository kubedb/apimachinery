package api

import "fmt"

const (
	DatabaseNamePrefix = "kubedb"
	LabelDatabaseKind  = "kubedb.com/kind"
	LabelDatabaseName  = "kubedb.com/name"

	PostgresDatabaseVersion = "postgres.kubedb.com/version"
	ElasticDatabaseVersion  = "elastic.kubedb.com/version"
)

func (p Postgres) OffshootName() string {
	return fmt.Sprintf("%v-%v", p.Name, ResourceCodePostgres)
}

func (p Postgres) ServiceName() string {
	return p.Name
}

func (p Postgres) OffshootLabels() map[string]string {
	return map[string]string{
		"origin":          "kubedb",
		LabelDatabaseKind: p.Name,
	}
}

func (p Postgres) StatefulSetLabels() map[string]string {
	labels := make(map[string]string)
	for key, val := range p.Labels {
		labels[key] = val
	}
	labels[LabelDatabaseKind] = ResourceKindPostgres
	return labels
}

func (p Postgres) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range p.Annotations {
		annotations[key] = val
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
		"origin":          "kubedb",
		LabelDatabaseKind: e.Name,
	}
}

func (p Elastic) StatefulSetLabels() map[string]string {
	labels := make(map[string]string)
	for key, val := range p.Labels {
		labels[key] = val
	}
	labels[LabelDatabaseKind] = ResourceKindElastic
	return labels
}

func (p Elastic) StatefulSetAnnotations() map[string]string {
	annotations := make(map[string]string)
	for key, val := range p.Annotations {
		annotations[key] = val
	}
	annotations[ElasticDatabaseVersion] = string(p.Spec.Version)
	return annotations
}
