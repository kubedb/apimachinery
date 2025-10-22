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

	"kubedb.dev/apimachinery/apis/kubedb"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	meta_util "kmodules.xyz/client-go/meta"
	// ofst "kmodules.xyz/offshoot-api/api/v1"
)

func (w *Weaviate) SetHealthCheckerDefaults() {
	if w.Spec.HealthChecker.PeriodSeconds == nil {
		w.Spec.HealthChecker.PeriodSeconds = pointer.Int32(10) // Int32P shows an error.
	}
	if w.Spec.HealthChecker.TimeoutSeconds == nil {
		w.Spec.HealthChecker.TimeoutSeconds = pointer.Int32(10)
	}
	if w.Spec.HealthChecker.FailureThreshold == nil {
		w.Spec.HealthChecker.FailureThreshold = pointer.Int32(3)
	}
}

func (w *Weaviate) ResourceKind() string {
	return ResourceKindWeaviate
}

// Owner returns owner reference to resources
func (w *Weaviate) Owner() *metav1.OwnerReference {
	return metav1.NewControllerRef(w, SchemeGroupVersion.WithKind(w.ResourceKind()))
}

func (w *Weaviate) OffshootName() string {
	return w.Name
}

func (w *Weaviate) ServiceName() string {
	return w.OffshootName()
}

// GetServiceTemplate returns a pointer to the desired serviceTemplate referred by "aliaS". Otherwise, it returns nil.
//func GetServiceTemplate(templates []NamedServiceTemplateSpec, alias ServiceAlias) ofst.ServiceTemplateSpec {
//	for i := range templates {
//		c := templates[i]
//		if c.Alias == alias {
//			return c.ServiceTemplateSpec
//		}
//	}
//	return ofst.ServiceTemplateSpec{}
//}

func (w *Weaviate) ResourcePlural() string {
	return ResourcePluralWeaviate
}

func (w *Weaviate) ResourceFQN() string {
	return fmt.Sprintf("%s.%s", w.ResourcePlural(), kubedb.GroupName)
}

func (w *Weaviate) offshootLabels(selector, override map[string]string) map[string]string {
	selector[meta_util.ComponentLabelKey] = kubedb.ComponentDatabase
	return meta_util.FilterKeys(kubedb.GroupName, selector, meta_util.OverwriteKeys(nil, w.Labels, override))
}

func (w *Weaviate) OffshootSelectors(extraSelectors ...map[string]string) map[string]string {
	selector := map[string]string{
		meta_util.NameLabelKey:      w.ResourceFQN(),
		meta_util.InstanceLabelKey:  w.Name,
		meta_util.ManagedByLabelKey: kubedb.GroupName,
	}
	return meta_util.OverwriteKeys(selector, extraSelectors...)
}

func (w *Weaviate) OffshootLabels() map[string]string {
	return w.offshootLabels(w.OffshootSelectors(), nil)
}

func (w *Weaviate) GoverningServiceName() string {
	return meta_util.NameWithSuffix(w.ServiceName(), "pods")
}

func (w *Weaviate) PodLabels(extraLabels ...map[string]string) map[string]string {
	return w.offshootLabels(meta_util.OverwriteKeys(w.OffshootSelectors(), extraLabels...), w.Spec.PodTemplate.Labels)
}

func (w *Weaviate) PetSetName() string {
	return w.OffshootName()
}
