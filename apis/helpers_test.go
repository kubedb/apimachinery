package apis

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

var (
	DefaultResources = core.ResourceRequirements{
		Requests: core.ResourceList{
			core.ResourceCPU:    resource.MustParse(".500"),
			core.ResourceMemory: resource.MustParse("1.5Gi"),
		},
		Limits: core.ResourceList{
			core.ResourceCPU:    resource.MustParse("1"),
			core.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
)

func TestSetDefaultResourceLimits(t *testing.T) {
	type args struct {
		req              *core.ResourceRequirements
		defaultResources core.ResourceRequirements
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "both set",
			args: args{
				req: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceCPU:    resource.MustParse(".200"),
						core.ResourceMemory: resource.MustParse("1.2Gi"),
					},
					Limits: core.ResourceList{
						core.ResourceCPU:    resource.MustParse(".2"),
						core.ResourceMemory: resource.MustParse("2.2Gi"),
					},
				},
				defaultResources: DefaultResources,
			},
		},
		{
			name: "no defaults",
		},
		{
			name: "no requests",
		},
		{
			name: "no limits",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.args.req
			SetDefaultResourceLimits(tt.args.req, tt.args.defaultResources)
			checkExpexcted(tt.args.req, old)
		})
	}
}

func checkExpexcted(req *core.ResourceRequirements, old *core.ResourceRequirements) bool {
	// implement
}
