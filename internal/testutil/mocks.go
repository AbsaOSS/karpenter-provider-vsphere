package testutil

import (
	"context"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// MockKwokInstanceTypesProvider is a mock implementation of KwokInstanceTypesProvider for testing.
type MockKwokInstanceTypesProvider struct {
	instances []*corecloudprovider.InstanceType
}

// List returns the mock instance types with kubelet config applied.
func (m *MockKwokInstanceTypesProvider) List(ctx context.Context, diskSize int64, kubeletConfig *v1alpha1.KubeletConfiguration) ([]*corecloudprovider.InstanceType, error) {
	return m.instances, nil
}

// NewMockKwokInstanceTypesProvider creates a new mock provider with the given instance types.
func NewMockKwokInstanceTypesProvider(instances []*corecloudprovider.InstanceType) *MockKwokInstanceTypesProvider {
	return &MockKwokInstanceTypesProvider{
		instances: instances,
	}
}
