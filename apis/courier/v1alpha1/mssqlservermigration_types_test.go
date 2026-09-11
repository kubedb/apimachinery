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
	"encoding/json"
	"reflect"
	"testing"
)

func TestMSSQLServerDatabaseSelectionJSONRoundTrip(t *testing.T) {
	want := MSSQLServerSchema{
		Enabled: true,
		Schema:  []string{"dbo", "sales"},
		Databases: []MSSQLServerDatabaseSelection{
			{
				Name:         "DatabaseA",
				Table:        []string{"dbo.Payload", "dbo.Users"},
				ExcludeTable: []string{"dbo.Audit"},
			},
			{
				Name:  "DatabaseB",
				Table: []string{"dbo.Payload"},
			},
		},
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var got MSSQLServerSchema
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestMSSQLServerDatabaseSelectionDeepCopy(t *testing.T) {
	original := &MSSQLServerSchema{
		Enabled: true,
		Databases: []MSSQLServerDatabaseSelection{
			{
				Name:         "DatabaseA",
				Table:        []string{"dbo.Payload"},
				ExcludeTable: []string{"dbo.Audit"},
			},
		},
	}

	copy := original.DeepCopy()
	copy.Databases[0].Name = "Changed"
	copy.Databases[0].Table[0] = "dbo.Other"
	copy.Databases[0].ExcludeTable[0] = "dbo.OtherAudit"

	if original.Databases[0].Name != "DatabaseA" {
		t.Fatalf("deep copy mutated original database name: %q", original.Databases[0].Name)
	}
	if original.Databases[0].Table[0] != "dbo.Payload" {
		t.Fatalf("deep copy mutated original table: %q", original.Databases[0].Table[0])
	}
	if original.Databases[0].ExcludeTable[0] != "dbo.Audit" {
		t.Fatalf("deep copy mutated original excluded table: %q", original.Databases[0].ExcludeTable[0])
	}
}
