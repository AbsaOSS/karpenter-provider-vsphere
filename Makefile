KARPENTER_CORE_DIR = $(shell go list -m -f '{{ .Dir }}' sigs.k8s.io/karpenter)
BINARY_FILENAME ?= "karpenter-provider-vsphere"

LDFLAGS ?= -ldflags=-X=sigs.k8s.io/karpenter/pkg/operator.Version=$(shell git describe --tags --always | cut -d"v" -f2)

GOFLAGS ?= $(LDFLAGS)
WITH_GOFLAGS = GOFLAGS="$(GOFLAGS)"
KO_DOCKER_REPO ?= ghcr.io/absaoss/karpenter-provider-vsphere
KOCACHE ?= ~/.ko

generate:
	go generate ./...
	cp  $(KARPENTER_CORE_DIR)/pkg/apis/crds/* pkg/apis/crds

image: ## Build the Karpenter controller images using ko build
	$(eval CONTROLLER_IMG=$(shell $(WITH_GOFLAGS) KOCACHE=$(KOCACHE) KO_DOCKER_REPO="$(KO_DOCKER_REPO)" ko build --bare github.com/absaoss/karpenter-provider-vsphere/cmd/controller))
	$(eval IMG_REPOSITORY=$(shell echo $(CONTROLLER_IMG) | cut -d "@" -f 1 | cut -d ":" -f 1))
	$(eval IMG_TAG=$(shell echo $(CONTROLLER_IMG) | cut -d "@" -f 1 | cut -d ":" -f 2 -s))
	$(eval IMG_DIGEST=$(shell echo $(CONTROLLER_IMG) | cut -d "@" -f 2))

binary: ## Build the Karpenter controller binary using go build
	go build $(GOFLAGS) -o $(BINARY_FILENAME) ./cmd/controller/...
