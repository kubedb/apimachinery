package v1alpha1

import (
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func GetFinalizer() string {
	return SchemeGroupVersion.Group
}

func (_ MongoDBArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource("MongoDBArchiver"))
}

func (_ PostgresArchiver) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource("PostgresArchiver"))
}
