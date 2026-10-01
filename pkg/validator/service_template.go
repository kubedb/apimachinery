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

package validator

import (
	"fmt"
	"slices"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func ServiceTemplateAliasWarnings(templates []dbapi.NamedServiceTemplateSpec, supported ...dbapi.ServiceAlias) admission.Warnings {
	var warnings admission.Warnings
	for i, t := range templates {
		if !slices.Contains(supported, t.Alias) {
			warnings = append(warnings, fmt.Sprintf("spec.serviceTemplates[%d].alias %q is not used by this database and will be ignored; supported aliases: %v", i, t.Alias, supported))
		}
	}
	return warnings
}
