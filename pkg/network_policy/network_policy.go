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

package network_policy

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cu "kmodules.xyz/client-go/client"
	meta_util "kmodules.xyz/client-go/meta"
	kubestashapi "kubestash.dev/apimachinery/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NetworkPolicyNameHealthCheck = "kubedb-healthcheck"
	NetworkPolicyNameDBInternal  = "kubedb-database-internal"
	NetworkPolicyNameDBBackup    = "kubedb-database-backup"
)

func EnsureNetworkPolicy(kbClient client.Client, dbNs string) error {
	err := ensureHealthCheckerNetworkPolicy(kbClient, dbNs)
	if err != nil {
		return err
	}
	err = ensureDBInternalNetworkPolicy(kbClient, dbNs)
	if err != nil {
		return err
	}
	return ensureBackupNetworkPolicy(kbClient, dbNs)
}

func ensureHealthCheckerNetworkPolicy(kbClient client.Client, dbNs string) error {
	netPol := netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyNameHealthCheck,
			Namespace: dbNs,
		},
	}
	_, err := cu.CreateOrPatch(context.TODO(), kbClient, &netPol, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*netv1.NetworkPolicy)
		in.Spec.PodSelector = metav1.LabelSelector{
			MatchLabels: api.GetSelectorForNetworkPolicy(),
		}
		in.Spec.Ingress = []netv1.NetworkPolicyIngressRule{
			{
				From: []netv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								corev1.LabelMetadataName: meta_util.PodNamespace(),
							},
						},
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								meta_util.InstanceLabelKey: "kubedb",
							},
						},
					},
				},
			},
		}
		in.Spec.PolicyTypes = []netv1.PolicyType{
			netv1.PolicyTypeIngress,
		}
		return in
	})
	return err
}

func ensureDBInternalNetworkPolicy(kbClient client.Client, dbNs string) error {
	netPol := netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyNameDBInternal,
			Namespace: dbNs,
		},
	}

	_, err := cu.CreateOrPatch(context.TODO(), kbClient, &netPol, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*netv1.NetworkPolicy)
		in.Spec.PodSelector = metav1.LabelSelector{
			MatchLabels: api.GetSelectorForNetworkPolicy(),
		}
		in.Spec.Ingress = []netv1.NetworkPolicyIngressRule{
			{
				From: []netv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								corev1.LabelMetadataName: dbNs,
							},
						},
					},
				},
			},
		}
		in.Spec.Egress = []netv1.NetworkPolicyEgressRule{
			{
				To: []netv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								corev1.LabelMetadataName: dbNs,
							},
						},
					},
					{
						IPBlock: &netv1.IPBlock{
							CIDR: "0.0.0.0/0",
						},
					},
				},
			},
		}
		in.Spec.PolicyTypes = []netv1.PolicyType{
			netv1.PolicyTypeIngress,
			netv1.PolicyTypeEgress,
		}
		return in
	})
	return err
}

func ensureBackupNetworkPolicy(kbClient client.Client, dbNs string) error {
	netPol := netv1.NetworkPolicy{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      NetworkPolicyNameDBBackup,
			Namespace: dbNs,
		},
	}

	_, err := cu.CreateOrPatch(context.TODO(), kbClient, &netPol, func(obj client.Object, createOp bool) client.Object {
		in := obj.(*netv1.NetworkPolicy)
		in.Spec.PodSelector = metav1.LabelSelector{
			MatchLabels: map[string]string{
				meta_util.ManagedByLabelKey: kubestashapi.KubeStashKey,
			},
		}
		in.Spec.Egress = []netv1.NetworkPolicyEgressRule{
			{
				To: []netv1.NetworkPolicyPeer{
					{
						NamespaceSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								corev1.LabelMetadataName: dbNs,
							},
						},
					},
					{
						IPBlock: &netv1.IPBlock{
							CIDR: "0.0.0.0/0",
						},
					},
				},
			},
		}
		in.Spec.PolicyTypes = []netv1.PolicyType{
			netv1.PolicyTypeEgress,
		}
		return in
	})
	return err
}
