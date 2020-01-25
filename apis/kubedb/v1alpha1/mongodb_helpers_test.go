/*
Copyright The KubeDB Authors.

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
	"testing"

	"kubedb.dev/apimachinery/apis/catalog/v1alpha1"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
)

var testTopology = &core_util.Topology{
	Regions: map[string][]string{
		"us-east-1": {"us-east-1a", "us-east-1b", "us-east-1c"},
	},
	TotalNodes: 100,
	InstanceTypes: map[string]int{
		"n1-standard-4": 100,
	},
	LabelZone:         core.LabelZoneFailureDomain,
	LabelRegion:       core.LabelZoneRegion,
	LabelInstanceType: core.LabelInstanceType,
}

func TestMongoDB_HostAddress(t *testing.T) {
	mongodb := &MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-name",
			Namespace: "demo",
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: MongoDBSpec{
			Version: "3.6-v2",
			ShardTopology: &MongoDBShardingTopology{
				Shard: MongoDBShardNode{
					Shards: 3,
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				ConfigServer: MongoDBConfigNode{
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				Mongos: MongoDBMongosNode{
					MongoDBNode: MongoDBNode{
						Replicas: 2,
					},
				},
			},
		},
	}

	mongodb.SetDefaults(&v1alpha1.MongoDBVersion{}, testTopology)

	shardDSN := mongodb.HostAddress()
	t.Log(shardDSN)

	mongodb.Spec.ShardTopology = nil
	mongodb.Spec.Replicas = types.Int32P(3)
	mongodb.Spec.ReplicaSet = &MongoDBReplicaSet{
		Name: "mgo-rs",
	}

	repsetDSN := mongodb.HostAddress()
	t.Log(repsetDSN)

}

func TestMongoDB_ShardDSN(t *testing.T) {
	mongodb := &MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-name",
			Namespace: "demo",
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: MongoDBSpec{
			Version: "3.6-v2",
			ShardTopology: &MongoDBShardingTopology{
				Shard: MongoDBShardNode{
					Shards: 3,
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				ConfigServer: MongoDBConfigNode{
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				Mongos: MongoDBMongosNode{
					MongoDBNode: MongoDBNode{
						Replicas: 2,
					},
				},
			},
		},
	}

	shardDSN := mongodb.ShardDSN(0)
	t.Log(shardDSN)

	mongodb.SetDefaults(&v1alpha1.MongoDBVersion{}, testTopology)
}

func TestMongoDB_ConfigSvrDSN(t *testing.T) {
	mongodb := &MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-name",
			Namespace: "demo",
			Labels: map[string]string{
				"app": "kubedb",
			},
		},
		Spec: MongoDBSpec{
			Version: "3.6-v2",
			ShardTopology: &MongoDBShardingTopology{
				Shard: MongoDBShardNode{
					Shards: 3,
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				ConfigServer: MongoDBConfigNode{
					MongoDBNode: MongoDBNode{
						Replicas: 3,
					},
					Storage: &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP("standard"),
					},
				},
				Mongos: MongoDBMongosNode{
					MongoDBNode: MongoDBNode{
						Replicas: 2,
					},
				},
			},
		},
	}

	configDSN := mongodb.ConfigSvrDSN()
	t.Log(configDSN)
}

func TestMongoDB_SetDefaults(t *testing.T) {
	mongodb := &MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-sample",
			Namespace: "demo",
		},
		Spec: MongoDBSpec{
			Version: "3.6-v2",
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: types.StringP("standard"),
			},
		},
	}

	mongodb.SetDefaults(&v1alpha1.MongoDBVersion{}, testTopology)
}
