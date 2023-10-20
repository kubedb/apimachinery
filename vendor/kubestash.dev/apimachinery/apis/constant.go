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

const (
	OwnerKey = ".metadata.controller"
)

const (
	KubeStashBackupComponent      = "kubestash-backup"
	KubeStashRestoreComponent     = "kubestash-restore"
	KubeStashInitializerComponent = "kubestash-initializer"
	KubeStashUploaderComponent    = "kubestash-uploader"
	KubeStashCleanerComponent     = "kubestash-cleaner"
	KubeStashHookComponent        = "kubestash-hook"
)

// Keys for offshoot labels
const (
	KubeStashInvokerName      = "kubestash.com/invoker-name"
	KubeStashInvokerNamespace = "kubestash.com/invoker-namespace"
	KubeStashInvokerKind      = "kubestash.com/invoker-kind"

	KubeStashApp = "kubestash.com/app"
)

// Keys for structure logging
const (
	KeyTargetKind      = "target_kind"
	KeyTargetName      = "target_name"
	KeyTargetNamespace = "target_namespace"
	KeyReason          = "reason"
	KeyName            = "name"
)

// Keys for BackupBlueprint
const (
	VariablesKey       = "variables.kubestash.com"
	BackupBlueprintKey = "blueprint.kubestash.com"

	KeyBlueprintName      = BackupBlueprintKey + "/name"
	KeyBlueprintNamespace = BackupBlueprintKey + "/namespace"
	KeyBlueprintSessions  = BackupBlueprintKey + "/sessions"
)

// RBAC related constants
const (
	KindClusterRole = "ClusterRole"
	KindRole        = "Role"

	KubeStashBackupJobClusterRole       = "kubestash-backup-job"
	KubeStashRestoreJobClusterRole      = "kubestash-restore-job"
	KubeStashCronJobClusterRole         = "kubestash-cron-job"
	KubeStashBackendJobClusterRole      = "kubestash-backend-job"
	KubeStashBackendAccessorClusterRole = "kubestash-backend-accessor"
)

// Reconciliation related constants
const (
	Requeue      = true
	DoNotRequeue = false
)

// Workload related constants
const (
	EnvComponentName = "COMPONENT_NAME"

	ComponentPod        = "pod"
	ComponentDeployment = "deployment"

	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindDeployment  = "Deployment"
)

// PersistentVolumeClaim related constants
const (
	KindPersistentVolumeClaim = "PersistentVolumeClaim"
	KeyPodOrdinal             = "POD_ORDINAL"
	ComponentPVC              = "pvc"
	PVCName                   = "PVC_NAME"
)

const (
	PrefixTrigger         = "trigger"
	PrefixInit            = "init"
	PrefixUpload          = "upload"
	PrefixCleanup         = "cleanup"
	PrefixRetentionPolicy = "retentionpolicy"
)

// InterimVolume related constants
const (
	InterimVolume = "interim-volume"
)

// Local Network Volume Accessor related constants
const (
	KubeStashNetVolAccessor = "kubestash-netvol-accessor"
	TempDirVolumeName       = "kubestash-temp-dir"
	TempDirMountPath        = "/tmp"
	OperatorContainer       = "operator"
)

const (
	DirRepository = "repository"
)
