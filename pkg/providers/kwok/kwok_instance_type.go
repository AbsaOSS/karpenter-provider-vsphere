package kwok

import (
	"context"
	"fmt"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

type instanceFamily string

const (
	InstanceFamilyCompute instanceFamily = "c"
	InstanceFamilyStorage instanceFamily = "s"
	InstanceFamilyMemory  instanceFamily = "m"

	maxPods           = 110
	amd64Architecture = "amd64"
	linuxOS           = "linux"
)

func (f instanceFamily) memoryToVCPURatio() int {
	switch f {
	case InstanceFamilyCompute:
		return 2
	case InstanceFamilyStorage:
		return 4
	case InstanceFamilyMemory:
		return 8
	default:
		return 0
	}
}

func (f instanceFamily) IsValid() bool {
	switch f {
	case InstanceFamilyCompute, InstanceFamilyStorage, InstanceFamilyMemory:
		return true
	default:
		return false
	}
}

type instanceProfile struct {
	Family instanceFamily
	Size   int
	VCPU   int
}

func (p instanceProfile) name() string {
	return fmt.Sprintf("%s-%dx", p.Family, p.Size)
}

func (p instanceProfile) price() float64 {
	return float64(p.VCPU)*0.25 + float64(p.memGiB())*0.01
}

type KwokInstanceTypesProvider interface {
	List(ctx context.Context, diskSize int64) ([]*cloudprovider.InstanceType, error)
}

func (p instanceProfile) memGiB() int {
	return p.VCPU * p.Family.memoryToVCPURatio()
}

type KwokInstanceTypesStaticProvider struct {
}

var instanceTypeSizes = [...]int{
	2, 4, 9,
	16, 32,
}

func sizeToVCPU(size int) int {
	return size / 2
}

func getProfilesCatalog() map[string]instanceProfile {
	catalog := make(map[string]instanceProfile)
	families := []instanceFamily{InstanceFamilyCompute, InstanceFamilyStorage, InstanceFamilyMemory}
	for _, family := range families {
		for _, size := range instanceTypeSizes {
			p := instanceProfile{
				Family: family,
				Size:   size,
				VCPU:   sizeToVCPU(size),
			}
			catalog[p.name()] = p
		}
	}
	return catalog
}

// Catalog contains all supported instance profiles.
var instanceProfilesCatalog = getProfilesCatalog()

func (r KwokInstanceTypesStaticProvider) List(ctx context.Context, diskSize int64) ([]*cloudprovider.InstanceType, error) {
	zone, region := getZoneAndRegionFromContext(ctx)
	result := make([]*cloudprovider.InstanceType, 0, len(instanceProfilesCatalog))
	for _, profile := range instanceProfilesCatalog {
		// TODO: specify os and resourcePods instead of hardcoding them
		result = append(result, enrichToInstanceType(&profile, linuxOS, amd64Architecture, diskSize, zone, region, maxPods))
	}
	return result, nil
}

func getZoneAndRegionFromContext(ctx context.Context) (string, string) {
	var zone, region string
	if opts := options.FromContext(ctx); opts != nil {
		zone = opts.Zone
		region = opts.Region
	}
	return zone, region
}

func enrichToInstanceType(profile *instanceProfile, os string, arch string, diskSize int64, zone string, region string, resourcePods int) *cloudprovider.InstanceType {
	return &cloudprovider.InstanceType{
		Name: profile.name(),
		Requirements: scheduling.NewRequirements(
			scheduling.NewRequirement(corev1.LabelInstanceTypeStable, corev1.NodeSelectorOpIn, profile.name()),
			scheduling.NewRequirement(corev1.LabelArchStable, corev1.NodeSelectorOpIn, arch),
			scheduling.NewRequirement(corev1.LabelOSStable, corev1.NodeSelectorOpIn, os),
		),
		Capacity: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse(fmt.Sprintf("%d", profile.VCPU)),
			corev1.ResourceMemory:           resource.MustParse(fmt.Sprintf("%dGi", profile.memGiB())),
			corev1.ResourcePods:             resource.MustParse(fmt.Sprintf("%d", resourcePods)),
			corev1.ResourceEphemeralStorage: resource.MustParse(utils.GiToByteAsString(diskSize)),
		},
		//TODO: compute kubelet overhead
		Overhead: &cloudprovider.InstanceTypeOverhead{},
		Offerings: []*cloudprovider.Offering{
			{
				Requirements: scheduling.NewRequirements(
					scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn, zone),
					scheduling.NewRequirement(corev1.LabelTopologyRegion, corev1.NodeSelectorOpIn, region),
					scheduling.NewRequirement(karpv1.CapacityTypeLabelKey, corev1.NodeSelectorOpIn, karpv1.CapacityTypeOnDemand),
				),
				Price:     profile.price(),
				Available: true,
			},
		},
	}
}
