package kwok

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
)

// evictionSignalKeys maps eviction signal names to corev1.ResourceName.
var evictionSignalKeys = map[string]corev1.ResourceName{
	"memory.available":  corev1.ResourceMemory,
	"nodefs.available":  corev1.ResourceEphemeralStorage,
	"imagefs.available": corev1.ResourceEphemeralStorage,
	"pid.available":     corev1.ResourcePods,
}

// parseResourceList converts a resource map (e.g. kubeReserved, systemReserved)
// into a corev1.ResourceList, treating each key as a corev1.ResourceName verbatim
// (allowing extended resources) and ignoring unparsable values.
func parseResourceList(resourceMap map[string]string) corev1.ResourceList {
	list := corev1.ResourceList{}
	for key, value := range resourceMap {
		if qty, err := resource.ParseQuantity(value); err == nil {
			list[corev1.ResourceName(key)] = qty
		}
	}
	return list
}

// parseEvictionThreshold converts eviction signal names (e.g. nodefs.available) into a
// corev1.ResourceList via evictionSignalKeys, ignoring unknown keys and unparsable values.
func parseEvictionThreshold(evictionHard map[string]string) corev1.ResourceList {
	list := corev1.ResourceList{}
	for key, value := range evictionHard {
		resourceName, ok := evictionSignalKeys[key]
		if !ok {
			continue
		}
		if qty, err := resource.ParseQuantity(value); err == nil {
			list[resourceName] = qty
		}
	}
	return list
}

// ToInstanceTypeOverhead extracts kubelet overhead from KubeletConfiguration
// and returns an InstanceTypeOverhead struct. It processes kubeReserved, systemReserved,
// and evictionHard fields from the kubelet configuration.
func ToInstanceTypeOverhead(kc *v1alpha1.KubeletConfiguration) *cloudprovider.InstanceTypeOverhead {
	if kc == nil {
		return &cloudprovider.InstanceTypeOverhead{}
	}

	overhead := &cloudprovider.InstanceTypeOverhead{
		KubeReserved:      parseResourceList(kc.KubeReserved),
		SystemReserved:    parseResourceList(kc.SystemReserved),
		EvictionThreshold: parseEvictionThreshold(kc.EvictionHard),
	}

	if len(overhead.KubeReserved) == 0 && len(overhead.SystemReserved) == 0 && len(overhead.EvictionThreshold) == 0 {
		return &cloudprovider.InstanceTypeOverhead{}
	}

	return overhead
}
