package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/apiextensions"
	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/ops"
	"kubedb.dev/apimachinery/crds"
)

func (r *FerretDBOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralFerretDBOpsRequest))
}

var _ apis.ResourceInfo = &FerretDBOpsRequest{}

func (r *FerretDBOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralFerretDBOpsRequest, ops.GroupName)
}

func (r *FerretDBOpsRequest) ResourceShortCode() string {
	return ResourceCodeFerretDBOpsRequest
}

func (r *FerretDBOpsRequest) ResourceKind() string {
	return ResourceKindFerretDBOpsRequest
}

func (r *FerretDBOpsRequest) ResourceSingular() string {
	return ResourceSingularFerretDBOpsRequest
}

func (r *FerretDBOpsRequest) ResourcePlural() string {
	return ResourcePluralFerretDBOpsRequest
}

var _ Accessor = &FerretDBOpsRequest{}

func (r *FerretDBOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return r.ObjectMeta
}

func (r *FerretDBOpsRequest) GetDBRefName() string {
	return r.Spec.DatabaseRef.Name
}

func (r *FerretDBOpsRequest) GetRequestType() any {
	return r.Spec.Type
}

func (r *FerretDBOpsRequest) GetStatus() OpsRequestStatus {
	return r.Status
}

func (r *FerretDBOpsRequest) SetStatus(s OpsRequestStatus) {
	r.Status = s
}
