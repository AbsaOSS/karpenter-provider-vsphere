package instance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/vim25/types"
)

func TestGenerateSpec(t *testing.T) {
	instanceType := &corecloudprovider.InstanceType{
		Name: "vsphere-vm.cpu-1.mem-64gb.os-linux",
		Capacity: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("1"),
			corev1.ResourceMemory:           resource.MustParse("64Gi"),
			corev1.ResourceEphemeralStorage: resource.MustParse(utils.GiToByteAsString(5)),
		},
	}
	expectedMemInMB := int64(65536) // 64 GiB in megabytes
	mem := utils.InstanceTypeToMegabytes(instanceType.Capacity.Memory())
	assert.Equal(t, expectedMemInMB, mem)
}

func TestGenerateVMName(t *testing.T) {
	require.Equal(t, "cluster-karp-claim", GenerateVMName("cluster", "claim"))
}

func TestNewInstance(t *testing.T) {
	created := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tags := map[string]string{
		v1alpha1.LabelInstanceSize: "small",
	}

	instance := NewInstance(nil, "vm-uuid", "template-path", "poweredOn", "vm-name", created, tags)

	require.Equal(t, created, instance.LaunchTime)
	require.Equal(t, "vm-uuid", instance.ID)
	require.Equal(t, "template-path", instance.Image)
	require.Equal(t, "poweredOn", instance.State)
	require.Equal(t, "vm-name", instance.Name)
	require.Equal(t, "small", instance.Type)
	require.Equal(t, tags, instance.Tags)
	require.Nil(t, instance.GetVM())
}

func TestConfigExtract(t *testing.T) {
	config := Config{
		&types.OptionValue{Key: "key", Value: "value"},
	}

	require.Equal(t, []types.BaseOptionValue(config), config.Extract())
	var nilConfig *Config
	require.Nil(t, nilConfig.Extract())
}

func TestImageFromAnnotation(t *testing.T) {
	tests := []struct {
		name   string
		config *types.VirtualMachineConfigInfo
		want   string
	}{
		{name: "nil config", want: ImageNotFound},
		{name: "empty annotation", config: &types.VirtualMachineConfigInfo{}, want: ""},
		{name: "legacy format without space", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from:/dc0/vm/flatcar-template"}, want: "/dc0/vm/flatcar-template"},
		{name: "legacy format with space", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from: /dc0/vm/flatcar-template"}, want: "/dc0/vm/flatcar-template"},
		{name: "yaml format with both keys", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from: /DC0/vm/flatcar-template\ninstanceType: s-2x\n"}, want: "/DC0/vm/flatcar-template"},
		{name: "yaml format reordered keys", config: &types.VirtualMachineConfigInfo{Annotation: "instanceType: s-2x\ncloned_from: /DC0/vm/flatcar-template\n"}, want: "/DC0/vm/flatcar-template"},
		{name: "yaml with only cloned_from", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from: /dc0/vm/flatcar-template\n"}, want: "/dc0/vm/flatcar-template"},
		{name: "no recognized key falls back to raw string", config: &types.VirtualMachineConfigInfo{Annotation: "/dc0/vm/flatcar-template"}, want: "/dc0/vm/flatcar-template"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, imageFromConfig(test.config))
		})
	}
}

func TestInstanceTypeFromAnnotation(t *testing.T) {
	tests := []struct {
		name   string
		config *types.VirtualMachineConfigInfo
		want   string
	}{
		{name: "nil config", want: ""},
		{name: "empty annotation", config: &types.VirtualMachineConfigInfo{}, want: ""},
		{name: "legacy format, no instance_type key without space", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from:/dc0/vm/flatcar-template"}, want: ""},
		{name: "legacy format, no instance_type key", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from: /dc0/vm/flatcar-template"}, want: ""},
		{name: "yaml format", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from: /DC0/vm/flatcar-template\ninstanceType: s-2x\n"}, want: "s-2x"},
		{name: "yaml format reordered", config: &types.VirtualMachineConfigInfo{Annotation: "instanceType: s-2x\ncloned_from: /DC0/vm/flatcar-template\n"}, want: "s-2x"},
		{name: "yaml with only instanceType", config: &types.VirtualMachineConfigInfo{Annotation: "instanceType: m-4x\n"}, want: "m-4x"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, instanceTypeFromConfig(test.config))
		})
	}
}

func TestBelongsToCluster(t *testing.T) {
	tests := []struct {
		name        string
		tags        map[string]string
		clusterName string
		want        bool
	}{
		{name: "matching cluster", tags: map[string]string{v1alpha1.ClusterNameTagKey: "cluster-a"}, clusterName: "cluster-a", want: true},
		{name: "different cluster", tags: map[string]string{v1alpha1.ClusterNameTagKey: "cluster-b"}, clusterName: "cluster-a", want: false},
		{name: "legacy misspelled key is ignored", tags: map[string]string{"karpneter.sh/clustername": "cluster-a"}, clusterName: "cluster-a", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, belongsToCluster(test.tags, test.clusterName))
		})
	}
}

// TestSelectInstanceType_SelectsByPriceFirst verifies that when instance types
// have different prices, the cheapest option is selected regardless of family.
// Using KwoK pricing with valid sizes [2, 4, 9, 16, 32]:
// - c-2x: 1 CPU, 2 GiB, price = 0.27 (cheapest)
// - s-2x: 1 CPU, 4 GiB, price = 0.29
// - m-2x: 1 CPU, 8 GiB, price = 0.33 (most expensive)
func TestSelectInstanceType_SelectsByPriceFirst(t *testing.T) {
	zone := "us-west-2a"
	ctx := context.Background()
	ctx = options.ToContext(ctx, &options.Options{Zone: zone})

	// Compute type (c-2x) is cheapest, so it should be selected even though
	// memory type (m-2x) might be more suitable for memory-optimized workload
	computeType := testutil.NewInstanceType().
		WithName("c-2x").
		WithCPU("1").
		WithMemory("2Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		Build() // Cheapest: 0.27
	standardType := testutil.NewInstanceType().
		WithName("s-2x").
		WithCPU("1").
		WithMemory("4Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		Build() // Mid: 0.29
	memoryType := testutil.NewInstanceType().
		WithName("m-2x").
		WithCPU("1").
		WithMemory("8Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		Build() // Most expensive: 0.33

	provider := &DefaultProvider{}
	selected, err := provider.selectInstanceType(ctx, []*corecloudprovider.InstanceType{
		memoryType,
		standardType,
		computeType,
	})

	require.NoError(t, err)
	require.Equal(t, "c-2x", selected.Name)
}

// TestSelectInstanceType_SelectsLowestCPUWhenPriceEqual verifies the tie-breaking
// logic: when prices are equal, lower CPU count wins.
// Using fixed prices to test the CPU tiebreaker when prices are identical.
// Using KwoK valid sizes [2, 4, 9, 16, 32]:
// - s-2x: 1 CPU (lowest)
// - m-4x: 2 CPU
// - c-9x: 4 CPU (highest)
func TestSelectInstanceType_SelectsLowestCPUWhenPriceEqual(t *testing.T) {
	zone := "us-west-2a"
	ctx := context.Background()
	ctx = options.ToContext(ctx, &options.Options{Zone: zone})

	// All have same fixed price, but different CPU counts
	// s-2x has lowest CPU (1), so it should be selected
	type1 := testutil.NewInstanceType().
		WithName("s-2x").
		WithCPU("1").
		WithMemory("4Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Lowest CPU: 1
	type2 := testutil.NewInstanceType().
		WithName("m-4x").
		WithCPU("2").
		WithMemory("16Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Mid CPU: 2
	type3 := testutil.NewInstanceType().
		WithName("c-9x").
		WithCPU("4").
		WithMemory("8Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Highest CPU: 4

	provider := &DefaultProvider{}
	selected, err := provider.selectInstanceType(ctx, []*corecloudprovider.InstanceType{
		type3,
		type1,
		type2,
	})

	require.NoError(t, err)
	require.Equal(t, "s-2x", selected.Name)
}

// TestSelectInstanceType_SelectsLowestMemoryWhenCPUEqual verifies the next tie-breaker:
// when prices and CPU counts are equal, lower memory count wins.
// Using fixed prices to test the memory tiebreaker logic.
// Using KwoK valid sizes [2, 4, 9, 16, 32] with 2 CPU (size-4x):
// - c-4x: 2 CPU, 4 GiB (lowest memory)
// - s-4x: 2 CPU, 8 GiB
// - m-4x: 2 CPU, 16 GiB (highest memory)
func TestSelectInstanceType_SelectsLowestMemoryWhenCPUEqual(t *testing.T) {
	zone := "us-west-2a"
	ctx := context.Background()
	ctx = options.ToContext(ctx, &options.Options{Zone: zone})

	// Same price and CPU, but different memory
	// c-4x has lowest memory (4 GiB), so it should be selected
	type1 := testutil.NewInstanceType().
		WithName("c-4x").
		WithCPU("2").
		WithMemory("4Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Lowest memory: 4 GiB
	type2 := testutil.NewInstanceType().
		WithName("s-4x").
		WithCPU("2").
		WithMemory("8Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Mid memory: 8 GiB
	type3 := testutil.NewInstanceType().
		WithName("m-4x").
		WithCPU("2").
		WithMemory("16Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(100.0).
		Build() // Highest memory: 16 GiB

	provider := &DefaultProvider{}
	selected, err := provider.selectInstanceType(ctx, []*corecloudprovider.InstanceType{
		type3,
		type1,
		type2,
	})

	require.NoError(t, err)
	require.Equal(t, "c-4x", selected.Name)
}

// TestSelectInstanceType_FiltersOutUnavailableZones verifies that instance types
// without offerings in the requested zone are filtered out.
// Using KwoK pricing model with valid sizes [2, 4, 9, 16, 32].
func TestSelectInstanceType_FiltersOutUnavailableZones(t *testing.T) {
	zone := "us-west-2a"
	ctx := context.Background()
	ctx = options.ToContext(ctx, &options.Options{Zone: zone})

	// Instance in correct zone
	correctZone := testutil.NewInstanceType().
		WithName("m-2x").
		WithCPU("1").
		WithMemory("8Gi").
		WithZone(zone).
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(0.33). // (1*0.25)+(8*0.01)
		Build()

	// Instance in different zone
	wrongZone := testutil.NewInstanceType().
		WithName("c-2x").
		WithCPU("1").
		WithMemory("2Gi").
		WithZone("us-west-2b").
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(0.27). // (1*0.25)+(2*0.01)
		Build()

	provider := &DefaultProvider{}
	selected, err := provider.selectInstanceType(ctx, []*corecloudprovider.InstanceType{
		wrongZone,
		correctZone,
	})

	require.NoError(t, err)
	require.Equal(t, "m-2x", selected.Name)
}

// TestSelectInstanceType_ErrorWhenNoAvailableTypes verifies that an error is returned
// when no instance types have offerings available in the requested zone.
// Using KwoK valid sizes [2, 4, 9, 16, 32].
func TestSelectInstanceType_ErrorWhenNoAvailableTypes(t *testing.T) {
	zone := "us-west-2a"
	ctx := context.Background()
	ctx = options.ToContext(ctx, &options.Options{Zone: zone})

	// Instance only available in different zone
	wrongZone := testutil.NewInstanceType().
		WithName("m-2x").
		WithCPU("1").
		WithMemory("8Gi").
		WithZone("us-west-2b").
		WithOfferingSchedulingCapacityTypeOnDemand().
		WithPrice(0.33). // (1*0.25)+(8*0.01)
		Build()

	provider := &DefaultProvider{}
	_, err := provider.selectInstanceType(ctx, []*corecloudprovider.InstanceType{wrongZone})

	require.Error(t, err)
	require.Contains(t, err.Error(), "no instance types have an available on-demand offering")
	require.Contains(t, err.Error(), zone)
}
