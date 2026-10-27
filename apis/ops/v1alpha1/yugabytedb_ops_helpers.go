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

func (n *YugabyteDBOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralYugabyteDBOpsRequest))
}

var _ apis.ResourceInfo = &YugabyteDBOpsRequest{}

func (n *YugabyteDBOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralYugabyteDBOpsRequest, ops.GroupName)
}

func (n *YugabyteDBOpsRequest) ResourceShortCode() string {
	return ResourceCodeYugabyteDBOpsRequest
}

func (n *YugabyteDBOpsRequest) ResourceKind() string {
	return ResourceKindYugabyteDBOpsRequest
}

func (n *YugabyteDBOpsRequest) ResourceSingular() string {
	return ResourceSingularYugabyteDBOpsRequest
}

func (n *YugabyteDBOpsRequest) ResourcePlural() string {
	return ResourcePluralYugabyteDBOpsRequest
}

var _ Accessor = &YugabyteDBOpsRequest{}

func (n *YugabyteDBOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return n.ObjectMeta
}

func (n *YugabyteDBOpsRequest) GetDBRefName() string {
	return n.Spec.DatabaseRef.Name
}

func (n *YugabyteDBOpsRequest) GetRequestType() string {
	return string(n.Spec.Type)
}

func (n *YugabyteDBOpsRequest) GetStatus() OpsRequestStatus {
	return n.Status
}

func (n *YugabyteDBOpsRequest) SetStatus(s OpsRequestStatus) {
	n.Status = s
}
