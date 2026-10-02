# Shared logging API

This is an additive API-library change. Generated CRDs expose the schema;
operators and admission-server binaries must adopt this library and implement
the selected engine adapter before new configurations are used in production.
CRD installation alone does not configure database stdout or a Collector.

## Ownership

- Missing collectionMode means Sidecar, preserving existing installations.
- Sidecar uses KubeDB-managed file receivers and an inline typed exporter.
- NodeAgent selects database log sources suitable for stdout/stderr.
  The platform installs and configures its own node Collector, including
  parsing, routing, authentication, TLS, queues, and upgrades.
- There is no LogRoute, routeRef, collectorRef, or node-agent configuration CR.
- Database audit/statement policies are separate from forwarding.
  Unsupported sources must be rejected by the engine adapter.

## NodeAgent

```yaml
spec:
  logForwarder:
    collectionMode: NodeAgent
    sources:
      - name: general
    rolloutPolicy: Manual
```

No exporter, stateStorage, sourceStorage, processors, delivery, resources,
securityContext, fileLog, or initialPosition is allowed in NodeAgent.
Externally collected logs may exist independently of this KubeDB field.
KubeDB cannot claim backend delivery for a platform-managed Collector.

## Sidecar

```yaml
spec:
  logForwarder:
    collectionMode: Sidecar
    sources:
      - name: general
        fileLog:
          startAt: end
          maxLogSize: 1MiB
    exporter:
      name: primary
      type: splunk_hec
      splunkHec:
        endpoint: https://splunk.example.com:8088/services/collector
        tokenSecretRef:
          name: splunk-ingest
          key: token
        sendingQueue:
          enabled: true
          queueSize: 1000
          numConsumers: 2
          blockOnOverflow: true
        retryOnFailure:
          enabled: true
          initialInterval: 2s
          maxInterval: 30s
          maxElapsedTime: 0s
    processors:
      batch:
        timeout: 2s
        sendBatchSize: 256
    stateStorage:
      accessModes: [ReadWriteOnce]
      volumeMode: Filesystem
      resources:
        requests:
          storage: 1Gi
    rolloutPolicy: Manual
```

Exporter types and matching blocks:

    otlphttp          otlpHttp
    splunk_hec        splunkHec
    elasticsearch     elasticsearch
    datadog           datadog
    syslog            syslog

Exactly one block matching type is required. Elasticsearch uses native retry
(enabled, initialInterval, maxInterval, maxRetries), not retryOnFailure.
Public camelCase maps to native Collector snake_case in the runtime renderer.
Only compiled, tested components and settings may be advertised by a runtime.

KubeDB owns paths, parser operators, pipeline wiring, checkpoint storage,
persistent queue component IDs, and memory limits. Request-count queue size
is not a byte bound. Blocking overflow is not an end-to-end delivery guarantee.
File startAt only applies when no checkpoint exists.
Per-source tuning for a shared physical reader must agree.

Credentials select non-optional keys from same-namespace Secrets, not broad
envFrom in new typed configurations. TLS defaults to Verify. HTTP plaintext
requires explicit Disabled. Existing TLS selectors remain compatible.

## Compatibility and activation

Legacy destination, delivery, and initialPosition remain accepted.
Exporter cannot be combined with populated legacy destination or delivery.
A source cannot specify both initialPosition and fileLog.startAt.
Legacy splunk/elastic profiles correspond to splunk_hec/elasticsearch types.
No stored spec, component identity, queue, checkpoint, or PVC is renamed
merely to adopt the new API field names. Legacy raw exporter configurations
remain Sidecar-only. Manual OpsRequest activation remains the policy.

The current database embedding is Postgres (v1 and v1alpha2) and Neo4j
(v1alpha2). Shared types are engine-neutral; other database operators need
separate capability/defaulting/runtime fan-out, not a blanket audit promise.
