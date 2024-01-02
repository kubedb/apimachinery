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

package v1alpha1

import "kubedb.dev/apimachinery/apis/kafka"

const (
	LabelRole = kafka.GroupName + "/role"
	RoleStats = "stats"

	ComponentKafka   = "kafka"
	DefaultStatsPath = "/metrics"
)

// ConnectCluster constants

const (
	ConnectClusterContainerName = "connect-cluster"
	ConnectClusterModeEnv       = "CONNECT_CLUSTER_MODE"

	ConnectClusterOperatorVolumeConfig  = "connect-operator-config"
	ConnectClusterCustomVolumeConfig    = "connect-custom-config"
	ConnectorPluginsVolumeName          = "connector-plugins"
	ConnectClusterAuthSecretVolumeName  = "connect-cluster-auth"
	ConnectClusterOffsetFileDirName     = "connect-stand-offset"
	KafkaClientCertVolumeName           = "kafka-client-ssl"
	ConnectClusterServerCertsVolumeName = "server-certs"

	ConnectClusterOperatorConfigPath   = "/opt/kafka/config/connect-operator-config"
	ConnectorPluginsVolumeDir          = "/opt/kafka/libs/connector-plugins"
	ConnectClusterAuthSecretVolumePath = "/var/private/basic-auth"
	ConnectClusterOffsetFileDir        = "/var/log/connect"
	ConnectClusterCustomConfigPath     = "/opt/kafka/config/connect-custom-config"
	KafkaClientCertDir                 = "/var/private/kafka-client-ssl"
	ConnectClusterServerCertVolumeDir  = "/var/private/ssl"
)
