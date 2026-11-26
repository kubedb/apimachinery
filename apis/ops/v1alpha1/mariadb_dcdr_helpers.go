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

// MariaDBDCDRSupportedOpsTypes are the ops request types with a DC-DR implementation, in the order
// they are listed to users. The webhook and the ops-manager both refuse every other type.
var MariaDBDCDRSupportedOpsTypes = []MariaDBOpsRequestType{
	MariaDBOpsRequestTypeHorizontalScaling,
	MariaDBOpsRequestTypeRestart,
	MariaDBOpsRequestTypeVerticalScaling,
	MariaDBOpsRequestTypeReconfigure,
	MariaDBOpsRequestTypeRotateAuth,
	MariaDBOpsRequestTypeReconfigureTLS,
	MariaDBOpsRequestTypeUpdateVersion,
	MariaDBOpsRequestTypeVolumeExpansion,
}

// mariadbDCDRNeverSupported are the types a DC-DR MariaDB refuses by design, with the reason.
var mariadbDCDRNeverSupported = map[MariaDBOpsRequestType]string{
	MariaDBOpsRequestTypeStorageMigration: "it moves the data through Jobs and PersistentVolumes of the hub cluster, " +
		"while the data of a DC-DR MariaDB lives in the spoke clusters",
	MariaDBOpsRequestTypeArchiverRestore: "it wipes and restores the data volumes in the hub cluster, " +
		"while the data of a DC-DR MariaDB lives in the spoke clusters; restore into a new DC-DR MariaDB instead",
}

// MariaDBDCDROpsRefusal returns why a DC-DR MariaDB cannot run this ops request, or "" when it can
// (or when the database is not DC-DR). The rules depend on the request and the database spec only.
func MariaDBDCDROpsRefusal(db *dbapi.MariaDB, req *MariaDBOpsRequest) string {
	if db == nil || req == nil || !db.IsDCDRDistributed() {
		return ""
	}
	t := req.Spec.Type
	if why, ok := mariadbDCDRNeverSupported[t]; ok {
		return fmt.Sprintf("%s is not supported for a DC-DR MariaDB: %s", t, why)
	}
	switch t {
	case MariaDBOpsRequestTypeHorizontalScaling:
		if hs := req.Spec.HorizontalScaling; hs != nil && hs.MaxScale {
			return "HorizontalScaling of MaxScale is not supported for a DC-DR MariaDB: it runs Galera, without MaxScale"
		}
	case MariaDBOpsRequestTypeVerticalScaling:
		if vs := req.Spec.VerticalScaling; vs != nil {
			if vs.Mode == VerticalScalingModeInPlace {
				return "VerticalScaling with mode InPlace is not supported for a DC-DR MariaDB: the pods run in the spoke " +
					"clusters and can only be replaced, not resized, from the hub; use mode Restart"
			}
			if vs.MaxScale != nil {
				return "VerticalScaling of MaxScale is not supported for a DC-DR MariaDB: it runs Galera, without MaxScale"
			}
		}
	case MariaDBOpsRequestTypeVolumeExpansion:
		if ve := req.Spec.VolumeExpansion; ve != nil {
			if ve.Mode == VolumeExpansionModeOffline {
				return "VolumeExpansion with mode Offline is not supported for a DC-DR MariaDB: it recreates pods and volumes " +
					"in the hub cluster; use mode Online"
			}
			if ve.MaxScale != nil {
				return "VolumeExpansion of MaxScale is not supported for a DC-DR MariaDB: it runs Galera, without MaxScale"
			}
		}
	case MariaDBOpsRequestTypeReconfigureTLS:
		if why := MariaDBDCDRTLSRefusal(db, req.Spec.TLS); why != "" {
			return "ReconfigureTLS of a DC-DR MariaDB: " + why
		}
	}
	for _, s := range MariaDBDCDRSupportedOpsTypes {
		if s == t {
			return ""
		}
	}
	names := make([]string, 0, len(MariaDBDCDRSupportedOpsTypes))
	for _, s := range MariaDBDCDRSupportedOpsTypes {
		names = append(names, string(s))
	}
	return fmt.Sprintf("%s is not supported for a DC-DR MariaDB yet (supported: %s): its pods run in the spoke clusters",
		t, strings.Join(names, ", "))
}

// MariaDBDCDRTLSRefusal returns why a DC-DR MariaDB cannot take these TLS settings in a rolling
// restart, or "". Galera encrypts its replication traffic with the server certificate
// (wsrep_provider_options socket.ssl_*) and every node verifies its peers against the CA, so:
//   - turning TLS on or off would need every node of a data center's Galera cluster restarted
//     together (a full cluster bootstrap): a TLS and a non TLS node cannot form one cluster;
//   - a new issuer may sign with a new CA, which the not yet restarted nodes reject, so a restarted
//     node could not rejoin its cluster.
//
// Rotating the certificates, changing them (e.g. their SANs) and changing requireSSL roll safely.
func MariaDBDCDRTLSRefusal(db *dbapi.MariaDB, tc *MariaDBTLSSpec) string {
	if db == nil || tc == nil {
		return ""
	}
	switch {
	case db.Spec.TLS == nil:
		return "turning TLS on is not supported: Galera nodes with and without TLS cannot form one cluster, so every node " +
			"of each data center would have to restart together (a full cluster bootstrap); provision a new DC-DR MariaDB with TLS instead"
	case tc.Remove:
		return "removing TLS is not supported: Galera nodes with and without TLS cannot form one cluster, so every node " +
			"of each data center would have to restart together (a full cluster bootstrap)"
	case tc.IssuerRef != nil && db.Spec.TLS.IssuerRef != nil && !sameIssuer(tc.IssuerRef.APIGroup, tc.IssuerRef.Kind, tc.IssuerRef.Name,
		db.Spec.TLS.IssuerRef.APIGroup, db.Spec.TLS.IssuerRef.Kind, db.Spec.TLS.IssuerRef.Name):
		return fmt.Sprintf("changing the issuer (%s -> %s) is not supported: a new CA would make the restarted Galera nodes "+
			"and the not yet restarted ones reject each other; rotate the certificates of the current issuer instead",
			db.Spec.TLS.IssuerRef.Name, tc.IssuerRef.Name)
	}
	return ""
}

func sameIssuer(g1 *string, k1, n1 string, g2 *string, k2, n2 string) bool {
	deref := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	return deref(g1) == deref(g2) && k1 == k2 && n1 == n2
}
