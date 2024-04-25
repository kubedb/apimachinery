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

func (m MSSQLVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralMSSQLVersion))
}

var _ apis.ResourceInfo = &MSSQLVersion{}

func (m MSSQLVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralMSSQLVersion, catalog.GroupName)
}

func (m MSSQLVersion) ResourceShortCode() string {
	return ResourceCodeMSSQLVersion
}

func (m MSSQLVersion) ResourceKind() string {
	return ResourceKindMSSQLVersion
}

func (m MSSQLVersion) ResourceSingular() string {
	return ResourceSingularMSSQLVersion
}

func (m MSSQLVersion) ResourcePlural() string {
	return ResourcePluralMSSQLVersion
}

func (m MSSQLVersion) ValidateSpecs() error {
	if m.Spec.Version == "" || m.Spec.DB.Image == "" || m.Spec.Coordinator.Image == "" {
		return fmt.Errorf(`at least one of the following specs is not set for MSSQLVersion "%v":
spec.version,
spec.coordinator.image,
spec.initContainer.image`, m.Name)
	}
	// TODO: add m.spec.exporter.image check FOR monitoring
	return nil
}
