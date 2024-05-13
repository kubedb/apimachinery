package v1alpha1

import (
	"fmt"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/autoscaling"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (s SinglestoreAutoscaler) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralSinglestoreAutoscaler))
}

var _ apis.ResourceInfo = &SinglestoreAutoscaler{}

func (s SinglestoreAutoscaler) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralSinglestoreAutoscaler, autoscaling.GroupName)
}

func (s SinglestoreAutoscaler) ResourceShortCode() string {
	return ResourceCodeSinglestoreAutoscaler
}

func (s SinglestoreAutoscaler) ResourceKind() string {
	return ResourceKindSinglestoreAutoscaler
}

func (s SinglestoreAutoscaler) ResourceSingular() string {
	return ResourceSingularSinglestoreAutoscaler
}

func (s SinglestoreAutoscaler) ResourcePlural() string {
	return ResourcePluralSinglestoreAutoscaler
}

func (s SinglestoreAutoscaler) ValidateSpecs() error {
	return nil
}

var _ StatusAccessor = &SinglestoreAutoscaler{}

func (s *SinglestoreAutoscaler) GetStatus() AutoscalerStatus {
	return s.Status
}

func (s *SinglestoreAutoscaler) SetStatus(m AutoscalerStatus) {
	s.Status = m
}
