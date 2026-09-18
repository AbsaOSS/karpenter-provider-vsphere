package testutil

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

// NodeClassBuilder builds VsphereNodeClass test fixtures with fluent API.
type NodeClassBuilder struct {
	obj *v1alpha1.VsphereNodeClass
}

// NewNodeClass creates a new NodeClass builder with sensible defaults.
func NewNodeClass() *NodeClassBuilder {
	return &NodeClassBuilder{
		obj: &v1alpha1.VsphereNodeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default",
				Namespace: "default",
			},
			Spec: v1alpha1.VsphereNodeClassSpec{
				DiskSize: 100,
				InstanceTypes: []v1alpha1.InstanceType{
					{
						CPU:     "4",
						Memory:  "16Gi",
						MaxPods: "110",
						OS:      "Linux",
						Zone:    "zone-a",
					},
				},
			},
		},
	}
}

// WithName sets the NodeClass name.
func (b *NodeClassBuilder) WithName(name string) *NodeClassBuilder {
	b.obj.Name = name
	return b
}

// WithNamespace sets the NodeClass namespace.
func (b *NodeClassBuilder) WithNamespace(namespace string) *NodeClassBuilder {
	b.obj.Namespace = namespace
	return b
}

// WithDiskSize sets the disk size in GB.
func (b *NodeClassBuilder) WithDiskSize(sizeGB int64) *NodeClassBuilder {
	b.obj.Spec.DiskSize = sizeGB
	return b
}

// WithInstanceType adds an instance type to the NodeClass.
func (b *NodeClassBuilder) WithInstanceType(cpu, memory string, maxPods int, zone, os string) *NodeClassBuilder {
	b.obj.Spec.InstanceTypes = append(b.obj.Spec.InstanceTypes, v1alpha1.InstanceType{
		CPU:     cpu,
		Memory:  memory,
		MaxPods: fmt.Sprintf("%d", maxPods),
		Zone:    zone,
		OS:      os,
	})
	return b
}

// WithInstanceTypes sets the instance types (replaces existing).
func (b *NodeClassBuilder) WithInstanceTypes(types []v1alpha1.InstanceType) *NodeClassBuilder {
	b.obj.Spec.InstanceTypes = types
	return b
}

// WithSingleInstanceType sets a single instance type.
func (b *NodeClassBuilder) WithSingleInstanceType(cpu, memory string, maxPods int, zone, os string) *NodeClassBuilder {
	b.obj.Spec.InstanceTypes = []v1alpha1.InstanceType{
		{
			CPU:     cpu,
			Memory:  memory,
			MaxPods: fmt.Sprintf("%d", maxPods),
			Zone:    zone,
			OS:      os,
		},
	}
	return b
}

// WithCPU sets the CPU on the first instance type.
func (b *NodeClassBuilder) WithCPU(cpu string) *NodeClassBuilder {
	if len(b.obj.Spec.InstanceTypes) > 0 {
		b.obj.Spec.InstanceTypes[0].CPU = cpu
	}
	return b
}

// WithMemory sets the memory on the first instance type.
func (b *NodeClassBuilder) WithMemory(memory string) *NodeClassBuilder {
	if len(b.obj.Spec.InstanceTypes) > 0 {
		b.obj.Spec.InstanceTypes[0].Memory = memory
	}
	return b
}

// Build returns the constructed NodeClass.
func (b *NodeClassBuilder) Build() *v1alpha1.VsphereNodeClass {
	return b.obj
}

// NodeClaimBuilder builds NodeClaim test fixtures with fluent API.
type NodeClaimBuilder struct {
	obj *karpv1.NodeClaim
}

// NewNodeClaim creates a new NodeClaim builder with sensible defaults.
func NewNodeClaim() *NodeClaimBuilder {
	return &NodeClaimBuilder{
		obj: &karpv1.NodeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-claim",
				Namespace: "default",
			},
			Spec: karpv1.NodeClaimSpec{},
		},
	}
}

// WithName sets the NodeClaim name.
func (b *NodeClaimBuilder) WithName(name string) *NodeClaimBuilder {
	b.obj.Name = name
	return b
}

// WithNamespace sets the NodeClaim namespace.
func (b *NodeClaimBuilder) WithNamespace(namespace string) *NodeClaimBuilder {
	b.obj.Namespace = namespace
	return b
}

// WithRequirements adds requirements to the NodeClaim.
func (b *NodeClaimBuilder) WithRequirements(requirements map[string][]string) *NodeClaimBuilder {
	// Note: NodeClaimSpec.Requirements handling depends on the karpenter version
	// For now, this method is a placeholder for future compatibility
	return b
}

// Build returns the constructed NodeClaim.
func (b *NodeClaimBuilder) Build() *karpv1.NodeClaim {
	return b.obj
}

// InstanceTypeBuilder builds corecloudprovider.InstanceType test fixtures.
type InstanceTypeBuilder struct {
	obj *corecloudprovider.InstanceType
}

// NewInstanceType creates a new InstanceType builder with sensible defaults.
func NewInstanceType() *InstanceTypeBuilder {
	return &InstanceTypeBuilder{
		obj: &corecloudprovider.InstanceType{
			Name: "default-instance",
			Requirements: scheduling.NewRequirements(
				scheduling.NewRequirement(
					corev1.LabelArchStable,
					corev1.NodeSelectorOpIn,
					"amd64",
				),
				scheduling.NewRequirement(
					corev1.LabelOSStable,
					corev1.NodeSelectorOpIn,
					"linux",
				),
			),
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("4"),
				corev1.ResourceMemory:           resource.MustParse("16Gi"),
				corev1.ResourcePods:             resource.MustParse("110"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
			Overhead: &corecloudprovider.InstanceTypeOverhead{},
		},
	}
}

// WithName sets the instance type name.
func (b *InstanceTypeBuilder) WithName(name string) *InstanceTypeBuilder {
	b.obj.Name = name
	return b
}

// WithCPU sets the CPU capacity.
func (b *InstanceTypeBuilder) WithCPU(cpu string) *InstanceTypeBuilder {
	b.obj.Capacity[corev1.ResourceCPU] = resource.MustParse(cpu)
	return b
}

// WithMemory sets the memory capacity.
func (b *InstanceTypeBuilder) WithMemory(memory string) *InstanceTypeBuilder {
	b.obj.Capacity[corev1.ResourceMemory] = resource.MustParse(memory)
	return b
}

// WithPods sets the maximum pods capacity.
func (b *InstanceTypeBuilder) WithPods(pods string) *InstanceTypeBuilder {
	b.obj.Capacity[corev1.ResourcePods] = resource.MustParse(pods)
	return b
}

// WithEphemeralStorage sets the ephemeral storage capacity.
func (b *InstanceTypeBuilder) WithEphemeralStorage(storage string) *InstanceTypeBuilder {
	b.obj.Capacity[corev1.ResourceEphemeralStorage] = resource.MustParse(storage)
	return b
}

// WithRequirement adds a requirement to the instance type.
func (b *InstanceTypeBuilder) WithRequirement(key string, operator corev1.NodeSelectorOperator, values ...string) *InstanceTypeBuilder {
	// Add the requirement to the instance type requirements map
	newReq := scheduling.NewRequirement(key, operator, values...)
	b.obj.Requirements[newReq.Key] = newReq
	return b
}

// WithOverhead sets empty overhead for the instance type.
func (b *InstanceTypeBuilder) WithOverhead() *InstanceTypeBuilder {
	b.obj.Overhead = &corecloudprovider.InstanceTypeOverhead{}
	return b
}

// WithZone sets the zone offering requirement.
// If offerings don't exist, creates a new offering. If offerings exist, adds zone to the last offering.
func (b *InstanceTypeBuilder) WithZone(zone string) *InstanceTypeBuilder {
	if b.obj.Offerings == nil {
		b.obj.Offerings = make([]*corecloudprovider.Offering, 0)
	}

	if len(b.obj.Offerings) == 0 {
		// Create new offering with zone
		offering := &corecloudprovider.Offering{
			Available: true,
			Requirements: scheduling.NewRequirements(
				scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn, zone),
			),
		}
		b.obj.Offerings = append(b.obj.Offerings, offering)
	} else {
		// Adjust existing last offering
		lastOffering := b.obj.Offerings[len(b.obj.Offerings)-1]
		newReq := scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn, zone)
		lastOffering.Requirements[newReq.Key] = newReq
	}
	return b
}

// WithOfferingSchedulingRequirement adds a scheduling requirement to the last offering.
// If offerings don't exist, creates a new empty offering first.
func (b *InstanceTypeBuilder) WithOfferingSchedulingRequirement(key string, operator corev1.NodeSelectorOperator, values ...string) *InstanceTypeBuilder {
	if b.obj.Offerings == nil {
		b.obj.Offerings = make([]*corecloudprovider.Offering, 0)
	}

	if len(b.obj.Offerings) == 0 {
		// Create new empty offering if none exist
		offering := &corecloudprovider.Offering{
			Available:    true,
			Requirements: scheduling.NewRequirements(),
		}
		b.obj.Offerings = append(b.obj.Offerings, offering)
	}

	// Add requirement to the last offering
	offering := b.obj.Offerings[len(b.obj.Offerings)-1]
	newReq := scheduling.NewRequirement(key, operator, values...)
	offering.Requirements[newReq.Key] = newReq

	return b
}

// WithOfferingSchedulingCapacityTypeOnDemand adds an on-demand capacity type requirement to the last offering.
// This is a convenience method that calls WithOfferingSchedulingRequirement with capacity-type=on-demand.
func (b *InstanceTypeBuilder) WithOfferingSchedulingCapacityTypeOnDemand() *InstanceTypeBuilder {
	return b.WithOfferingSchedulingRequirement(karpv1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, karpv1.CapacityTypeOnDemand)
}

// WithPrice sets the price on the last offering (must be called after WithZone).
func (b *InstanceTypeBuilder) WithPrice(price float64) *InstanceTypeBuilder {
	if len(b.obj.Offerings) > 0 {
		b.obj.Offerings[len(b.obj.Offerings)-1].Price = price
	}
	return b
}

// Build returns the constructed InstanceType.
func (b *InstanceTypeBuilder) Build() *corecloudprovider.InstanceType {
	return b.obj
}

// ToBuilder converts an existing InstanceType into a builder for modification.
func ToBuilder(it *corecloudprovider.InstanceType) *InstanceTypeBuilder {
	return &InstanceTypeBuilder{
		obj: it,
	}
}
