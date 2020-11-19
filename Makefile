# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL:=help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

# This option is for running docker manifest command
export DOCKER_CLI_EXPERIMENTAL := enabled

# Directories.
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin


# Binaries.
# Need to use abspath so we can invoke these from subdirectories
KUSTOMIZE := $(abspath $(TOOLS_BIN_DIR)/kustomize)
CONTROLLER_GEN := $(abspath $(TOOLS_BIN_DIR)/controller-gen)
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/golangci-lint)

# Bindata.
GOBINDATA := $(abspath $(TOOLS_BIN_DIR)/go-bindata)
GOBINDATA_UNDISTRO_DIR := config
GOBINDATA_TEMPLATES_DIR := templates

GINKGO_NODES ?= 1

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif

# Set build time variables including version details
LDFLAGS := $(shell hack/version-local.sh)

all: test undistro

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: ## Run tests
	source ./scripts/fetch_ext_bins.sh; fetch_tools; setup_envs; go test -race -v ./... $(TEST_ARGS)

.PHONY: test-cover
test-cover: ## Run tests with code coverage and code generate reports
	source ./scripts/fetch_ext_bins.sh; fetch_tools; setup_envs; go test -v -coverprofile=out/coverage.out ./... $(TEST_ARGS)
	go tool cover -func=out/coverage.out -o out/coverage.txt
	go tool cover -html=out/coverage.out -o out/coverage.html

## --------------------------------------
## Binaries
## --------------------------------------

.PHONY: undistro
undistro: ## Build undistro binary
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/undistro ./cmd/undistro 

# $(KUSTOMIZE): $(TOOLS_DIR)/go.mod # Build kustomize from tools folder.
# 	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/kustomize sigs.k8s.io/kustomize/kustomize/v3

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Build controller-gen from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Build golangci-lint from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOBINDATA): $(TOOLS_DIR)/go.mod # Build go-bindata from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/go-bindata github.com/go-bindata/go-bindata/go-bindata


## --------------------------------------
## Dev Environment
## --------------------------------------

.PHONY: dev
dev: ## start dev environment
	./hack/kind.sh
	kubectl cluster-info --context kind-kind
	tilt up --host 0.0.0.0

## --------------------------------------
## Linting
## --------------------------------------

.PHONY: lint lint-full
lint: $(GOLANGCI_LINT) ## Lint codebase
	$(GOLANGCI_LINT) run -v

lint-full: $(GOLANGCI_LINT) ## Run slower linters to detect possible issues
	$(GOLANGCI_LINT) run -v --fast=false

## --------------------------------------
## Generate / Manifests
## --------------------------------------

.PHONY: generate
generate: ## Generate code
	$(MAKE) generate-manifests
	$(MAKE) generate-go-core
	$(MAKE) generate-bindata

.PHONY: generate-go-core
generate-go-core: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) \
		object:headerFile=./hack/boilerplate.go.txt \
		paths=./api/...

.PHONY: generate-bindata
generate-bindata: $(KUSTOMIZE) $(GOBINDATA) clean-bindata  ## Generate code for embedding the undistro api manifest
	# Package manifest YAML into a single file.
	mkdir -p $(GOBINDATA_UNDISTRO_DIR)/manifest/
	$(KUSTOMIZE) build $(GOBINDATA_UNDISTRO_DIR)/crd > $(GOBINDATA_UNDISTRO_DIR)/manifest/undistro-api.yaml
	# Generate go-bindata, add boilerplate, then cleanup.
	$(GOBINDATA) -mode=420 -modtime=1 -pkg=config -o=$(GOBINDATA_UNDISTRO_DIR)/zz_generated.bindata.go $(GOBINDATA_UNDISTRO_DIR)/manifest/ $(GOBINDATA_UNDISTRO_DIR)/assets
	cat ./hack/boilerplate.go.txt $(GOBINDATA_UNDISTRO_DIR)/zz_generated.bindata.go > $(GOBINDATA_UNDISTRO_DIR)/manifest/manifests.go
	cp $(GOBINDATA_UNDISTRO_DIR)/manifest/manifests.go $(GOBINDATA_UNDISTRO_DIR)/zz_generated.bindata.go
	# Cleanup the manifest folder.
	$(MAKE) clean-bindata

PHONY: generate-manifests
generate-manifests: $(CONTROLLER_GEN) ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./api/... \
		paths=./controllers/... \
		crd:crdVersions=v1 \
		rbac:roleName=manager-role \
		output:crd:dir=./config/crd/bases \
		output:webhook:dir=./config/webhook \
		webhook

# Install CRDs into a cluster
install: generate
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: generate
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	go mod tidy
	cd $(TOOLS_DIR); go mod tidy

## --------------------------------------
## Cleanup / Verification
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated files
	$(MAKE) clean-bin
	$(MAKE) clean-bindata

.PHONY: clean-bin
clean-bin: ## Remove all generated binaries
	rm -rf bin
	rm -rf hack/tools/bin

.PHONY: clean-bindata
clean-bindata: ## Remove bindata generated folder
	rm -rf $(GOBINDATA_UNDISTRO_DIR)/manifest

.PHONY: clean-manifests ## Reset manifests in config directories back to master
clean-manifests:
	@read -p "WARNING: This will reset all config directories to local master. Press [ENTER] to continue."
	git checkout master config 

.PHONY: verify
verify:
	$(MAKE) verify-modules
	$(MAKE) verify-gen

.PHONY: verify-modules
verify-modules: modules
	@if !(git diff --quiet HEAD -- go.sum go.mod hack/tools/go.mod hack/tools/go.sum); then \
		git diff; \
		echo "go module files are out of date"; exit 1; \
	fi
	@if (find . -name 'go.mod' | xargs -n1 grep -q -i 'k8s.io/client-go.*+incompatible'); then \
		find . -name "go.mod" -exec grep -i 'k8s.io/client-go.*+incompatible' {} \; -print; \
		echo "go module contains an incompatible client-go version"; exit 1; \
	fi

.PHONY: verify-gen
verify-gen: generate
	@if !(git diff --quiet HEAD); then \
		git diff; \
		echo "generated files are out of date, run make generate"; exit 1; \
	fi
