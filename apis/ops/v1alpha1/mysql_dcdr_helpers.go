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
	"fmt"
	"strings"

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
)

// MySQLDCDRSupportedOpsTypes are the ops request types with a DC-DR implementation, in the order
// they are listed to users. The webhook and the ops-manager both refuse every other type.
var MySQLDCDRSupportedOpsTypes = []MySQLOpsRequestType{
	MySQLOpsRequestTypeHorizontalScaling,
	MySQLOpsRequestTypeRestart,
	MySQLOpsRequestTypeVerticalScaling,
	MySQLOpsRequestTypeReconfigure,
	MySQLOpsRequestTypeRotateAuth,
	MySQLOpsRequestTypeReconfigureTLS,
	MySQLOpsRequestTypeUpdateVersion,
	MySQLOpsRequestTypeVolumeExpansion,
}

// mysqlDCDRNeverSupported are the types a DC-DR MySQL refuses by design, with the reason.
var mysqlDCDRNeverSupported = map[MySQLOpsRequestType]string{
	MySQLOpsRequestTypeStorageMigration: "it moves the data through Jobs and PersistentVolumes of the hub cluster, " +
		"while the data of a DC-DR MySQL lives in the spoke clusters",
	MySQLOpsRequestTypeArchiverRestore: "it wipes and restores the data volumes in the hub cluster, " +
		"while the data of a DC-DR MySQL lives in the spoke clusters; restore into a new DC-DR MySQL instead",
	MySQLOpsRequestTypeReplicationModeTransformation: "a DC-DR MySQL must stay in single-primary Group Replication",
}

// MySQLDCDROpsRefusal returns why a DC-DR MySQL cannot run this ops request, or "" when it can (or
// when the database is not DC-DR).
func MySQLDCDROpsRefusal(db *dbapi.MySQL, req *MySQLOpsRequest) string {
	if db == nil || req == nil || !db.IsDCDRDistributed() {
		return ""
	}
	t := req.Spec.Type
	if why, ok := mysqlDCDRNeverSupported[t]; ok {
		return fmt.Sprintf("%s is not supported for a DC-DR MySQL: %s", t, why)
	}
	switch t {
	case MySQLOpsRequestTypeVerticalScaling:
		if vs := req.Spec.VerticalScaling; vs != nil && vs.Mode == VerticalScalingModeInPlace {
			return "VerticalScaling with mode InPlace is not supported for a DC-DR MySQL: the pods run in the spoke " +
				"clusters and can only be replaced, not resized, from the hub; use mode Restart"
		}
	case MySQLOpsRequestTypeVolumeExpansion:
		if ve := req.Spec.VolumeExpansion; ve != nil && ve.Mode == VolumeExpansionModeOffline {
			return "VolumeExpansion with mode Offline is not supported for a DC-DR MySQL: it recreates pods and volumes " +
				"in the hub cluster; use mode Online"
		}
	}
	for _, s := range MySQLDCDRSupportedOpsTypes {
		if s == t {
			return ""
		}
	}
	names := make([]string, 0, len(MySQLDCDRSupportedOpsTypes))
	for _, s := range MySQLDCDRSupportedOpsTypes {
		names = append(names, string(s))
	}
	return fmt.Sprintf("%s is not supported for a DC-DR MySQL yet (supported: %s): its pods run in the spoke clusters",
		t, strings.Join(names, ", "))
}
