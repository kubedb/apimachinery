# Copyright 2019 AppsCode Inc.
# Copyright 2016 The Kubernetes Authors.
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

SHELL=/bin/bash -o pipefail

GO_PKG   := kubedb.dev
REPO     := $(notdir $(shell pwd))
BIN      := apimachinery

CRD_OPTIONS          ?= "crd:maxDescLen=0,generateEmbeddedObjectMeta=true,allowDangerousTypes=true"
# https://github.com/appscodelabs/gengo-builder
CODE_GENERATOR_IMAGE ?= appscode/gengo:release-1.25
CORE_API_GROUPS      ?= kubedb:v1alpha1 kubedb:v1alpha2 postgres:v1alpha1 catalog:v1alpha1 config:v1alpha1 ops:v1alpha1 autoscaling:v1alpha1 dashboard:v1alpha1 schema:v1alpha1
API_GROUPS           ?= $(CORE_API_GROUPS) ui:v1alpha1

# This version-strategy uses git tags to set the version string
git_branch       := $(shell git rev-parse --abbrev-ref HEAD)
git_tag          := $(shell git describe --exact-match --abbrev=0 2>/dev/null || echo "")
commit_hash      := $(shell git rev-parse --verify HEAD)
commit_timestamp := $(shell date --date="@$$(git show -s --format=%ct)" --utc +%FT%T)

VERSION          := $(shell git describe --tags --always --dirty)
version_strategy := commit_hash
ifdef git_tag
	VERSION := $(git_tag)
	version_strategy := tag
else
	ifeq (,$(findstring $(git_branch),master HEAD))
		ifneq (,$(patsubst release-%,,$(git_branch)))
			VERSION := $(git_branch)
			version_strategy := branch
		endif
	endif
endif

###
### These variables should not need tweaking.
###

SRC_PKGS := apis client crds pkg # directories which hold app source (not vendored)
SRC_DIRS := $(SRC_PKGS) hack/gencrd

DOCKER_PLATFORMS := linux/amd64 linux/arm linux/arm64
BIN_PLATFORMS    := $(DOCKER_PLATFORMS) windows/amd64 darwin/amd64

# Used internally.  Users should pass GOOS and/or GOARCH.
OS   := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

BASEIMAGE_PROD   ?= gcr.io/distroless/static-debian11
BASEIMAGE_DBG    ?= debian:bullseye

GO_VERSION       ?= 1.18
BUILD_IMAGE      ?= appscode/golang-dev:$(GO_VERSION)

OUTBIN = bin/$(OS)_$(ARCH)/$(BIN)
ifeq ($(OS),windows)
  OUTBIN = bin/$(OS)_$(ARCH)/$(BIN).exe
endif

# Directories that we need created to build/test.
BUILD_DIRS  := bin/$(OS)_$(ARCH)     \
               .go/bin/$(OS)_$(ARCH) \
               .go/cache             \
               hack/config           \
               $(HOME)/.credentials  \
               $(HOME)/.kube         \
               $(HOME)/.minikube

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: fmt build

# For the following OS/ARCH expansions, we transform OS/ARCH into OS_ARCH
# because make pattern rules don't match with embedded '/' characters.

build-%:
	@$(MAKE) build                        \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

all-build: $(addprefix build-, $(subst /,_, $(BIN_PLATFORMS)))

version:
	@echo version=$(VERSION)
	@echo version_strategy=$(version_strategy)
	@echo git_tag=$(git_tag)
	@echo git_branch=$(git_branch)
	@echo commit_hash=$(commit_hash)
	@echo commit_timestamp=$(commit_timestamp)

DOCKER_REPO_ROOT := /go/src/$(GO_PKG)/$(REPO)

# Generate a typed clientset
.PHONY: clientset
clientset:
	@docker run --rm	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		/go/src/k8s.io/code-generator/generate-groups.sh \
			all                                          \
			$(GO_PKG)/$(REPO)/client                     \
			$(GO_PKG)/$(REPO)/apis                       \
			"$(API_GROUPS)" \
			--go-header-file "./hack/license/go.txt"

.PHONY: gen-conversion
gen-conversion:
	rm -rf ./apis/kubedb/v1alpha1/zz_generated.conversion.go
	@docker run --rm                                   \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		/go/bin/conversion-gen --go-header-file ./hack/license/go.txt \
			--input-dirs $(GO_PKG)/$(REPO)/apis/kubedb/v1alpha1 \
			--extra-peer-dirs "kmodules.xyz/monitoring-agent-api/api/v1alpha1" \
			-O zz_generated.conversion

# Generate openapi schema
.PHONY: openapi
openapi: $(addprefix openapi-, $(subst :,_, $(API_GROUPS)))
	@echo "Generating openapi/swagger.json"
	@docker run --rm	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		--env GO111MODULE=on                             \
		--env GOFLAGS="-mod=vendor"                      \
		$(BUILD_IMAGE)                                   \
		go run hack/gencrd/main.go

openapi-%:
	@echo "Generating openapi schema for $(subst _,/,$*)"
	@mkdir -p .config/api-rules
	@docker run --rm	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		openapi-gen                                      \
			--v 1 --logtostderr                          \
			--go-header-file "./hack/license/go.txt" \
			--input-dirs "$(GO_PKG)/$(REPO)/apis/$(subst _,/,$*),k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1,k8s.io/api/apps/v1,kmodules.xyz/offshoot-api/api/v1,kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1,kmodules.xyz/monitoring-agent-api/api/v1,k8s.io/api/rbac/v1,k8s.io/api/autoscaling/v2beta2,kmodules.xyz/objectstore-api/api/v1,kmodules.xyz/client-go/api/v1" \
			--output-package "$(GO_PKG)/$(REPO)/apis/$(subst _,/,$*)" \
			--report-filename .config/api-rules/violation_exceptions.list

# Generate CRD manifests
.PHONY: gen-crds
gen-crds:
	@echo "Generating CRD manifests"
	@docker run --rm	                    \
		-u $$(id -u):$$(id -g)              \
		-v /tmp:/.cache                     \
		-v $$(pwd):$(DOCKER_REPO_ROOT)      \
		-w $(DOCKER_REPO_ROOT)              \
	    --env HTTP_PROXY=$(HTTP_PROXY)      \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)    \
		$(CODE_GENERATOR_IMAGE)             \
		controller-gen                      \
			$(CRD_OPTIONS)                  \
			paths="./apis/..."              \
			output:crd:artifacts:config=crds

crds_to_patch := kubedb.com_elasticsearches.yaml \
					kubedb.com_etcds.yaml \
					kubedb.com_mariadbs.yaml \
					kubedb.com_memcacheds.yaml \
					kubedb.com_mongodbs.yaml \
					kubedb.com_mysqls.yaml \
					kubedb.com_perconaxtradbs.yaml \
					kubedb.com_pgbouncers.yaml \
					kubedb.com_postgreses.yaml \
					kubedb.com_proxysqls.yaml \
					kubedb.com_redises.yaml

.PHONY: patch-crds
patch-crds: $(addprefix patch-crd-, $(crds_to_patch))
patch-crd-%: $(BUILD_DIRS)
	@echo "patching $*"
	@kubectl patch -f crds/$* -p "$$(cat hack/crd-patch.json)" --type=json --local=true -o yaml > bin/$*
	@mv bin/$* crds/$*

.PHONY: label-crds
label-crds: $(BUILD_DIRS)
	@for f in crds/*.yaml; do \
		echo "applying app.kubernetes.io/name=kubedb label to $$f"; \
		kubectl label --overwrite -f $$f --local=true -o yaml app.kubernetes.io/name=kubedb > bin/crd.yaml; \
		mv bin/crd.yaml $$f; \
	done

.PHONY: gen-crd-protos
gen-crd-protos: $(addprefix gen-crd-protos-, $(subst :,_, $(CORE_API_GROUPS))) gen-crd-protos-ui-v1alpha1
	@rm -rf apis/kubedb/v1alpha1/generated.pb.go
	@rm -rf apis/kubedb/v1alpha1/generated.proto
	@rm -rf vendor/sigs.k8s.io/controller-runtime/pkg/scheme/generated.pb.go
	@rm -rf vendor/sigs.k8s.io/controller-runtime/pkg/scheme/generated.proto

gen-crd-protos-%:
	@echo "Generating protobuf for $(subst _,/,$*)"
	@docker run --rm                                     \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		go-to-protobuf                                   \
			--go-header-file "./hack/license/go.txt"     \
			--proto-import=$(DOCKER_REPO_ROOT)/vendor    \
			--proto-import=$(DOCKER_REPO_ROOT)/third_party/protobuf \
			--apimachinery-packages=-k8s.io/apimachinery/pkg/api/resource,-k8s.io/apimachinery/pkg/apis/meta/v1,-k8s.io/apimachinery/pkg/apis/meta/v1beta1,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/util/intstr \
			--packages=+sigs.k8s.io/controller-runtime/pkg/scheme,-k8s.io/api/core/v1,-k8s.io/api/apps/v1,-k8s.io/api/autoscaling/v2beta2,-kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1,-kmodules.xyz/monitoring-agent-api/api/v1,-kmodules.xyz/objectstore-api/api/v1,-kmodules.xyz/offshoot-api/api/v1,-kmodules.xyz/client-go/api/v1,kubedb.dev/apimachinery/apis/$(subst _,/,$*)

.PHONY: gen-crd-protos-ui-v1alpha1
gen-crd-protos-ui-v1alpha1:
	@echo "Generating protobuf for ui/v1alpha1"
	@docker run --rm                                     \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		go-to-protobuf                                   \
			--go-header-file "./hack/license/go.txt"     \
			--proto-import=$(DOCKER_REPO_ROOT)/vendor    \
			--proto-import=$(DOCKER_REPO_ROOT)/third_party/protobuf \
			--apimachinery-packages=-k8s.io/apimachinery/pkg/api/resource,-k8s.io/apimachinery/pkg/apis/meta/v1,-k8s.io/apimachinery/pkg/apis/meta/v1beta1,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/util/intstr \
			--packages=+sigs.k8s.io/controller-runtime/pkg/scheme,-k8s.io/api/core/v1,-k8s.io/api/apps/v1,-k8s.io/api/autoscaling/v2beta2,-kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1,-kmodules.xyz/monitoring-agent-api/api/v1,-kmodules.xyz/objectstore-api/api/v1,-kmodules.xyz/offshoot-api/api/v1,-kmodules.xyz/client-go/api/v1,kubedb.dev/apimachinery/apis/ui/v1alpha1,kubedb.dev/apimachinery/apis/kubedb/v1alpha2

.PHONY: manifests
manifests: gen-crds patch-crds label-crds

.PHONY: gen
gen: clientset manifests # openapi gen-crd-protos

fmt: $(BUILD_DIRS)
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        REPO_PKG=$(GO_PKG)                                  \
	        ./hack/fmt.sh $(SRC_DIRS)                           \
	    "

build: $(OUTBIN)

.PHONY: .go/$(OUTBIN)
$(OUTBIN): $(BUILD_DIRS)
	@echo "making $(OUTBIN)"
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        version_strategy=$(version_strategy)                \
	        git_branch=$(git_branch)                            \
	        git_tag=$(git_tag)                                  \
	        commit_hash=$(commit_hash)                          \
	        commit_timestamp=$(commit_timestamp)                \
	        ./hack/build.sh                                     \
	    "
	@echo

.PHONY: test
test: unit-tests

unit-tests: $(BUILD_DIRS)
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        ./hack/test.sh $(SRC_PKGS)                          \
	    "

ADDTL_LINTERS   := goconst,gofmt,goimports,unparam

.PHONY: lint
lint: $(BUILD_DIRS)
	@echo "running linter"
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    --env GO111MODULE=on                                    \
	    --env GOFLAGS="-mod=vendor"                             \
	    $(BUILD_IMAGE)                                          \
	    golangci-lint run --enable $(ADDTL_LINTERS) --timeout=10m --skip-files="generated.*\.go$\" --skip-dirs-use-default --skip-dirs=client,vendor

$(BUILD_DIRS):
	@mkdir -p $@

.PHONY: dev
dev: gen fmt push

.PHONY: verify
verify: verify-gen verify-modules

.PHONY: verify-modules
verify-modules:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
	@if !(git diff --exit-code HEAD); then \
		echo "go module files are out of date"; exit 1; \
	fi

.PHONY: verify-gen
verify-gen: gen fmt
	@if !(git diff --exit-code HEAD); then \
		echo "generated files are out of date, run make gen"; exit 1; \
	fi

.PHONY: add-license
add-license:
	@echo "Adding license header"
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(BUILD_IMAGE)                                   \
		ltag -t "./hack/license" --excludes "vendor contrib libbuild" -v

.PHONY: check-license
check-license:
	@echo "Checking files for license header"
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(BUILD_IMAGE)                                   \
		ltag -t "./hack/license" --excludes "vendor contrib libbuild" --check -v

.PHONY: ci
ci: verify check-license lint build unit-tests #cover

.PHONY: clean
clean:
	rm -rf .go bin
