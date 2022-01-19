package v1alpha1

import (
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (_ RedisDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourceRedisDatabases))
}

var _ Interface = &RedisDatabase{}

func (in *RedisDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *RedisDatabase) GetStatus() DatabaseStatus {
	return in.Status
}
