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

package archiver

import (
	"context"
	"slices"
	"sort"

	archiverapi "kubedb.dev/apimachinery/apis/archiver/v1alpha1"
	"kubedb.dev/apimachinery/pkg/double_optin"

	authv1 "k8s.io/api/authorization/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/client-go/cluster"
	identityapi "kmodules.xyz/resource-metadata/apis/identity/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCorrespondingArchiver(kbClient client.Client, dbMeta metav1.ObjectMeta, archiverList []archiverapi.Accessor) (*metav1.ObjectMeta, error) {
	projectNSList, err := GetSameProjectNamespaces(kbClient, dbMeta.Namespace)
	if err != nil {
		return nil, err
	}

	var priorityList []priority
	for _, archiver := range archiverList {
		var archiverNs core.Namespace
		err := kbClient.Get(context.TODO(), types.NamespacedName{
			Name: archiver.GetObjectMeta().Namespace,
		}, &archiverNs)
		if err != nil {
			return nil, err
		}

		var dbNs core.Namespace
		err = kbClient.Get(context.TODO(), types.NamespacedName{
			Name: dbMeta.Namespace,
		}, &dbNs)
		if err != nil {
			return nil, err
		}

		possible, err := double_optin.CheckIfDoubleOptInPossible(dbMeta, dbNs.ObjectMeta, archiverNs.ObjectMeta, archiver.GetConsumers())
		if err != nil {
			return nil, err
		}
		if possible {
			priorityList = append(priorityList, getPriority(archiver.GetObjectMeta(), projectNSList, dbMeta.Namespace))
		}
	}
	if priorityList == nil {
		return nil, err
	}
	sort.Slice(priorityList, func(i, j int) bool {
		return priorityList[i].index < priorityList[j].index
	})
	return &priorityList[0].archiver, nil
}

func GetSameProjectNamespaces(kbClient client.Client, dbNs string) ([]string, error) {
	if isUIServerAndRancher(kbClient.RESTMapper()) {
		nsReview := identityapi.SelfSubjectNamespaceAccessReview{
			Spec: identityapi.SelfSubjectNamespaceAccessReviewSpec{
				ResourceAttributes: []authv1.ResourceAttributes{
					{
						Namespace: "",
						Verb:      "get",
						Group:     archiverapi.SchemeGroupVersion.Group,
						Version:   archiverapi.SchemeGroupVersion.Version,
						Resource:  "*",
					},
				},
				NonResourceAttributes: nil,
			},
		}
		err := kbClient.Create(context.TODO(), &nsReview)
		if err != nil {
			return nil, err
		}
		for _, nsList := range nsReview.Status.Projects {
			if slices.Contains(nsList, dbNs) {
				return nsList, nil
			}
		}
	}
	return nil, nil
}

func isUIServerAndRancher(mapper meta.RESTMapper) bool {
	_, err := mapper.RESTMapping(schema.GroupKind{
		Group: identityapi.GroupName,
		Kind:  identityapi.ResourceKindSelfSubjectNamespaceAccessReview,
	})
	return err == nil && cluster.IsRancherManaged(mapper)
}

type priority struct {
	archiver metav1.ObjectMeta
	index    int
}

func getPriority(archiver metav1.ObjectMeta, projectNSList []string, dbNs string) priority {
	idx := 2
	if archiver.Namespace == dbNs {
		idx = 0
	} else if slices.Contains(projectNSList, archiver.Namespace) {
		idx = 1
	}
	return priority{archiver, idx}
}
