/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/autoscaling"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (r DurableWorkflowAutoscaler) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralDurableWorkflowAutoscaler))
}

var _ apis.ResourceInfo = &DurableWorkflowAutoscaler{}

func (r DurableWorkflowAutoscaler) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralDurableWorkflowAutoscaler, autoscaling.GroupName)
}

func (r DurableWorkflowAutoscaler) ResourceShortCode() string {
	return ResourceCodeDurableWorkflowAutoscaler
}

func (r DurableWorkflowAutoscaler) ResourceKind() string {
	return ResourceKindDurableWorkflowAutoscaler
}

func (r DurableWorkflowAutoscaler) ResourceSingular() string {
	return ResourceSingularDurableWorkflowAutoscaler
}

func (r DurableWorkflowAutoscaler) ResourcePlural() string {
	return ResourcePluralDurableWorkflowAutoscaler
}

func (r DurableWorkflowAutoscaler) ValidateSpecs() error {
	return nil
}

var _ StatusAccessor = &DurableWorkflowAutoscaler{}

func (r *DurableWorkflowAutoscaler) GetStatus() AutoscalerStatus {
	return r.Status
}

func (r *DurableWorkflowAutoscaler) SetStatus(s AutoscalerStatus) {
	r.Status = s
}

// ComputeSpecForRole returns the compute autoscaler spec of a server role, nil if the role is not configured.
// The roles are the values of kubedb.DurableWorkflowRole*.
func (s *DurableWorkflowComputeAutoscalerSpec) ComputeSpecForRole(role string) *ComputeAutoscalerSpec {
	if s == nil {
		return nil
	}
	switch role {
	case kubedb.DurableWorkflowRoleFrontend:
		return s.Frontend
	case kubedb.DurableWorkflowRoleHistory:
		return s.History
	case kubedb.DurableWorkflowRoleMatching:
		return s.Matching
	case kubedb.DurableWorkflowRoleWorker:
		return s.Worker
	case kubedb.DurableWorkflowRoleInternalFrontend:
		return s.InternalFrontend
	}
	return nil
}

// IsEmpty tells that no role has a compute autoscaler spec.
func (s *DurableWorkflowComputeAutoscalerSpec) IsEmpty() bool {
	return s == nil || (s.Frontend == nil && s.History == nil && s.Matching == nil && s.Worker == nil && s.InternalFrontend == nil)
}
