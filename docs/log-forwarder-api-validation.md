# Logging API verification

Date: 2026-10-02
VM: 10.2.0.146
Repository: /home/ubuntu/go/src/kubedb.dev/apimachinery
Branch: feat/native-log-forwarder
Scope: shared API and pure helpers, not a deployed operator change.

Checks:

    make gen                      Deepcopy, conversions, CRDs and OpenAPI generated.
    make fmt                      Handwritten and generated sources formatted.
    make build                    Library packages compiled.
    Focused Go tests              Both API versions, strict parser and round trips.
    CRD server-side dry run       Postgres and Neo4j schemas accepted, not applied.
    Collector 0.153.0 validate     Typed processor pipeline accepted.
    Collector 0.153.0 validate     Custom basicauth extension accepted.
    Collector 0.153.0 validate     Invalid OTTL rejected.
    Collector Docker startup      Percentage limiter detected 256 MiB cgroup budget.

Observed percentage defaults for a 256 MiB container:

    hard percentage               80
    spike percentage              15
    computed hard threshold       204 MiB
    computed spike allowance      38 MiB

The pinned Collector rounds these thresholds. They are based on the container
limit, not the host memory. The temporary test container was stopped afterward.

Deployment boundary:

    No CRDs were applied.
    No database or operator workload was replaced.
    Only dummy credentials were used by the local Collector fixtures.
    Each database renderer must still consume the new helpers.
    Stored POC specs and compatible operator binaries need coordinated migration
    before applying the schemas that remove the unreleased legacy fields.

See log-forwarder-api.md for the API and controlled-rollout contract.
