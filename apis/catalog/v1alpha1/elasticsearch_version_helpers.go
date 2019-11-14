/*
Copyright The KubeDB Authors.

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

	"kubedb.dev/apimachinery/api/crds"
	"kubedb.dev/apimachinery/apis"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/yaml"
)

func (_ ElasticsearchVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	data := crds.MustAsset("catalog.kubedb.com_elasticsearchversions.yaml")
	var out apiextensions.CustomResourceDefinition
	utilruntime.Must(yaml.Unmarshal(data, &out))
	return &out
}

var _ apis.ResourceInfo = &ElasticsearchVersion{}

func (e ElasticsearchVersion) ResourceShortCode() string {
	return ResourceCodeElasticsearchVersion
}

func (e ElasticsearchVersion) ResourceKind() string {
	return ResourceKindElasticsearchVersion
}

func (e ElasticsearchVersion) ResourceSingular() string {
	return ResourceSingularElasticsearchVersion
}

func (e ElasticsearchVersion) ResourcePlural() string {
	return ResourcePluralElasticsearchVersion
}

func (e ElasticsearchVersion) ValidateSpecs() error {
	if e.Spec.AuthPlugin == "" ||
		e.Spec.Version == "" ||
		e.Spec.DB.Image == "" ||
		e.Spec.Tools.Image == "" ||
		e.Spec.Exporter.Image == "" ||
		e.Spec.InitContainer.YQImage == "" ||
		e.Spec.InitContainer.Image == "" {
		return fmt.Errorf(`atleast one of the following specs is not set for elasticsearchVersion "%v":
spec.authPlugin,
spec.version,
spec.db.image,
spec.tools.image,
spec.exporter.image,
spec.initContainer.yqImage,
spec.initContainer.image.`, e.Name)
	}
	return nil
}
