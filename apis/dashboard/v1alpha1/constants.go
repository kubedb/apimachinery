package v1alpha1

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	DefaultResources = core.ResourceRequirements{
		Requests: core.ResourceList{
			core.ResourceCPU:    resource.MustParse(".500"),
			core.ResourceMemory: resource.MustParse("1024Mi"),
		},
		Limits: core.ResourceList{
			core.ResourceMemory: resource.MustParse("1024Mi"),
		},
	}
)
