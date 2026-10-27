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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	core "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func logInt(v int32) *int32 { return &v }
func validLogForwarder() *LogForwarderSpec {
	return &LogForwarderSpec{
		Exporter:     &LogExporterSpec{Type: "otlphttp", OTLPHTTP: &LogOTLPHTTPExporter{Endpoint: "https://example.com"}},
		Resources:    core.ResourceRequirements{Limits: core.ResourceList{core.ResourceMemory: resource.MustParse("256Mi")}},
		StateStorage: core.PersistentVolumeClaimSpec{AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce}, Resources: core.VolumeResourceRequirements{Requests: core.ResourceList{core.ResourceStorage: resource.MustParse("1Gi")}}},
		Sources:      []LogSourceSpec{{Name: "general", FileLog: &LogFileLogReceiver{}}},
	}
}

func TestLogDefaultsAndMemoryPairs(t *testing.T) {
	lf := validLogForwarder()
	lf.SetCollectionDefaults()
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	m := lf.Processors.MemoryLimiter
	if *m.LimitPercentage != 80 || *m.SpikeLimitPercentage != 15 || m.LimitMiB != nil {
		t.Fatal("percentage defaults")
	}
	before := lf.DeepCopy()
	lf.SetCollectionDefaults()
	if !reflect.DeepEqual(before, lf) {
		t.Fatal("non-idempotent defaults")
	}
	absolute := validLogForwarder()
	absolute.Processors = &LogProcessors{MemoryLimiter: &LogMemoryLimiterProcessor{LimitMiB: logInt(180), SpikeLimitMiB: logInt(32)}}
	absolute.SetCollectionDefaults()
	if absolute.Processors.MemoryLimiter.LimitPercentage != nil {
		t.Fatal("mixed defaults")
	}
	if err := absolute.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	cases := []*LogMemoryLimiterProcessor{
		{LimitPercentage: logInt(80)},
		{SpikeLimitPercentage: logInt(15)},
		{LimitMiB: logInt(180)},
		{LimitMiB: logInt(180), SpikeLimitMiB: logInt(32), LimitPercentage: logInt(80), SpikeLimitPercentage: logInt(15)},
		{LimitPercentage: logInt(100), SpikeLimitPercentage: logInt(15)},
		{LimitPercentage: logInt(80), SpikeLimitPercentage: logInt(80)},
		{LimitMiB: logInt(256), SpikeLimitMiB: logInt(32)},
	}
	for i, m := range cases {
		lf := validLogForwarder()
		lf.Processors = &LogProcessors{MemoryLimiter: m}
		lf.SetCollectionDefaults()
		if err := lf.ValidateCollection(); err == nil {
			t.Fatalf("accepted invalid pair %d", i)
		}
	}
}

func TestLogNodeAgentOwnership(t *testing.T) {
	lf := &LogForwarderSpec{CollectionMode: LogCollectionModeNodeAgent, Sources: []LogSourceSpec{{Name: "general"}}}
	lf.SetCollectionDefaults()
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	if lf.Processors != nil || lf.Exporter != nil {
		t.Fatal("sidecar defaults leaked")
	}
	for _, change := range []func(*LogForwarderSpec){
		func(l *LogForwarderSpec) { l.Processors = &LogProcessors{} },
		func(l *LogForwarderSpec) { l.Extensions = &LogExtensions{} },
		func(l *LogForwarderSpec) { l.Sources[0].FileLog = &LogFileLogReceiver{} },
		func(l *LogForwarderSpec) { l.Exporter = validLogForwarder().Exporter },
	} {
		copy := lf.DeepCopy()
		change(copy)
		if copy.ValidateCollection() == nil {
			t.Fatal("accepted node configuration")
		}
	}
}

func TestLogProcessorOrderAndExtensions(t *testing.T) {
	lf := validLogForwarder()
	lf.Processors = &LogProcessors{
		Attributes: &LogAttributesProcessor{Actions: []LogAttributeAction{{Key: "environment", Action: "upsert", Value: &apiextensionsv1.JSON{Raw: []byte(`"test"`)}}}},
		Transform:  &LogTransformProcessor{LogStatements: []LogTransformStatements{{Context: "log", Statements: []string{`set(attributes["normalized"], true)`}}}},
		Filter:     &LogFilterProcessor{Logs: LogFilterLogs{LogRecord: []string{`attributes["drop"] == true`}}},
	}
	cfg, order, err := lf.ProcessorConfiguration(map[string]string{"k8s.namespace.name": "demo", "db.system": "postgresql"})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"memory_limiter", "attributes", "transform", "filter", "resource/kubedb", "batch"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("wrong order %v", order)
	}
	if cfg["memory_limiter"]["limit_percentage"] != int32(80) {
		t.Fatal("native memory mapping")
	}
	lf.Processors.ExtraConfig = "resource/custom:\n  attributes:\n    - key: team\n      value: database\n      action: upsert\n"
	if _, _, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql"}); err == nil {
		t.Fatal("missing explicit order")
	}
	lf.Processors.Order = []string{"filter", "resource/custom", "attributes", "transform"}
	if _, _, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql"}); err != nil {
		t.Fatal(err)
	}
	for _, order := range [][]string{{"filter"}, {"filter", "filter", "resource/custom", "attributes", "transform"}, {"memory_limiter", "filter", "resource/custom", "attributes", "transform"}} {
		lf.Processors.Order = order
		if _, _, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql"}); err == nil {
			t.Fatal("invalid order")
		}
	}
	lf.Extensions = &LogExtensions{
		ExtraConfig: "basicauth/client:\n  client_auth:\n    username: " + "${env:EXT_USER}" + "\n    password: " + "${env:EXT_PASSWORD}" + "\n",
		SecretEnv: map[string]core.SecretKeySelector{
			"EXT_USER":     {LocalObjectReference: core.LocalObjectReference{Name: "credentials"}, Key: "username"},
			"EXT_PASSWORD": {LocalObjectReference: core.LocalObjectReference{Name: "credentials"}, Key: "password"},
		},
	}
	lf.Processors.Order = []string{"filter", "resource/custom", "attributes", "transform"}
	lf.Exporter.OTLPHTTP.Auth = &LogExporterAuth{Type: "Extension", ExtensionRef: "basicauth/client"}
	if err := lf.ValidateCollection(); err != nil {
		t.Fatal(err)
	}
	lf.Exporter.OTLPHTTP.Auth.ExtensionRef = "basicauth/missing"
	if lf.ValidateCollection() == nil {
		t.Fatal("unresolved auth")
	}
	lf.Extensions.SecretEnv["OTEL_RESOURCE_ATTRIBUTES"] = core.SecretKeySelector{}
	if _, err := lf.ExtensionConfiguration(); err == nil {
		t.Fatal("reserved environment name")
	}
}

func TestLogRuntimeMemoryBudget(t *testing.T) {
	lf := validLogForwarder()
	lf.Resources = core.ResourceRequirements{}
	if _, _, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql"}); err == nil {
		t.Fatal("missing memory limit")
	}
	lf = validLogForwarder()
	if _, _, err := lf.ProcessorConfiguration(nil); err == nil {
		t.Fatal("missing trusted identity")
	}
}

// TestLogCollectorFixture emits a complete config for validation by the pinned image when requested.
func TestLogCollectorFixture(t *testing.T) {
	directory := os.Getenv("KUBEDB_COLLECTOR_FIXTURE_DIR")
	if directory == "" {
		return
	}
	lf := validLogForwarder()
	lf.Processors = &LogProcessors{
		Attributes: &LogAttributesProcessor{Actions: []LogAttributeAction{{Key: "environment", Action: "upsert", Value: &apiextensionsv1.JSON{Raw: []byte(`"test"`)}}}},
		Transform:  &LogTransformProcessor{LogStatements: []LogTransformStatements{{Context: "log", Statements: []string{`set(attributes["normalized"], true)`}}}},
		Filter:     &LogFilterProcessor{Logs: LogFilterLogs{LogRecord: []string{`attributes["drop"] == true`}}},
	}
	cfg, order, err := lf.ProcessorConfiguration(map[string]string{"db.system": "postgresql", "k8s.namespace.name": "demo"})
	if err != nil {
		t.Fatal(err)
	}
	fixture := map[string]interface{}{
		"receivers":  map[string]interface{}{"filelog": map[string]interface{}{"include": []string{"/tmp/test.log"}}},
		"processors": cfg, "exporters": map[string]interface{}{"debug": map[string]interface{}{}},
		"service": map[string]interface{}{"pipelines": map[string]interface{}{"logs": map[string]interface{}{"receivers": []string{"filelog"}, "processors": order, "exporters": []string{"debug"}}}},
	}
	data, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "processors.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	writeFixture := func(name string) {
		data, err := json.Marshal(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Validate custom authentication and extension activation using dummy credentials.
	lf.Extensions = &LogExtensions{
		ExtraConfig: "basicauth/client:\n  client_auth:\n    username: " + "${env:EXT_USER}" + "\n    password: " + "${env:EXT_PASSWORD}" + "\n",
		SecretEnv: map[string]core.SecretKeySelector{
			"EXT_USER":     {LocalObjectReference: core.LocalObjectReference{Name: "credentials"}, Key: "username"},
			"EXT_PASSWORD": {LocalObjectReference: core.LocalObjectReference{Name: "credentials"}, Key: "password"},
		},
	}
	extensions, err := lf.ExtensionConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	fixture["extensions"] = extensions
	fixture["exporters"] = map[string]interface{}{"otlphttp": map[string]interface{}{"endpoint": "https://example.com", "auth": map[string]interface{}{"authenticator": "basicauth/client"}}}
	service := fixture["service"].(map[string]interface{})
	service["extensions"] = []string{"basicauth/client"}
	pipeline := service["pipelines"].(map[string]interface{})["logs"].(map[string]interface{})
	pipeline["exporters"] = []string{"otlphttp"}
	writeFixture("extensions.json")
	// Structural checks leave OTTL semantics to the pinned binary.
	cfg["transform"]["log_statements"] = []interface{}{map[string]interface{}{"context": "log", "statements": []string{"NoSuchKubeDBFunction()"}}}
	writeFixture("invalid-ottl.json")
}

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
	if lf.Exporter != nil || lf.Processors != nil || lf.Extensions != nil {
		t.Fatal("node defaulting injected sidecar configuration")
	}
	for _, mutate := range []func(*LogForwarderSpec){
		func(l *LogForwarderSpec) { l.Exporter = validLogForwarderForTest().Exporter },
		func(l *LogForwarderSpec) { l.Processors = &LogProcessors{} },
		func(l *LogForwarderSpec) { l.Sources[0].FileLog = &LogFileLogReceiver{} },
		func(l *LogForwarderSpec) { l.StateStorage = validLogForwarderForTest().StateStorage },
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
		{"mismatched block", func(l *LogForwarderSpec) { l.Exporter.Type = "otlphttp" }},
		{"multiple blocks", func(l *LogForwarderSpec) { l.Exporter.OTLPHTTP = &LogOTLPHTTPExporter{Endpoint: "https://example.com"} }},
		{"zero queue", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.SendingQueue = &LogSendingQueue{QueueSize: &zero} }},
		{"bad startAt", func(l *LogForwarderSpec) { l.Sources[0].FileLog.StartAt = "End" }},
		{"bad size", func(l *LogForwarderSpec) { l.Sources[0].FileLog.MaxLogSize = "0MiB" }},
		{"optional credential", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.TokenSecretRef.Optional = &optional }},
		{"URL credentials", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.Endpoint = "https://user:password@example.com" }},
		{"plaintext without opt in", func(l *LogForwarderSpec) { l.Exporter.SplunkHEC.Endpoint = "http://example.com" }},
		{"negative storage", func(l *LogForwarderSpec) {
			l.StateStorage.Resources.Requests[core.ResourceStorage] = resource.MustParse("-1Gi")
		}},
		{"automatic rollout", func(l *LogForwarderSpec) { l.RolloutPolicy = LogForwarderRolloutPolicy("Automatic") }},
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
