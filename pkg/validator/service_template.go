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
	"context"
	"fmt"
	"slices"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type serviceTemplateAliasValidator struct {
	admission.CustomValidator
	supported []dbapi.ServiceAlias
}

// WithServiceTemplateAliasWarnings warns about spec.serviceTemplates entries whose alias has no
// matching Service for this database, since the operator silently ignores them.
func WithServiceTemplateAliasWarnings(v admission.CustomValidator, supported ...dbapi.ServiceAlias) admission.CustomValidator {
	return serviceTemplateAliasValidator{CustomValidator: v, supported: supported}
}

func (v serviceTemplateAliasValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	warnings, err := v.CustomValidator.ValidateCreate(ctx, obj)
	return v.withAliasWarnings(obj, warnings, err)
}

func (v serviceTemplateAliasValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	warnings, err := v.CustomValidator.ValidateUpdate(ctx, oldObj, newObj)
	return v.withAliasWarnings(newObj, warnings, err)
}

func (v serviceTemplateAliasValidator) withAliasWarnings(obj runtime.Object, warnings admission.Warnings, err error) (admission.Warnings, error) {
	aliasWarnings, convErr := v.aliasWarnings(obj)
	if convErr != nil && err == nil {
		err = convErr
	}
	return append(warnings, aliasWarnings...), err
}

func (v serviceTemplateAliasValidator) aliasWarnings(obj runtime.Object) (admission.Warnings, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	templates, _, err := unstructured.NestedSlice(content, "spec", "serviceTemplates")
	if err != nil {
		return nil, err
	}
	var warnings admission.Warnings
	for i, t := range templates {
		m, ok := t.(map[string]any)
		if !ok {
			continue
		}
		alias, _ := m["alias"].(string)
		if !slices.Contains(v.supported, dbapi.ServiceAlias(alias)) {
			warnings = append(warnings, fmt.Sprintf("spec.serviceTemplates[%d].alias %q is not used by this database and will be ignored; supported aliases: %v", i, alias, v.supported))
		}
	}
	return warnings, nil
}
