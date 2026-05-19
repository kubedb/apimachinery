# AGENTS.md - KubeDB apimachinery

This file provides instructions for AI coding agents working in the KubeDB `apimachinery` Go module.

## Project Overview

Shared Go module that defines the KubeDB API types (CRDs), generated clientsets/informers/listers, admission webhooks, and helper libraries used across the KubeDB operator ecosystem. This is the foundational `kubedb.dev/apimachinery` library; all KubeDB operators (provisioner, ops-manager, autoscaler, schema-manager, etc.) depend on it. The module is library-only - it has no `main` entrypoint binary (only `hack/gencrd/main.go` for code generation).

## Build & Development Commands

All build/codegen targets run inside Docker images (`ghcr.io/appscode/golang-dev:1.25` and `ghcr.io/appscode/gengo:release-1.32`).

```bash
# Compile all packages (no top-level binary; produces no useful artifact)
make build

# Format Go sources
make fmt

# Run unit tests
make test            # equivalent to: make unit-tests
make unit-tests      # runs ./hack/test.sh on SRC_PKGS (apis client crds pkg)

# Lint (golangci-lint via vendored mode)
make lint

# Full CI pipeline: verify check-license lint build unit-tests
make ci

# Verify that go.mod / vendor and generated code are up to date
make verify          # runs verify-gen + verify-modules

# License headers
make add-license
make check-license

# Cleanup
make clean
```

### Code Generation

The repo is generation-heavy. Run `make gen` after any change to `apis/`:

```bash
make gen             # clientset + gen-enum + manifests + openapi + gen-conversion
make clientset       # client/{clientset,informers,listers}
make gen-conversion  # zz_generated.conversion.go between kubedb/v1 and v1alpha2
make gen-enum        # go generate ./apis/... (enum stringers)
make openapi         # apis/.../openapi_generated.go + openapi/swagger.json
make gen-crds        # crds/*.yaml from kubebuilder markers (controller-gen)
make manifests       # gen-crds + patch-crds + label-crds
make gen-crd-protos  # *.pb.go (currently optional, excluded from `make gen`)
```

`API_GROUPS` is the canonical list of generated groups - see top of `Makefile`:

```
kubedb:v1alpha1 kubedb:v1alpha2 kubedb:v1 gitops:v1alpha1 postgres:v1alpha1
catalog:v1alpha1 config:v1alpha1 ops:v1alpha1 autoscaling:v1alpha1
elasticsearch:v1alpha1 schema:v1alpha1 archiver:v1alpha1 kafka:v1alpha1
migrator:v1alpha1 ui:v1alpha1
```

## Project Structure

```
apis/                          # API type definitions (CRDs) - one subdir per group
  kubedb/                      # core group kubedb.com (databases)
    v1/                        # current served version (Postgres, MongoDB, MySQL, ...)
    v1alpha2/                  # legacy/storage version, full database set (35 dbs)
    v1alpha1/                  # oldest version, kept for conversion
    install/                   # scheme registration (Install + roundtrip/pruning tests)
    fuzzer/                    # fuzz helpers consumed by install tests
    constants.go               # shared constants (labels, finalizers, container names)
    register.go                # GroupName = "kubedb.com"
  ops/v1alpha1/                # ops.kubedb.com - *OpsRequest types per database
  autoscaling/v1alpha1/        # autoscaling.kubedb.com - *Autoscaler types
  catalog/v1alpha1/            # catalog.kubedb.com - *Version (image catalog)
  archiver/v1alpha1/           # archiver.kubedb.com - backup archivers
  schema/v1alpha1/             # schema.kubedb.com - schema-manager CRDs
  config/v1alpha1/             # config.kubedb.com - in-cluster config types
  postgres/v1alpha1/           # postgres.kubedb.com - publisher/subscriber
  elasticsearch/v1alpha1/      # elasticsearch.kubedb.com - elasticsearch dashboard
  kafka/v1alpha1/              # kafka.kubedb.com - connectors, schemaregistry, etc.
  gitops/v1alpha1/             # gitops.kubedb.com - read-only GitOps mirrors
  migrator/v1alpha1/           # migrator.kubedb.com
  ui/v1alpha1/                 # ui.kubedb.com - dashboard support types
  helpers.go                   # cross-group helpers (also helpers_test.go)

client/                        # generated typed client (do not edit)
  clientset/versioned/         # clientset + fake + scheme + typed
  informers/                   # shared informers
  listers/                     # listers

crds/                          # generated YAML CRD manifests (one per Kind, ~186 files)
openapi/swagger.json           # aggregated OpenAPI spec
hack/                          # codegen drivers and shell scripts
  gencrd/main.go               # builds openapi/swagger.json
  build.sh, test.sh, fmt.sh    # invoked by Makefile docker targets
  license/, scripts/, config/  # license headers and helper scripts
  crd-patch.json               # JSONPatch applied to large CRDs via `patch-crds`

pkg/                           # runtime helpers consumed by operators
  webhooks/                    # admission webhook handlers per API group
    kubedb/v1, kubedb/v1alpha2 # per-database validators/mutators
    autoscaling/, elasticsearch/, kafka/, ops/, schema/
  controller/                  # reusable controller plumbing (PetSet, OCM, opsrequest)
  factory/client.go            # controller-runtime client construction
  eventer/recorder.go          # event recorder
  lib/                         # cross-cutting helpers (stash, kubestash, topology, ...)
  validator/, phase/           # status phase helpers + validators
  archiver/, network_policy/   # archiver + NetworkPolicy generation
  config_generator/            # database config-file rendering
  double_optin/                # cross-namespace selector consent checks
  features/, license/          # feature gates + license verifier wiring
  utils/                       # grpc, raft, generic resource utilities
  openapi/                     # OpenAPI rendering helpers
  admission/namespace/         # legacy admission helpers
  yq3/                         # vendored mikefarah/yq v3 wrapper

third_party/protobuf/          # proto includes for go-to-protobuf
vendor/                        # `go mod vendor` is required (GOFLAGS=-mod=vendor)
.config/api-rules/             # openapi-gen violation exception list
```

## Key Packages / APIs

- `apis/kubedb/v1alpha2` - storage version for all 35 database kinds (Cassandra, ClickHouse, Druid, Elasticsearch, FerretDB, Kafka, MariaDB, MongoDB, MySQL, Postgres, Redis, etc.). Each database has `<db>_types.go` + `<db>_helpers.go`. `helpers.go` and per-DB helpers expose `SetDefaults`, `OffshootSelectors`, `StatefulSet*Name`, `ServiceName`, etc.
- `apis/kubedb/v1` - newer served version (subset of databases promoted to GA, e.g. `postgres_types.go`, `mongodb_types.go`, `mysql_types.go`, `elasticsearch_types.go`, `redis_types.go`, `kafka_types.go`, etc.). `conversion.go` and generated `zz_generated.conversion.go` bridge to `v1alpha2`.
- `apis/kubedb/install/install.go` - registers both `v1` and `v1alpha2` and sets version priority (v1 over v1alpha2). Every API group has an analogous `install/` package.
- `apis/kubedb/constants.go` - canonical constants (labels, annotations, container names, ports, sidekick names). Reused across operators.
- `apis/ops/v1alpha1` - `*OpsRequest` types for declarative day-2 operations (restart, upgrade, reconfigure, scale, volume expansion). Generated enum stringers in `*_enum.go`.
- `apis/autoscaling/v1alpha1` - per-database `*Autoscaler` + VPA checkpoint plumbing.
- `apis/catalog/v1alpha1` - `*Version` catalog (container image versions per database).
- `apis/archiver/v1alpha1` - WAL/backup archiver CRDs (MySQL, MongoDB, Postgres, MariaDB, MSSQLServer).
- `apis/schema/v1alpha1` - schema-manager CRDs that provision databases inside running clusters.
- `apis/gitops/v1alpha1` - read-only mirror types for GitOps tooling.
- `client/clientset/versioned` - generated typed clientset; use `versioned.NewForConfig(restConfig)` in downstream operators.
- `pkg/webhooks/...` - admission validators/mutators wired into operator manager startup; per-database files (e.g. `webhooks/kubedb/v1alpha2/postgres.go`).
- `pkg/factory/client.go`, `pkg/eventer/recorder.go` - shared controller-runtime client and event recorder factories.
- `pkg/lib/*` - integration shims (Stash, KubeStash, topology spread, reconfigure merging) used by ops-request controllers.
- `pkg/openapi/lib.go`, `hack/gencrd/main.go` - assemble `openapi/swagger.json`.

## Testing

- Unit tests live next to their packages. Notable suites:
  - `apis/*/install/roundtrip_test.go` + `pruning_test.go` - scheme roundtripping (uses `sigs.k8s.io/randfill` + fuzzers under `apis/*/fuzzer/`).
  - `apis/kubedb/v1/postgres_helpers_test.go`, `mongodb_helpers_test.go` and matching `v1alpha2` tests - helper invariants.
  - `apis/helpers_test.go`, `pkg/phase/phase_test.go`, `pkg/config_generator/lib_test.go`.
- Run everything: `make unit-tests`. The runner shells out to `./hack/test.sh` which forces `GOFLAGS=-mod=vendor` and `CGO_ENABLED=0`.
- CI (`.github/workflows/ci.yml`) runs `make ci` on every PR plus a `kubectl create -R -f ./crds` smoke test on KinD clusters spanning k8s `v1.29.14`, `v1.31.14`, `v1.33.7`, `v1.35.0`.

## Dependencies

### Internal (other KubeDB / AppsCode modules)

- `kmodules.xyz/client-go`, `kmodules.xyz/custom-resources`, `kmodules.xyz/monitoring-agent-api`, `kmodules.xyz/objectstore-api`, `kmodules.xyz/offshoot-api`, `kmodules.xyz/webhook-runtime`, `kmodules.xyz/resource-metadata`, `kmodules.xyz/crd-schema-fuzz`
- `kubeops.dev/petset`, `kubeops.dev/sidekick`, `kubeops.dev/operator-shard-manager`, `kubeops.dev/csi-driver-cacerts`
- `kubestash.dev/apimachinery`, `stash.appscode.dev/apimachinery` - backup integrations
- `go.bytebuilders.dev/audit`, `go.bytebuilders.dev/license-verifier/kubernetes` - audit + licensing
- `go.virtual-secrets.dev/apimachinery`
- `gomodules.xyz/*` - shared utilities (`encoding`, `pointer`, `runtime`, `stow`, `x`, `wait`)

### External

- `k8s.io/{api,apimachinery,apiserver,client-go,component-base,kube-aggregator,kube-openapi,metrics}` v0.34.3 - go 1.25
- `sigs.k8s.io/controller-runtime` v0.22.4
- `github.com/prometheus-operator/prometheus-operator/pkg/{apis/monitoring,client}` v0.87.1 - ServiceMonitor wiring
- `github.com/cert-manager/cert-manager` v1.19.4 - certificate types
- `github.com/kubernetes-csi/external-snapshotter/client/v8` v8.4.0 - VolumeSnapshot
- `open-cluster-management.io/api` v1.2.0 - OCM support via `pkg/controller/ocm`
- `go.etcd.io/etcd/...` - etcd type imports for embedded etcd database; pinned via `replace` directives (server/pkg at v3.5.27, raft replaced with `kubedb/etcd-io/raft v3.5.0-beta.4`)
- `github.com/mikefarah/yq/v3` - wrapped under `pkg/yq3` for config rendering

## Code Conventions

- API directories use the standard Kubernetes layout: `register.go` (GroupVersion + `AddToScheme`), `doc.go` (`+k8s:deepcopy-gen` / `+k8s:conversion-gen` / `+k8s:openapi-gen` / `+groupName=` markers), `<kind>_types.go` (struct + kubebuilder markers), `<kind>_helpers.go` (methods on the type), and generated `zz_generated.deepcopy.go` / `openapi_generated.go` / `zz_generated.conversion.go`.
- Per-database files are split: every database has a matching `<db>_types.go` + `<db>_helpers.go` pair across `kubedb/`, `ops/`, `autoscaling/`, `catalog/`, etc. When adding a database, mirror the pattern across each group.
- Group name constants are in `apis/<group>/register.go` (e.g. `kubedb.GroupName = "kubedb.com"`). Reuse them; do not hard-code group strings.
- Shared label/annotation/container-name constants belong in `apis/kubedb/constants.go`, not in helper files.
- Generated files (`zz_generated.*.go`, `openapi_generated.go`, `generated.pb.go`, `crds/*.yaml`, `openapi/swagger.json`, `client/`) must be regenerated via `make gen`, never hand-edited.
- License header in `hack/license/go.txt` is enforced by `make check-license`. New files must start with this header (use `make add-license`).
- The linter (`.golangci.yml`) rewrites `interface{}` to `any` on format.
- `make verify` must pass: `go mod tidy && go mod vendor` cleanly, and `make gen && make fmt` produces no diff.
- Build/test scripts hard-code `GOFLAGS=-mod=vendor` - run `go mod vendor` after touching `go.mod`.

## Common Mistakes to Avoid

- Don't hand-edit anything under `client/`, `crds/`, `openapi/`, or any `zz_generated.*` / `openapi_generated.go` / `generated.pb.go` file - rerun `make gen`.
- Don't add a new database without updating all five tiers: `apis/kubedb/v1alpha2`, `apis/catalog/v1alpha1`, `apis/ops/v1alpha1`, `apis/autoscaling/v1alpha1`, and (where applicable) `apis/archiver/v1alpha1` + `apis/gitops/v1alpha1`. Also extend the relevant `pkg/webhooks/` package.
- Don't bypass `make verify` - generated code drifting from sources will fail CI.
- Don't bump `go.etcd.io/etcd/...` without checking the `replace` block at the bottom of `go.mod`; pins exist because of the embedded-etcd database type.
- Don't add CRD YAML files manually - controller-gen reads kubebuilder markers from `apis/...` and writes `crds/*.yaml`; large CRDs are post-processed via `hack/crd-patch.json` (see `crd_to_patch` list in the `Makefile`).
- Don't run `go build ./...` directly without `-mod=vendor` when verifying changes that other Make targets touch; mismatched module mode can cause spurious errors.
