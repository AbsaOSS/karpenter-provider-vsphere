package instance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/vim25/types"
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

func TestGenerateVMName(t *testing.T) {
	require.Equal(t, "cluster-karp-claim", GenerateVMName("cluster", "claim"))
}

func TestNewInstance(t *testing.T) {
	created := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tags := map[string]string{
		v1alpha1.LabelInstanceSize: "small",
	}

	instance := NewInstance(nil, "vm-uuid", "template-path", "poweredOn", "vm-name", created, tags)

	require.Equal(t, created, instance.LaunchTime)
	require.Equal(t, "vm-uuid", instance.ID)
	require.Equal(t, "template-path", instance.Image)
	require.Equal(t, "poweredOn", instance.State)
	require.Equal(t, "vm-name", instance.Name)
	require.Equal(t, "small", instance.Type)
	require.Equal(t, tags, instance.Tags)
	require.Nil(t, instance.GetVM())
}

func TestConfigExtract(t *testing.T) {
	config := Config{
		&types.OptionValue{Key: "key", Value: "value"},
	}

	require.Equal(t, []types.BaseOptionValue(config), config.Extract())
	var nilConfig *Config
	require.Nil(t, nilConfig.Extract())
}

func TestImageFromAnnotation(t *testing.T) {
	tests := []struct {
		name   string
		config *types.VirtualMachineConfigInfo
		want   string
	}{
		{name: "nil config", want: "image_not_found"},
		{name: "empty annotation", config: &types.VirtualMachineConfigInfo{}, want: ""},
		{name: "image path", config: &types.VirtualMachineConfigInfo{Annotation: "/dc0/vm/flatcar-template"}, want: "/dc0/vm/flatcar-template"},
		{name: "prefixed image path", config: &types.VirtualMachineConfigInfo{Annotation: "cloned_from:/dc0/vm/flatcar-template"}, want: "/dc0/vm/flatcar-template"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, imageFromConfig(test.config))
		})
	}
}

func TestBelongsToCluster(t *testing.T) {
	tests := []struct {
		name        string
		tags        map[string]string
		clusterName string
		want        bool
	}{
		{name: "matching cluster", tags: map[string]string{v1alpha1.ClusterNameTagKey: "cluster-a"}, clusterName: "cluster-a", want: true},
		{name: "different cluster", tags: map[string]string{v1alpha1.ClusterNameTagKey: "cluster-b"}, clusterName: "cluster-a", want: false},
		{name: "legacy misspelled key is ignored", tags: map[string]string{"karpneter.sh/clustername": "cluster-a"}, clusterName: "cluster-a", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, belongsToCluster(test.tags, test.clusterName))
		})
	}
}
