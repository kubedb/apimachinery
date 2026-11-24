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
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	olddbapi "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	core "k8s.io/api/core/v1"
	ofstv2 "kmodules.xyz/offshoot-api/api/v2"
)

func dcdrDataCenter(name, etcd, secret string) olddbapi.MilvusDataCenterSpec {
	return olddbapi.MilvusDataCenterSpec{
		Name:          name,
		MetaStorage:   &olddbapi.MetaStorageSpec{ExternallyManaged: true, Endpoints: []string{etcd}},
		ObjectStorage: &olddbapi.ObjectStorageSpec{ConfigSecret: &core.LocalObjectReference{Name: secret}},
	}
}

func dcdrMilvus() *olddbapi.Milvus {
	db := &olddbapi.Milvus{}
	db.Name = "mv"
	db.Spec.Distributed = true
	db.Spec.PodTemplate = &ofstv2.PodTemplateSpec{}
	db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "milvus-pp"}
	db.Spec.DisasterRecovery = &olddbapi.MilvusDCDRSpec{DataCenters: []olddbapi.MilvusDataCenterSpec{
		dcdrDataCenter("spoke-1", "http://etcd-a:2379", "s3-a"),
		dcdrDataCenter("spoke-2", "http://etcd-b:2379", "s3-b"),
	}}
	return db
}

func TestMilvusValidateDCDR(t *testing.T) {
	v := func(s string) *catalog.MilvusVersion {
		return &catalog.MilvusVersion{Spec: catalog.MilvusVersionSpec{Version: s}}
	}
	tests := []struct {
		name    string
		mutate  func(db *olddbapi.Milvus)
		version string
		wantErr bool
	}{
		{"valid", func(db *olddbapi.Milvus) {}, "2.6.26", false},
		{"old version", func(db *olddbapi.Milvus) {}, "2.6.15", true},
		{"no placement policy", func(db *olddbapi.Milvus) { db.Spec.PodTemplate.Spec.PodPlacementPolicy = nil }, "2.6.26", true},
		{"no disasterRecovery", func(db *olddbapi.Milvus) { db.Spec.DisasterRecovery = nil }, "2.6.26", true},
		{"one data center", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters = db.Spec.DisasterRecovery.DataCenters[:1]
		}, "2.6.26", true},
		{"duplicate data center", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters[1].Name = "spoke-1"
		}, "2.6.26", true},
		{"shared etcd", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters[1].MetaStorage.Endpoints = []string{"http://etcd-a:2379"}
		}, "2.6.26", true},
		{"internal etcd", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters[0].MetaStorage.ExternallyManaged = false
		}, "2.6.26", true},
		{"no object storage", func(db *olddbapi.Milvus) { db.Spec.DisasterRecovery.DataCenters[0].ObjectStorage = nil }, "2.6.26", true},
		{"archiver", func(db *olddbapi.Milvus) { db.Spec.Archiver = &olddbapi.Archiver{} }, "2.6.26", true},
		{"security disabled", func(db *olddbapi.Milvus) { db.Spec.DisableSecurity = true }, "2.6.26", true},
		{"disasterRecovery without distributed", func(db *olddbapi.Milvus) { db.Spec.Distributed = false }, "2.6.26", true},
		{"top-level object storage", func(db *olddbapi.Milvus) {
			db.Spec.ObjectStorage = &olddbapi.ObjectStorageSpec{ConfigSecret: &core.LocalObjectReference{Name: "s3"}}
		}, "2.6.26", true},
		{"not DC-DR without object storage", func(db *olddbapi.Milvus) {
			db.Spec.Distributed = false
			db.Spec.DisasterRecovery = nil
		}, "2.6.26", true},
		{"not DC-DR with object storage", func(db *olddbapi.Milvus) {
			db.Spec.Distributed = false
			db.Spec.DisasterRecovery = nil
			db.Spec.ObjectStorage = &olddbapi.ObjectStorageSpec{ConfigSecret: &core.LocalObjectReference{Name: "s3"}}
		}, "2.6.26", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := dcdrMilvus()
			tt.mutate(db)
			errs := milvusValidateDCDR(db, v(tt.version))
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got %v", tt.wantErr, errs)
			}
		})
	}
}

func TestMilvusValidateDCDRImmutability(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(db *olddbapi.Milvus)
		wantErr bool
	}{
		{"unchanged", func(db *olddbapi.Milvus) {}, false},
		{"tunable changed", func(db *olddbapi.Milvus) {
			n := int64(60)
			db.Spec.DisasterRecovery.MaxFailoverLagSeconds = &n
		}, false},
		{"distributed turned off", func(db *olddbapi.Milvus) { db.Spec.Distributed = false }, true},
		{"data center added", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters = append(db.Spec.DisasterRecovery.DataCenters, dcdrDataCenter("spoke-3", "http://etcd-c:2379", "s3-c"))
		}, true},
		{"object storage changed", func(db *olddbapi.Milvus) {
			db.Spec.DisasterRecovery.DataCenters[0].ObjectStorage.ConfigSecret.Name = "other"
		}, true},
		{"placement changed", func(db *olddbapi.Milvus) {
			db.Spec.PodTemplate.Spec.PodPlacementPolicy = &core.LocalObjectReference{Name: "other"}
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old, db := dcdrMilvus(), dcdrMilvus()
			tt.mutate(db)
			errs := milvusValidateDCDRImmutability(old, db)
			if (len(errs) > 0) != tt.wantErr {
				t.Fatalf("wantErr=%v, got %v", tt.wantErr, errs)
			}
		})
	}
}

func TestMilvusDCDRNames(t *testing.T) {
	db := dcdrMilvus()
	db.Namespace = "demo"
	if got := db.DCDRClusterID("Spoke_1"); got != "mv-spoke-1" {
		t.Errorf("DCDRClusterID = %q", got)
	}
	if got := db.DCDRPetSetName("spoke-1", olddbapi.MilvusNodeRoleProxy, ""); got != "mv-spoke-1-proxy" {
		t.Errorf("DCDRPetSetName = %q", got)
	}
	if got := db.DCDRPetSetName("spoke-1", "", ""); got != "mv-spoke-1" {
		t.Errorf("standalone DCDRPetSetName = %q", got)
	}
	if got := db.DCDRServiceAddress("spoke-2"); got != "mv-spoke-2.demo.svc.slice.local:19530" {
		t.Errorf("DCDRServiceAddress = %q", got)
	}
	if got := db.DCDRPchannels("spoke-1"); len(got) != olddbapi.MilvusDCDRDMLChannelNum || got[3] != "mv-spoke-1-rootcoord-dml_3" {
		t.Errorf("DCDRPchannels = %v", got)
	}
	dcView, err := db.ForDataCenter("spoke-2")
	if err != nil || dcView.Spec.ObjectStorage.ConfigSecret.Name != "s3-b" {
		t.Errorf("ForDataCenter = %v, %v", dcView, err)
	}
	if _, err := db.ForDataCenter("nope"); err == nil {
		t.Errorf("ForDataCenter of an unknown data center must fail")
	}
}
