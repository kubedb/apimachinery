/*
Copyright AppsCode Inc. and Contributors.

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
	"context"
	"fmt"
	"os"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kmapi "kmodules.xyz/client-go/api/v1"
	kapiutil "kmodules.xyz/client-go/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ShouldEnqueueObjectForShard(kbClient client.Client, shardConfig string, labels map[string]string) bool {
	if shardConfig == "" {
		return true
	}
	if labels == nil {
		klog.Warningf("shard-config provided, but labels is nil, skip enqueuing object")
		return false
	}
	shardId := ExtractShardKeyFromLabels(labels, shardConfig)
	if shardId == "" {
		klog.Warningf("shard-config provided, but no shardId found in the labels, skip enqueuing object")
		return false
	}
	requeue, err := ShouldReconcileByShard(shardId, shardConfig, kbClient)
	if err != nil {
		klog.Warningf("ShouldReconcileByShard failed with err: %v", err)
		return false
	}
	return requeue
}

func ExtractShardKeyFromLabels(labels map[string]string, shardConfigName string) string {
	// klog.Infof("got pg labels: %v", labels)
	shardKey := fmt.Sprintf("shard-index.%s/%s", SchemeGroupVersion.Group, shardConfigName)
	val, ok := labels[shardKey]
	if !ok {
		return ""
	}
	return val
}

func ShouldReconcileByShard(shardId, shardConfigName string, kbClient client.Client) (bool, error) {
	pods, err := GetPodListsFromShardConfig(kbClient, shardConfigName)
	if err != nil {
		return false, err
	}
	return isShardIdAndHostnameMatched(shardId, pods), nil
}

func GetPodListsFromShardConfig(kbClient client.Client, shardConfigName string) ([]string, error) {
	hostName, err := getPodHostname()
	if err != nil {
		return nil, err
	}
	ns, err := getPodNamespace()
	if err != nil {
		return nil, err
	}
	pod := &v1.Pod{}
	err = kbClient.Get(context.TODO(), types.NamespacedName{
		Name:      hostName,
		Namespace: ns,
	}, pod)
	if err != nil {
		return nil, err
	}
	pod.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "",
		Kind:  "Pod",
	})
	lineage, err := kapiutil.DetectLineage(context.TODO(), kbClient, pod)
	if err != nil {
		return nil, err
	}
	if len(lineage) == 0 {
		return nil, fmt.Errorf("no owner found for pod %s/%s", pod.Namespace, pod.Name)
	}

	shardConfig, err := fetchShardConfiguration(shardConfigName, kbClient)
	if err != nil {
		return nil, err
	}
	return getPodNamesFromShardConfig(lineage[0], shardConfig), nil
}

func getPodHostname() (string, error) {
	hostName := os.Getenv("HOSTNAME")
	if hostName == "" {
		return "", fmt.Errorf("HOSTNAME environment variable is empty")
	}
	return hostName, nil
}

func getPodNamespace() (string, error) {
	out, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", fmt.Errorf("failed to read namespace file: %w", err)
	}
	return string(out), nil
}

func fetchShardConfiguration(shardConfigName string, c client.Client) (*ShardConfiguration, error) {
	shardConfig := &ShardConfiguration{}
	err := c.Get(context.TODO(), types.NamespacedName{
		Name: shardConfigName,
	}, shardConfig)
	if err != nil {
		return nil, err
	}
	return shardConfig, nil
}

func getPodNamesFromShardConfig(objectInfo kmapi.ObjectInfo, shardConfig *ShardConfiguration) []string {
	var pods []string
	for _, ca := range shardConfig.Status.Controllers {
		if ca.APIGroup == objectInfo.Resource.Group && ca.Kind == objectInfo.Resource.Kind && ca.Name == objectInfo.Ref.Name && ca.Namespace == objectInfo.Ref.Namespace {
			pods = ca.Pods
			break
		}
	}
	return pods
}

func isShardIdAndHostnameMatched(shardId string, pods []string) bool {
	hostName := os.Getenv("HOSTNAME")
	for i, pod := range pods {
		if pod == hostName && strconv.Itoa(i) == shardId {
			return true
		}
	}
	return false
}
