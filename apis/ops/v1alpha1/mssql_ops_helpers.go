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

func (_ *MSSQLOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMSSQLOpsRequest))
}

var _ apis.ResourceInfo = &MSSQLOpsRequest{}

func (k *MSSQLOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMSSQLOpsRequest, ops.GroupName)
}

func (k *MSSQLOpsRequest) ResourceShortCode() string {
	return ResourceCodeMSSQLOpsRequest
}

func (k *MSSQLOpsRequest) ResourceKind() string {
	return ResourceKindMSSQLOpsRequest
}

func (k *MSSQLOpsRequest) ResourceSingular() string {
	return ResourceSingularMSSQLOpsRequest
}

func (k *MSSQLOpsRequest) ResourcePlural() string {
	return ResourcePluralMSSQLOpsRequest
}

var _ Accessor = &MSSQLOpsRequest{}

func (k *MSSQLOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return k.ObjectMeta
}

func (k *MSSQLOpsRequest) GetDBRefName() string {
	return k.Spec.DatabaseRef.Name
}

func (k *MSSQLOpsRequest) GetRequestType() any {
	return k.Spec.Type
}

func (k *MSSQLOpsRequest) GetStatus() OpsRequestStatus {
	return k.Status
}

func (k *MSSQLOpsRequest) SetStatus(s OpsRequestStatus) {
	k.Status = s
}
