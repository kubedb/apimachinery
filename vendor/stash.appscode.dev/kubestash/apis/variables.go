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

const (
	KeyNodeName = "NODE_NAME"

	DefaultHost = "host-0"

	DockerRegistry = "dockerRegistry"
	DockerImage    = "dockerImage"
	ImageTag       = "imageTag"

	AddonImage  = "addonImage"
	PodName     = "podName"
	InvokerKind = "invokerKind"
	InvokerName = "invokerName"

	Namespace      = "namespace"
	BackupSession  = "backupSession"
	RestoreSession = "restoreSession"

	RepositoryName      = "repositoryName"
	RepositoryNamespace = "repositoryNamespace"

	PushgatewayURL    = "prometheusPushgatewayURL"
	PrometheusJobName = "prometheusJobName"

	TargetName      = "targetName"
	TargetKind      = "targetKind"
	TargetNamespace = "targetNamespace"
	TargetMountPath = "targetMountPath"
	TargetPaths     = "targetPaths"

	TargetAppReplicas = "targetAppReplicas"

	// default true
	// false when TmpDir.DisableCaching is true in backupConfig/restoreSession
	EnableCache    = "enableCache"
	InterimDataDir = "interimDataDir"

	InterimDataDirPath        = "/kubestash-interim-volume/data"
	KubeStashTmpDirMountPath  = "/kubestash-tmp"
	KubeStashDefaultMountPath = "/kubestash-data"

	// License related constants
	LicenseApiService = "licenseApiservice"
)
