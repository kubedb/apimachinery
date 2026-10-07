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
	"reflect"
	"testing"
)

func TestDB2HADRDatabases(t *testing.T) {
	cases := []struct {
		name string
		hadr *DB2HADRSpec
		want []DB2HADRDatabase
	}{
		{
			"standalone defaults to the image database", nil,
			[]DB2HADRDatabase{{"TESTDB", 55000}},
		},
		{
			"deprecated databaseName is a one-entry list", &DB2HADRSpec{DatabaseName: "abc"},
			[]DB2HADRDatabase{{"ABC", 55000}},
		},
		{
			"databases wins over databaseName", &DB2HADRSpec{DatabaseName: "abc", Databases: []DB2HADRDatabase{{Name: "xyz"}}},
			[]DB2HADRDatabase{{"XYZ", 55000}},
		},
		{
			"explicit ports are kept, the rest take the lowest free",
			&DB2HADRSpec{Databases: []DB2HADRDatabase{{Name: "a"}, {Name: "b", Port: 55000}, {Name: "c"}}},
			[]DB2HADRDatabase{{"A", 55001}, {"B", 55000}, {"C", 55002}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := &DB2{Spec: DB2Spec{HADR: c.hadr}}
			if got := db.HADRDatabases(); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestDB2HADRDefaultsAreStable(t *testing.T) {
	db := &DB2{Spec: DB2Spec{HADR: &DB2HADRSpec{Databases: []DB2HADRDatabase{{Name: "a"}, {Name: "b"}, {Name: "c"}}}}}
	db.setHADRDefaults()
	// Removing the first database must not move the others.
	db.Spec.HADR.Databases = db.Spec.HADR.Databases[1:]
	db.setHADRDefaults()
	want := []DB2HADRDatabase{{"B", 55001}, {"C", 55002}}
	if !reflect.DeepEqual(db.Spec.HADR.Databases, want) {
		t.Fatalf("got %+v, want %+v", db.Spec.HADR.Databases, want)
	}
	// A database added later takes the freed port.
	db.Spec.HADR.Databases = append(db.Spec.HADR.Databases, DB2HADRDatabase{Name: "d"})
	db.setHADRDefaults()
	if got := db.Spec.HADR.Databases[2]; got != (DB2HADRDatabase{"D", 55000}) {
		t.Fatalf("new database got %+v, want D:55000", got)
	}
}

func TestDB2HADRDeepCopy(t *testing.T) {
	a := &DB2{Spec: DB2Spec{HADR: &DB2HADRSpec{Databases: []DB2HADRDatabase{{Name: "A", Port: 55000}}}}}
	b := a.DeepCopy()
	b.Spec.HADR.Databases[0].Port = 1
	if a.Spec.HADR.Databases[0].Port != 55000 {
		t.Fatal("DeepCopy shares the databases slice")
	}
}
