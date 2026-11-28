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

func (DurableWorkflowVersion) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralDurableWorkflowVersion))
}

var _ apis.ResourceInfo = &DurableWorkflowVersion{}

func (r DurableWorkflowVersion) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", ResourcePluralDurableWorkflowVersion, catalog.GroupName)
}

func (r DurableWorkflowVersion) ResourceShortCode() string {
	return ResourceCodeDurableWorkflowVersion
}

func (r DurableWorkflowVersion) ResourceKind() string {
	return ResourceKindDurableWorkflowVersion
}

func (r DurableWorkflowVersion) ResourceSingular() string {
	return ResourceSingularDurableWorkflowVersion
}

func (r DurableWorkflowVersion) ResourcePlural() string {
	return ResourcePluralDurableWorkflowVersion
}

func (r DurableWorkflowVersion) IsTemporal() bool {
	return r.Spec.Distribution == DurableWorkflowDistroTemporal
}

func (r DurableWorkflowVersion) IsCadence() bool {
	return r.Spec.Distribution == DurableWorkflowDistroCadence
}

// OnlineSchemaUpdate reports whether this version's schema update can run while the previous version is serving.
func (r DurableWorkflowVersion) OnlineSchemaUpdate() bool {
	return r.Spec.Rollout.OnlineSchemaUpdate == nil || *r.Spec.Rollout.OnlineSchemaUpdate
}

// SupportsStore reports whether the engine can be used as the default or visibility store.
func (r DurableWorkflowVersion) SupportsStore(engine string, visibility bool) bool {
	for _, s := range r.Spec.Schema.Stores {
		if strings.EqualFold(s.Type, engine) {
			if visibility {
				return s.Visibility
			}
			return s.Default
		}
	}
	return false
}

func (r DurableWorkflowVersion) ValidateSpecs() error {
	if r.Spec.Version == "" ||
		r.Spec.Distribution == "" ||
		r.Spec.Server.Image == "" ||
		r.Spec.AdminTools.Image == "" {
		fields := []string{
			"spec.version",
			"spec.distribution",
			"spec.server.image",
			"spec.adminTools.image",
		}
		return fmt.Errorf("atleast one of the following specs is not set for DurableWorkflowVersion %q: %s", r.Name, strings.Join(fields, ", "))
	}
	return nil
}
