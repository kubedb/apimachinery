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

func (_ *MSSQLServerOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMSSQLServerOpsRequest))
}

var _ apis.ResourceInfo = &MSSQLServerOpsRequest{}

func (k *MSSQLServerOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMSSQLServerOpsRequest, ops.GroupName)
}

func (k *MSSQLServerOpsRequest) ResourceShortCode() string {
	return ResourceCodeMSSQLServerOpsRequest
}

func (k *MSSQLServerOpsRequest) ResourceKind() string {
	return ResourceKindMSSQLServerOpsRequest
}

func (k *MSSQLServerOpsRequest) ResourceSingular() string {
	return ResourceSingularMSSQLServerOpsRequest
}

func (k *MSSQLServerOpsRequest) ResourcePlural() string {
	return ResourcePluralMSSQLServerOpsRequest
}

var _ Accessor = &MSSQLServerOpsRequest{}

func (k *MSSQLServerOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return k.ObjectMeta
}

func (k *MSSQLServerOpsRequest) GetDBRefName() string {
	return k.Spec.DatabaseRef.Name
}

func (k *MSSQLServerOpsRequest) GetRequestType() any {
	return k.Spec.Type
}

func (k *MSSQLServerOpsRequest) GetStatus() OpsRequestStatus {
	return k.Status
}

func (k *MSSQLServerOpsRequest) SetStatus(s OpsRequestStatus) {
	k.Status = s
}
