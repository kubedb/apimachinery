package api

import "fmt"

func (p Postgres) OffshootName() string {
	return fmt.Sprintf("%v-%v", p.Name, ResourceCodePostgres)
}

func (p Postgres) OffshootLabels() map[string]string {
	return map[string]string{
		"origin":          "kubedb",
		"kubedb.com/name": p.Name,
	}
}

func (e Elastic) OffshootName() string {
	return fmt.Sprintf("%v-%v", e.Name, ResourceCodeElastic)
}

func (e Elastic) OffshootLabels() map[string]string {
	return map[string]string{
		"origin":          "kubedb",
		"kubedb.com/name": e.Name,
	}
}
