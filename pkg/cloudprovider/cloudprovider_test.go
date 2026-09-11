package cloudprovider

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

func TestInstanceToNodeClaim(t *testing.T) {
	launchTime := time.Date(
		2026,
		time.January,
		10,
		12,
		30,
		0,
		0,
		time.UTC,
	)

	i := &instance.Instance{
		ID:         "vm-1234",
		Name:       "worker-1234",
		Image:      "ubuntu-2404",
		State:      "poweredOn",
		LaunchTime: launchTime,
		Tags: map[string]string{
			v1alpha1.ClusterNameTagKey: "test-cluster",
			corev1.LabelTopologyZone:   "zone-a",
			karpv1.NodePoolLabelKey:    "default-nodepool",
		},
	}

	instanceType := &corecloudprovider.InstanceType{
		Name: "vsphere-vm.cpu-4.mem-16gb.os-ubuntu",
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
			corev1.ResourceCPU:                      resource.MustParse("4"),
			corev1.ResourceMemory:                   resource.MustParse("16Gi"),
			corev1.ResourcePods:                     resource.MustParse("110"),
			corev1.ResourceEphemeralStorage:         resource.MustParse("100Gi"),
			corev1.ResourceName("example.com/zero"): resource.MustParse("0"),
		},
		Overhead: &corecloudprovider.InstanceTypeOverhead{},
	}

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
	i := &instance.Instance{
		ID:    "vm-powered-off",
		Name:  "worker-powered-off",
		Image: "ubuntu-2404",
		State: "powerOff",
		LaunchTime: time.Date(
			2026,
			time.January,
			10,
			12,
			30,
			0,
			0,
			time.UTC,
		),
		Tags: map[string]string{
			v1alpha1.ClusterNameTagKey: "test-cluster",
			corev1.LabelTopologyZone:   "zone-a",
			karpv1.NodePoolLabelKey:    "default-nodepool",
		},
	}

	instanceType := &corecloudprovider.InstanceType{
		Name: "vsphere-vm.cpu-2.mem-4gb.os-ubuntu",
		Capacity: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
		Overhead: &corecloudprovider.InstanceTypeOverhead{},
	}

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
	nodeClass := &v1alpha1.VsphereNodeClass{
		Spec: v1alpha1.VsphereNodeClassSpec{
			DiskSize: 100,
			InstanceTypes: []v1alpha1.InstanceType{
				{
					CPU:     "4",
					Memory:  "16Gi",
					MaxPods: "110",
					Zone:    "zone-a",
					OS:      "Linux",
				},
				{
					CPU:     "2",
					Memory:  "8Gi",
					MaxPods: "55",
					Zone:    "zone-b",
					OS:      "linux",
				},
			},
		},
	}

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
