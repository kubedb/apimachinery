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
	"encoding/json"
	"testing"

	v1 "kubedb.dev/apimachinery/apis/kubedb/v1"
)

// TestPostgresReplicationConversionIsFieldByField pins the fix for the webhook-server
// crash loop on dc-c.
//
// PostgresReplication used to be converted with
//
//	out.Replication = (*v1.PostgresReplication)(unsafe.Pointer(in.Replication))
//
// which is only sound while both structs have identical layouts. v1 gained
// BestEffortCrossDCLagBytesForFailover and v1alpha2 did not, so that cast reinterpreted a
// five field v1alpha2 allocation as the six field v1 type. Reading the sixth field read
// past the end of the allocation and produced a non nil pointer to garbage, which survived
// IsNil() and then faulted inside encoding/json with
// "invalid memory address or nil pointer dereference".
//
// The assertion that matters is that the v1 only field comes out nil rather than garbage.
func TestPostgresReplicationConversionIsFieldByField(t *testing.T) {
	keep := int32(512)
	in := &PostgresSpec{
		Replication: &PostgresReplication{
			WALLimitPolicy:         WALKeepSize,
			WalKeepSizeInMegaBytes: &keep,
		},
	}

	var out v1.PostgresSpec
	if err := Convert_v1alpha2_PostgresSpec_To_v1_PostgresSpec(in, &out, nil); err != nil {
		t.Fatalf("conversion failed: %v", err)
	}
	if out.Replication == nil {
		t.Fatal("Replication was dropped by conversion")
	}
	if out.Replication.WalKeepSizeInMegaBytes == nil || *out.Replication.WalKeepSizeInMegaBytes != keep {
		t.Errorf("WalKeepSizeInMegaBytes did not survive: %v", out.Replication.WalKeepSizeInMegaBytes)
	}
	// The out of bounds read. v1alpha2 has nothing to put here, so it must be nil.
	if out.Replication.BestEffortCrossDCLagBytesForFailover != nil {
		t.Fatalf("v1-only field is not nil after converting from v1alpha2: %d (this is the out-of-bounds read)",
			*out.Replication.BestEffortCrossDCLagBytesForFailover)
	}

	// Marshalling is what actually faulted in the webhook conversion self-test.
	if _, err := json.Marshal(&out); err != nil {
		t.Fatalf("marshalling the converted spec failed: %v", err)
	}
}

// TestPostgresReplicationDownConversionDropsV1OnlyField documents the lossy direction:
// v1alpha2 cannot express the budget, so it is dropped rather than silently reinterpreted.
func TestPostgresReplicationDownConversionDropsV1OnlyField(t *testing.T) {
	budget := int64(64 << 20)
	in := &v1.PostgresSpec{
		Replication: &v1.PostgresReplication{
			WALLimitPolicy:                       v1.WALKeepSize,
			BestEffortCrossDCLagBytesForFailover: &budget,
		},
	}

	var out PostgresSpec
	if err := Convert_v1_PostgresSpec_To_v1alpha2_PostgresSpec(in, &out, nil); err != nil {
		t.Fatalf("conversion failed: %v", err)
	}
	if out.Replication == nil {
		t.Fatal("Replication was dropped entirely")
	}
	if string(out.Replication.WALLimitPolicy) != string(v1.WALKeepSize) {
		t.Errorf("WALLimitPolicy did not survive: %q", out.Replication.WALLimitPolicy)
	}
	if _, err := json.Marshal(&out); err != nil {
		t.Fatalf("marshalling the converted spec failed: %v", err)
	}
}
