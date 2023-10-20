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

package apis

import "time"

const (
	RequeueTimeInterval = 10 * time.Second

	KubeStashCleanupFinalizer = "kubestash.com/cleanup"
	KubeStashKey              = "kubestash.com"
	KubeDBGroupName           = "kubedb.com"
)

// =================== Keys for structure logging =====================
const (
	KeyTargetKind      = "target_kind"
	KeyTargetName      = "target_name"
	KeyTargetNamespace = "target_namespace"
	KeyReason          = "reason"
	KeyName            = "name"
)

const (
	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindClusterRole = "ClusterRole"
)

const (
	Requeue      = true
	DoNotRequeue = false
)

const (
	OwnerKey = ".metadata.controller"
)
