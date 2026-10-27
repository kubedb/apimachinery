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

func (r NATSAutoscaler) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralNATSAutoscaler))
}

var _ apis.ResourceInfo = &NATSAutoscaler{}

func (r NATSAutoscaler) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralNATSAutoscaler, autoscaling.GroupName)
}

func (r NATSAutoscaler) ResourceShortCode() string {
	return ResourceCodeNATSAutoscaler
}

func (r NATSAutoscaler) ResourceKind() string {
	return ResourceKindNATSAutoscaler
}

func (r NATSAutoscaler) ResourceSingular() string {
	return ResourceSingularNATSAutoscaler
}

func (r NATSAutoscaler) ResourcePlural() string {
	return ResourcePluralNATSAutoscaler
}

func (r NATSAutoscaler) ValidateSpecs() error {
	return nil
}

var _ StatusAccessor = &NATSAutoscaler{}

func (r *NATSAutoscaler) GetStatus() AutoscalerStatus {
	return r.Status
}

func (r *NATSAutoscaler) SetStatus(s AutoscalerStatus) {
	r.Status = s
}
