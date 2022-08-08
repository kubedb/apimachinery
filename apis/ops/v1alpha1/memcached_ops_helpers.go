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

func (_ MemcachedOpsRequest) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMemcachedOpsRequest))
}

var _ apis.ResourceInfo = &MemcachedOpsRequest{}

func (m MemcachedOpsRequest) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMemcachedOpsRequest, ops.GroupName)
}

func (m MemcachedOpsRequest) ResourceShortCode() string {
	return ResourceCodeMemcachedOpsRequest
}

func (m MemcachedOpsRequest) ResourceKind() string {
	return ResourceKindMemcachedOpsRequest
}

func (m MemcachedOpsRequest) ResourceSingular() string {
	return ResourceSingularMemcachedOpsRequest
}

func (m MemcachedOpsRequest) ResourcePlural() string {
	return ResourcePluralMemcachedOpsRequest
}

func (m MemcachedOpsRequest) ValidateSpecs() error {
	return nil
}

var _ Accessor = &MemcachedOpsRequest{}

func (e *MemcachedOpsRequest) GetObjectMeta() metav1.ObjectMeta {
	return e.ObjectMeta
}

func (e *MemcachedOpsRequest) GetRequestType() OpsRequestType {
	return e.Spec.Type
}

func (e *MemcachedOpsRequest) GetDBRefName() string {
	return e.Spec.DatabaseRef.Name
}

func (e *MemcachedOpsRequest) GetStatus() OpsRequestStatus {
	return e.Status
}

func (e *MemcachedOpsRequest) SetStatus(s OpsRequestStatus) {
	e.Status = s
}
