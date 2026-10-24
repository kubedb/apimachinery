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
	"kubedb.dev/apimachinery/apis/ops"
	"kubedb.dev/apimachinery/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/apiextensions"
)

func (n *DurableWorkflowOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralDurableWorkflowOpsRequest))
}

var _ apis.ResourceInfo = &DurableWorkflowOpsRequest{}

func (n *DurableWorkflowOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralDurableWorkflowOpsRequest, ops.GroupName)
}

func (n *DurableWorkflowOpsRequest) ResourceShortCode() string {
	return ResourceCodeDurableWorkflowOpsRequest
}

func (n *DurableWorkflowOpsRequest) ResourceKind() string {
	return ResourceKindDurableWorkflowOpsRequest
}

func (n *DurableWorkflowOpsRequest) ResourceSingular() string {
	return ResourceSingularDurableWorkflowOpsRequest
}

func (n *DurableWorkflowOpsRequest) ResourcePlural() string {
	return ResourcePluralDurableWorkflowOpsRequest
}

var _ Accessor = &DurableWorkflowOpsRequest{}

func (n *DurableWorkflowOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return n.ObjectMeta
}

func (n *DurableWorkflowOpsRequest) GetDBRefName() string {
	return n.Spec.DatabaseRef.Name
}

func (n *DurableWorkflowOpsRequest) GetRequestType() string {
	return string(n.Spec.Type)
}

func (n *DurableWorkflowOpsRequest) GetStatus() OpsRequestStatus {
	return n.Status
}

func (n *DurableWorkflowOpsRequest) SetStatus(s OpsRequestStatus) {
	n.Status = s
}
