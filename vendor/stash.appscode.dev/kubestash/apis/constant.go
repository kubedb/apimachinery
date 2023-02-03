/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

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
)

// =================== Keys for structure logging =====================
const (
	KeyTargetKind      = "target_kind"
	KeyTargetName      = "target_name"
	KeyTargetNamespace = "target_namespace"
	KeyReason          = "reason"
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
