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
	"strings"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/catalog"
	"kubedb.dev/apimachinery/crds"

	"kmodules.xyz/client-go/apiextensions"
)

func (YugabyteDBVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralYugabyteDBVersion))
}

var _ apis.ResourceInfo = &YugabyteDBVersion{}

func (r YugabyteDBVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralYugabyteDBVersion, catalog.GroupName)
}

func (r YugabyteDBVersion) ResourceShortCode() string {
	return ResourceCodeYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourceKind() string {
	return ResourceKindYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourceSingular() string {
	return ResourceSingularYugabyteDBVersion
}

func (r YugabyteDBVersion) ResourcePlural() string {
	return ResourcePluralYugabyteDBVersion
}

func (r YugabyteDBVersion) ValidateSpecs() error {
	if r.Spec.Version == "" ||
		r.Spec.DB.Image == "" ||
		r.Spec.YSQLMajorVersion == 0 {
		fields := []string{
			"spec.version",
			"spec.db.image",
			"spec.ysqlMajorVersion",
		}
		return fmt.Errorf("atleast one of the following specs is not set for YugabyteDBVersion %q: %s", r.Name, strings.Join(fields, ", "))
	}
	return nil
}
