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

package v1

import (
	"fmt"

	"kmodules.xyz/resource-metrics/api"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	api.Register(schema.GroupVersionKind{
		Group:   "kubedb.com",
		Version: "v1",
		Kind:    "MySQL",
	}, MySQL{}.ResourceCalculator())
}

type MySQL struct{}

func (r MySQL) ResourceCalculator() api.ResourceCalculator {
	return &api.ResourceCalculatorFuncs{
		AppRoles:               []api.PodRole{api.PodRoleDefault, api.PodRoleRouter},
		RuntimeRoles:           []api.PodRole{api.PodRoleDefault, api.PodRoleSidecar, api.PodRoleExporter, api.PodRoleRouter},
		RoleReplicasFn:         r.roleReplicasFn,
		ModeFn:                 r.modeFn,
		UsesTLSFn:              r.usesTLSFn,
		RoleResourceLimitsFn:   r.roleResourceFn(api.ResourceLimits),
		RoleResourceRequestsFn: r.roleResourceFn(api.ResourceRequests),
	}
}

func (r MySQL) roleReplicasFn(obj map[string]interface{}) (api.ReplicaList, error) {
	result := api.ReplicaList{}

	// Standalone or GroupReplication Mode
	replicas, found, err := unstructured.NestedInt64(obj, "spec", "replicas")
	if err != nil {
		return nil, fmt.Errorf("failed to read spec.replicas %v: %w", obj, err)
	}
	if !found {
		result[api.PodRoleDefault] = 1
	} else {
		result[api.PodRoleDefault] = replicas
	}

	// InnoDB Router
	mode, found, err := unstructured.NestedString(obj, "spec", "topology", "mode")
	if err != nil {
		return nil, err
	}
	if found && mode == "InnoDBCluster" {
		replicas, found, err := unstructured.NestedInt64(obj, "spec", "topology", "innoDBCluster", "router", "replicas")
		if err != nil {
			return nil, err
		}
		if !found {
			result[api.PodRoleRouter] = 1
		} else {
			result[api.PodRoleRouter] = replicas
		}
	}

	return result, nil
}

func (r MySQL) modeFn(obj map[string]interface{}) (string, error) {
	mode, found, err := unstructured.NestedString(obj, "spec", "topology", "mode")
	if err != nil {
		return "", err
	}
	if found {
		return mode, nil
	}
	return DBModeStandalone, nil
}

func (r MySQL) usesTLSFn(obj map[string]interface{}) (bool, error) {
	_, found, err := unstructured.NestedFieldNoCopy(obj, "spec", "tls")
	return found, err
}

func (r MySQL) roleResourceFn(fn func(rr core.ResourceRequirements) core.ResourceList) func(obj map[string]interface{}) (map[api.PodRole]api.PodInfo, error) {
	return func(obj map[string]interface{}) (map[api.PodRole]api.PodInfo, error) {
		container, replicas, err := api.AppNodeResourcesV2(obj, fn, MySQLContainerName, "spec")
		if err != nil {
			return nil, err
		}

		exporter, err := api.ContainerResources(obj, fn, "spec", "monitor", "prometheus", "exporter")
		if err != nil {
			return nil, err
		}

		result := map[api.PodRole]api.PodInfo{
			api.PodRoleDefault:  {Resource: container, Replicas: replicas},
			api.PodRoleExporter: {Resource: exporter, Replicas: replicas},
		}

		if replicas > 1 {
			sidecar, err := api.SidecarNodeResourcesV2(obj, fn, MySQLSidecarContainerName, "spec")
			if err != nil {
				return nil, err
			}
			result[api.PodRoleSidecar] = api.PodInfo{Resource: sidecar, Replicas: replicas}
		}

		// InnoDB Router
		mode, found, err := unstructured.NestedString(obj, "spec", "topology", "mode")
		if err != nil {
			return nil, err
		}

		if found && mode == "InnoDBCluster" {
			router, replicas, err := api.AppNodeResourcesV2(obj, fn, MySQLRouterContainerName, "spec", "topology", "innoDBCluster", "router")
			if err != nil {
				return nil, err
			}
			result[api.PodRoleRouter] = api.PodInfo{Resource: router, Replicas: replicas}
		}
		return result, nil
	}
}
