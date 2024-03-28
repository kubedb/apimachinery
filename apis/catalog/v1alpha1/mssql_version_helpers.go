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
	"kubedb.dev/apimachinery/apis/catalog"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (m MsSQLVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMsSQLVersion))
}

var _ apis.ResourceInfo = &MsSQLVersion{}

func (m MsSQLVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMsSQLVersion, catalog.GroupName)
}

func (m MsSQLVersion) ResourceShortCode() string {
	return ResourceCodeMsSQLVersion
}

func (m MsSQLVersion) ResourceKind() string {
	return ResourceKindMsSQLVersion
}

func (m MsSQLVersion) ResourceSingular() string {
	return ResourceSingularMsSQLVersion
}

func (m MsSQLVersion) ResourcePlural() string {
	return ResourcePluralMsSQLVersion
}

func (m MsSQLVersion) ValidateSpecs() error {
	if m.Spec.Version == "" || m.Spec.DB.Image == "" || m.Spec.Coordinator.Image == "" {
		return fmt.Errorf(`at least one of the following specs is not set for MsSQLVersion "%v":
spec.version,
spec.coordinator.image,
spec.initContainer.image`, m.Name)
	}
	return nil
}
