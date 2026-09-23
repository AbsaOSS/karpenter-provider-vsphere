package cloudprovider

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil"
	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil/instancefixture"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

func TestInstanceToNodeClaim(t *testing.T) {
	launchTime := instancefixture.StandardLaunchTime()

	i := instancefixture.New().
		WithID("vm-1234").
		WithName("worker-1234").
		WithImage("ubuntu-2404").
		WithState("poweredOn").
		WithLaunchTime(launchTime).
		WithClusterName("test-cluster").
		WithZone("zone-a").
		WithNodePool("default-nodepool").
		Build()

	instanceType := testutil.StandardInstanceType8X()

	provider := &CloudProvider{}

	nodeClaim := provider.instanceToNodeClaim(i, instanceType)

	require.NotNil(t, nodeClaim)

	t.Run("sets metadata", func(t *testing.T) {
		assert.Equal(
			t,
			GenerateNodeClaimName(
				i.Name,
				i.Tags[v1alpha1.ClusterNameTagKey],
			),
			nodeClaim.Name,
		)

		assert.Equal(t, launchTime, nodeClaim.CreationTimestamp.Time)
		assert.Nil(t, nodeClaim.DeletionTimestamp)
		assert.NotNil(t, nodeClaim.Labels)
		assert.NotNil(t, nodeClaim.Annotations)
		assert.Empty(t, nodeClaim.Annotations)
	})

	t.Run("sets provider status", func(t *testing.T) {
		assert.Equal(
			t,
			"vsphere://vm-1234",
			nodeClaim.Status.ProviderID,
		)

		assert.Equal(
			t,
			"ubuntu-2404",
			nodeClaim.Status.ImageID,
		)
	})

	t.Run("sets instance type label", func(t *testing.T) {
		actual, exists := nodeClaim.Labels[corev1.LabelInstanceTypeStable]

		assert.True(
			t,
			exists,
			"expected label %q to be present",
			corev1.LabelInstanceTypeStable,
		)

		assert.Equal(
			t,
			instanceType.Name,
			actual,
		)
	})

	t.Run("sets on-demand capacity type label", func(t *testing.T) {
		actual, exists := nodeClaim.Labels[karpv1.CapacityTypeLabelKey]

		assert.True(
			t,
			exists,
			"expected label %q to be present",
			karpv1.CapacityTypeLabelKey,
		)

		assert.Equal(
			t,
			karpv1.CapacityTypeOnDemand,
			actual,
			"NodeClaim should have the on-demand capacity type",
		)
	})

	t.Run("sets topology zone label", func(t *testing.T) {
		actual, exists := nodeClaim.Labels[corev1.LabelTopologyZone]

		assert.True(
			t,
			exists,
			"expected label %q to be present",
			corev1.LabelTopologyZone,
		)

		assert.Equal(t, "zone-a", actual)
	})

	t.Run("sets node pool label", func(t *testing.T) {
		actual, exists := nodeClaim.Labels[karpv1.NodePoolLabelKey]

		assert.True(
			t,
			exists,
			"expected label %q to be present",
			karpv1.NodePoolLabelKey,
		)

		assert.Equal(t, "default-nodepool", actual)
	})

	t.Run("sets resolved requirement labels", func(t *testing.T) {
		assert.Equal(
			t,
			"amd64",
			nodeClaim.Labels[corev1.LabelArchStable],
		)

		assert.Equal(
			t,
			"linux",
			nodeClaim.Labels[corev1.LabelOSStable],
		)
	})

	t.Run("sets non-zero capacity", func(t *testing.T) {
		assertResourceQuantity(
			t,
			nodeClaim.Status.Capacity,
			corev1.ResourceCPU,
			"4",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Capacity,
			corev1.ResourceMemory,
			"16Gi",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Capacity,
			corev1.ResourcePods,
			"110",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Capacity,
			corev1.ResourceEphemeralStorage,
			"100Gi",
		)
	})

	t.Run("sets non-zero allocatable resources", func(t *testing.T) {
		assertResourceQuantity(
			t,
			nodeClaim.Status.Allocatable,
			corev1.ResourceCPU,
			"4",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Allocatable,
			corev1.ResourceMemory,
			"16Gi",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Allocatable,
			corev1.ResourcePods,
			"110",
		)

		assertResourceQuantity(
			t,
			nodeClaim.Status.Allocatable,
			corev1.ResourceEphemeralStorage,
			"100Gi",
		)
	})

	t.Run("filters zero-valued resources", func(t *testing.T) {
		zeroResource := corev1.ResourceName("example.com/zero")

		assert.NotContains(
			t,
			nodeClaim.Status.Capacity,
			zeroResource,
		)

		assert.NotContains(
			t,
			nodeClaim.Status.Allocatable,
			zeroResource,
		)
	})
}

func TestInstanceToNodeClaimPoweredOff(t *testing.T) {
	i := instancefixture.New().
		WithID("vm-powered-off").
		WithName("worker-powered-off").
		WithImage("ubuntu-2404").
		WithState("powerOff").
		WithClusterName("test-cluster").
		WithZone("zone-a").
		WithNodePool("default-nodepool").
		Build()

	instanceType := testutil.ComputeInstanceType4X()

	provider := &CloudProvider{}

	nodeClaim := provider.instanceToNodeClaim(i, instanceType)
	require.NotNil(t, nodeClaim.DeletionTimestamp)

	assert.WithinDuration(
		t,
		time.Now(),
		nodeClaim.DeletionTimestamp.Time,
		time.Minute,
		"DeletionTimestamp should be within one minute of the current time",
	)

	assert.Equal(
		t,
		karpv1.CapacityTypeOnDemand,
		nodeClaim.Labels[karpv1.CapacityTypeLabelKey],
	)

	assert.Equal(
		t,
		karpv1.CapacityTypeOnDemand,
		nodeClaim.Labels[karpv1.CapacityTypeLabelKey],
	)
}

func TestInstanceTypesFromNodeClass(t *testing.T) {
	nodeClass := testutil.NewNodeClass().
		WithSingleInstanceType("4", "16Gi", 110, "zone-a", "Linux").
		WithInstanceType("2", "8Gi", 55, "zone-b", "linux").
		Build()

	instanceTypes := instanceTypesFromNodeClass(nodeClass)

	require.Len(t, instanceTypes, 2, "expected 2 instance types")

	t.Run("creates instance type with correct name format", func(t *testing.T) {
		assert.Equal(
			t,
			"vsphere-vm.cpu-4.mem-16gb.os-linux",
			instanceTypes[0].Name,
		)
		assert.Equal(
			t,
			"vsphere-vm.cpu-2.mem-8gb.os-linux",
			instanceTypes[1].Name,
		)
	})

	t.Run("sets instance type requirements", func(t *testing.T) {
		requirements := instanceTypes[0].Requirements
		require.NotNil(t, requirements)

		// Verify instance type requirement
		instanceTypeReq, ok := requirements[corev1.LabelInstanceTypeStable]
		require.True(t, ok, "expected instance type requirement")
		assert.Len(t, instanceTypeReq.Values(), 1)
		assert.Equal(t, "vsphere-vm.cpu-4.mem-16gb.os-linux", instanceTypeReq.Values()[0])

		// Verify architecture requirement
		archReq, ok := requirements[corev1.LabelArchStable]
		require.True(t, ok, "expected architecture requirement")
		assert.Len(t, archReq.Values(), 1)
		assert.Equal(t, "amd64", archReq.Values()[0])

		// Verify OS requirement
		osReq, ok := requirements[corev1.LabelOSStable]
		require.True(t, ok, "expected OS requirement")
		assert.Len(t, osReq.Values(), 1)
		assert.Equal(t, "linux", osReq.Values()[0])
	})

	t.Run("sets offering requirements with capacity type", func(t *testing.T) {
		offerings := instanceTypes[0].Offerings
		require.Len(t, offerings, 1, "expected 1 offering per instance type")

		offering := offerings[0]
		require.NotNil(t, offering.Requirements)

		// Verify topology zone requirement
		zoneReq, ok := offering.Requirements[corev1.LabelTopologyZone]
		require.True(t, ok, "expected topology zone requirement in offering")
		assert.Len(t, zoneReq.Values(), 1)
		assert.Equal(t, "zone-a", zoneReq.Values()[0])

		// Verify capacity type requirement - this is the fix we're testing
		capacityTypeReq, ok := offering.Requirements[karpv1.CapacityTypeLabelKey]
		require.True(t, ok, "expected capacity type requirement in offering")
		assert.Len(t, capacityTypeReq.Values(), 1)
		assert.Equal(t, karpv1.CapacityTypeOnDemand, capacityTypeReq.Values()[0])
	})

	t.Run("sets offering properties", func(t *testing.T) {
		offerings := instanceTypes[0].Offerings
		offering := offerings[0]

		assert.Equal(t, float64(100.0), offering.Price)
		assert.True(t, offering.Available)
	})

	t.Run("sets capacity resources", func(t *testing.T) {
		capacity := instanceTypes[0].Capacity
		require.NotNil(t, capacity)

		assertResourceQuantity(t, capacity, corev1.ResourceCPU, "4")
		assertResourceQuantity(t, capacity, corev1.ResourceMemory, "16Gi")
		assertResourceQuantity(t, capacity, corev1.ResourcePods, "110")
		// Disk size is converted from Gi to bytes (100Gi)
		assertResourceQuantity(t, capacity, corev1.ResourceEphemeralStorage, "107374182400")
	})

	t.Run("normalizes OS to lowercase", func(t *testing.T) {
		// First instance has "Linux" (capitalized), second has "linux" (lowercase)
		// Both should result in lowercase "linux" in name and requirements
		osReq0, ok := instanceTypes[0].Requirements[corev1.LabelOSStable]
		require.True(t, ok)
		assert.Equal(t, "linux", osReq0.Values()[0])

		osReq1, ok := instanceTypes[1].Requirements[corev1.LabelOSStable]
		require.True(t, ok)
		assert.Equal(t, "linux", osReq1.Values()[0])
	})

	t.Run("handles multiple instance types independently", func(t *testing.T) {
		// Verify second instance type has different zone
		offerings := instanceTypes[1].Offerings
		zoneReq, ok := offerings[0].Requirements[corev1.LabelTopologyZone]
		require.True(t, ok)
		assert.Equal(t, "zone-b", zoneReq.Values()[0])

		// Both should have same capacity type (OnDemand)
		capacityTypeReq, ok := offerings[0].Requirements[karpv1.CapacityTypeLabelKey]
		require.True(t, ok)
		assert.Equal(t, karpv1.CapacityTypeOnDemand, capacityTypeReq.Values()[0])
	})
}

func assertResourceQuantity(
	t *testing.T,
	resourceList corev1.ResourceList,
	resourceName corev1.ResourceName,
	expected string,
) {
	t.Helper()

	actual, exists := resourceList[resourceName]
	require.True(
		t,
		exists,
		"expected resource %q to be present",
		resourceName,
	)

	expectedQuantity := resource.MustParse(expected)

	assert.Zero(
		t,
		actual.Cmp(expectedQuantity),
		"resource %q was %q, expected %q",
		resourceName,
		actual.String(),
		expectedQuantity.String(),
	)
}

func TestResolveInstanceTypes(t *testing.T) {
	ctx := context.Background()

	// Create test data using builders from testutil
	mockKwokInstanceTypes := testutil.MockKwokInstanceTypesSlice()
	mockKwokProvider := testutil.NewMockKwokProvider(mockKwokInstanceTypes)
	nodeClass := testutil.NewNodeClass().Build()
	nodeClaim := testutil.NewNodeClaim().Build()
	provider := newTestCloudProvider(mockKwokProvider)

	// Call resolveInstanceTypes
	instanceTypes, err := provider.resolveInstanceTypes(ctx, nodeClaim, nodeClass)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, instanceTypes)
	require.GreaterOrEqual(t, len(instanceTypes), 2,
		"expected at least 2 instance types (1 from CR, 1 from KWOK)")

	// Extract instance type names for easier assertions
	names := make([]string, len(instanceTypes))
	for i, it := range instanceTypes {
		names[i] = it.Name
	}

	const kwokInstanceTypeName = "c-4x"
	require.Contains(t, names, "vsphere-vm.cpu-4.mem-16gb.os-linux", "CR-based instance type should be present")
	require.Contains(t, names, kwokInstanceTypeName, "KWOK instance type should be present")
	assert.Contains(t, names, "vsphere-vm.cpu-4.mem-16gb.os-linux", "CR-based instance type should be in the list")
	assert.Contains(t, names, kwokInstanceTypeName, "KWOK instance type should be in the list")
}

// newTestCloudProvider creates a CloudProvider with the given KWOK provider for testing.
// This is kept here because CloudProvider fields are private and cannot be directly accessed from testutil.
func newTestCloudProvider(kwokProvider *testutil.MockKwokInstanceTypesProvider) *CloudProvider {
	return &CloudProvider{
		kwokInstanceTypesProvider: kwokProvider,
	}
}
