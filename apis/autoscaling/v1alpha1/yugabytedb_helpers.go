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

func (r YugabyteDBAutoscaler) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralYugabyteDBAutoscaler))
}

var _ apis.ResourceInfo = &YugabyteDBAutoscaler{}

func (r YugabyteDBAutoscaler) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralYugabyteDBAutoscaler, autoscaling.GroupName)
}

func (r YugabyteDBAutoscaler) ResourceShortCode() string {
	return ResourceCodeYugabyteDBAutoscaler
}

func (r YugabyteDBAutoscaler) ResourceKind() string {
	return ResourceKindYugabyteDBAutoscaler
}

func (r YugabyteDBAutoscaler) ResourceSingular() string {
	return ResourceSingularYugabyteDBAutoscaler
}

func (r YugabyteDBAutoscaler) ResourcePlural() string {
	return ResourcePluralYugabyteDBAutoscaler
}

func (r YugabyteDBAutoscaler) ValidateSpecs() error {
	return nil
}

var _ StatusAccessor = &YugabyteDBAutoscaler{}

func (r *YugabyteDBAutoscaler) GetStatus() AutoscalerStatus {
	return r.Status
}

func (r *YugabyteDBAutoscaler) SetStatus(s AutoscalerStatus) {
	r.Status = s
}
