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
	"strings"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/pkg/errors"
	"gomodules.xyz/x/arrays"
	core "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateStorage(kc client.Client, storageType olddbapi.StorageType, spec *core.PersistentVolumeClaimSpec, storageSpecPath ...string) error {
	if storageType == olddbapi.StorageTypeEphemeral {
		return nil
	}

	storagePath := "spec.storage"
	if len(storageSpecPath) != 0 {
		storagePath = strings.Join(storageSpecPath, ".")
	}

	if spec == nil {
		return fmt.Errorf(`%v is missing for durable storage type`, storagePath)
	}

	if spec.StorageClassName != nil {
		var sc storagev1.StorageClass
		if err := kc.Get(context.TODO(), types.NamespacedName{
			Name: *spec.StorageClassName,
		}, &sc); err != nil {
			if kerr.IsNotFound(err) {
				return fmt.Errorf(`%v.storageClassName "%v" not found`, storagePath, *spec.StorageClassName)
			}
			return err
		}
	}

	if val, found := spec.Resources.Requests[core.ResourceStorage]; found {
		if val.Value() <= 0 {
			return errors.New("invalid ResourceStorage request")
		}
	} else {
		return errors.New("missing ResourceStorage request")
	}

	return nil
}

// ValidateMonitorSpec validates the Monitoring spec after all the defaulting is done.
func ValidateMonitorSpec(agent *mona.AgentSpec) error {
	if agent.Agent == "" {
		return fmt.Errorf(`object 'Agent' is missing in '%+v'`, agent)
	}

	if !mona.IsKnownAgentType(agent.Agent) {
		return fmt.Errorf("unknown monitoring agent type %s", agent.Agent)
	}

	if agent.Agent.Vendor() == mona.VendorPrometheus {
		if agent.Prometheus != nil &&
			agent.Prometheus.Exporter.Port >= 1024 &&
			agent.Prometheus.Exporter.Port <= 65535 {
			return nil
		}
		return fmt.Errorf(`invalid 'monitor.prometheus' in '%+v'. prometheus.exporter.port value must be between 1024 and 65535, inclusive`, agent)
	}

	return fmt.Errorf(`invalid 'Agent' in '%+v'`, agent)
}

func IsStorageTypeCompatibleWithSpec(storageType olddbapi.StorageType, storage *core.PersistentVolumeClaimSpec, ephemeralStorage *core.EmptyDirVolumeSource) error {
	if storageType == olddbapi.StorageTypeEphemeral && storage != nil {
		return fmt.Errorf("'spec.storage' is not supported for Ephemeral storage type, use 'spec.ephemeralStorage' to configure Ephemeral storage type")
	}
	if storageType == olddbapi.StorageTypeDurable && ephemeralStorage != nil {
		return fmt.Errorf("'spec.ephemeralStorage' is not supported for Durable storage type, use 'spec.storage' to configure Durable storage type")
	}

	return nil
}

func IsStorageTypeCompatibleWithSpecV1(storageType dbapi.StorageType, storage *core.PersistentVolumeClaimSpec, ephemeralStorage *core.EmptyDirVolumeSource) error {
	if storageType == dbapi.StorageTypeEphemeral && storage != nil {
		return fmt.Errorf("'spec.storage' is not supported for Ephemeral storage type, use 'spec.ephemeralStorage' to configure Ephemeral storage type")
	}
	if storageType == dbapi.StorageTypeDurable && ephemeralStorage != nil {
		return fmt.Errorf("'spec.ephemeralStorage' is not supported for Durable storage type, use 'spec.storage' to configure Durable storage type")
	}

	return nil
}

func CheckSecretsExist(client client.Client, secNames []string, namespace string) error {
	var secret core.Secret
	for _, sec := range secNames {
		err := client.Get(context.TODO(), types.NamespacedName{
			Name:      sec,
			Namespace: namespace,
		}, &secret)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateEnvVar(envs []core.EnvVar, forbiddenEnvs []string, resourceType string) error {
	for _, env := range envs {
		present, _ := arrays.Contains(forbiddenEnvs, env.Name)
		if present {
			return fmt.Errorf("environment variable %s is forbidden to use in %s spec", env.Name, resourceType)
		}
	}
	return nil
}

func ValidateInternalUsers(users map[string]olddbapi.ElasticsearchUserSpec, allowedInternalUsers []string, resourceType string) error {
	for user := range users {
		present, _ := arrays.Contains(allowedInternalUsers, user)
		if !present {
			return fmt.Errorf("Internal user %s is forbidden to use in %s spec", user, resourceType)
		}
	}
	return nil
}

func ValidateInternalUsersV1(users map[string]dbapi.ElasticsearchUserSpec, allowedInternalUsers []string, resourceType string) error {
	for user := range users {
		present, _ := arrays.Contains(allowedInternalUsers, user)
		if !present {
			return fmt.Errorf("Internal user %s is forbidden to use in %s spec", user, resourceType)
		}
	}
	return nil
}

func ValidateHealth(health *kmapi.HealthCheckSpec) error {
	if health.PeriodSeconds != nil && *health.PeriodSeconds <= 0 {
		return fmt.Errorf(`spec.healthCheck.periodSeconds: can not be less than 1`)
	}

	if health.TimeoutSeconds != nil && *health.TimeoutSeconds <= 0 {
		return fmt.Errorf(`spec.healthCheck.timeoutSeconds: can not be less than 1`)
	}

	if health.FailureThreshold != nil && *health.FailureThreshold <= 0 {
		return fmt.Errorf(`spec.healthCheck.failureThreshold: can not be less than 1`)
	}
	return nil
}

func ValidateVolumes(volumes []core.Volume, reservedVolumeNames []string) error {
	for _, rv := range reservedVolumeNames {
		for _, ugv := range volumes {
			if ugv.Name == rv {
				return errors.New("Cannot use a reserve volume name: " + rv)
			}
		}
	}
	return nil
}

func ValidateMountPaths(volumeMounts []core.VolumeMount, reservedMountPaths []string) error {
	trimPrefixSuffix := func(s string) string {
		return strings.TrimSuffix(strings.TrimPrefix(s, "/"), "/") + "/"
	}
	for _, vm := range volumeMounts {
		givenPath := trimPrefixSuffix(vm.MountPath)
		for _, our := range reservedMountPaths {
			ourPath := trimPrefixSuffix(our)
			if strings.HasPrefix(givenPath, ourPath) {
				return errors.New("Cannot use mountPath " + vm.MountPath + " because of reservedMountPaths path " + our)
			}
		}
	}
	return nil
}
