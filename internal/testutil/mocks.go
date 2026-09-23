package testutil

import (
	"context"

	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// MockKwokInstanceTypesProvider is a mock implementation of KwokInstanceTypesProvider for testing.
type MockKwokInstanceTypesProvider struct {
	instances []*corecloudprovider.InstanceType
}

// List returns the mock instance types.
func (m *MockKwokInstanceTypesProvider) List(ctx context.Context, diskSize int64) ([]*corecloudprovider.InstanceType, error) {
	return m.instances, nil
}

// NewMockKwokInstanceTypesProvider creates a new mock provider with the given instance types.
func NewMockKwokInstanceTypesProvider(instances []*corecloudprovider.InstanceType) *MockKwokInstanceTypesProvider {
	return &MockKwokInstanceTypesProvider{
		instances: instances,
	}
}
