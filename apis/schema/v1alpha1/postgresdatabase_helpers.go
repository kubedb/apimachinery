package v1alpha1

import (
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ PostgresDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePostgresDatabases))
}

var _ Interface = &PostgresDatabase{}

func (in *PostgresDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *PostgresDatabase) GetStatus() DatabaseStatus {
	return in.Status
}
