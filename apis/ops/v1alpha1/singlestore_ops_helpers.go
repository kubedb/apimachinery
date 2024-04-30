package v1alpha1

import (
	"fmt"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/ops"
	"kubedb.dev/apimachinery/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/apiextensions"
)

func (r *SinglestoreOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralSinglestoreOpsRequest))
}

var _ apis.ResourceInfo = &SinglestoreOpsRequest{}

func (r *SinglestoreOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralSinglestoreOpsRequest, ops.GroupName)
}

func (r *SinglestoreOpsRequest) ResourceShortCode() string {
	return ResourceCodeSinglestoreOpsRequest
}

func (r *SinglestoreOpsRequest) ResourceKind() string {
	return ResourceKindSinglestoreOpsRequest
}

func (r *SinglestoreOpsRequest) ResourceSingular() string {
	return ResourceSingularSinglestoreOpsRequest
}

func (r *SinglestoreOpsRequest) ResourcePlural() string {
	return ResourcePluralSinglestoreOpsRequest
}

var _ Accessor = &SinglestoreOpsRequest{}

func (r *SinglestoreOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return r.ObjectMeta
}

func (r *SinglestoreOpsRequest) GetDBRefName() string {
	return r.Spec.DatabaseRef.Name
}

func (r *SinglestoreOpsRequest) GetRequestType() any {
	return r.Spec.Type
}

func (r *SinglestoreOpsRequest) GetStatus() OpsRequestStatus {
	return r.Status
}

func (r *SinglestoreOpsRequest) SetStatus(s OpsRequestStatus) {
	r.Status = s
}
