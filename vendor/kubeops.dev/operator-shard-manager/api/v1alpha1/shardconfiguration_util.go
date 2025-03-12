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
	"strings"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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
	shardKey := fmt.Sprintf("shard.%s/%s-ID", SchemeGroupVersion.Group, shardConfigName)
	val, ok := labels[shardKey]
	if !ok {
		return ""
	}
	return val
}

func ShouldReconcileByShard(shardId, shardConfigName string, c client.Client) (bool, error) {
	hostName, err := getPodHostname()
	if err != nil {
		return false, err
	}
	ns, err := getPodNamespace()
	if err != nil {
		return false, err
	}

	deploymentName := deploymentNameFromHostname(hostName)
	shardConfig, err := fetchShardConfiguration(shardConfigName, c)
	if err != nil {
		return false, err
	}
	pods := getPodNamesFromShardConfig(deploymentName, ns, shardConfig)
	return isShardIdAndHostnameMatched(hostName, shardId, pods), nil
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

func deploymentNameFromHostname(hostName string) string {
	parts := strings.Split(hostName, "-")
	return strings.Join(parts[:len(parts)-2], "-")
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

func getPodNamesFromShardConfig(deploymentName string, ns string, shardConfig *ShardConfiguration) []string {
	var pods []string
	for _, ca := range shardConfig.Status.Controllers {
		if ca.APIGroup == apps.GroupName && ca.Kind == "Deployment" && ca.Name == deploymentName && ca.Namespace == ns {
			pods = ca.Pods
			break
		}
	}
	return pods
}

func isShardIdAndHostnameMatched(hostName, shardId string, pods []string) bool {
	for i, pod := range pods {
		if pod == hostName && strconv.Itoa(i) == shardId {
			return true
		}
	}
	return false
}
