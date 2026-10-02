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
	"fmt"
	"regexp"
	"strings"
	"time"

	lfconfig "kubedb.dev/apimachinery/pkg/logforwarder"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetDefaults preserves complete user memory pairs; partial pairs stay invalid.
func (p *LogProcessors) SetDefaults() {
	if p.MemoryLimiter == nil {
		p.MemoryLimiter = &LogMemoryLimiterProcessor{}
	}
	m := p.MemoryLimiter
	if m.CheckInterval == nil {
		m.CheckInterval = &metav1.Duration{Duration: time.Second}
	}
	if m.LimitPercentage == nil && m.SpikeLimitPercentage == nil && m.LimitMiB == nil && m.SpikeLimitMiB == nil {
		hard, spike := int32(80), int32(15)
		m.LimitPercentage = &hard
		m.SpikeLimitPercentage = &spike
	}
	if p.Batch == nil {
		p.Batch = &LogBatchProcessor{}
	}
	if p.Batch.Timeout == nil {
		p.Batch.Timeout = &metav1.Duration{Duration: 2 * time.Second}
	}
	if p.Batch.SendBatchSize == nil {
		v := int32(256)
		p.Batch.SendBatchSize = &v
	}
	if p.Filter != nil && p.Filter.ErrorMode == "" {
		p.Filter.ErrorMode = "propagate"
	}
	if p.Transform != nil && p.Transform.ErrorMode == "" {
		p.Transform.ErrorMode = "propagate"
	}
}

func (lf *LogForwarderSpec) validateProcessing() error {
	p := lf.Processors
	if p != nil {
		if _, _, err := p.UserProcessorConfiguration(); err != nil {
			return err
		}
		if _, err := p.memoryConfiguration(lf.Resources); err != nil {
			return err
		}
		defaults := p.DeepCopy()
		defaults.SetDefaults()
		b := defaults.Batch
		if b.Timeout.Duration <= 0 || *b.SendBatchSize < 1 || b.SendBatchMaxSize != nil && (*b.SendBatchMaxSize < 0 || *b.SendBatchMaxSize > 0 && *b.SendBatchMaxSize < *b.SendBatchSize) {
			return fmt.Errorf("batch.sendBatchMaxSize must be zero or at least the effective sendBatchSize")
		}
	}
	extensions, err := lf.ExtensionConfiguration()
	if err != nil {
		return err
	}
	if lf.Exporter != nil && lf.Exporter.OTLPHTTP != nil {
		a := lf.Exporter.OTLPHTTP.Auth
		if a != nil && a.Type == "Extension" {
			if _, ok := extensions[a.ExtensionRef]; !ok {
				return fmt.Errorf("exporter.auth.extensionRef must name a defined extension")
			}
		}
	}
	return nil
}

// ExtensionConfiguration returns native extension objects; controllers activate every returned ID.
// SecretEnv references must be resolved into the same namespace by the controller.
func (lf *LogForwarderSpec) ExtensionConfiguration() (map[string]map[string]interface{}, error) {
	if lf.Extensions == nil {
		return map[string]map[string]interface{}{}, nil
	}
	e := lf.Extensions
	cfg, err := lfconfig.ParseComponents(e.ExtraConfig, map[string]bool{"file_storage/*": true, "health_check/*": true})
	if err != nil {
		return nil, err
	}
	for name, s := range e.SecretEnv {
		if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(name) || strings.HasPrefix(name, "KUBEDB_") || strings.HasPrefix(name, "OTEL_") || strings.HasPrefix(name, "K8S_") || strings.HasPrefix(name, "POD_") || strings.HasPrefix(name, "NODE_") || strings.HasPrefix(name, "DB_") || strings.HasPrefix(name, "DATABASE_") || name == "GOMEMLIMIT" || name == "PATH" || name == "HOME" || name == "LD_PRELOAD" || name == "LD_LIBRARY_PATH" {
			return nil, fmt.Errorf("extension secretEnv name is invalid or reserved")
		}
		if err := validateLogSecret(&s); err != nil {
			return nil, err
		}
	}
	for _, match := range regexp.MustCompile(`\$\{env:([A-Za-z_][A-Za-z0-9_]*)\}`).FindAllStringSubmatch(e.ExtraConfig, -1) {
		if _, ok := e.SecretEnv[match[1]]; !ok {
			return nil, fmt.Errorf("extension environment reference requires a secretEnv binding")
		}
	}
	return cfg, nil
}

// UserProcessorConfiguration converts typed options to native Collector keys.
// It checks structure only; the pinned Collector must validate OTTL and component support.
func (p *LogProcessors) UserProcessorConfiguration() (map[string]map[string]interface{}, []string, error) {
	if p == nil {
		return map[string]map[string]interface{}{}, nil, nil
	}
	cfg, err := lfconfig.ParseComponents(p.ExtraConfig, map[string]bool{"memory_limiter/*": true, "batch/*": true, "resource/kubedb": true, "attributes": true, "transform": true, "filter": true})
	// Named extra attributes/transform/filter components are permitted, but bare IDs are reserved.
	if err != nil {
		return nil, nil, err
	}
	if p.Attributes != nil {
		if len(p.Attributes.Actions) == 0 {
			return nil, nil, fmt.Errorf("attributes.actions must not be empty")
		}
		actions := []interface{}{}
		for _, a := range p.Attributes.Actions {
			if a.Key == "" {
				return nil, nil, fmt.Errorf("attribute action requires key")
			}
			item := map[string]interface{}{"key": a.Key, "action": a.Action}
			inputs := 0
			if a.Value != nil {
				var v interface{}
				if err := json.Unmarshal(a.Value.Raw, &v); err != nil {
					return nil, nil, fmt.Errorf("attribute value must be valid scalar JSON")
				}
				switch v.(type) {
				case string, bool, float64:
				default:
					return nil, nil, fmt.Errorf("attribute value must be a non-null scalar")
				}
				item["value"] = v
				inputs++
			}
			if a.FromAttribute != "" {
				item["from_attribute"] = a.FromAttribute
				inputs++
			}
			if a.FromContext != "" {
				item["from_context"] = a.FromContext
				inputs++
			}
			switch a.Action {
			case "insert", "update", "upsert":
				if inputs != 1 || a.Pattern != "" || a.ConvertedType != "" {
					return nil, nil, fmt.Errorf("attribute assignment requires exactly one value source")
				}
			case "delete", "hash":
				if inputs != 0 || a.Pattern != "" || a.ConvertedType != "" {
					return nil, nil, fmt.Errorf("delete/hash action does not accept value sources")
				}
			case "extract":
				if inputs != 0 || a.Pattern == "" || a.ConvertedType != "" {
					return nil, nil, fmt.Errorf("extract action requires only pattern")
				}
				r, err := regexp.Compile(a.Pattern)
				if err != nil || len(r.SubexpNames()) < 2 {
					return nil, nil, fmt.Errorf("extract pattern requires named captures")
				}
				named := false
				for _, n := range r.SubexpNames() {
					if n != "" {
						named = true
					}
				}
				if !named {
					return nil, nil, fmt.Errorf("extract pattern requires named captures")
				}
				item["pattern"] = a.Pattern
			case "convert":
				if inputs != 0 || a.Pattern != "" || (a.ConvertedType != "int" && a.ConvertedType != "double" && a.ConvertedType != "string") {
					return nil, nil, fmt.Errorf("convert action requires only convertedType")
				}
				item["converted_type"] = a.ConvertedType
			default:
				return nil, nil, fmt.Errorf("unsupported attribute action")
			}
			actions = append(actions, item)
		}
		cfg["attributes"] = map[string]interface{}{"actions": actions}
	}
	validMode := func(m LogProcessorErrorMode) bool { return m == "" || m == "propagate" || m == "ignore" }
	mode := func(m LogProcessorErrorMode) string {
		if m == "" {
			return "propagate"
		}
		return string(m)
	}
	if p.Filter != nil {
		f := p.Filter
		if !validMode(f.ErrorMode) || len(f.Logs.LogRecord) == 0 {
			return nil, nil, fmt.Errorf("filter requires supported errorMode and logRecord conditions")
		}
		for _, s := range f.Logs.LogRecord {
			if strings.TrimSpace(s) == "" {
				return nil, nil, fmt.Errorf("filter condition must not be empty")
			}
		}
		cfg["filter"] = map[string]interface{}{"error_mode": mode(f.ErrorMode), "logs": map[string]interface{}{"log_record": f.Logs.LogRecord}}
	}
	if p.Transform != nil {
		f := p.Transform
		if !validMode(f.ErrorMode) || len(f.LogStatements) == 0 {
			return nil, nil, fmt.Errorf("transform requires supported errorMode and logStatements")
		}
		groups := []interface{}{}
		for _, g := range f.LogStatements {
			if (g.Context != "log" && g.Context != "resource" && g.Context != "scope") || len(g.Statements) == 0 {
				return nil, nil, fmt.Errorf("transform group requires supported context and statements")
			}
			for _, s := range append(append([]string{}, g.Statements...), g.Conditions...) {
				if strings.TrimSpace(s) == "" {
					return nil, nil, fmt.Errorf("transform statement or condition must not be empty")
				}
			}
			group := map[string]interface{}{"context": g.Context, "statements": g.Statements}
			if len(g.Conditions) > 0 {
				group["conditions"] = g.Conditions
			}
			groups = append(groups, group)
		}
		cfg["transform"] = map[string]interface{}{"error_mode": mode(f.ErrorMode), "log_statements": groups}
	}
	order := p.Order
	if len(order) == 0 && strings.TrimSpace(p.ExtraConfig) == "" {
		for _, id := range []string{"attributes", "transform", "filter"} {
			if _, ok := cfg[id]; ok {
				order = append(order, id)
			}
		}
	}
	if len(cfg) == 0 && len(order) == 0 {
		return cfg, nil, nil
	}
	ordered, err := lfconfig.OrderedStages(cfg, order)
	return cfg, ordered, err
}

func (p *LogProcessors) memoryConfiguration(resources core.ResourceRequirements) (map[string]interface{}, error) {
	copy := p.DeepCopy()
	copy.SetDefaults()
	m := copy.MemoryLimiter
	if m.CheckInterval.Duration <= 0 {
		return nil, fmt.Errorf("memoryLimiter.checkInterval must be positive")
	}
	percentage := m.LimitPercentage != nil || m.SpikeLimitPercentage != nil
	absolute := m.LimitMiB != nil || m.SpikeLimitMiB != nil
	if percentage == absolute {
		return nil, fmt.Errorf("choose exactly one memory threshold pair")
	}
	cfg := map[string]interface{}{"check_interval": m.CheckInterval.Duration.String()}
	limit, limited := resources.Limits[core.ResourceMemory]
	if percentage {
		if m.LimitPercentage == nil || m.SpikeLimitPercentage == nil || *m.LimitPercentage < 1 || *m.LimitPercentage > 99 || *m.SpikeLimitPercentage < 1 || *m.SpikeLimitPercentage >= *m.LimitPercentage {
			return nil, fmt.Errorf("percentage thresholds require a valid hard/spike pair")
		}
		cfg["limit_percentage"] = *m.LimitPercentage
		cfg["spike_limit_percentage"] = *m.SpikeLimitPercentage
	} else {
		if m.LimitMiB == nil || m.SpikeLimitMiB == nil || *m.LimitMiB < 1 || *m.SpikeLimitMiB < 1 || *m.SpikeLimitMiB >= *m.LimitMiB {
			return nil, fmt.Errorf("MiB thresholds require a valid hard/spike pair")
		}
		if limited && int64(*m.LimitMiB)*1024*1024 >= limit.Value() {
			return nil, fmt.Errorf("memoryLimiter.limitMiB must be below the sidecar memory limit")
		}
		cfg["limit_mib"] = *m.LimitMiB
		cfg["spike_limit_mib"] = *m.SpikeLimitMiB
	}
	return cfg, nil
}

// ProcessorConfiguration creates the managed pipeline; callers supply trusted database/pod identity.
// An explicit Sidecar memory limit is mandatory at runtime, including percentage mode.
func (lf *LogForwarderSpec) ProcessorConfiguration(identity map[string]string) (map[string]map[string]interface{}, []string, error) {
	if lf.CollectionMode == LogCollectionModeNodeAgent {
		return nil, nil, fmt.Errorf("NodeAgent processors are platform-owned")
	}
	memory, ok := lf.Resources.Limits[core.ResourceMemory]
	if !ok || memory.Sign() <= 0 {
		return nil, nil, fmt.Errorf("Sidecar requires an explicit positive memory limit before rendering")
	}
	if len(identity) == 0 {
		return nil, nil, fmt.Errorf("trusted database/pod resource identity is required")
	}
	p := lf.Processors
	if p == nil {
		p = &LogProcessors{}
	}
	p = p.DeepCopy()
	p.SetDefaults()
	cfg, order, err := p.UserProcessorConfiguration()
	if err != nil {
		return nil, nil, err
	}
	mem, err := p.memoryConfiguration(lf.Resources)
	if err != nil {
		return nil, nil, err
	}
	cfg["memory_limiter"] = mem
	attrs := []interface{}{}
	ids := make(map[string]map[string]interface{}, len(identity))
	for k := range identity {
		ids[k] = nil
	}
	for _, k := range lfconfig.SortedIDs(ids) {
		attrs = append(attrs, map[string]interface{}{"key": k, "value": identity[k], "action": "upsert"})
	}
	cfg["resource/kubedb"] = map[string]interface{}{"attributes": attrs}
	b := p.Batch
	batch := map[string]interface{}{"timeout": b.Timeout.Duration.String(), "send_batch_size": *b.SendBatchSize}
	if b.SendBatchMaxSize != nil {
		batch["send_batch_max_size"] = *b.SendBatchMaxSize
	}
	if b.Timeout.Duration <= 0 || *b.SendBatchSize <= 0 || b.SendBatchMaxSize != nil && (*b.SendBatchMaxSize < 0 || *b.SendBatchMaxSize > 0 && *b.SendBatchMaxSize < *b.SendBatchSize) {
		return nil, nil, fmt.Errorf("invalid batch thresholds")
	}
	cfg["batch"] = batch
	pipeline := append([]string{"memory_limiter"}, order...)
	pipeline = append(pipeline, "resource/kubedb", "batch")
	return cfg, pipeline, nil
}
