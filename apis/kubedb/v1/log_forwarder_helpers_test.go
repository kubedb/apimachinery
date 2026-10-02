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

package v1

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func validLogForwarderForTest() *LogForwarderSpec {
	return &LogForwarderSpec{
		Exporter: &LogExporterSpec{Type: "splunk_hec", SplunkHEC: &LogSplunkHECExporter{
			Endpoint:       "https://splunk.example.com:8088/services/collector",
			TokenSecretRef: core.SecretKeySelector{LocalObjectReference: core.LocalObjectReference{Name: "ingest"}, Key: "token"},
		}},
		StateStorage: core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce},
			Resources:   core.VolumeResourceRequirements{Requests: core.ResourceList{core.ResourceStorage: resource.MustParse("1Gi")}},
		},
		Sources: []LogSourceSpec{{Name: "general", FileLog: &LogFileLogReceiver{}}},
	}
}

func TestCanonicalLogForwarderDefaults(t *testing.T) {
	lf := validLogForwarderForTest()
	lf.SetCollectionDefaults()
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	if lf.CollectionMode != LogCollectionModeSidecar || lf.Exporter.Name != "primary" || lf.Sources[0].FileLog.StartAt != "end" {
		t.Fatalf("incorrect defaults: %#v", lf)
	}
	if lf.Delivery != nil || !reflect.DeepEqual(lf.Destination, LogDestinationSpec{}) || lf.Sources[0].InitialPosition != "" {
		t.Fatal("canonical defaulting populated deprecated fields")
	}
	q := lf.Exporter.SplunkHEC.SendingQueue
	if !*q.Enabled || *q.QueueSize != 1000 || *q.NumConsumers != 2 || !*q.BlockOnOverflow {
		t.Fatalf("incorrect queue defaults: %#v", q)
	}
	if lf.Exporter.SplunkHEC.TLS.Mode != LogForwarderTLSModeVerify {
		t.Fatal("TLS must default to verification")
	}
	copy := lf.DeepCopy()
	copy.Exporter.SplunkHEC.TokenSecretRef.Name = "different"
	*copy.Exporter.SplunkHEC.SendingQueue.QueueSize = 20
	copy.Sources[0].FileLog.StartAt = "beginning"
	if lf.Exporter.SplunkHEC.TokenSecretRef.Name != "ingest" || *q.QueueSize != 1000 || lf.Sources[0].FileLog.StartAt != "end" {
		t.Fatal("deep copy aliases nested settings")
	}
	before, err := json.Marshal(lf)
	if err != nil {
		t.Fatal(err)
	}
	lf.SetCollectionDefaults()
	after, err := json.Marshal(lf)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("defaulting is not idempotent")
	}
}

func TestNodeAgentPlatformOwnership(t *testing.T) {
	lf := &LogForwarderSpec{CollectionMode: LogCollectionModeNodeAgent, Sources: []LogSourceSpec{{Name: "general"}}}
	lf.SetCollectionDefaults()
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	if lf.Exporter != nil || lf.Delivery != nil || lf.Processors != nil || lf.Sources[0].InitialPosition != "" || !reflect.DeepEqual(lf.Destination, LogDestinationSpec{}) {
		t.Fatal("node defaulting injected sidecar configuration")
	}
	for _, mutate := range []func(*LogForwarderSpec){
		func(l *LogForwarderSpec) { l.Exporter = validLogForwarderForTest().Exporter },
		func(l *LogForwarderSpec) { l.Delivery = &LogDeliverySpec{} },
		func(l *LogForwarderSpec) { l.Processors = &LogProcessors{} },
		func(l *LogForwarderSpec) { l.Sources[0].FileLog = &LogFileLogReceiver{} },
		func(l *LogForwarderSpec) { l.StateStorage = validLogForwarderForTest().StateStorage },
		func(l *LogForwarderSpec) { l.Destination.Profile = "splunk" },
		func(l *LogForwarderSpec) {
			l.Resources.Requests = core.ResourceList{core.ResourceCPU: resource.MustParse("10m")}
		},
	} {
		copy := lf.DeepCopy()
		mutate(copy)
		if err := copy.ValidateCollection(); err == nil {
			t.Fatalf("NodeAgent accepted sidecar config: %#v", copy)
		}
	}
}

func TestLogForwarderValidation(t *testing.T) {
	zero := int32(0)
	optional := true
	cases := []struct {
		name   string
		mutate func(*LogForwarderSpec)
	}{
		{"mixed destination", func(l *LogForwarderSpec) { l.Destination.Profile = "splunk" }},
		{"mixed delivery", func(l *LogForwarderSpec) { l.Delivery = &LogDeliverySpec{} }},
		{"mismatched block", func(l *LogForwarderSpec) { l.Exporter.Type = "otlphttp" }},
		{"multiple blocks", func(l *LogForwarderSpec) { l.Exporter.OTLPHTTP = &LogOTLPHTTPExporter{Endpoint: "https://example.com"} }},
		{"zero queue", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.SendingQueue = &LogSendingQueue{QueueSize: &zero} }},
		{"file aliases", func(l *LogForwarderSpec) {
			l.Sources[0].InitialPosition = LogInitialPositionEnd
			l.Sources[0].FileLog.StartAt = "end"
		}},
		{"bad startAt", func(l *LogForwarderSpec) { l.Sources[0].FileLog.StartAt = "End" }},
		{"bad size", func(l *LogForwarderSpec) { l.Sources[0].FileLog.MaxLogSize = "0MiB" }},
		{"optional credential", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.TokenSecretRef.Optional = &optional }},
		{"URL credentials", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.Endpoint = "https://user:password@example.com" }},
		{"plaintext without opt in", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.Endpoint = "http://example.com" }},
		{"negative storage", func(l *LogForwarderSpec) {
			l.StateStorage.Resources.Requests[core.ResourceStorage] = resource.MustParse("-1Gi")
		}},
		{"automatic rollout", func(l *LogForwarderSpec) { l.RolloutPolicy = LogForwarderRolloutAutomatic }},
		{"missing cert pair", func(l *LogForwarderSpec) {
			l.Exporter.SplunkHEC.TLS = &LogForwarderTLS{ClientCertSecretRef: &core.SecretKeySelector{}}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lf := validLogForwarderForTest()
			tc.mutate(lf)
			if err := lf.ValidateCollection(); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func TestLegacyAndElasticsearchLogForwarder(t *testing.T) {
	lf := validLogForwarderForTest()
	lf.Exporter = nil
	lf.Destination = LogDestinationSpec{Name: "primary", Profile: "splunk", Endpoint: "https://example.com"}
	lf.Sources[0] = LogSourceSpec{Name: "general", InitialPosition: LogInitialPositionEnd}
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(lf)
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(lf)
	if string(before) != string(after) {
		t.Fatal("validation changed legacy state")
	}
	lf.Exporter = &LogExporterSpec{Type: "elasticsearch", Elasticsearch: &LogElasticsearchExporter{Endpoints: []string{"https://elastic.example.com"}, Retry: &LogElasticsearchRetry{}}}
	lf.Destination = LogDestinationSpec{}
	lf.SetCollectionDefaults()
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(lf)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "retryOnFailure") || strings.Contains(string(data), "maxElapsedTime") {
		t.Fatal("Elasticsearch uses native retry, not exporterhelper fields")
	}
}
