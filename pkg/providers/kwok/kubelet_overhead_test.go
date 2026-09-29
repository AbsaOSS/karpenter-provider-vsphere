package kwok

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
)

func TestParseResourceList(t *testing.T) {
	t.Run("parses known keys", func(t *testing.T) {
		list := parseResourceList(map[string]string{"cpu": "500m", "memory": "256Mi"})
		assert.Equal(t, resource.MustParse("500m"), list[corev1.ResourceCPU])
		assert.Equal(t, resource.MustParse("256Mi"), list[corev1.ResourceMemory])
	})

	t.Run("accepts extended resource names verbatim", func(t *testing.T) {
		list := parseResourceList(map[string]string{"example.com/foo": "1"})
		assert.Equal(t, resource.MustParse("1"), list[corev1.ResourceName("example.com/foo")])
	})

	t.Run("ignores unparsable values", func(t *testing.T) {
		list := parseResourceList(map[string]string{"cpu": "not-a-quantity"})
		assert.Empty(t, list)
	})

	t.Run("empty map returns empty list", func(t *testing.T) {
		list := parseResourceList(map[string]string{})
		assert.Empty(t, list)
	})
}

func TestToInstanceTypeOverhead(t *testing.T) {
	t.Run("nil KubeletConfiguration returns empty overhead", func(t *testing.T) {
		overhead := ToInstanceTypeOverhead(nil)
		assert.Empty(t, overhead.KubeReserved)
		assert.Empty(t, overhead.SystemReserved)
		assert.Empty(t, overhead.EvictionThreshold)
	})

	t.Run("empty KubeletConfiguration returns empty overhead", func(t *testing.T) {
		overhead := ToInstanceTypeOverhead(&v1alpha1.KubeletConfiguration{})
		assert.Empty(t, overhead.KubeReserved)
		assert.Empty(t, overhead.SystemReserved)
		assert.Empty(t, overhead.EvictionThreshold)
	})

	t.Run("populates KubeReserved and SystemReserved from resource names", func(t *testing.T) {
		kc := &v1alpha1.KubeletConfiguration{
			KubeReserved:   map[string]string{"cpu": "1", "memory": "1Gi"},
			SystemReserved: map[string]string{"cpu": "500m", "ephemeral-storage": "2Gi"},
		}

		overhead := ToInstanceTypeOverhead(kc)

		assert.Equal(t, resource.MustParse("1"), overhead.KubeReserved[corev1.ResourceCPU])
		assert.Equal(t, resource.MustParse("1Gi"), overhead.KubeReserved[corev1.ResourceMemory])
		assert.Equal(t, resource.MustParse("500m"), overhead.SystemReserved[corev1.ResourceCPU])
		assert.Equal(t, resource.MustParse("2Gi"), overhead.SystemReserved[corev1.ResourceEphemeralStorage])
	})

	t.Run("populates EvictionThreshold from eviction signal names, not resource names", func(t *testing.T) {
		// nodefs.available and imagefs.available both map to ResourceEphemeralStorage,
		// so they are tested separately to avoid asserting on a map-key collision.
		kc := &v1alpha1.KubeletConfiguration{
			EvictionHard: map[string]string{
				"memory.available": "100Mi",
				"pid.available":    "500",
			},
		}

		overhead := ToInstanceTypeOverhead(kc)

		assert.Equal(t, resource.MustParse("100Mi"), overhead.EvictionThreshold[corev1.ResourceMemory])
		assert.Equal(t, resource.MustParse("500"), overhead.EvictionThreshold[corev1.ResourcePods])
	})

	t.Run("maps nodefs.available to ephemeral-storage", func(t *testing.T) {
		kc := &v1alpha1.KubeletConfiguration{EvictionHard: map[string]string{"nodefs.available": "1Gi"}}
		overhead := ToInstanceTypeOverhead(kc)
		assert.Equal(t, resource.MustParse("1Gi"), overhead.EvictionThreshold[corev1.ResourceEphemeralStorage])
	})

	t.Run("maps imagefs.available to ephemeral-storage", func(t *testing.T) {
		kc := &v1alpha1.KubeletConfiguration{EvictionHard: map[string]string{"imagefs.available": "2Gi"}}
		overhead := ToInstanceTypeOverhead(kc)
		assert.Equal(t, resource.MustParse("2Gi"), overhead.EvictionThreshold[corev1.ResourceEphemeralStorage])
	})

	t.Run("eviction signal keys are unrelated to kubeReserved/systemReserved resource keys", func(t *testing.T) {
		// "cpu" is a valid kubeReserved/systemReserved key but not a recognized eviction signal.
		kc := &v1alpha1.KubeletConfiguration{
			EvictionHard: map[string]string{"cpu": "100m"},
		}

		overhead := ToInstanceTypeOverhead(kc)

		assert.Empty(t, overhead.EvictionThreshold)
	})
}
