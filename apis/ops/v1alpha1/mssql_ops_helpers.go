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

func (_ *MsSQLOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMsSQLOpsRequest))
}

var _ apis.ResourceInfo = &MsSQLOpsRequest{}

func (k *MsSQLOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMsSQLOpsRequest, ops.GroupName)
}

func (k *MsSQLOpsRequest) ResourceShortCode() string {
	return ResourceCodeMsSQLOpsRequest
}

func (k *MsSQLOpsRequest) ResourceKind() string {
	return ResourceKindMsSQLOpsRequest
}

func (k *MsSQLOpsRequest) ResourceSingular() string {
	return ResourceSingularMsSQLOpsRequest
}

func (k *MsSQLOpsRequest) ResourcePlural() string {
	return ResourcePluralMsSQLOpsRequest
}

var _ Accessor = &MsSQLOpsRequest{}

func (k *MsSQLOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return k.ObjectMeta
}

func (k *MsSQLOpsRequest) GetDBRefName() string {
	return k.Spec.DatabaseRef.Name
}

func (k *MsSQLOpsRequest) GetRequestType() any {
	return k.Spec.Type
}

func (k *MsSQLOpsRequest) GetStatus() OpsRequestStatus {
	return k.Status
}

func (k *MsSQLOpsRequest) SetStatus(s OpsRequestStatus) {
	k.Status = s
}
