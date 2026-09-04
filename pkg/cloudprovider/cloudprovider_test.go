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
