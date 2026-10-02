package overhead

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/karpenter/pkg/utils/resources"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
)

func TestMemory(t *testing.T) {
	tests := []struct {
		name string
		opts *options.Options
		want string
	}{
		{name: "no options in context", opts: nil, want: "16Gi"},
		{name: "zero perent", opts: &options.Options{VMMemoryOverheadPercent: 0}, want: "16Gi"},
		{name: "default percent", opts: &options.Options{VMMemoryOverheadPercent: 0.075}, want: "15155Mi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.opts != nil {
				ctx = options.ToContext(ctx, tt.opts)
			}
			got := Memory(ctx, resource.MustParse("16Gi"))
			require.Zero(t, got.Cmp(resource.MustParse(tt.want)), "got %s, want %s", got.String(), tt.want)
		})
	}
}

func TestDefaultKubeReservedCPU(t *testing.T) {
	tests := []struct {
		vcpu string
		want string
	}{
		{vcpu: "1", want: "60m"},
		{vcpu: "2", want: "70m"},
		{vcpu: "3", want: "75m"},
		{vcpu: "4", want: "80m"},
		{vcpu: "8", want: "90m"},
		{vcpu: "16", want: "110m"},
	}
	for _, tt := range tests {
		t.Run(tt.vcpu+" vCPU", func(t *testing.T) {
			got := defaultKubeReserved(resources.Quantity(tt.vcpu), resources.Quantity("110"))
			require.Equal(t, tt.want, got["cpu"])
		})
	}
}

func TestDefaultKubeReservedMemory(t *testing.T) {
	tests := []struct {
		pods string
		want string
	}{
		{pods: "1", want: "266Mi"},
		{pods: "50", want: "805Mi"},
		{pods: "110", want: "1465Mi"},
		{pods: "250", want: "3005Mi"},
	}
	for _, tt := range tests {
		t.Run(tt.pods+" pods", func(t *testing.T) {
			got := defaultKubeReserved(resources.Quantity("4"), resources.Quantity(tt.pods))
			require.Equal(t, tt.want, got["memory"])
			require.Equal(t, "1Gi", got["ephemeral-storage"])
		})
	}
}

func testCapacity() corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse("4"),
		corev1.ResourceMemory:           resource.MustParse("15155Mi"),
		corev1.ResourcePods:             resource.MustParse("110"),
		corev1.ResourceEphemeralStorage: resource.MustParse("50Gi"),
	}
}

func TestResolve(t *testing.T) {
	t.Run("nil config gives the defaults", func(t *testing.T) {
		r := Resolve(nil, testCapacity())
		require.Equal(t, int64(110), r.MaxPods)
		require.Equal(t, map[string]string{"cpu": "80m", "memory": "1465Mi", "ephemeral-storage": "1Gi"}, r.KubeReserved)
		require.Empty(t, r.SystemReserved)
		require.Equal(t, defaultEvictionHard, r.EvictionHard)
	})

	t.Run("an override replaces only its own key", func(t *testing.T) {
		r := Resolve(&v1alpha1.KubeletConfiguration{
			KubeReserved: map[string]string{"memory": "1Gi"},
			EvictionHard: map[string]string{"memory.available": "5%"},
		}, testCapacity())
		require.Equal(t, map[string]string{"cpu": "80m", "memory": "1Gi", "ephemeral-storage": "1Gi"}, r.KubeReserved)
		require.Equal(t, "5%", r.EvictionHard["memory.available"])
		require.Equal(t, "10%", r.EvictionHard["nodefs.available"])
		require.Len(t, r.EvictionHard, 5)
	})

	t.Run("does not modify the package defaults", func(t *testing.T) {
		Resolve(&v1alpha1.KubeletConfiguration{EvictionHard: map[string]string{"memory.available": "1Gi"}}, testCapacity())
		require.Equal(t, "100Mi", defaultEvictionHard["memory.available"])
	})
}

func TestComputeEvictionSignal(t *testing.T) {
	capacity := resource.MustParse("50Gi")
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "quantity", value: "500Mi", want: "500Mi"},
		{name: "percentage", value: "10%", want: "5Gi"},
		{name: "fractional percentage rounds up", value: "0.001%", want: "536871"},
		{name: "100% disables the threshold", value: "100%", want: "0"},
		{name: "invalid percentage", value: "ab12%", wantErr: true},
		{name: "invalid quantity", value: "ab12", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeEvictionSignal(capacity, tt.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Zero(t, got.Cmp(resource.MustParse(tt.want)), "got %s, want %s", got.String(), tt.want)
		})
	}
}

func TestNewOverhead(t *testing.T) {
	allocatable := func(capacity corev1.ResourceList, kc *v1alpha1.KubeletConfiguration) corev1.ResourceList {
		o, err := NewOverhead(Resolve(kc, capacity), capacity)
		require.NoError(t, err)
		return resources.Subtract(capacity, o.Total())
	}
	requireQuantity := func(t *testing.T, want string, got resource.Quantity) {
		require.Zero(t, got.Cmp(resource.MustParse(want)), "got %s, want %s", got.String(), want)
	}

	t.Run("example: defaults only", func(t *testing.T) {
		got := allocatable(testCapacity(), nil)
		requireQuantity(t, "3920m", got[corev1.ResourceCPU])
		requireQuantity(t, "13590Mi", got[corev1.ResourceMemory])
		requireQuantity(t, "44Gi", got[corev1.ResourceEphemeralStorage])
		requireQuantity(t, "110", got[corev1.ResourcePods])
	})

	t.Run("example: with spec.kubelet", func(t *testing.T) {
		capacity := testCapacity()
		capacity[corev1.ResourcePods] = resource.MustParse("50")
		got := allocatable(capacity, &v1alpha1.KubeletConfiguration{
			KubeReserved:   map[string]string{"memory": "1Gi"},
			SystemReserved: map[string]string{"cpu": "100m", "memory": "256Mi"},
			EvictionHard:   map[string]string{"memory.available": "5%", "nodefs.available": "15%"},
		})
		requireQuantity(t, "3820m", got[corev1.ResourceCPU])
		requireQuantity(t, "13117.25Mi", got[corev1.ResourceMemory])
		requireQuantity(t, "41.5Gi", got[corev1.ResourceEphemeralStorage])
	})

	t.Run("invalid value returns an error, no panic", func(t *testing.T) {
		r := Resolve(&v1alpha1.KubeletConfiguration{KubeReserved: map[string]string{"memory": "abc"}}, testCapacity())
		_, err := NewOverhead(r, testCapacity())
		require.ErrorContains(t, err, "kubeReserved")
	})

	t.Run("a resource with zero capacity is not reserved", func(t *testing.T) {
		capacity := testCapacity()
		capacity[corev1.ResourceEphemeralStorage] = resource.MustParse("0")

		o, err := NewOverhead(Resolve(nil, capacity), capacity)
		require.NoError(t, err)
		require.NotContains(t, o.KubeReserved, corev1.ResourceEphemeralStorage)
		require.NotContains(t, o.EvictionThreshold, corev1.ResourceEphemeralStorage)

		got := resources.Subtract(capacity, o.Total())
		requireQuantity(t, "0", got[corev1.ResourceEphemeralStorage])
		requireQuantity(t, "13590Mi", got[corev1.ResourceMemory])
		require.True(t, resources.Fits(corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}, got))
	})
}
