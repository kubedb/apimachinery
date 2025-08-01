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
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
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
