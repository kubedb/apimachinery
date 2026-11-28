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

package v1alpha2

import (
	"fmt"
	"time"

	"kubedb.dev/apimachinery/apis"
	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/apiextensions"
)

func (*WorkflowNamespace) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(SchemeGroupVersion.WithResource(ResourcePluralWorkflowNamespaces))
}

var _ apis.ResourceInfo = &WorkflowNamespace{}

func (r *WorkflowNamespace) ResourceShortCode() string { return ResourceCodeWorkflowNamespace }

func (r *WorkflowNamespace) ResourceKind() string { return ResourceKindWorkflowNamespace }

func (r *WorkflowNamespace) ResourceSingular() string { return ResourceSingularWorkflowNamespace }

func (r *WorkflowNamespace) ResourcePlural() string { return ResourcePluralWorkflowNamespaces }

func (r *WorkflowNamespace) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", r.ResourcePlural(), kubedb.GroupName)
}

// EngineName is the name of the namespace or domain in the engine.
func (r *WorkflowNamespace) EngineName() string {
	if r.Spec.Name != "" {
		return r.Spec.Name
	}
	return r.Name
}

// RetentionDuration is spec.retention, or the default.
func (r *WorkflowNamespace) RetentionDuration() time.Duration {
	if r.Spec.Retention == nil {
		return DefaultWorkflowNamespaceRetention
	}
	return r.Spec.Retention.Duration
}

// IsArchivalEnabled reports whether archival is requested.
func (r *WorkflowNamespace) IsArchivalEnabled() bool {
	return r.Spec.Archival != nil && r.Spec.Archival.Enabled
}

// SetDefaults fills spec.name, spec.retention and spec.deletionPolicy.
func (r *WorkflowNamespace) SetDefaults() {
	if r.Spec.Name == "" {
		r.Spec.Name = r.Name
	}
	if r.Spec.Retention == nil {
		r.Spec.Retention = &metav1.Duration{Duration: DefaultWorkflowNamespaceRetention}
	}
	if r.Spec.DeletionPolicy == "" {
		r.Spec.DeletionPolicy = DeletionPolicyDelete
	}
}
