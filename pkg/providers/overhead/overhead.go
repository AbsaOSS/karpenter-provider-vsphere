package overhead

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/utils/resources"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
)

const (
	evictionMemoryAvailable = "memory.available"
	evictionNodeFSAvailable = "nodefs.available"
)

var defaultEvictionHard = map[string]string{
	evictionMemoryAvailable: "100Mi",
	evictionNodeFSAvailable: "10%",
	"nodefs.inodesFree":     "5%",
	"imagefs.available":     "15%",
	"imagefs.inodesFree":    "5%",
}

type Resolved struct {
	MaxPods        int64
	KubeReserved   map[string]string
	SystemReserved map[string]string
	EvictionHard   map[string]string
}

func Memory(ctx context.Context, configured resource.Quantity) *resource.Quantity {
	var percent float64
	if opts := options.FromContext(ctx); opts != nil {
		percent = opts.VMMemoryOverheadPercent
	}
	mib := configured.Value() / (1024 * 1024)
	overheadMiB := int64(math.Ceil(float64(mib) * percent))
	return resources.Quantity(fmt.Sprintf("%dMi", mib-overheadMiB))
}

func defaultKubeReserved(cpus, pods *resource.Quantity) map[string]string {
	cpu := resource.NewMilliQuantity(0, resource.DecimalSI)
	for _, tier := range []struct {
		start, end int64
		percentage float64
	}{
		{start: 0, end: 1000, percentage: 0.06},
		{start: 1000, end: 2000, percentage: 0.01},
		{start: 2000, end: 4000, percentage: 0.005},
		{start: 4000, end: 1 << 31, percentage: 0.0025},
	} {
		if m := cpus.MilliValue(); m >= tier.start {
			span := float64(tier.end - tier.start)
			if m < tier.end {
				span = float64(m - tier.start)
			}
			cpu.Add(*resource.NewMilliQuantity(int64(span*tier.percentage), resource.DecimalSI))
		}
	}
	return map[string]string{
		string(corev1.ResourceCPU):              string(cpu.String()),
		string(corev1.ResourceMemory):           fmt.Sprintf("%dMi", 11*pods.Value()+255),
		string(corev1.ResourceEphemeralStorage): "1Gi",
	}
}

func Resolve(kc *v1alpha1.KubeletConfiguration, capacity corev1.ResourceList) Resolved {
	return Resolved{
		MaxPods:        capacity.Pods().Value(),
		KubeReserved:   lo.Assign(defaultKubeReserved(capacity.Cpu(), capacity.Pods()), kc.GetKubeReserved()),
		SystemReserved: lo.Assign(kc.GetSystemReserved()),
		EvictionHard:   lo.Assign(defaultEvictionHard, kc.GetEvictionHard()),
	}
}

func computeEvictionSignal(capacity resource.Quantity, value string) (resource.Quantity, error) {
	if strings.HasSuffix(value, "%") {
		p, err := strconv.ParseFloat(strings.TrimSuffix(value, "%"), 64)
		if err != nil {
			return resource.Quantity{}, fmt.Errorf("invalid percentage %q: %w", value, err)
		}
		if p == 100 {
			p = 0
		}
		return *resource.NewQuantity(int64(math.Ceil(capacity.AsApproximateFloat64()/100*p)), resource.BinarySI), nil
	}
	return resource.ParseQuantity(value)
}

func parseResourceMap(m map[string]string) (corev1.ResourceList, error) {
	out := corev1.ResourceList{}
	for k, v := range m {
		q, err := resource.ParseQuantity(v)
		if err != nil {
			return nil, fmt.Errorf("%s=%q: %w", k, v, err)
		}
		out[corev1.ResourceName(k)] = q
	}
	return out, nil
}

func NewOverhead(r Resolved, capacity corev1.ResourceList) (*cloudprovider.InstanceTypeOverhead, error) {
	kube, err := parseResourceMap(r.KubeReserved)
	if err != nil {
		return nil, fmt.Errorf("kubeReserved: %w", err)
	}
	sys, err := parseResourceMap(r.SystemReserved)
	if err != nil {
		return nil, fmt.Errorf("systemReserved: %w", err)
	}
	memory, err := computeEvictionSignal(*capacity.Memory(), r.EvictionHard[evictionMemoryAvailable])
	if err != nil {
		return nil, fmt.Errorf("evictionHard: %s: %w", evictionMemoryAvailable, err)
	}
	storage, err := computeEvictionSignal(*capacity.StorageEphemeral(), r.EvictionHard[evictionNodeFSAvailable])
	if err != nil {
		return nil, fmt.Errorf("evictionHard: %s: %w", evictionNodeFSAvailable, err)
	}
	return &cloudprovider.InstanceTypeOverhead{
		KubeReserved:   kube,
		SystemReserved: sys,
		EvictionThreshold: corev1.ResourceList{
			corev1.ResourceMemory:           memory,
			corev1.ResourceEphemeralStorage: storage,
		},
	}, nil
}
