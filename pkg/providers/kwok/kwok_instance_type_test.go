package kwok

import (
	"context"
	"fmt"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

func TestSizeToVCPU(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "size 2 gives 1 VCPU",
			size:     2,
			expected: 1,
		},
		{
			name:     "size 4 gives 2 VCPU",
			size:     4,
			expected: 2,
		},
		{
			name:     "size 9 gives 4 VCPU",
			size:     9,
			expected: 4,
		},
		{
			name:     "size 16 gives 8 VCPU",
			size:     16,
			expected: 8,
		},
		{
			name:     "size 32 gives 16 VCPU",
			size:     32,
			expected: 16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sizeToVCPU(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstanceProfileMemGiB(t *testing.T) {
	tests := []struct {
		name     string
		profile  instanceProfile
		expected int
		desc     string
	}{
		{
			name: "Compute family with size 2 (VCPU 1)",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   2,
				VCPU:   1,
			},
			expected: 2, // 1 * 2 = 2
			desc:     "1 VCPU * 2 ratio = 2 GiB",
		},
		{
			name: "Compute family with size 4 (VCPU 2)",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   4,
				VCPU:   2,
			},
			expected: 4, // 2 * 2 = 4
			desc:     "2 VCPU * 2 ratio = 4 GiB",
		},
		{
			name: "Storage family with size 2 (VCPU 1)",
			profile: instanceProfile{
				Family: InstanceFamilyStorage,
				Size:   2,
				VCPU:   1,
			},
			expected: 4, // 1 * 4 = 4
			desc:     "1 VCPU * 4 ratio = 4 GiB",
		},
		{
			name: "Memory family with size 4 (VCPU 2)",
			profile: instanceProfile{
				Family: InstanceFamilyMemory,
				Size:   4,
				VCPU:   2,
			},
			expected: 16, // 2 * 8 = 16
			desc:     "2 VCPU * 8 ratio = 16 GiB",
		},
		{
			name: "Compute with size 9 (VCPU 4)",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   9,
				VCPU:   4,
			},
			expected: 8, // 4 * 2 = 8
			desc:     "4 VCPU * 2 ratio = 8 GiB",
		},
		{
			name: "Storage with size 9 (VCPU 4)",
			profile: instanceProfile{
				Family: InstanceFamilyStorage,
				Size:   9,
				VCPU:   4,
			},
			expected: 16, // 4 * 4 = 16
			desc:     "4 VCPU * 4 ratio = 16 GiB",
		},
		{
			name: "Memory with size 16 (VCPU 8)",
			profile: instanceProfile{
				Family: InstanceFamilyMemory,
				Size:   16,
				VCPU:   8,
			},
			expected: 64, // 8 * 8 = 64
			desc:     "8 VCPU * 8 ratio = 64 GiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.memGiB()
			assert.Equal(t, tt.expected, result, tt.desc)
		})
	}
}

func TestInstanceProfilePrice(t *testing.T) {
	tests := []struct {
		name             string
		profile          instanceProfile
		expectedMinPrice float64
		expectedMaxPrice float64
		desc             string
	}{
		{
			name: "Compute size 2 (VCPU 1)",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   2,
				VCPU:   1,
			},
			// price = 1*0.25 + 2*0.01 = 0.25 + 0.02 = 0.27
			expectedMinPrice: 0.269,
			expectedMaxPrice: 0.271,
			desc:             "1 * 0.25 + 2 * 0.01 = 0.27",
		},
		{
			name: "Compute size 4 (VCPU 2)",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   4,
				VCPU:   2,
			},
			// price = 2*0.25 + 4*0.01 = 0.5 + 0.04 = 0.54
			expectedMinPrice: 0.539,
			expectedMaxPrice: 0.541,
			desc:             "2 * 0.25 + 4 * 0.01 = 0.54",
		},
		{
			name: "Memory size 32 (VCPU 16)",
			profile: instanceProfile{
				Family: InstanceFamilyMemory,
				Size:   32,
				VCPU:   16,
			},
			// price = 16*0.25 + 128*0.01 = 4.0 + 1.28 = 5.28
			expectedMinPrice: 5.27,
			expectedMaxPrice: 5.29,
			desc:             "16 * 0.25 + 128 * 0.01 = 5.28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.price()
			assert.GreaterOrEqual(t, result, tt.expectedMinPrice, tt.desc)
			assert.LessOrEqual(t, result, tt.expectedMaxPrice, tt.desc)
		})
	}
}

func TestInstanceProfileName(t *testing.T) {
	tests := []struct {
		name     string
		profile  instanceProfile
		expected string
	}{
		{
			name: "Compute family",
			profile: instanceProfile{
				Family: InstanceFamilyCompute,
				Size:   2,
				VCPU:   1,
			},
			expected: "c-2x",
		},
		{
			name: "Storage family",
			profile: instanceProfile{
				Family: InstanceFamilyStorage,
				Size:   4,
				VCPU:   2,
			},
			expected: "s-4x",
		},
		{
			name: "Memory family",
			profile: instanceProfile{
				Family: InstanceFamilyMemory,
				Size:   16,
				VCPU:   8,
			},
			expected: "m-16x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.name()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProfilesCatalog(t *testing.T) {
	catalog := getProfilesCatalog()
	const expectedProfileCount = 15

	t.Run(fmt.Sprintf("catalog contains %d profiles (3 families * 5 sizes)", expectedProfileCount), func(t *testing.T) {
		assert.Equal(t, expectedProfileCount, len(catalog), fmt.Sprintf("catalog should contain %d profiles", expectedProfileCount))
	})

	t.Run("all profiles in catalog have correct VCPU conversion", func(t *testing.T) {
		expectedSizes := []int{2, 4, 9, 16, 32}
		for _, size := range expectedSizes {
			expectedVCPU := size / 2
			for _, fam := range []instanceFamily{InstanceFamilyCompute, InstanceFamilyStorage, InstanceFamilyMemory} {
				profileName := fmt.Sprintf("%s-%dx", fam, size)
				profile, exists := catalog[profileName]
				require.True(t, exists, "profile %s should exist in catalog", profileName)
				assert.Equal(t, expectedVCPU, profile.VCPU, "VCPU for size %d should be %d", size, expectedVCPU)
			}
		}
	})

	t.Run("compute family has correct memory ratio (2x VCPU)", func(t *testing.T) {
		tests := []struct {
			size        int
			vcpu        int
			expectedMem int
		}{
			{2, 1, 2},
			{4, 2, 4},
			{9, 4, 8},
			{16, 8, 16},
			{32, 16, 32},
		}

		for _, tt := range tests {
			profile := instanceProfile{Family: InstanceFamilyCompute, Size: tt.size, VCPU: tt.vcpu}
			assert.Equal(t, tt.expectedMem, profile.memGiB(), "compute size %d: expected %d GiB", tt.size, tt.expectedMem)
		}
	})

	t.Run("storage family has correct memory ratio (4x VCPU)", func(t *testing.T) {
		tests := []struct {
			size        int
			vcpu        int
			expectedMem int
		}{
			{2, 1, 4},
			{4, 2, 8},
			{9, 4, 16},
			{16, 8, 32},
			{32, 16, 64},
		}

		for _, tt := range tests {
			profile := instanceProfile{Family: InstanceFamilyStorage, Size: tt.size, VCPU: tt.vcpu}
			assert.Equal(t, tt.expectedMem, profile.memGiB(), "storage size %d: expected %d GiB", tt.size, tt.expectedMem)
		}
	})

	t.Run("memory family has correct memory ratio (8x VCPU)", func(t *testing.T) {
		tests := []struct {
			size        int
			vcpu        int
			expectedMem int
		}{
			{2, 1, 8},
			{4, 2, 16},
			{9, 4, 32},
			{16, 8, 64},
			{32, 16, 128},
		}

		for _, tt := range tests {
			profile := instanceProfile{Family: InstanceFamilyMemory, Size: tt.size, VCPU: tt.vcpu}
			assert.Equal(t, tt.expectedMem, profile.memGiB(), "memory size %d: expected %d GiB", tt.size, tt.expectedMem)
		}
	})
}

func TestGetZoneAndRegionFromContext(t *testing.T) {
	t.Run("without options in context, returns empty strings", func(t *testing.T) {
		ctx := context.Background()
		zone, region := getZoneAndRegionFromContext(ctx)
		assert.Equal(t, "", zone)
		assert.Equal(t, "", region)
	})

	t.Run("with options in context, returns zone and region", func(t *testing.T) {
		opts := &options.Options{
			Zone:   "us-west-1a",
			Region: "us-west-1",
		}
		ctx := options.ToContext(context.Background(), opts)
		zone, region := getZoneAndRegionFromContext(ctx)
		assert.Equal(t, "us-west-1a", zone)
		assert.Equal(t, "us-west-1", region)
	})
}

func TestEnrichToInstanceType(t *testing.T) {
	profile := &instanceProfile{
		Family: InstanceFamilyCompute,
		Size:   2,
		VCPU:   1,
	}

	instanceType := enrichToInstanceType(profile, "linux", "amd64", 100, "us-west-1a", "us-west-1", 110)

	t.Run("has correct name", func(t *testing.T) {
		assert.Equal(t, "c-2x", instanceType.Name)
	})

	t.Run("has instance type requirement", func(t *testing.T) {
		req := instanceType.Requirements.Get(corev1.LabelInstanceTypeStable)
		assert.Equal(t, 1, len(req.Values()), "should have one instance type value")
		assert.Contains(t, req.Values(), "c-2x")
	})

	t.Run("has architecture requirement", func(t *testing.T) {
		req := instanceType.Requirements.Get(corev1.LabelArchStable)
		assert.Equal(t, 1, len(req.Values()), "should have one architecture value")
		assert.Contains(t, req.Values(), "amd64")
	})

	t.Run("has OS requirement", func(t *testing.T) {
		req := instanceType.Requirements.Get(corev1.LabelOSStable)
		assert.Equal(t, 1, len(req.Values()), "should have one OS value")
		assert.Contains(t, req.Values(), "linux")
	})

	t.Run("has correct capacity", func(t *testing.T) {
		assert.NotNil(t, instanceType.Capacity[corev1.ResourceCPU])
		assert.NotNil(t, instanceType.Capacity[corev1.ResourceMemory])
		assert.NotNil(t, instanceType.Capacity[corev1.ResourcePods])
		assert.NotNil(t, instanceType.Capacity[corev1.ResourceEphemeralStorage])
	})

	t.Run("has one offering with zone and region", func(t *testing.T) {
		require.Equal(t, 1, len(instanceType.Offerings))
		offering := instanceType.Offerings[0]

		zoneReq := offering.Requirements.Get(corev1.LabelTopologyZone)
		assert.Equal(t, 1, len(zoneReq.Values()))
		assert.Equal(t, "us-west-1a", zoneReq.Values()[0])

		regionReq := offering.Requirements.Get(corev1.LabelTopologyRegion)
		assert.Equal(t, 1, len(regionReq.Values()))
		assert.Equal(t, "us-west-1", regionReq.Values()[0])

		capacityTypeReq := offering.Requirements.Get(karpv1.CapacityTypeLabelKey)
		assert.Equal(t, 1, len(capacityTypeReq.Values()))
		assert.Equal(t, karpv1.CapacityTypeOnDemand, capacityTypeReq.Values()[0])
	})

	t.Run("offering is available", func(t *testing.T) {
		require.Equal(t, 1, len(instanceType.Offerings))
		assert.True(t, instanceType.Offerings[0].Available)
	})

	t.Run("price is calculated correctly", func(t *testing.T) {
		require.Equal(t, 1, len(instanceType.Offerings))
		// price = 1*0.25 + 2*0.01 = 0.27
		assert.GreaterOrEqual(t, instanceType.Offerings[0].Price, 0.269)
		assert.LessOrEqual(t, instanceType.Offerings[0].Price, 0.271)
	})
}

func TestKwokInstanceTypesStaticProviderList(t *testing.T) {
	provider := KwokInstanceTypesStaticProvider{}

	t.Run("returns instance types for different disk sizes", func(t *testing.T) {
		opts1 := &options.Options{Zone: "us-west-1a", Region: "us-west-1"}
		ctx1 := options.ToContext(context.Background(), opts1)

		instanceTypes1, err := provider.List(ctx1, 50)
		require.NoError(t, err)
		assert.Equal(t, 15, len(instanceTypes1), "should return 15 instance types (5 sizes * 3 families)")

		opts2 := &options.Options{Zone: "us-west-1b", Region: "us-west-1"}
		ctx2 := options.ToContext(context.Background(), opts2)

		instanceTypes2, err := provider.List(ctx2, 100)
		require.NoError(t, err)
		assert.Equal(t, 15, len(instanceTypes2), "should return 15 instance types")
	})

	t.Run("zone and region are correctly injected into each instance type", func(t *testing.T) {
		opts := &options.Options{Zone: "us-west-1c", Region: "us-west-1"}
		ctx := options.ToContext(context.Background(), opts)

		instanceTypes, err := provider.List(ctx, 50)
		require.NoError(t, err)

		for _, it := range instanceTypes {
			require.Equal(t, 1, len(it.Offerings), "each instance type should have 1 offering")
			offering := it.Offerings[0]

			zoneReq := offering.Requirements.Get(corev1.LabelTopologyZone)
			assert.Equal(t, "us-west-1c", zoneReq.Values()[0])

			regionReq := offering.Requirements.Get(corev1.LabelTopologyRegion)
			assert.Equal(t, "us-west-1", regionReq.Values()[0])
		}
	})
}

func TestFamilyMemoryToVCPURatio(t *testing.T) {
	tests := []struct {
		family   instanceFamily
		expected int
	}{
		{InstanceFamilyCompute, 2},
		{InstanceFamilyStorage, 4},
		{InstanceFamilyMemory, 8},
		{instanceFamily("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("family %s", tt.family), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.family.memoryToVCPURatio())
		})
	}
}

func TestFamilyIsValid(t *testing.T) {
	tests := []struct {
		family   instanceFamily
		expected bool
	}{
		{InstanceFamilyCompute, true},
		{InstanceFamilyStorage, true},
		{InstanceFamilyMemory, true},
		{instanceFamily("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("family %s", tt.family), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.family.IsValid())
		})
	}
}
