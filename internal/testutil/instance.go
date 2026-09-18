package testutil

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
)

// NodeClassInstanceTypeBuilder builds v1alpha1.InstanceType test fixtures.
// Note: This is for v1alpha1.InstanceType (used in NodeClass spec), not corecloudprovider.InstanceType
type NodeClassInstanceTypeBuilder struct {
	obj *v1alpha1.InstanceType
}

// NewNodeClassInstanceType creates a new NodeClassInstanceType builder with defaults.
func NewNodeClassInstanceType() *NodeClassInstanceTypeBuilder {
	return &NodeClassInstanceTypeBuilder{
		obj: &v1alpha1.InstanceType{
			CPU:     "4",
			Memory:  "16Gi",
			MaxPods: "110",
			Zone:    "zone-a",
			OS:      "Linux",
		},
	}
}

// WithCPU sets the CPU.
func (b *NodeClassInstanceTypeBuilder) WithCPU(cpu string) *NodeClassInstanceTypeBuilder {
	b.obj.CPU = cpu
	return b
}

// WithMemory sets the memory.
func (b *NodeClassInstanceTypeBuilder) WithMemory(memory string) *NodeClassInstanceTypeBuilder {
	b.obj.Memory = memory
	return b
}

// WithMaxPods sets the max pods.
func (b *NodeClassInstanceTypeBuilder) WithMaxPods(maxPods string) *NodeClassInstanceTypeBuilder {
	b.obj.MaxPods = maxPods
	return b
}

// WithZone sets the zone.
func (b *NodeClassInstanceTypeBuilder) WithZone(zone string) *NodeClassInstanceTypeBuilder {
	b.obj.Zone = zone
	return b
}

// WithOS sets the OS.
func (b *NodeClassInstanceTypeBuilder) WithOS(os string) *NodeClassInstanceTypeBuilder {
	b.obj.OS = os
	return b
}

// WithArch sets the architecture.
func (b *NodeClassInstanceTypeBuilder) WithArch(arch string) *NodeClassInstanceTypeBuilder {
	b.obj.Arch = arch
	return b
}

// Build returns the constructed v1alpha1.InstanceType.
func (b *NodeClassInstanceTypeBuilder) Build() *v1alpha1.InstanceType {
	return b.obj
}
