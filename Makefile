GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

TAG := `git describe --tags --always`
VERSION :=

AZURE_SUBSCRIPTION_ID ?= $(shell az account show --query id -o tsv)
AZURE_RESOURCE_GROUP ?=
AZURE_CLUSTER_NAME ?=

# Adds a '-dirty' suffix to version string if there are uncommitted changes
changes := $(shell git status --porcelain)
ifeq ($(changes),)
	VERSION := $(TAG)
else
	VERSION := $(TAG)-dirty
endif

LINTER_VERSION ?= v1.53.2

include ie.mk

LDFLAGS := "-X github.com/Azure/kubectl-aks/cmd.version=$(VERSION) -extldflags '-static'"

.DEFAULT_GOAL := kubectl-aks

# Build
KUBECTL_AKS_TARGETS = \
	kubectl-aks-linux-amd64 \
	kubectl-aks-linux-arm64 \
	kubectl-aks-darwin-amd64 \
	kubectl-aks-darwin-arm64 \
	kubectl-aks-windows-amd64

.PHONY: list-kubectl-aks-targets
list-kubectl-aks-targets:
	@echo $(KUBECTL_AKS_TARGETS)

.PHONY: kubectl-aks-all
kubectl-aks-all: $(KUBECTL_AKS_TARGETS)

.PHONY: kubectl-aks
kubectl-aks: kubectl-aks-$(GOHOSTOS)-$(GOHOSTARCH)
	mv kubectl-aks-$(GOHOSTOS)-$(GOHOSTARCH) kubectl-aks

# make does not allow implicit rules (with '%') to be phony so let's use
# the 'phony_explicit' dependency to make implicit rules inherit the phony
# attribute
.PHONY: phony_explicit
phony_explicit:

.PHONY: kubectl-aks-%
kubectl-aks-%: phony_explicit
	export GO111MODULE=on CGO_ENABLED=0 && \
	export GOOS=$(shell echo $* |cut -f1 -d-) GOARCH=$(shell echo $* |cut -f2 -d-) && \
	go build -ldflags $(LDFLAGS) \
		-o kubectl-aks-$${GOOS}-$${GOARCH} \
		github.com/Azure/kubectl-aks

# Lint
.PHONY: lint
lint:
	docker run --rm --env XDG_CACHE_HOME=/tmp/xdg_home_cache \
		--env GOLANGCI_LINT_CACHE=/tmp/golangci_lint_cache \
		--user $(shell id -u):$(shell id -g) -v $(shell pwd):/app -w /app \
		golangci/golangci-lint:$(LINTER_VERSION) golangci-lint run

# Install
.PHONY: install
install: kubectl-aks
	mkdir -p ~/.local/bin/
	cp kubectl-aks ~/.local/bin/

# Uninstall
.PHONY: uninstall
uninstall:
	rm -f ~/.local/bin/kubectl-aks

# Run unit tests
.PHONY: unit-test
unit-test:
	go test -v ./...

# Run integration tests
.PHONY: integration-test
integration-test: kubectl-aks
	KUBECTL_AKS="$(shell pwd)/kubectl-aks" \
	AZURE_SUBSCRIPTION_ID=$(AZURE_SUBSCRIPTION_ID) \
	AZURE_RESOURCE_GROUP=$(AZURE_RESOURCE_GROUP) \
	AZURE_CLUSTER_NAME=$(AZURE_CLUSTER_NAME) \
		go test -v ./test/integration/... -integration

# Run documentation tests
.PHONY: documentation-test-readme
documentation-test-readme: install-ie
	ie execute README.md

DOCUMENTATION_TEST_FILES ?= \
	./docs/run-command.md

.PHONY: documentation-test-commands
documentation-test-commands: install install-ie
	for file in $(DOCUMENTATION_TEST_FILES); do \
		ie execute $$file; \
	done

# Clean
.PHONY: clean
clean:
	rm -f kubectl-aks

.PHONY: cleanall
cleanall: clean
	rm -f $(KUBECTL_AKS_TARGETS)
