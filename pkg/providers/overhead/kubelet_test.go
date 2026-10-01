package overhead

import (
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestKubeletConfigFile(t *testing.T) {
	t.Run("defaults only", func(t *testing.T) {
		got, err := KubeletConfigFile(Resolve(nil, testCapacity()))
		require.NoError(t, err)
		require.Equal(t, `apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
maxPods: 110
kubeReserved:
    cpu: 80m
    ephemeral-storage: 1Gi
    memory: 1465Mi
evictionHard:
    imagefs.available: 15%
    imagefs.inodesFree: 5%
    memory.available: 100Mi
    nodefs.available: 10%
    nodefs.inodesFree: 5%
`, got)
	})

	t.Run("with spec.kubelet", func(t *testing.T) {
		got, err := KubeletConfigFile(Resolve(&v1alpha1.KubeletConfiguration{
			KubeReserved:   map[string]string{"memory": "1Gi"},
			SystemReserved: map[string]string{"cpu": "100m", "memory": "256Mi"},
			EvictionHard:   map[string]string{"memory.available": "5%"},
		}, testCapacity()))
		require.NoError(t, err)

		var parsed kubeletConfig
		require.NoError(t, yaml.Unmarshal([]byte(got), &parsed))
		require.Equal(t, "kubelet.config.k8s.io/v1beta1", parsed.APIVersion)
		require.Equal(t, "KubeletConfiguration", parsed.Kind)
		require.Equal(t, int64(110), parsed.MaxPods)
		require.Equal(t, map[string]string{"cpu": "80m", "memory": "1Gi", "ephemeral-storage": "1Gi"}, parsed.KubeReserved)
		require.Equal(t, map[string]string{"cpu": "100m", "memory": "256Mi"}, parsed.SystemReserved)
		require.Equal(t, "5%", parsed.EvictionHard["memory.available"])
		require.Equal(t, "10%", parsed.EvictionHard["nodefs.available"])
	})

	t.Run("output is stable", func(t *testing.T) {
		first, err := KubeletConfigFile(Resolve(nil, testCapacity()))
		require.NoError(t, err)
		for range 20 {
			again, err := KubeletConfigFile(Resolve(nil, testCapacity()))
			require.NoError(t, err)
			require.Equal(t, first, again)
		}
	})
}
