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

	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"
)

func TestPostgresLogForwarderVersionRoundTrip(t *testing.T) {
	cases := []*dbapi.LogForwarderSpec{
		{CollectionMode: dbapi.LogCollectionModeNodeAgent, Sources: []dbapi.LogSourceSpec{{Name: "general"}}},
		{Exporter: &dbapi.LogExporterSpec{Name: "primary", Type: "otlphttp", OTLPHTTP: &dbapi.LogOTLPHTTPExporter{Endpoint: "https://example.com"}}, Sources: []dbapi.LogSourceSpec{{Name: "general", FileLog: &dbapi.LogFileLogReceiver{StartAt: "beginning", MaxLogSize: "1MiB"}}}},
		{Destination: dbapi.LogDestinationSpec{Name: "primary", Profile: "splunk"}, Sources: []dbapi.LogSourceSpec{{Name: "general", InitialPosition: dbapi.LogInitialPositionEnd}}},
	}
	for _, lf := range cases {
		input := dbapi.PostgresSpec{LogForwarder: lf}
		var legacy PostgresSpec
		if err := Convert_v1_PostgresSpec_To_v1alpha2_PostgresSpec(&input, &legacy, nil); err != nil {
			t.Fatal(err)
		}
		var output dbapi.PostgresSpec
		if err := Convert_v1alpha2_PostgresSpec_To_v1_PostgresSpec(&legacy, &output, nil); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(input.LogForwarder, output.LogForwarder) {
			t.Fatalf("logForwarder lost in conversion: %#v", output.LogForwarder)
		}
	}
}
