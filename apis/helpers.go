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

package apis

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
)

const (
	Finalizer = "kubedb.com"
)

type ResourceInfo interface {
	ResourceFQN() string
	ResourceShortCode() string
	ResourceKind() string
	ResourceSingular() string
	ResourcePlural() string
	CustomResourceDefinition() *apiextensions.CustomResourceDefinition
}

func SetDefaultResourceLimits(req *core.ResourceRequirements, defaultResources core.ResourceRequirements) {
	// if request is set,
	//		- limit set:
	//			- return max(limit,request)
	// else if limit set:
	//		- return limit
	// else
	//		- return default
	calLimit := func(name core.ResourceName, defaultValue resource.Quantity) resource.Quantity {
		if r, ok := req.Requests[name]; ok {
			// l is greater than r == 1.
			if l, exist := req.Limits[name]; exist && l.Cmp(r) == 1 {
				return l
			}
			return r
		}
		if l, ok := req.Limits[name]; ok {
			return l
		}
		return defaultValue
	}
	// if request is not set,
	//		- if limit exists:
	//				- copy limit
	//		- else
	//				- set default
	// else
	// 		- return request
	// endif
	calRequest := func(name core.ResourceName, defaultValue resource.Quantity) resource.Quantity {
		if r, ok := req.Requests[name]; !ok {
			if l, exist := req.Limits[name]; exist {
				return l
			}
			return defaultValue
		} else {
			return r
		}
	}

	if req.Limits == nil {
		req.Limits = core.ResourceList{}
	}
	if req.Requests == nil {
		req.Requests = core.ResourceList{}
	}

	klog.Infof("1111111 %+v %+v \n", req.Requests, req.Limits)

	// Calculate the limits first
	for l := range defaultResources.Limits {
		req.Limits[l] = calLimit(l, defaultResources.Limits[l])
	}

	// Once the limit is calculated, Calculate requests
	for r, q := range defaultResources.Requests {
		// If no req is specified, Use the default one. Otherwise, limits will be used
		if _, exists := req.Requests[r]; !exists {
			req.Requests[r] = q
		}
		req.Requests[r] = calRequest(r, q)
	}
	klog.Infof("22222222 %+v %+v \n", req.Requests, req.Limits)
}
