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

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	RegisterOpsPathMapperToPlugins(&RedisOpsRequest{})
}

type RedisOpsRequest struct{}

var _ OpsPathMapper = (*RedisOpsRequest)(nil)

func (m *RedisOpsRequest) HorizontalPathMapping() map[OpsReqPath]ReferencedObjPath {
	return map[OpsReqPath]ReferencedObjPath{
		"spec.horizontalScaling.master":   "spec.cluster.master",
		"spec.horizontalScaling.replicas": "spec.cluster.replicas",
	}
}

func (m *RedisOpsRequest) VerticalPathMapping() map[OpsReqPath]ReferencedObjPath {
	return map[OpsReqPath]ReferencedObjPath{
		"spec.verticalScaling.redis":       "spec.podTemplate.spec.resources",
		"spec.verticalScaling.exporter":    "spec.monitor.prometheus.exporter.resources",
		"spec.verticalScaling.coordinator": "spec.coordinator.resources",
	}
}

func (m *RedisOpsRequest) VolumeExpansionPathMapping() map[OpsReqPath]ReferencedObjPath {
	return map[OpsReqPath]ReferencedObjPath{
		"spec.volumeExpansion.redis": "spec.storage.resources.requests.storage",
	}
}

func (m *RedisOpsRequest) GetReferencedDbObjectPath() []string {
	return []string{"spec", "databaseRef", "referencedDB"}
}

func (m *RedisOpsRequest) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "ops.kubedb.com",
		Version: "v1alpha1",
		Kind:    "RedisOpsRequest",
	}
}
