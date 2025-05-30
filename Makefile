KARPENTER_CORE_DIR = $(shell go list -m -f '{{ .Dir }}' sigs.k8s.io/karpenter)

generate:
	go generate ./...
	cp  $(KARPENTER_CORE_DIR)/pkg/apis/crds/* pkg/apis/crds
