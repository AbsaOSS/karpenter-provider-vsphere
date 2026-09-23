package testutil

import (
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// MockKwokInstanceTypesSlice returns a slice of mock KWOK instance types.
func MockKwokInstanceTypesSlice() []*corecloudprovider.InstanceType {
	return []*corecloudprovider.InstanceType{
		MockKwokInstanceType(),
	}
}

// NewMockKwokProvider creates a mock KWOK instance types provider.
func NewMockKwokProvider(instances []*corecloudprovider.InstanceType) *MockKwokInstanceTypesProvider {
	return NewMockKwokInstanceTypesProvider(instances)
}

// DefaultMockKwokProvider creates a mock KWOK provider with default instance types.
func DefaultMockKwokProvider() *MockKwokInstanceTypesProvider {
	return NewMockKwokProvider(MockKwokInstanceTypesSlice())
}
