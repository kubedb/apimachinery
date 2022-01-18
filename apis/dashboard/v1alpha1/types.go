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

import (
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DashboardPhase string

const (
	// used for Dashboards that are currently provisioning
	DashboardPhaseProvisioning DashboardPhase = "Provisioning"
	// used for Dashboards that are provisionined / deployments and serices are ready
	DashboardPhaseProvisioned DashboardPhase = "Provisioned"
	// used for Dashboards that are currently ReplicaReady, AcceptingConnection and Ready
	DashboardPhaseReady DashboardPhase = "Ready"
	// used for Dashboards that can connect, ReplicaReady == false || Ready == false (eg, ES yellow)
	DashboardPhaseCritical DashboardPhase = "Critical"
	// used for Dashboards that can't connect
	DashboardPhaseNotReady DashboardPhase = "NotReady"
)

type DashboardConditionType string

const (
	DashboardConditionInitialized         DashboardConditionType = "Initializing"
	DashboardConditionDeploymentAvailable DashboardConditionType = "DeploymentAvailable"
	DashboardConditionServiceReady        DashboardConditionType = "ServiceReady"
	DashboardConditionAcceptingConnection DashboardConditionType = "AcceptingConnection"
	DashboardConditionStateGreenOrYellow  DashboardConditionType = "ServerReady"
	DashboardConditionStateRed            DashboardConditionType = "ServerNotReady"
)

type DashboardConditionReason string

const (
	DashboardDeploymentAvailable                DashboardConditionReason = "MinimumReplicasAvailable"
	DashboardDeploymentNotAvailable             DashboardConditionReason = "MinimumReplicasNotAvailable"
	DashboardServiceReady                       DashboardConditionReason = "ServiceAcceptingRequests"
	DashboardServiceNotReady                    DashboardConditionReason = "ServiceNotAcceptingRequests"
	DashboardAcceptingConnectionRequest         DashboardConditionReason = "DashboardAcceptingConnectionRequests"
	DashboardNotAcceptingConnectionRequest      DashboardConditionReason = "DashboardNotAcceptingConnectionRequests"
	DashboardReadinessCheckSucceeded            DashboardConditionReason = "DashboardReadinessCheckSucceeded"
	DashboardReadinessCheckSucceededWithWarning DashboardConditionReason = "DashboardReadinessCheckSucceeded"
	DashboardReadinessCheckFailed               DashboardConditionReason = "DashboardReadinessCheckFailed"
)

type DashboardStatus string

const (
	Available     DashboardStatus = "Available"
	StatusOK      DashboardStatus = "OK"
	StatusWarning DashboardStatus = "Warning"
	StatusError   DashboardStatus = "Error"
)

const (
	ResourceCodeElasticsearchDashboard     = "ed"
	ResourceKindElasticsearchDashboard     = "ElasticsearchDashboard"
	ResourceSingularElasticsearchDashboard = "elasticsearchdashboard"
	ResourcePluralElasticsearchDashboard   = "elasticsearchdashboards"
)

type ElasticsearchDashboardCertificateAlias string

const (
	ElasticsearchDashboardCACert           ElasticsearchDashboardCertificateAlias = "ca"
	ElasticsearchDatabaseClient            ElasticsearchDashboardCertificateAlias = "database-client"
	ElasticsearchDashboardKibanaServerCert ElasticsearchDashboardCertificateAlias = "kibana-server"
	ElasticsearchDashboardConfig           ElasticsearchDashboardCertificateAlias = "dashboard-config"
)

type ServiceAlias string

const (
	PrimaryServiceAlias ServiceAlias = "primary"
	StatsServiceAlias   ServiceAlias = "stats"
)

type ServerState string

var (
	StateGreen  ServerState = "green"
	StateYellow ServerState = "yellow"
	StateRed    ServerState = "red"
)

var (
	ElasticsearchDashboardKibanaConfigDir = "/usr/share/kibana/config"
	ElasticsearchDashboardDefaultPort     = (int32)(5601)
	ElasticsearchDashboardCpuReq          = "500m"
	ElasticsearchDashboardMemReq          = "1Gi"
	ElasticsearchDashboardMemLimit        = "1Gi"
)

const CaCertKey string = "ca.crt"
const ComponentDashboard string = "dashboard"
const KibanaHealthStatusAPi string = "/api/status"
const KibanaSecretDataKey string = "kibana.yml"
const DefaultDatabaseClientCertSuffix = "archiver-cert"
const HealthCheckInterval = time.Second * 30
const GlobalHost = "0.0.0.0"

const (
	ES_USER_ENV     = "ELASTICSEARCH_USERNAME"
	ES_PASSWORD_ENV = "ELASTICSEARCH_PASSWORD"
)

var (
	DeletionPolicy         = meta.DeletePropagationForeground
	GracefulDeletionPeriod = (int64)(time.Duration(time.Second * 3))
)

type CertSecrets struct {
	Ca  *core.Secret
	Crt *core.Secret
}
