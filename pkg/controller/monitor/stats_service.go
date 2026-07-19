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

package monitor

import (
	"context"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kutil "kmodules.xyz/client-go"
	clientutil "kmodules.xyz/client-go/client"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatsServiceDB interface {
	client.Object
	StatsService() mona.StatsAccessor
	StatsServiceLabels() map[string]string
	AsOwner() *metav1.OwnerReference
}

type StatsServiceOptions struct {
	KBClient client.Client
	DB       StatsServiceDB
	Monitor  *mona.AgentSpec
	// Selectors is passed in because OffshootSelectors has an incompatible signature across API
	// versions (v1 non-variadic, v1alpha2 variadic).
	Selectors       map[string]string
	ServiceTemplate ofst.ServiceTemplateSpec
	// ExtraPorts are appended to the exporter port before templating (e.g. postgres raft metrics).
	ExtraPorts []core.ServicePort
}

func (o StatsServiceOptions) Ensure(ctx context.Context) (kutil.VerbType, error) {
	if o.Monitor == nil || o.Monitor.Agent.Vendor() != mona.VendorPrometheus {
		return kutil.VerbUnchanged, nil
	}

	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.DB.StatsService().ServiceName(),
			Namespace: o.DB.GetNamespace(),
		},
	}

	ports := append([]core.ServicePort{
		{
			Name:       mona.PrometheusExporterPortName,
			Port:       o.Monitor.Prometheus.Exporter.Port,
			TargetPort: intstr.FromString(mona.PrometheusExporterPortName),
		},
	}, o.ExtraPorts...)

	vt, err := clientutil.CreateOrPatch(ctx, o.KBClient, svc, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*core.Service)
		core_util.EnsureOwnerReference(&in.ObjectMeta, o.DB.AsOwner())
		in.Labels = o.DB.StatsServiceLabels()
		in.Annotations = meta_util.OverwriteKeys(in.Annotations, o.ServiceTemplate.Annotations)

		in.Spec.Selector = o.Selectors
		in.Spec.Ports = ofst.PatchServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, ports),
			o.ServiceTemplate.Spec.Ports,
		)
		if o.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = o.ServiceTemplate.Spec.ClusterIP
		}
		if o.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = o.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = o.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = o.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = o.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = o.ServiceTemplate.Spec.ExternalTrafficPolicy
		in.Spec.SessionAffinityConfig = o.ServiceTemplate.Spec.SessionAffinityConfig
		if o.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = o.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return vt, nil
}
