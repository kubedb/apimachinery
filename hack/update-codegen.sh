#!/usr/bin/env bash

# Copyright AppsCode Inc. and Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Regenerates the deepcopy/conversion helpers under ./apis and the
# clientset/listers/informers under ./client, replacing the old
# generate-groups.sh based "clientset" target and the ad hoc conversion-gen
# invocation in "gen-conversion", using the generator binaries from
# k8s.io/code-generator's kube_codegen.sh toolchain
# (https://github.com/kubernetes/code-generator/blob/master/kube_codegen.sh).
#
# This deliberately does NOT call kube_codegen.sh's gen_helpers/gen_client
# wrapper functions: those auto-discover their scope by grepping the whole
# input tree for +k8s:deepcopy-gen / +k8s:defaulter-gen / +k8s:conversion-gen
# / +genclient markers, which doesn't match this repo's actual, narrower
# generation footprint --
#   - every apis/<group>/<version> package carries a +k8s:defaulter-gen
#     marker, but defaulter-gen has never actually been run here (no
#     zz_generated.defaults.go exists in the repo);
#   - apis/kubedb/v1alpha1 and apis/{autoscaling,ops,catalog,config}/v1alpha1
#     all carry +k8s:conversion-gen markers, but only apis/kubedb/v1alpha2
#     has ever actually had conversion-gen run against it;
#   - apis/config/v1alpha1 and apis/kubedb/v1alpha1 have zero +genclient
#     types, so gen_client's auto-discovery would silently drop their (all
#     but empty) typed clientset packages.
# so the generator binaries are invoked directly instead, with the exact
# same $(API_GROUPS)-driven scope the old Makefile targets used, to keep
# generated-code output unchanged from before this migration.
#
# This is meant to be run inside the gengo-builder image
# (https://github.com/appscodelabs/gengo-builder), which clones
# k8s.io/code-generator (as the kmodules/code-generator fork) to
# /go/src/k8s.io/code-generator. Run it via `make update-codegen`.

set -o errexit
set -o nounset
set -o pipefail

# deepcopy-gen (as shipped by k8s.io/code-generator) fatals on plain Go type
# aliases (`type X = Y`, e.g. apis/ui/v1alpha1's
# `MariaDBSchemaOverviewSpec = GenericSchemaOverviewSpec`) once go/types
# represents them as *types.Alias nodes, which is opt-in prior to Go 1.24 and
# the default from Go 1.24 on. Forcing the pre-1.22 substitution behavior
# here sidesteps the crash. See https://go.dev/issue/69772.
export GODEBUG="${GODEBUG:+${GODEBUG},}gotypesalias=0"

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
CODEGEN_PKG="${CODEGEN_PKG:-/go/src/k8s.io/code-generator}"
THIS_PKG="kubedb.dev/apimachinery"
BOILERPLATE="${SCRIPT_ROOT}/hack/license/go.txt"

# Same "group:v1,v2 ..." format the old generate-groups.sh took; defaults to
# the Makefile's $(API_GROUPS) (passed through as an env var by `make
# update-codegen`/`make verify-codegen`) so the scope isn't duplicated
# between the Makefile and this script.
API_GROUPS="${API_GROUPS:?API_GROUPS must be set, e.g. from the Makefile \$(API_GROUPS) variable}"

# Expand "group:v1,v2 group2:v3 ..." into a "group/version" array, e.g.
# "kubedb:v1alpha1,v1alpha2" -> kubedb/v1alpha1 kubedb/v1alpha2.
group_versions=()
for group_and_versions in ${API_GROUPS}; do
    group="${group_and_versions%%:*}"
    IFS=',' read -r -a versions <<<"${group_and_versions#*:}"
    for version in "${versions[@]}"; do
        group_versions+=("${group}/${version}")
    done
done

input_pkgs=()
for gv in "${group_versions[@]}"; do
    input_pkgs+=("${THIS_PKG}/apis/${gv}")
done

# Install the generator binaries the same way kube_codegen.sh does: cd into
# the code-generator module and `go install` fully-qualified package names,
# so they resolve against the module already on disk instead of as an
# out-of-module dependency.
(
    cd "${CODEGEN_PKG}"
    GO111MODULE=on go install \
        k8s.io/code-generator/cmd/deepcopy-gen \
        k8s.io/code-generator/cmd/conversion-gen \
        k8s.io/code-generator/cmd/client-gen \
        k8s.io/code-generator/cmd/lister-gen \
        k8s.io/code-generator/cmd/informer-gen
)
GOBIN="${GOBIN:-$(go env GOPATH)/bin}"

# Deepcopy helpers for every group/version in $(API_GROUPS).
echo "Generating deepcopy code for ${#input_pkgs[@]} targets"
find "${SCRIPT_ROOT}/apis" -name zz_generated.deepcopy.go -delete
"${GOBIN}/deepcopy-gen" \
    --output-file zz_generated.deepcopy.go \
    --go-header-file "${BOILERPLATE}" \
    "${input_pkgs[@]}"

# Conversion helpers between apis/kubedb/v1alpha2 and apis/kubedb/v1, the
# same scope the old "gen-conversion" Makefile target used. (As noted above,
# apis/kubedb/v1alpha1 and apis/{autoscaling,ops,catalog,config}/v1alpha1
# also carry +k8s:conversion-gen markers, but were never wired into any
# Makefile target and are left alone here too.)
echo "Generating conversion code for apis/kubedb/v1alpha2"
"${GOBIN}/conversion-gen" \
    --go-header-file "${BOILERPLATE}" \
    --extra-peer-dirs "kmodules.xyz/monitoring-agent-api/api/v1" \
    --output-file zz_generated.conversion.go \
    "${THIS_PKG}/apis/kubedb/v1alpha2"

# Typed clientset, listers and informers for every group/version in
# $(API_GROUPS) -- including apis/config/v1alpha1 and apis/kubedb/v1alpha1,
# which have no +genclient resource types of their own but still get a
# (near-empty) typed client, matching today's output.
echo "Generating client code for ${#group_versions[@]} targets"
inputs=()
for gv in "${group_versions[@]}"; do
    inputs+=(--input "${gv}")
done
find "${SCRIPT_ROOT}/client/clientset" -name '*.go' -exec grep -l '^// Code generated by client-gen. DO NOT EDIT.$' {} + 2>/dev/null | xargs -r rm -f
"${GOBIN}/client-gen" \
    --go-header-file "${BOILERPLATE}" \
    --output-dir "${SCRIPT_ROOT}/client/clientset" \
    --output-pkg "${THIS_PKG}/client/clientset" \
    --clientset-name versioned \
    --input-base "${SCRIPT_ROOT}/apis" \
    "${inputs[@]}"

echo "Generating lister code for ${#input_pkgs[@]} targets"
find "${SCRIPT_ROOT}/client/listers" -name '*.go' -exec grep -l '^// Code generated by lister-gen. DO NOT EDIT.$' {} + 2>/dev/null | xargs -r rm -f
"${GOBIN}/lister-gen" \
    --go-header-file "${BOILERPLATE}" \
    --output-dir "${SCRIPT_ROOT}/client/listers" \
    --output-pkg "${THIS_PKG}/client/listers" \
    "${input_pkgs[@]}"

echo "Generating informer code for ${#input_pkgs[@]} targets"
find "${SCRIPT_ROOT}/client/informers" -name '*.go' -exec grep -l '^// Code generated by informer-gen. DO NOT EDIT.$' {} + 2>/dev/null | xargs -r rm -f
"${GOBIN}/informer-gen" \
    --go-header-file "${BOILERPLATE}" \
    --output-dir "${SCRIPT_ROOT}/client/informers" \
    --output-pkg "${THIS_PKG}/client/informers" \
    --versioned-clientset-package "${THIS_PKG}/client/clientset/versioned" \
    --listers-package "${THIS_PKG}/client/listers" \
    "${input_pkgs[@]}"
