package v1alpha1

import (
	"fmt"
	"sort"
	"strconv"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

func (spec *ComputeAutoscalerSpec) SetDefaults() {
	if spec == nil {
		return
	}
	if spec.Trigger == "" {
		spec.Trigger = AutoscalerTriggerOff
	}
	if spec.ControlledResources == nil {
		spec.ControlledResources = []core.ResourceName{core.ResourceCPU, core.ResourceMemory}
	}
	if spec.ContainerControlledValues == nil {
		reqAndLim := ContainerControlledValuesRequestsAndLimits
		spec.ContainerControlledValues = &reqAndLim
	}
	if spec.ResourceDiffPercentage == 0 {
		spec.ResourceDiffPercentage = DefaultResourceDiffPercentage
	}
	if spec.PodLifeTimeThreshold.Duration == 0 {
		spec.PodLifeTimeThreshold = metav1.Duration{Duration: DefaultPodLifeTimeThreshold}
	}
}

func (spec *StorageAutoscalerSpec) SetDefaults() {
	if spec == nil {
		return
	}
	if spec.Trigger == "" {
		spec.Trigger = AutoscalerTriggerOff
	}
	if spec.UsageThreshold == nil {
		spec.UsageThreshold = ptr.To[int32](DefaultStorageUsageThreshold)
	}
	if spec.ScalingThreshold == nil {
		spec.ScalingThreshold = ptr.To[int32](DefaultStorageScalingThreshold)
	}
	spec.setDefaultScalingRules()
}

type quantity struct {
	InString   string
	InQuantity resource.Quantity
	Threshold  string
}

func (spec *StorageAutoscalerSpec) setDefaultScalingRules() {
	if spec.ScalingRules == nil {
		spec.ScalingRules = []StorageScalingRule{
			{
				AppliesUpto: "",
				Threshold:   fmt.Sprintf("%spc", strconv.Itoa(int(*spec.ScalingThreshold))),
			},
		}
	}
	var quantities []quantity
	var zeroQuantityThresholds []string
	for _, sr := range spec.ScalingRules {
		if sr.AppliesUpto == "" {
			zeroQuantityThresholds = append(zeroQuantityThresholds, sr.Threshold)
			continue
		}
		quantities = append(quantities, quantity{
			InString:   sr.AppliesUpto,
			InQuantity: resource.MustParse(sr.AppliesUpto),
			Threshold:  sr.Threshold,
		})
	}
	sort.Slice(quantities, func(i, j int) bool {
		return quantities[i].InQuantity.Cmp(quantities[j].InQuantity) < 0
	})

	spec.ScalingRules = make([]StorageScalingRule, 0)
	for _, q := range quantities {
		spec.ScalingRules = append(spec.ScalingRules, StorageScalingRule{
			AppliesUpto: q.InString,
			Threshold:   q.Threshold,
		})
	}
	for _, threshold := range zeroQuantityThresholds {
		spec.ScalingRules = append(spec.ScalingRules, StorageScalingRule{
			AppliesUpto: "",
			Threshold:   threshold,
		})
	}
	klog.Infof("scaling Rules = %v \n", spec.ScalingRules)
}
