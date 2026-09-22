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

package mongodb

import (
	"context"

	"kubedb.dev/apimachinery/apis/kubedb"
	"kubedb.dev/apimachinery/pkg/reconfigure"
)

// OpaqueFiles are configuration files that are executed rather than parsed as parameters, so any
// change to them is Static regardless of content.
var OpaqueFiles = []string{kubedb.MongoDBConfigurationJSFile}

// locked are parameters the provisioner rewrites on every reconcile. A user override is either
// ignored or breaks the cluster, so it is rejected at admission with the CRD field that does work.
var locked = map[string]string{
	"replication.replSetName":    "spec.replicaSet.name",
	"replication.oplogSizeMB":    "",
	"sharding.clusterRole":       "spec.shardTopology",
	"net.port":                   "",
	"net.bindIp":                 "",
	"net.tls.mode":               "spec.sslMode",
	"net.tls.certificateKeyFile": "spec.tls",
	"net.tls.CAFile":             "spec.tls",
	"net.tls.clusterFile":        "spec.tls",
	"storage.engine":             "spec.storageEngine",
	"storage.dbPath":             "",
	"security.keyFile":           "spec.keyFileSecret",
	"security.authorization":     "",
	"security.clusterAuthMode":   "spec.clusterAuthMode",
}

// lockedReason carries the "why", for the paths where naming a CRD field is not the whole answer.
var lockedReason = map[string]string{
	"replication.oplogSizeMB": "the operator sizes the oplog at provisioning time; resize a running member with a Reconfigure ops request that sets storage.oplogMinRetentionHours, or recreate the member",
	"net.port":                "the operator fixes the mongod/mongos listening port and wires every Service, probe and connection string to it",
	"net.bindIp":              "the operator binds the pod IP and localhost so that peers, probes and the exporter can all reach the member",
	"storage.dbPath":          "the operator mounts the data volume at a fixed path",
	"security.authorization":  "the operator always provisions MongoDB with authorization enabled",
}

// dynamic lists the parameters that can be applied to a running server. Everything absent from this
// map is Static: there are far too many MongoDB options to enumerate, so the default must be safe.
// Seeded from the MongoDB 8.0 tuning reference; the setParameter subtree is refined per server at
// reconfigure time by asking getParameter with showDetails.
var dynamic = map[string]bool{
	"operationProfiling.mode":                     true,
	"operationProfiling.slowOpThresholdMs":        true,
	"storage.oplogMinRetentionHours":              true,
	"storage.wiredTiger.engineConfig.cacheSizeGB": true,

	"setParameter.ttlMonitorEnabled":                            true,
	"setParameter.ttlMonitorSleepSecs":                          true,
	"setParameter.cursorTimeoutMillis":                          true,
	"setParameter.notablescan":                                  true,
	"setParameter.maxIndexBuildMemoryUsageMegabytes":            true,
	"setParameter.internalQueryMaxBlockingSortMemoryUsageBytes": true,
	"setParameter.internalQueryStatsRateLimit":                  true,
	"setParameter.transactionLifetimeLimitSeconds":              true,
	"setParameter.storageEngineConcurrentReadTransactions":      true,
	"setParameter.storageEngineConcurrentWriteTransactions":     true,
	"setParameter.wiredTigerConcurrentReadTransactions":         true,
	"setParameter.wiredTigerConcurrentWriteTransactions":        true,
	"setParameter.wiredTigerEngineRuntimeConfig":                true,
	"setParameter.oplogBatchDelayMillis":                        true,
	"setParameter.enableFlowControl":                            true,
	"setParameter.flowControlTargetLagSeconds":                  true,
	"setParameter.slowOpSampleRate":                             true,
	"setParameter.slowms":                                       true,
	"setParameter.logComponentVerbosity":                        true,
	"setParameter.ShardingTaskExecutorPoolMaxSize":              true,
	"setParameter.ShardingTaskExecutorPoolMinSize":              true,
}

// aliases map the names people write in tickets and spreadsheets onto the canonical config path,
// for classification lookup only. What the user wrote is never rewritten.
var aliases = map[string]string{
	"maxConns":                          "net.maxIncomingConnections",
	"journal.enabled":                   "storage.journal.enabled",
	"slowOpThresholdMs":                 "operationProfiling.slowOpThresholdMs",
	"slowms":                            "operationProfiling.slowOpThresholdMs",
	"profile":                           "operationProfiling.mode",
	"cursorTimeoutMillis":               "setParameter.cursorTimeoutMillis",
	"maxIndexBuildMemoryUsageMegabytes": "setParameter.maxIndexBuildMemoryUsageMegabytes",
	"notablescan":                       "setParameter.notablescan",
	"cacheSizeGB":                       "storage.wiredTiger.engineConfig.cacheSizeGB",
}

func Canonical(path string) string {
	if c, ok := aliases[path]; ok {
		return c
	}
	return path
}

// ClassOf is the table tier of classification. It is the only tier available to the validating
// webhook, which cannot reach a running server.
func ClassOf(path string) reconfigure.Class {
	path = Canonical(path)
	if _, ok := locked[path]; ok {
		return reconfigure.ClassLocked
	}
	if dynamic[path] {
		return reconfigure.ClassDynamic
	}
	return reconfigure.ClassStatic
}

// TableClassifier classifies from the static table alone.
type TableClassifier struct{}

func (TableClassifier) Classify(_ context.Context, changes []reconfigure.Change) (map[string]reconfigure.Class, error) {
	out := make(map[string]reconfigure.Class, len(changes))
	for _, c := range changes {
		if c.Path == "" { // an opaque file; any change to it needs a restart
			out[c.Key()] = reconfigure.ClassStatic
			continue
		}
		out[c.Key()] = ClassOf(c.Path)
	}
	return out, nil
}
