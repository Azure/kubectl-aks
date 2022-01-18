GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

TAG := `git describe --tags --always`
VERSION :=

# Adds a '-dirty' suffix to version string if there are uncommitted changes
changes := $(shell git status --porcelain)
ifeq ($(changes),)
	VERSION := $(TAG)
else
	VERSION := $(TAG)-dirty
endif

LDFLAGS := "-X github.com/Azure/kubectl-az/cmd.version=$(VERSION) -extldflags '-static'"

.DEFAULT_GOAL := kubectl-az

KUBECTL_AZ_TARGETS = \
	kubectl-az-linux-amd64 \
	kubectl-az-linux-arm64 \
	kubectl-az-darwin-amd64 \
	kubectl-az-darwin-arm64 \
	kubectl-az-windows-amd64

.PHONY: list-kubectl-az-targets
list-kubectl-az-targets:
	@echo $(KUBECTL_AZ_TARGETS)

.PHONY: kubectl-az-all
kubectl-az-all: $(KUBECTL_AZ_TARGETS)

.PHONY: kubectl-az
kubectl-az: kubectl-az-$(GOHOSTOS)-$(GOHOSTARCH)
	mv kubectl-az-$(GOHOSTOS)-$(GOHOSTARCH) kubectl-az

# make does not allow implicit rules (with '%') to be phony so let's use
# the 'phony_explicit' dependency to make implicit rules inherit the phony
# attribute
.PHONY: phony_explicit
phony_explicit:

.PHONY: kubectl-az-%
kubectl-az-%: phony_explicit
	export GO111MODULE=on CGO_ENABLED=0 && \
	export GOOS=$(shell echo $* |cut -f1 -d-) GOARCH=$(shell echo $* |cut -f2 -d-) && \
	go build -ldflags $(LDFLAGS) \
		-o kubectl-az-$${GOOS}-$${GOARCH} \
		github.com/Azure/kubectl-az

.PHONY: install
install: kubectl-az
	mkdir -p ~/.local/bin/
	cp kubectl-az ~/.local/bin/

.PHONY: clean
clean:
	rm -f kubectl-az

.PHONY: cleanall
cleanall: clean
	rm -f $(KUBECTL_AZ_TARGETS)
