# KubeDB logging API

Status: unreleased API. This replaces the old destination/delivery shape.
Scope: shared apimachinery types and pure configuration helpers, not operator deployment.

## Ownership

Sidecar: KubeDB selects sources, mounts log/state storage, renders the Collector,
and activates configuration through an OpsRequest.

NodeAgent: KubeDB configures supported database streams for container stdout/stderr.
The platform installs and configures the DaemonSet, including parsing, processing,
routing, credentials, buffering and retention. File-only, binary and table-resident
audit sources cannot be promised as stdout streams without an engine adapter.

Neither mode creates audit facilities missing from the selected image or edition.
Operators must reject unsupported sources/modes using their version capabilities.

## Sidecar example

Every Collector-specific field below is Sidecar-only.

```yaml
spec:
  # Configure database log forwarding.
  logForwarder:
    # Select an operator-managed per-pod collector.
    collectionMode: Sidecar
    # Activate log forwarding.
    enabled: true
    # Require controlled activation through an OpsRequest.
    rolloutPolicy: Manual
    # Select version-advertised database streams.
    sources:
      # Identify the general database log stream.
      - name: general
        # Tune the managed file reader; paths and parsers remain engine-owned.
        fileLog:
          # Read new files from their beginning; checkpoints take precedence.
          startAt: beginning
    # Set the collector container budget.
    resources:
      # Bound container resources explicitly.
      limits:
        # Supply the cgroup memory budget used by percentage thresholds.
        memory: 256Mi
    # Allocate persistent checkpoints and delivery state per pod.
    stateStorage:
      # Use one-writer state storage.
      accessModes: [ReadWriteOnce]
      # Use filesystem-backed state.
      volumeMode: Filesystem
      # Request bounded state capacity.
      resources:
        # Declare the storage request.
        requests:
          # Set the persistent state volume capacity.
          storage: 1Gi
    # Select exactly one destination/exporter variant.
    exporter:
      # Keep the persistent delivery identity stable.
      name: primary
      # Select the compiled OTLP/HTTP exporter.
      type: otlphttp
      # Configure the selected exporter.
      otlpHttp:
        # Supply a base URL; the exporter appends /v1/logs.
        endpoint: https://collector.example.com
        # Select an authentication extension.
        auth:
          # Delegate authentication to a native extension.
          type: Extension
          # Reference a defined extension component ID.
          extensionRef: basicauth/client
    # Configure the five typed processor options.
    processors:
      # Keep the always-enabled memory limiter first.
      memoryLimiter:
        # Check memory every second.
        checkInterval: 1s
        # Set the hard threshold; pairs with spikeLimitPercentage.
        limitPercentage: 80
        # Set the spike allowance; pairs with limitPercentage.
        spikeLimitPercentage: 15
      # Modify record attributes.
      attributes:
        # Apply native actions in this order.
        actions:
          # Identify the attribute to modify.
          - key: environment
            # Insert or replace its value.
            action: upsert
            # Supply a scalar value.
            value: production
      # Normalize log records with native OTTL.
      transform:
        # Propagate errors instead of silently bypassing normalization.
        errorMode: propagate
        # Group statements by context.
        logStatements:
          # Evaluate in log context.
          - context: log
            # Execute statements in order.
            statements:
              - 'set(attributes["normalized"], true)'
      # Drop unwanted records with native OTTL conditions.
      filter:
        # Propagate errors instead of silently bypassing filtering.
        errorMode: propagate
        # Configure log-only filtering.
        logs:
          # Drop records when any listed condition is true.
          logRecord:
            - 'attributes["drop"] == true'
      # Declare the user-stage sequence; managed stages are excluded.
      order: [attributes, transform, filter]
      # Keep batching last.
      batch:
        # Flush batches within two seconds.
        timeout: 2s
        # Trigger a batch at this record count.
        sendBatchSize: 256
    # Define native extensions and their credential bindings.
    extensions:
      # Supply component definitions without a top-level extensions wrapper.
      extraConfig: |
        basicauth/client:
          client_auth:
            username: ${env:EXT_USER}
            password: ${env:EXT_PASSWORD}
      # Bind each referenced environment variable to a non-optional Secret key.
      secretEnv:
        # Bind the extension username.
        EXT_USER:
          # Select the Secret in the database namespace.
          name: forwarding-auth
          # Select its username key.
          key: username
        # Bind the extension password.
        EXT_PASSWORD:
          # Select the Secret in the database namespace.
          name: forwarding-auth
          # Select its password key.
          key: password
```

## Memory pairs

Use either percentage settings OR absolute MiB settings, never both.
Each selected pair must contain both fields; partial pairs are rejected.

    Percentage pair                 Absolute pair
    ------------------------------  -------------------------
    limitPercentage: 80             limitMiB: 180
    spikeLimitPercentage: 15        spikeLimitMiB: 32

If neither pair is supplied, Go defaulting inserts the percentage pair 80/15.
Absolute settings never receive percentage defaults.
The soft threshold is hard limit minus spike allowance.
Percentage settings use the collector's detected container/cgroup budget.
The runtime helper requires an explicit positive container memory limit;
an absolute hard threshold must be strictly below that limit.
This processor is admission/backpressure protection, not a guarantee against OOM
or log loss. The pinned image's cgroup detection must be certified on supported nodes.

## Processing and extensions

Five typed fields: memoryLimiter, batch, attributes, filter, transform.
Without an explicit order, configured user stages run attributes -> transform -> filter.
Raw processors belong in processors.extraConfig, using native component IDs and
snake_case keys, without a processors wrapper. Raw extras require an explicit order
listing every typed and raw user component exactly once.

Managed pipeline:

    memory_limiter
        -> ordered user stages
        -> resource/kubedb (trusted identity restamp)
        -> batch

Native IDs memory_limiter/*, batch/* and resource/kubedb cannot be replaced.
Bare attributes/transform/filter IDs belong to the typed fields; named extras such as
attributes/custom are permitted. Custom stages can filter records, including all
records; the final restamp protects surviving records' operator-owned identity,
not the audit completeness of arbitrary user processing.

Extension helpers return native definitions. The operator must register every
returned ID in service.extensions, using deterministic ordering. file_storage/*
and health_check/* remain operator-owned. Extension authentication currently
belongs to the OTLP/HTTP exporter only. Secret selectors must resolve in the DB
namespace; values should never enter rendered Secrets, logs or status messages.
Use Secret-backed environment references instead of literal credentials in raw YAML.

No connector API, LogRoute, custom receiver, custom pipeline or replacement exporter
escape hatch is exposed.

## Validation and rollout boundary

Shared validation checks memory pairs, actions, ordering, duplicate YAML keys,
reserved component names, Secret selectors and authentication references.
It does NOT parse OTTL or establish availability of arbitrary native components.
The downstream renderer must validate the complete configuration against the exact
catalog-pinned Collector build before activation, and must fail without rolling out
an invalid configuration. This apimachinery library supplies pure helpers; each
database operator still needs to wire them into its renderer and OpsRequest workflow.

## Unreleased cleanup and migration

Removed: destination, delivery, initialPosition, legacy profile/raw-exporter
configuration and old extraProcessors/extraExtensions locations.
Use exporter, exporter-local sendingQueue/retry options, sources[].fileLog.startAt,
processors.extraConfig and extensions.extraConfig instead.

Do not apply these CRDs over old POC resources while old operator binaries still
consume removed fields. First migrate stored manifests/specs and build compatible
provisioner, webhook and OpsRequest components. Apply the generated CRDs only as
part of that coordinated rollout.
