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

package network_policy

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cu "kmodules.xyz/client-go/client"
	meta_util "kmodules.xyz/client-go/meta"
	kubestashapi "kubestash.dev/apimachinery/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ciliumNetworkPolicyGVK is the GroupVersionKind of CiliumNetworkPolicy as
// installed by Cilium. We use unstructured.Unstructured rather than vendoring
// github.com/cilium/cilium to keep the apimachinery dependency tree small.
var ciliumNetworkPolicyGVK = schema.GroupVersionKind{
	Group:   "cilium.io",
	Version: "v2",
	Kind:    "CiliumNetworkPolicy",
}

func ensureCiliumPolicies(kbClient client.Client, dbNs string) error {
	if err := ensureCiliumHealthCheckerPolicy(kbClient, dbNs); err != nil {
		return err
	}
	if err := ensureCiliumDBInternalPolicy(kbClient, dbNs); err != nil {
		return err
	}
	if err := ensureCiliumKubeAPIPolicy(kbClient, dbNs); err != nil {
		return err
	}
	return ensureCiliumBackupPolicy(kbClient, dbNs)
}

// ensureCiliumNetworkPolicy upserts a CiliumNetworkPolicy with the given name,
// namespace, and spec. The spec is provided as a plain Go map matching the
// cilium.io/v2 CRD schema; using unstructured avoids vendoring cilium.
func ensureCiliumNetworkPolicy(kbClient client.Client, namespace, name string, spec map[string]interface{}) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(ciliumNetworkPolicyGVK)
	u.SetName(name)
	u.SetNamespace(namespace)
	_, err := cu.CreateOrPatch(context.TODO(), kbClient, u, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*unstructured.Unstructured)
		in.SetGroupVersionKind(ciliumNetworkPolicyGVK)
		_ = unstructured.SetNestedMap(in.Object, spec, "spec")
		return in
	})
	return err
}

func ensureCiliumHealthCheckerPolicy(kbClient client.Client, dbNs string) error {
	spec := map[string]interface{}{
		"endpointSelector": map[string]interface{}{
			"matchLabels": stringMapToInterface(api.GetSelectorForNetworkPolicy()),
		},
		"ingress": []interface{}{
			map[string]interface{}{
				"fromEndpoints": []interface{}{
					map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"k8s:" + corev1.LabelMetadataName:   meta_util.PodNamespace(),
							"k8s:" + meta_util.InstanceLabelKey: "kubedb",
						},
					},
				},
			},
		},
	}
	return ensureCiliumNetworkPolicy(kbClient, dbNs, NetworkPolicyNameHealthCheck, spec)
}

func ensureCiliumDBInternalPolicy(kbClient client.Client, dbNs string) error {
	spec := map[string]interface{}{
		"endpointSelector": map[string]interface{}{
			"matchLabels": stringMapToInterface(api.GetSelectorForNetworkPolicy()),
		},
		"ingress": []interface{}{
			map[string]interface{}{
				"fromEndpoints": []interface{}{
					map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"k8s:" + corev1.LabelMetadataName: dbNs,
						},
					},
				},
			},
		},
		"egress": []interface{}{
			// Intra-namespace egress to peer DB pods (replication, gossip, etc).
			map[string]interface{}{
				"toEndpoints": []interface{}{
					map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"k8s:" + corev1.LabelMetadataName: dbNs,
						},
					},
				},
			},
			// World egress (image pulls, telemetry, etc).
			map[string]interface{}{
				"toEntities": []interface{}{"world"},
			},
		},
	}
	return ensureCiliumNetworkPolicy(kbClient, dbNs, NetworkPolicyNameDBInternal, spec)
}

// ensureCiliumKubeAPIPolicy allows DB pods to reach the Kubernetes API server.
// Cluster databases (>1 replica) need this for leader election and operator
// SDK interactions; standalone DBs are unaffected by its presence.
func ensureCiliumKubeAPIPolicy(kbClient client.Client, dbNs string) error {
	spec := map[string]interface{}{
		"endpointSelector": map[string]interface{}{
			"matchLabels": stringMapToInterface(api.GetSelectorForNetworkPolicy()),
		},
		"egress": []interface{}{
			map[string]interface{}{
				"toEntities": []interface{}{"kube-apiserver"},
			},
		},
	}
	return ensureCiliumNetworkPolicy(kbClient, dbNs, "kubedb-kube-apiserver", spec)
}

func ensureCiliumBackupPolicy(kbClient client.Client, dbNs string) error {
	spec := map[string]interface{}{
		"endpointSelector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"k8s:" + meta_util.ManagedByLabelKey: kubestashapi.KubeStashKey,
			},
		},
		"egress": []interface{}{
			// Reach the DB pods in this namespace.
			map[string]interface{}{
				"toEndpoints": []interface{}{
					map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"k8s:" + corev1.LabelMetadataName: dbNs,
						},
					},
				},
			},
			// Reach object storage / external endpoints.
			map[string]interface{}{
				"toEntities": []interface{}{"world"},
			},
		},
	}
	return ensureCiliumNetworkPolicy(kbClient, dbNs, NetworkPolicyNameDBBackup, spec)
}

func stringMapToInterface(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
