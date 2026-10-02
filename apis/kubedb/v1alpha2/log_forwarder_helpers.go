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
	"net/url"
	"regexp"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetCollectionDefaults never populates deprecated fields for canonical or node specs.
func (lf *LogForwarderSpec) SetCollectionDefaults() {
	if lf.CollectionMode == "" {
		lf.CollectionMode = LogCollectionModeSidecar
	}
	if lf.Enabled == nil {
		v := true
		lf.Enabled = &v
	}
	if lf.RolloutPolicy == "" {
		lf.RolloutPolicy = LogForwarderRolloutManual
	}
	if lf.CollectionMode == LogCollectionModeNodeAgent {
		return
	}
	for i := range lf.Sources {
		if lf.Sources[i].FileLog != nil && lf.Sources[i].FileLog.StartAt == "" && lf.Sources[i].InitialPosition == "" {
			lf.Sources[i].FileLog.StartAt = "end"
		}
	}
	if lf.Exporter == nil {
		return
	}
	if lf.Exporter.Name == "" {
		lf.Exporter.Name = "primary"
	}
	options, retry := lf.Exporter.options()
	if options != nil {
		if options.TLS == nil {
			options.TLS = &LogForwarderTLS{Mode: LogForwarderTLSModeVerify}
		} else if options.TLS.Mode == "" {
			options.TLS.Mode = LogForwarderTLSModeVerify
		}
		if options.SendingQueue == nil {
			options.SendingQueue = &LogSendingQueue{}
		}
		q := options.SendingQueue
		if q.Enabled == nil {
			v := true
			q.Enabled = &v
		}
		if q.QueueSize == nil {
			v := int32(1000)
			q.QueueSize = &v
		}
		if q.NumConsumers == nil {
			v := int32(2)
			q.NumConsumers = &v
		}
		if q.BlockOnOverflow == nil {
			v := true
			q.BlockOnOverflow = &v
		}
	}
	if retry != nil {
		if *retry == nil {
			*retry = &LogRetryOnFailure{}
		}
		r := *retry
		if r.Enabled == nil {
			v := true
			r.Enabled = &v
		}
		if r.InitialInterval == nil {
			r.InitialInterval = &metav1.Duration{Duration: 2 * time.Second}
		}
		if r.MaxInterval == nil {
			r.MaxInterval = &metav1.Duration{Duration: 30 * time.Second}
		}
		if r.MaxElapsedTime == nil {
			r.MaxElapsedTime = &metav1.Duration{}
		}
	}
	// Elasticsearch's native retry defaults are not exporterhelper defaults.
	// Leave unspecified fields to the pinned Elasticsearch builder.
}

func (e *LogExporterSpec) options() (*LogExporterOptions, **LogRetryOnFailure) {
	switch e.Type {
	case "otlphttp":
		if e.OTLPHTTP != nil {
			return &e.OTLPHTTP.LogExporterOptions, &e.OTLPHTTP.RetryOnFailure
		}
	case "splunk_hec":
		if e.SplunkHEC != nil {
			return &e.SplunkHEC.LogExporterOptions, &e.SplunkHEC.RetryOnFailure
		}
	case "elasticsearch":
		if e.Elasticsearch != nil {
			return &e.Elasticsearch.LogExporterOptions, nil
		}
	case "datadog":
		if e.Datadog != nil {
			return &e.Datadog.LogExporterOptions, &e.Datadog.RetryOnFailure
		}
	case "syslog":
		if e.Syslog != nil {
			return &e.Syslog.LogExporterOptions, &e.Syslog.RetryOnFailure
		}
	}
	return nil, nil
}

// ValidateCollection checks the shared API independently of engine capabilities.
// Engine controllers must separately implement and certify selected sources/modes.
func (lf *LogForwarderSpec) ValidateCollection() error {
	if lf == nil {
		return nil
	}
	mode := lf.CollectionMode
	if mode == "" {
		mode = LogCollectionModeSidecar
	}
	if mode != LogCollectionModeSidecar && mode != LogCollectionModeNodeAgent {
		return fmt.Errorf("unsupported collectionMode %q", mode)
	}
	if lf.RolloutPolicy != "" && lf.RolloutPolicy != LogForwarderRolloutManual {
		return fmt.Errorf("rolloutPolicy must be Manual; use an OpsRequest")
	}
	d := lf.Destination
	legacy := d.Profile != "" || d.ExporterConfig != "" || d.Endpoint != "" || d.SecretRef != nil || d.TLS != nil || d.ExtraProcessors != "" || d.ExtraExtensions != ""
	for _, s := range lf.Sources {
		if s.Name == "" {
			return fmt.Errorf("source name is required")
		}
		if s.FileLog != nil {
			if mode == LogCollectionModeNodeAgent {
				return fmt.Errorf("NodeAgent does not accept per-database fileLog tuning")
			}
			if s.InitialPosition != "" && s.FileLog.StartAt != "" {
				return fmt.Errorf("initialPosition and fileLog.startAt are mutually exclusive")
			}
			if s.FileLog.StartAt != "" && s.FileLog.StartAt != "end" && s.FileLog.StartAt != "beginning" {
				return fmt.Errorf("fileLog.startAt must be end or beginning")
			}
			if s.FileLog.PollInterval != nil && s.FileLog.PollInterval.Duration <= 0 {
				return fmt.Errorf("fileLog.pollInterval must be positive")
			}
			if s.FileLog.MaxLogSize != "" && !regexp.MustCompile(`^[1-9][0-9]*(B|KiB|MiB|GiB|KB|MB|GB)?$`).MatchString(s.FileLog.MaxLogSize) {
				return fmt.Errorf("fileLog.maxLogSize must be a positive Collector size, for example 1MiB")
			}
		}
		if mode == LogCollectionModeNodeAgent && s.InitialPosition != "" {
			return fmt.Errorf("NodeAgent does not accept initialPosition")
		}
	}
	if mode == LogCollectionModeNodeAgent {
		if legacy || lf.Exporter != nil || lf.Delivery != nil || lf.Processors != nil || lf.SourceStorage != nil || lf.SecurityContext != nil || len(lf.Resources.Requests) > 0 || len(lf.Resources.Limits) > 0 || len(lf.StateStorage.AccessModes) > 0 || lf.StateStorage.VolumeMode != nil || len(lf.StateStorage.Resources.Requests) > 0 {
			return fmt.Errorf("NodeAgent collector configuration belongs to the platform; sidecar-only fields are not allowed")
		}
		return nil
	}
	if lf.Exporter != nil && (legacy || lf.Delivery != nil) {
		return fmt.Errorf("exporter cannot be combined with destination or delivery")
	}
	enabled := lf.Enabled == nil || *lf.Enabled
	if lf.Exporter == nil {
		if d.Profile != "" && d.ExporterConfig != "" {
			return fmt.Errorf("destination.profile and exporterConfig are mutually exclusive")
		}
		if enabled && (d.Profile != "") == (d.ExporterConfig != "") {
			return fmt.Errorf("Sidecar requires exporter or a legacy destination")
		}
	} else if err := lf.Exporter.validate(); err != nil {
		return err
	}
	if enabled {
		if len(lf.StateStorage.AccessModes) != 1 || lf.StateStorage.AccessModes[0] != core.ReadWriteOnce {
			return fmt.Errorf("Sidecar stateStorage must use exactly [ReadWriteOnce]")
		}
		if lf.StateStorage.VolumeMode != nil && *lf.StateStorage.VolumeMode != core.PersistentVolumeFilesystem {
			return fmt.Errorf("stateStorage must use Filesystem")
		}
		q, ok := lf.StateStorage.Resources.Requests[core.ResourceStorage]
		if !ok || q.Sign() <= 0 {
			return fmt.Errorf("stateStorage must request positive storage")
		}
	}
	if s := lf.SourceStorage; s != nil {
		if (s.EmptyDir != nil) == (s.VolumeClaimTemplate != nil) {
			return fmt.Errorf("sourceStorage requires exactly one volume type")
		}
		if s.EmptyDir != nil && (s.EmptyDir.SizeLimit == nil || s.EmptyDir.SizeLimit.Sign() <= 0) {
			return fmt.Errorf("sourceStorage.emptyDir requires a positive sizeLimit")
		}
	}
	if lf.Processors != nil && lf.Processors.Batch != nil {
		b := lf.Processors.Batch
		if b.Timeout != nil && b.Timeout.Duration <= 0 {
			return fmt.Errorf("batch.timeout must be positive")
		}
		if b.SendBatchSize != nil && *b.SendBatchSize < 1 {
			return fmt.Errorf("batch.sendBatchSize must be positive")
		}
		if b.SendBatchMaxSize != nil && (*b.SendBatchMaxSize < 0 || (*b.SendBatchMaxSize > 0 && b.SendBatchSize != nil && *b.SendBatchMaxSize < *b.SendBatchSize)) {
			return fmt.Errorf("batch.sendBatchMaxSize must be zero or at least sendBatchSize")
		}
	}
	return nil
}

func (e *LogExporterSpec) validate() error {
	count := 0
	for _, present := range []bool{e.OTLPHTTP != nil, e.SplunkHEC != nil, e.Elasticsearch != nil, e.Datadog != nil, e.Syslog != nil} {
		if present {
			count++
		}
	}
	o, r := e.options()
	if count != 1 || o == nil {
		return fmt.Errorf("exporter requires exactly one block matching type")
	}
	if e.Name != "" && !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`).MatchString(e.Name) {
		return fmt.Errorf("exporter.name must be a stable component name")
	}
	if q := o.SendingQueue; q != nil {
		if q.QueueSize != nil && *q.QueueSize < 1 || q.NumConsumers != nil && *q.NumConsumers < 1 {
			return fmt.Errorf("sendingQueue queueSize and numConsumers must be positive")
		}
	}
	if r != nil && *r != nil {
		if err := validateLogRetry((*r).InitialInterval, (*r).MaxInterval, (*r).MaxElapsedTime); err != nil {
			return err
		}
	}
	if tls := o.TLS; tls != nil {
		if tls.Mode != "" && tls.Mode != LogForwarderTLSModeVerify && tls.Mode != LogForwarderTLSModeDisabled {
			return fmt.Errorf("unsupported TLS mode")
		}
		if (tls.ClientCertSecretRef != nil) != (tls.ClientKeySecretRef != nil) {
			return fmt.Errorf("both client certificate and key selectors are required")
		}
		if tls.Mode == LogForwarderTLSModeDisabled && (tls.CASecretRef != nil || tls.ClientCertSecretRef != nil) {
			return fmt.Errorf("Disabled TLS cannot use CA or client certificates")
		}
		for _, s := range []*core.SecretKeySelector{tls.CASecretRef, tls.ClientCertSecretRef, tls.ClientKeySecretRef} {
			if err := validateLogSecret(s); err != nil {
				return err
			}
		}
	}
	switch e.Type {
	case "otlphttp":
		c := e.OTLPHTTP
		if (c.Endpoint != "") == (c.LogsEndpoint != "") {
			return fmt.Errorf("otlpHttp requires exactly one endpoint or logsEndpoint")
		}
		endpoint := c.Endpoint
		if endpoint == "" {
			endpoint = c.LogsEndpoint
		}
		if err := validateLogURL(endpoint, o.TLS); err != nil {
			return err
		}
		if err := validateLogAuth(c.Auth); err != nil {
			return err
		}
	case "splunk_hec":
		if err := validateLogURL(e.SplunkHEC.Endpoint, o.TLS); err != nil {
			return err
		}
		if err := validateLogSecret(&e.SplunkHEC.TokenSecretRef); err != nil {
			return err
		}
	case "elasticsearch":
		c := e.Elasticsearch
		if len(c.Endpoints) == 0 {
			return fmt.Errorf("elasticsearch.endpoints must not be empty")
		}
		for _, endpoint := range c.Endpoints {
			if err := validateLogURL(endpoint, o.TLS); err != nil {
				return err
			}
		}
		if err := validateLogSecret(c.APIKeySecretRef); err != nil {
			return err
		}
		if c.Retry != nil {
			if err := validateLogRetry(c.Retry.InitialInterval, c.Retry.MaxInterval, nil); err != nil {
				return err
			}
			if c.Retry.MaxRetries != nil && *c.Retry.MaxRetries < 0 {
				return fmt.Errorf("retry.maxRetries cannot be negative")
			}
		}
	case "datadog":
		if e.Datadog.API.Site == "" {
			return fmt.Errorf("datadog.api.site is required")
		}
		if err := validateLogSecret(&e.Datadog.API.KeySecretRef); err != nil {
			return err
		}
	case "syslog":
		c := e.Syslog
		if c.Endpoint == "" || c.Port < 1 || c.Port > 65535 || c.Network != "tcp" || (c.Protocol != "rfc5424" && c.Protocol != "rfc3164") {
			return fmt.Errorf("syslog requires endpoint, valid port, tcp, and a supported protocol")
		}
		if c.Facility != nil && (*c.Facility < 0 || *c.Facility > 23) {
			return fmt.Errorf("syslog.facility must be between 0 and 23")
		}
	}
	return nil
}

func validateLogRetry(initial, max, elapsed *metav1.Duration) error {
	if initial != nil && initial.Duration <= 0 || max != nil && max.Duration <= 0 || elapsed != nil && elapsed.Duration < 0 {
		return fmt.Errorf("retry intervals must be positive and elapsed time non-negative")
	}
	if initial != nil && max != nil && initial.Duration > max.Duration {
		return fmt.Errorf("retry.maxInterval must be at least initialInterval")
	}
	return nil
}

func validateLogSecret(s *core.SecretKeySelector) error {
	if s != nil && (s.Name == "" || s.Key == "" || s.Optional != nil && *s.Optional) {
		return fmt.Errorf("credential selectors require name, key, and non-optional Secrets")
	}
	return nil
}

func validateLogURL(endpoint string, tls *LogForwarderTLS) error {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return fmt.Errorf("endpoint must be an HTTP(S) URL without embedded credentials or fragments")
	}
	mode := LogForwarderTLSModeVerify
	if tls != nil && tls.Mode != "" {
		mode = tls.Mode
	}
	if (mode == LogForwarderTLSModeDisabled) != (u.Scheme == "http") {
		return fmt.Errorf("endpoint scheme conflicts with TLS mode; plaintext requires Disabled")
	}
	return nil
}

func validateLogAuth(a *LogExporterAuth) error {
	if a == nil {
		return nil
	}
	token := a.TokenSecretRef != nil
	user := a.UsernameSecretRef != nil
	pass := a.PasswordSecretRef != nil
	value := a.ValueSecretRef != nil
	header := a.HeaderName != ""
	valid := false
	switch a.Type {
	case "None":
		valid = !token && !user && !pass && !value && !header
	case "Token", "Bearer":
		valid = token && !user && !pass && !value && !header
	case "Basic":
		valid = user && pass && !token && !value && !header
	case "Header":
		valid = value && header && !token && !user && !pass
	}
	if !valid {
		return fmt.Errorf("auth fields must match type")
	}
	for _, s := range []*core.SecretKeySelector{a.TokenSecretRef, a.UsernameSecretRef, a.PasswordSecretRef, a.ValueSecretRef} {
		if err := validateLogSecret(s); err != nil {
			return err
		}
	}
	return nil
}
