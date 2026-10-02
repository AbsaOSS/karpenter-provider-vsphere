package overhead

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// kubeletConfig is the part of kubelet.config.k8s.io/v1beta1 KubeletConfiguration that we set.
type kubeletConfig struct {
	APIVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	MaxPods        int64             `yaml:"maxPods,omitempty"`
	KubeReserved   map[string]string `yaml:"kubeReserved,omitempty"`
	SystemReserved map[string]string `yaml:"systemReserved,omitempty"`
	EvictionHard   map[string]string `yaml:"evictionHard,omitempty"`
}

// KubeletConfigFile renders the merged values as a KubeletConfiguration file for the node.
// It is built from the same Resolved as NewOverhead, so the node reserves what Karpenter assumed.
func KubeletConfigFile(r Resolved) (string, error) {
	out, err := yaml.Marshal(kubeletConfig{
		APIVersion:     "kubelet.config.k8s.io/v1beta1",
		Kind:           "KubeletConfiguration",
		MaxPods:        r.MaxPods,
		KubeReserved:   r.KubeReserved,
		SystemReserved: r.SystemReserved,
		EvictionHard:   r.EvictionHard,
	})
	if err != nil {
		return "", fmt.Errorf("rendering kubelet config: %w", err)
	}
	return string(out), nil
}
