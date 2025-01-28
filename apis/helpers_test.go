package apis

import (
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var DefaultResources = core.ResourceRequirements{
	Requests: core.ResourceList{
		core.ResourceCPU:    resource.MustParse("1"),
		core.ResourceMemory: resource.MustParse("2Gi"),
	},
	Limits: core.ResourceList{
		core.ResourceCPU:    resource.MustParse("2"),
		core.ResourceMemory: resource.MustParse("4Gi"),
	},
}

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
			name: "Both the requests and limits are set",
			args: args{
				req: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceCPU:    resource.MustParse(".500"),
						core.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: core.ResourceList{
						core.ResourceCPU:    resource.MustParse("1"),
						core.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				defaultResources: DefaultResources,
			},
		},
		{
			name: "Only requests are set - limits should be set from requests",
			args: args{
				req: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceCPU:    resource.MustParse(".500"),
						core.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
				defaultResources: DefaultResources,
			},
		},
		{
			name: "Only limits are set - requests should be set from limits",
			args: args{
				req: &core.ResourceRequirements{
					Limits: core.ResourceList{
						core.ResourceCPU:    resource.MustParse("1"),
						core.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
				defaultResources: DefaultResources,
			},
		},
		{
			name: "Nothing is set - should use default values",
			args: args{
				req:              &core.ResourceRequirements{},
				defaultResources: DefaultResources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := tt.args.req
			SetDefaultResourceLimits(tt.args.req, tt.args.defaultResources)
			if !checkExpexcted(tt.args.req, old) {
				t.Errorf("Expected SetDefaultResourceLimits to set default limits correctly")
			}
		})
	}
}

func checkExpexcted(req *core.ResourceRequirements, old *core.ResourceRequirements) bool {
	// Check if requests and limits are properly initialized
	if req.Requests == nil || req.Limits == nil {
		return false
	}

	// If old requests existed but no limits, verify limits are set from requests
	if old != nil && old.Requests != nil && (old.Limits == nil || len(old.Limits) == 0) {
		for name, oldReq := range old.Requests {
			if newLim, exists := req.Limits[name]; !exists || newLim.Cmp(oldReq) != 0 {
				return false
			}
			// Also verify that request value is preserved
			if newReq, exists := req.Requests[name]; !exists || newReq.Cmp(oldReq) != 0 {
				return false
			}
		}
	}

	// If old limits existed but no requests, verify requests are set from limits
	if old != nil && old.Limits != nil && (old.Requests == nil || len(old.Requests) == 0) {
		for name, oldLim := range old.Limits {
			if newReq, exists := req.Requests[name]; !exists || newReq.Cmp(oldLim) != 0 {
				return false
			}
			// Also verify that limit value is preserved
			if newLim, exists := req.Limits[name]; !exists || newLim.Cmp(oldLim) != 0 {
				return false
			}
		}
	}

	// If neither requests nor limits existed, verify default values are used
	if old != nil && (old.Requests == nil || len(old.Requests) == 0) && (old.Limits == nil || len(old.Limits) == 0) {
		// CPU check
		if cpuReq, exists := req.Requests[core.ResourceCPU]; !exists || cpuReq.String() != "1" {
			return false
		}
		if cpuLim, exists := req.Limits[core.ResourceCPU]; !exists || cpuLim.String() != "2" {
			return false
		}
		// Memory check
		if memReq, exists := req.Requests[core.ResourceMemory]; !exists || memReq.String() != "2Gi" {
			return false
		}
		if memLim, exists := req.Limits[core.ResourceMemory]; !exists || memLim.String() != "4Gi" {
			return false
		}
	}

	// For all cases, ensure limits are not less than requests
	for name, reqVal := range req.Requests {
		if limVal, exists := req.Limits[name]; !exists || limVal.Cmp(reqVal) < 0 {
			return false
		}
	}

	return true
}
