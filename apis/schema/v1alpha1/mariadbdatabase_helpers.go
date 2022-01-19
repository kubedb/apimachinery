package v1alpha1

import (
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ MariaDBDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceMariaDBDatabases))
}

var _ Interface = &MariaDBDatabase{}

func (in *MariaDBDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *MariaDBDatabase) GetStatus() DatabaseStatus {
	return in.Status
}
