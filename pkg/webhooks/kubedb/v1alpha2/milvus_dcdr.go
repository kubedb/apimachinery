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

package v1alpha2

import (
	"fmt"
	"reflect"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"k8s.io/apimachinery/pkg/util/validation/field"
	utilversion "k8s.io/apimachinery/pkg/util/version"
)

// milvusDCDRMinVersion is the first Milvus release whose CDC supports switchover and force promote.
var milvusDCDRMinVersion = utilversion.MustParseGeneric("2.6.16")

func milvusValidateDCDR(db *olddbapi.Milvus, milvusVersion *catalog.MilvusVersion) field.ErrorList {
	spec := field.NewPath("spec")
	drPath := spec.Child("disasterRecovery")
	if !db.IsDCDR() {
		var errs field.ErrorList
		if db.Spec.DisasterRecovery != nil {
			errs = append(errs, field.Forbidden(drPath, "spec.disasterRecovery is only used with spec.distributed"))
		}
		if db.Spec.ObjectStorage == nil {
			errs = append(errs, field.Required(spec.Child("objectStorage"), "spec.objectStorage is required"))
		}
		return errs
	}

	var errs field.ErrorList
	if db.Spec.ObjectStorage != nil {
		errs = append(errs, field.Forbidden(spec.Child("objectStorage"),
			"a DC-DR Milvus sets the object storage per data center in spec.disasterRecovery.dataCenters"))
	}
	if milvusVersion != nil {
		if v, err := utilversion.ParseGeneric(milvusVersion.Spec.Version); err != nil {
			errs = append(errs, field.Invalid(spec.Child("version"), db.Spec.Version, fmt.Sprintf("cannot parse Milvus version %q: %v", milvusVersion.Spec.Version, err)))
		} else if v.LessThan(milvusDCDRMinVersion) {
			errs = append(errs, field.Invalid(spec.Child("version"), db.Spec.Version, fmt.Sprintf("DC-DR needs Milvus %s or later", milvusDCDRMinVersion)))
		}
	}
	if db.Spec.PodTemplate == nil || db.Spec.PodTemplate.Spec.PodPlacementPolicy == nil || db.Spec.PodTemplate.Spec.PodPlacementPolicy.Name == "" {
		errs = append(errs, field.Required(spec.Child("podTemplate", "spec", "podPlacementPolicy"),
			"a DC-DR Milvus names the PlacementPolicy that lists its data centers"))
	}
	if db.GetAnnotations()[kubedb.MilvusWALAnnotation] == kubedb.MilvusWALRocksMQ {
		errs = append(errs, field.Forbidden(field.NewPath("metadata", "annotations").Key(kubedb.MilvusWALAnnotation),
			"DC-DR needs the Woodpecker write-ahead log"))
	}
	// Not supported for a DC-DR Milvus yet: each needs per data center handling.
	unsupported := []struct {
		set  bool
		path *field.Path
	}{
		{db.Spec.Archiver != nil, spec.Child("archiver")},
		{db.Spec.Init != nil && db.Spec.Init.Archiver != nil, spec.Child("init", "archiver")},
		{db.Spec.Monitor != nil, spec.Child("monitor")},
		{db.Spec.TLS != nil, spec.Child("tls")},
		{db.Spec.Halted, spec.Child("halted")},
		{olddbapi.IsVirtualAuthSecretReferred(db.Spec.AuthSecret), spec.Child("authSecret")},
	}
	for _, u := range unsupported {
		if u.set {
			errs = append(errs, field.Forbidden(u.path, "not supported for a DC-DR Milvus (spec.distributed)"))
		}
	}
	if db.Spec.DisableSecurity {
		errs = append(errs, field.Forbidden(spec.Child("disableSecurity"),
			"a DC-DR Milvus needs authentication: CDC connects to the peer data centers with a dedicated user"))
	}

	dr := db.Spec.DisasterRecovery
	if dr == nil {
		return append(errs, field.Required(drPath, "spec.disasterRecovery is required with spec.distributed"))
	}
	dcsPath := drPath.Child("dataCenters")
	if len(dr.DataCenters) < 2 {
		errs = append(errs, field.Invalid(dcsPath, len(dr.DataCenters), "a DC-DR Milvus needs at least two data centers"))
	}
	seen := map[string]bool{}
	etcds := map[string]string{}
	for i, dc := range dr.DataCenters {
		p := dcsPath.Index(i)
		if dc.Name == "" {
			errs = append(errs, field.Required(p.Child("name"), "name is the OCM managed cluster of the data center"))
		} else if seen[dc.Name] {
			errs = append(errs, field.Duplicate(p.Child("name"), dc.Name))
		}
		seen[dc.Name] = true

		ms := dc.MetaStorage
		switch {
		case ms == nil:
			errs = append(errs, field.Required(p.Child("metaStorage"), "each data center needs its own meta etcd"))
		case !ms.ExternallyManaged || len(ms.Endpoints) == 0:
			errs = append(errs, field.Invalid(p.Child("metaStorage"), dc.Name,
				"the meta etcd of a data center must be externally managed, with endpoints reachable from that data center"))
		default:
			if olddbapi.IsVirtualAuthSecretReferred(ms.AuthSecret) {
				errs = append(errs, field.Forbidden(p.Child("metaStorage", "authSecret"), "virtual secrets are not supported for an external meta etcd"))
			}
			key := fmt.Sprintf("%v", ms.Endpoints)
			if other, ok := etcds[key]; ok {
				errs = append(errs, field.Invalid(p.Child("metaStorage", "endpoints"), ms.Endpoints,
					fmt.Sprintf("data center %s uses the same meta etcd; each data center needs its own", other)))
			}
			etcds[key] = dc.Name
		}
		if dc.ObjectStorage == nil || dc.ObjectStorage.ConfigSecret == nil || dc.ObjectStorage.ConfigSecret.Name == "" {
			errs = append(errs, field.Required(p.Child("objectStorage", "configSecret"), "each data center needs its own object storage"))
		} else if dc.ObjectStorage.TLS != nil {
			errs = append(errs, milvusValidateObjectStorageTLS(dc.ObjectStorage.TLS, p.Child("objectStorage", "tls"))...)
		}
	}
	return errs
}

// milvusValidateDCDRImmutability keeps the data centers of a DC-DR Milvus fixed: each data
// center's cluster id prefixes its channel names, and its storage holds its copy of the data.
func milvusValidateDCDRImmutability(oldDB, db *olddbapi.Milvus) field.ErrorList {
	spec := field.NewPath("spec")
	var errs field.ErrorList
	if oldDB.Spec.Distributed != db.Spec.Distributed {
		errs = append(errs, field.Forbidden(spec.Child("distributed"), "spec.distributed cannot be changed"))
	}
	if !oldDB.IsDCDR() || !db.IsDCDR() || oldDB.Spec.DisasterRecovery == nil || db.Spec.DisasterRecovery == nil {
		return errs
	}
	dcsPath := spec.Child("disasterRecovery", "dataCenters")
	if !reflect.DeepEqual(oldDB.DCDRDataCenterNames(), db.DCDRDataCenterNames()) {
		return append(errs, field.Forbidden(dcsPath, "the data centers of a DC-DR Milvus cannot be added, removed or reordered"))
	}
	for i := range db.Spec.DisasterRecovery.DataCenters {
		o, n := oldDB.Spec.DisasterRecovery.DataCenters[i], db.Spec.DisasterRecovery.DataCenters[i]
		if !reflect.DeepEqual(o.MetaStorage, n.MetaStorage) {
			errs = append(errs, field.Forbidden(dcsPath.Index(i).Child("metaStorage"), "the meta etcd of a data center cannot be changed"))
		}
		if !reflect.DeepEqual(o.ObjectStorage, n.ObjectStorage) {
			errs = append(errs, field.Forbidden(dcsPath.Index(i).Child("objectStorage"), "the object storage of a data center cannot be changed"))
		}
	}
	if oldDB.Spec.PodTemplate != nil && db.Spec.PodTemplate != nil &&
		!reflect.DeepEqual(oldDB.Spec.PodTemplate.Spec.PodPlacementPolicy, db.Spec.PodTemplate.Spec.PodPlacementPolicy) {
		errs = append(errs, field.Forbidden(spec.Child("podTemplate", "spec", "podPlacementPolicy"), "the PlacementPolicy of a DC-DR Milvus cannot be changed"))
	}
	return errs
}
