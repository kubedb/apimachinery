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

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"kmodules.xyz/client-go/cluster"
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
	if cluster.IsRancherManaged(kbClient.RESTMapper()) {
		namespaces, err := cluster.ListSiblingNamespaces(kbClient, dbNs)
		if err != nil {
			return nil, err
		}
		ret := make([]string, 0, len(namespaces))
		for i, namespace := range namespaces {
			ret[i] = namespace.Name
		}
		return ret, nil
	}
	return nil, nil
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
