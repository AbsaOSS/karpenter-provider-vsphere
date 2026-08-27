package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func TestGenerateSpec(t *testing.T) {
	instanceType := &corecloudprovider.InstanceType{
		Name: "vsphere-vm.cpu-1.mem-64gb.os-linux",
		Capacity: corev1.ResourceList{
			corev1.ResourceCPU:              resource.MustParse("1"),
			corev1.ResourceMemory:           resource.MustParse("64Gi"),
			corev1.ResourceEphemeralStorage: resource.MustParse(utils.GiToByteAsString(5)),
		},
	}
	expectedMemInMB := int64(65536) // 64 GiB in megabytes
	mem := utils.InstanceTypeToMegabytes(instanceType.Capacity.Memory())
	assert.Equal(t, expectedMemInMB, mem)
}
