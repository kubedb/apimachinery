package v1alpha1

import (
	"fmt"
	"kmodules.xyz/client-go/apiextensions"
	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/autoscaling"
	"kubedb.dev/apimachinery/crds"
)

func (r HazelcastAutoscaler) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralHazelcastAutoscaler))
}

var _ apis.ResourceInfo = &HazelcastAutoscaler{}

func (p HazelcastAutoscaler) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralHazelcastAutoscaler, autoscaling.GroupName)
}

func (p HazelcastAutoscaler) ResourceShortCode() string {
	return ResourceCodeHazelcastAutoscaler
}

func (p HazelcastAutoscaler) ResourceKind() string {
	return ResourceKindHazelcastAutoscaler
}

func (p HazelcastAutoscaler) ResourceSingular() string {
	return ResourceSingularHazelcastAutoscaler
}

func (p HazelcastAutoscaler) ResourcePlural() string {
	return ResourcePluralHazelcastAutoscaler
}

func (p HazelcastAutoscaler) ValidateSpecs() error {
	return nil
}

var _ StatusAccessor = &HazelcastAutoscaler{}

func (e *HazelcastAutoscaler) GetStatus() AutoscalerStatus {
	return e.Status
}

func (e *HazelcastAutoscaler) SetStatus(s AutoscalerStatus) {
	e.Status = s
}
