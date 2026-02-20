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

package hanadb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	apiutils "kubedb.dev/apimachinery/pkg/utils"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SystemReplicationStatus struct {
	Status        string
	Details       string
	ReplayBacklog string
}

type SystemReplicationHealthSummary struct {
	AllHealthy bool
	HasActive  bool
	HasSyncing bool
	HasError   bool
	Summary    string
}

const (
	SystemReplicationStatusColumn        = "REPLICATION_STATUS"
	SystemReplicationStatusDetailsColumn = "REPLICATION_STATUS_DETAILS"
	SystemReplicationReplayBacklogColumn = "REPLAY_BACKLOG"
)

const SystemReplicationStatusQuery = `
SELECT REPLICATION_STATUS, REPLICATION_STATUS_DETAILS,
       (LAST_LOG_POSITION - REPLAYED_LOG_POSITION) AS REPLAY_BACKLOG
FROM SYS.M_SERVICE_REPLICATION`

// GetGoverningServiceDNSName returns the pod DNS record under the governing service.
func GetGoverningServiceDNSName(podName string, db *api.HanaDB) string {
	return fmt.Sprintf("%s.%s.%s.svc.%s", podName, db.GoverningServiceName(), db.Namespace, apiutils.FindDomain())
}

// GetPrimaryServiceDNSName returns the primary service DNS name.
func GetPrimaryServiceDNSName(db *api.HanaDB) string {
	return fmt.Sprintf("%s.%s.svc", db.ServiceName(), db.Namespace)
}

// GetPodFQDN joins pod and governing service DNS names.
func GetPodFQDN(podName, governingServiceDNSName string) string {
	return fmt.Sprintf("%s.%s", podName, governingServiceDNSName)
}

func NewSystemReplicationStatus(status, details, replayBacklog string) SystemReplicationStatus {
	return SystemReplicationStatus{
		Status:        strings.ToUpper(strings.TrimSpace(status)),
		Details:       strings.TrimSpace(details),
		ReplayBacklog: strings.TrimSpace(replayBacklog),
	}
}

func ParseSystemReplicationStatuses(rows []map[string]string) []SystemReplicationStatus {
	statuses := make([]SystemReplicationStatus, 0, len(rows))
	for _, row := range rows {
		status := NewSystemReplicationStatus(
			row[SystemReplicationStatusColumn],
			row[SystemReplicationStatusDetailsColumn],
			row[SystemReplicationReplayBacklogColumn],
		)
		if status.Status == "" {
			continue
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func EvaluateSystemReplicationHealth(statuses []SystemReplicationStatus) SystemReplicationHealthSummary {
	summary := SystemReplicationHealthSummary{
		AllHealthy: true,
	}
	if len(statuses) == 0 {
		summary.AllHealthy = false
		summary.Summary = "no replication status found"
		return summary
	}

	statusParts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		replStatus := strings.ToUpper(strings.TrimSpace(status.Status))
		replDetails := strings.TrimSpace(status.Details)
		backlog := strings.TrimSpace(status.ReplayBacklog)
		if replStatus == "" {
			continue
		}

		statusPart := replStatus
		if backlog != "" && backlog != "0" {
			statusPart += "(backlog:" + backlog + ")"
		}
		if replDetails != "" && replStatus != "ACTIVE" {
			statusPart += "[" + replDetails + "]"
		}
		statusParts = append(statusParts, statusPart)

		switch replStatus {
		case "ACTIVE":
			summary.HasActive = true
		case "SYNCING", "INITIALIZING", "UNKNOWN":
			summary.HasSyncing = true
		case "ERROR":
			summary.HasError = true
		default:
			summary.HasSyncing = true
		}

		if !isSystemReplicationMemberHealthy(replStatus, replDetails) {
			summary.AllHealthy = false
		}
	}

	if len(statusParts) == 0 {
		summary.AllHealthy = false
		summary.Summary = "no replication status found"
		return summary
	}

	summary.Summary = strings.Join(statusParts, ", ")
	return summary
}

func isSystemReplicationMemberHealthy(status, details string) bool {
	if status != "ACTIVE" {
		return false
	}

	if details == "" {
		return true
	}

	normalizedDetails := strings.ToUpper(details)
	if strings.Contains(normalizedDetails, "DISCONNECT") ||
		strings.Contains(normalizedDetails, "ERROR") ||
		strings.Contains(normalizedDetails, "FAIL") ||
		strings.Contains(normalizedDetails, "SYNCING") ||
		strings.Contains(normalizedDetails, "INITIALIZ") ||
		strings.Contains(normalizedDetails, "UNKNOWN") {
		return false
	}

	// If details mention connectivity state, require connected.
	if strings.Contains(normalizedDetails, "CONNECT") &&
		!strings.Contains(normalizedDetails, "CONNECTED") {
		return false
	}

	return true
}

// GetAuthCredentialsFromSecret reads SYSTEM user/password from the auth secret.
func GetAuthCredentialsFromSecret(ctx context.Context, kc client.Client, db *api.HanaDB) (string, string, error) {
	secret := &core.Secret{}
	if err := kc.Get(ctx, types.NamespacedName{
		Namespace: db.Namespace,
		Name:      db.GetAuthSecretName(),
	}, secret); err != nil {
		return "", "", err
	}

	user := kubedb.HanaDBSystemUser
	if usernameBytes, ok := secret.Data[core.BasicAuthUsernameKey]; ok && len(usernameBytes) > 0 {
		user = string(usernameBytes)
	}

	if passwordBytes, ok := secret.Data[core.BasicAuthPasswordKey]; ok && len(passwordBytes) > 0 {
		return user, string(passwordBytes), nil
	}

	passwordJSON, ok := secret.Data[kubedb.HanaDBPasswordFileKey]
	if !ok {
		return "", "", fmt.Errorf("secret %s/%s missing %s key", secret.Namespace, secret.Name, kubedb.HanaDBPasswordFileKey)
	}

	var passwordData struct {
		MasterPassword string `json:"master_password"`
	}
	if err := json.Unmarshal(passwordJSON, &passwordData); err != nil {
		return "", "", fmt.Errorf("failed to parse %s in secret %s/%s: %v", kubedb.HanaDBPasswordFileKey, secret.Namespace, secret.Name, err)
	}
	if passwordData.MasterPassword == "" {
		return "", "", fmt.Errorf("master password not specified in secret %s/%s", secret.Namespace, secret.Name)
	}

	return user, passwordData.MasterPassword, nil
}
