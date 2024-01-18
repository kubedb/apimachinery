package v1alpha1

import (
	"fmt"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/catalog"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (f FerretDBVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralFerretDBVersion))
}

var _ apis.ResourceInfo = &FerretDBVersion{}

func (f FerretDBVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralFerretDBVersion, catalog.GroupName)
}

func (f FerretDBVersion) ResourceShortCode() string {
	return ResourceCodeFerretDBVersion
}

func (f FerretDBVersion) ResourceKind() string {
	return ResourceKindFerretDBVersion
}

func (f FerretDBVersion) ResourceSingular() string {
	return ResourceSingularFerretDBVersion
}

func (f FerretDBVersion) ResourcePlural() string {
	return ResourcePluralFerretDBVersion
}

func (f FerretDBVersion) ValidateSpecs() error {
	if f.Spec.Version == "" ||
		f.Spec.DB.Image == "" {
		return fmt.Errorf(`atleast one of the following specs is not set for ferretdbVersion "%v":
spec.version,
spec.db.image`, f.Name)
	}
	return nil
}
