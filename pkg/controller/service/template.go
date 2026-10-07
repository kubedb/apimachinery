/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	core "k8s.io/api/core/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

// ApplyServiceTemplate syncs a non-headless Service with its serviceTemplate, so removing a field
// from the template reverts the Service instead of leaving the previous value in place.
func ApplyServiceTemplate(in *core.Service, template ofst.ServiceTemplateSpec, ports []core.ServicePort) {
	in.Annotations = template.Annotations
	in.Spec.Ports = ofst.PatchServicePorts(core_util.MergeServicePorts(in.Spec.Ports, ports), template.Spec.Ports)

	// clusterIP is immutable once allocated by the apiserver
	if template.Spec.ClusterIP != "" {
		in.Spec.ClusterIP = template.Spec.ClusterIP
	}

	in.Spec.Type = template.Spec.Type
	if in.Spec.Type == "" {
		in.Spec.Type = core.ServiceTypeClusterIP
	}
	in.Spec.ExternalIPs = template.Spec.ExternalIPs
	in.Spec.LoadBalancerIP = template.Spec.LoadBalancerIP
	in.Spec.LoadBalancerSourceRanges = template.Spec.LoadBalancerSourceRanges

	// the apiserver defaults externalTrafficPolicy to Cluster for externally accessible Services;
	// leaving it empty there would produce a patch on every reconcile
	in.Spec.ExternalTrafficPolicy = template.Spec.ExternalTrafficPolicy
	if in.Spec.ExternalTrafficPolicy == "" && (in.Spec.Type == core.ServiceTypeNodePort || in.Spec.Type == core.ServiceTypeLoadBalancer) {
		in.Spec.ExternalTrafficPolicy = core.ServiceExternalTrafficPolicyCluster
	}

	// healthCheckNodePort is allocated by the apiserver when not set
	if template.Spec.HealthCheckNodePort > 0 {
		in.Spec.HealthCheckNodePort = template.Spec.HealthCheckNodePort
	}
	in.Spec.SessionAffinityConfig = template.Spec.SessionAffinityConfig
}
