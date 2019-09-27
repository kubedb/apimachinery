package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestElasticsearch_SetDefaults(t *testing.T) {
	elasticsearch := &Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-sample",
			Namespace: "demo",
		},
		Spec: ElasticsearchSpec{
			Version: "7.2.0",
			EnableSecurity: types.BoolP(false),
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

	elasticsearch.Spec.SetDefaults()

	fmt.Println(elasticsearch)
}
