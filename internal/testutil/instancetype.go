package testutil

import (
	corev1 "k8s.io/api/core/v1"

	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// DefaultInstanceType returns an instance type with sensible defaults.
func DefaultInstanceType() *corecloudprovider.InstanceType {
	return NewInstanceType().Build()
}

// StandardInstanceType8X returns a standard s-8x instance type for testing.
func StandardInstanceType8X() *corecloudprovider.InstanceType {
	return NewInstanceType().
		WithName("s-8x").
		WithCPU("4").
		WithMemory("16Gi").
		WithPods("110").
		Build()
}

// ComputeInstanceType4X returns a standard c-4x instance type for testing.
func ComputeInstanceType4X() *corecloudprovider.InstanceType {
	return NewInstanceType().
		WithName("c-4x").
		WithCPU("2").
		WithMemory("4Gi").
		WithPods("110").
		Build()
}

// MockKwokInstanceType returns a mock KWOK instance type.
func MockKwokInstanceType() *corecloudprovider.InstanceType {
	return NewInstanceType().
		WithName("c-4x").
		WithCPU("2").
		WithMemory("4Gi").
		WithPods("110").
		WithRequirement(corev1.LabelInstanceTypeStable, corev1.NodeSelectorOpIn, "c-4x").
		WithRequirement(corev1.LabelArchStable, corev1.NodeSelectorOpIn, "amd64").
		WithRequirement(corev1.LabelOSStable, corev1.NodeSelectorOpIn, "linux").
		WithOverhead().
		WithZone("zone-a").
		WithOfferingSchedulingCapacityTypeOnDemand().
		Build()
}
