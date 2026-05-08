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
	"strconv"
	"strings"

	autoscalingapi "kubedb.dev/apimachinery/apis/autoscaling/v1alpha1"
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func setDefaultStorageValues(storageSpec *autoscalingapi.StorageAutoscalerSpec) {
	storageSpec.SetDefaults()
}

func setDefaultComputeValues(computeSpec *autoscalingapi.ComputeAutoscalerSpec) {
	computeSpec.SetDefaults()
}

func setInMemoryDefaults(computeSpec *autoscalingapi.ComputeAutoscalerSpec, storageEngine dbapi.StorageEngine) {
	if computeSpec == nil || storageEngine != dbapi.StorageEngineInMemory {
		return
	}
	if computeSpec.InMemoryStorage == nil {
		// assigning a dummy pointer to set the defaults
		computeSpec.InMemoryStorage = &autoscalingapi.ComputeInMemoryStorageSpec{}
	}
	if computeSpec.InMemoryStorage.UsageThresholdPercentage == 0 {
		computeSpec.InMemoryStorage.UsageThresholdPercentage = autoscalingapi.DefaultInMemoryStorageUsageThresholdPercentage
	}
	if computeSpec.InMemoryStorage.ScalingFactorPercentage == 0 {
		computeSpec.InMemoryStorage.ScalingFactorPercentage = autoscalingapi.DefaultInMemoryStorageScalingFactorPercentage
	}
}

func validateScalingRules(storageSpec *autoscalingapi.StorageAutoscalerSpec) error {
	var zeroQuantityThresholds []string
	for _, sr := range storageSpec.ScalingRules {
		if sr.AppliesUpto == "" {
			zeroQuantityThresholds = append(zeroQuantityThresholds, sr.Threshold)
		}
		th := sr.Threshold
		if before, ok := strings.CutSuffix(th, "%"); ok {
			if !isNum(before) {
				return fmt.Errorf("%v is not a valid percentage value", th)
			}
		} else if before, ok := strings.CutSuffix(th, "pc"); ok {
			if !isNum(before) {
				return fmt.Errorf("%v is not a valid percentage value", th)
			}
		} else {
			_, err := resource.ParseQuantity(sr.Threshold)
			if err != nil {
				return fmt.Errorf("%v is not a valid quatity", sr.Threshold)
			}
		}
	}
	if len(zeroQuantityThresholds) > 1 {
		return fmt.Errorf("%v appliesUpto value are empty in %v", zeroQuantityThresholds, storageSpec.ScalingRules)
	}
	return nil
}

func isNum(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
