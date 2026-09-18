package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/cloudprovider"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kwok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// setupCloudProvider builds a CloudProvider on top of the same vcsim-backed
// instance.DefaultProvider used by the instance provider e2e tests (see
// setupInstanceProvider in instance_test.go), adding a fake kube client
// seeded with a Ready NodeClass and the real KWOK instance type catalog so
// instance-type resolution (CloudProvider.resolveInstanceTypes) and
// selection (instance.DefaultProvider.selectInstanceType) can be exercised
// together through CloudProvider.Create.
func setupCloudProvider(t *testing.T) (*cloudprovider.CloudProvider, *v1alpha1.VsphereNodeClass, context.Context) {
	t.Helper()

	instanceProvider, class, ctx := setupInstanceProvider(t)

	// Mark the NodeClass Ready so CloudProvider.Create doesn't short-circuit
	// on NodeClassReadinessUnknown / NodeClassNotReady before ever reaching
	// instance type resolution.
	class.Status.KubernetesVersion = "v1.30.0"
	class.StatusConditions().SetTrue(v1alpha1.ConditionTypeKubernetesVersionReady)

	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(class).
		Build()

	cp := cloudprovider.New(instanceProvider, kubeClient, kwok.KwokInstanceTypesStaticProvider{})
	return cp, class, ctx
}

// sumPodResources aggregates a set of pod resource requests, mirroring what
// Karpenter's scheduler would compute for the pods it bin-packed onto a
// single node before calling CloudProvider.Create with that NodeClaim.
func sumPodResources(pods ...corev1.ResourceList) corev1.ResourceList {
	cpu := resource.Quantity{}
	mem := resource.Quantity{}
	for _, pod := range pods {
		cpu.Add(*pod.Cpu())
		mem.Add(*pod.Memory())
	}
	return corev1.ResourceList{
		corev1.ResourceCPU:    cpu,
		corev1.ResourceMemory: mem,
	}
}

// TestCloudProviderCreate_ChoosesMemoryOptimizedFamily exercises
// CloudProvider.Create end-to-end with a NodeClaim shaped like the result of
// Karpenter bin-packing 8 pods requesting 1 CPU/8Gi and 3 pods requesting 1
// CPU/9Gi onto a single node (11 CPU, 91Gi memory total requested).
//
// Per the KWOK pricing model (see
// https://github.com/kubernetes-sigs/karpenter/blob/1228db7c7ff5254a43808eb2ba90b72cc27fcd30/designs/kwok-provider.md#pricing),
// instance types are named "<family>-<size>x", where size halves to the
// vCPU count and family sets the memory:vCPU ratio: c(ompute)=2,
// s(tandard)=4, m(emory)=8. With the supported sizes {2,4,9,16,32}, the
// largest catalog entries are c-32x (16 vCPU / 32Gi), s-32x (16 vCPU / 64Gi)
// and m-32x (16 vCPU / 128Gi) - only m-32x has enough memory (128Gi >= 91Gi)
// to fit this NodeClaim's aggregate request, so CloudProvider.resolveInstanceTypes
// must filter every other candidate out, leaving
// instance.DefaultProvider.selectInstanceType with the memory family as the
// only viable pick.
func TestCloudProviderCreate_ChoosesMemoryOptimizedFamily(t *testing.T) {
	cp, class, ctx := setupCloudProvider(t)

	pods := make([]corev1.ResourceList, 0, 11)
	for i := 0; i < 8; i++ {
		pods = append(pods, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		})
	}
	for i := 0; i < 3; i++ {
		pods = append(pods, corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("9Gi"),
		})
	}

	nodeClaim := &karpv1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "claim-memory-family",
			Labels: map[string]string{karpv1.NodePoolLabelKey: "default"},
		},
		Spec: karpv1.NodeClaimSpec{
			NodeClassRef: &karpv1.NodeClassReference{
				Kind: "VsphereNodeClass",
				Name: class.Name,
			},
			Resources: karpv1.ResourceRequirements{
				Requests: sumPodResources(pods...),
			},
			Requirements: []karpv1.NodeSelectorRequirementWithMinValues{
				{
					Key:      corev1.LabelInstanceTypeStable,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"m-32x", "m-16x", "m-9x", "m-4x", "m-2x", "c-32x", "c-16x", "c-9x", "c-4x", "c-2x", "s-32x", "s-16x", "s-9x", "s-4x", "s-2x", "vsphere-vm.cpu-10.mem-64gb.os-linux", "vsphere-vm.cpu-16.mem-64gb.os-linux"},
				},
			},
		},
	}

	created, err := cp.Create(ctx, nodeClaim)
	require.NoError(t, err)
	require.NotNil(t, created)

	instanceType := created.Labels[corev1.LabelInstanceTypeStable]
	require.NotEmpty(t, instanceType, "expected the instance-type label to be set on the created NodeClaim")

	assert.True(t, strings.HasPrefix(instanceType, string(kwok.InstanceFamilyMemory)+"-"),
		"expected a memory-optimized instance type (m-<size>x) to be chosen for a workload whose memory:vCPU ratio (~8.3) matches the memory family, got %q", instanceType)
	assert.Equal(t, "m-32x", instanceType)
}

// Checks that the cloud provider correctly filters instance types based on the requirements specified in the NodeClaim.
// If node claim has minimum 9x size instance type, then not 2x size selected when 1 CPU and 1Gi memory pod requested.
func TestCloudProviderCreate_FiltersInstanceTypesWithRequirementsFromNodeClaim(t *testing.T) {
	cp, class, ctx := setupCloudProvider(t)

	nodeClaim := &karpv1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "claim-minimal",
			Labels: map[string]string{karpv1.NodePoolLabelKey: "default"},
		},
		Spec: karpv1.NodeClaimSpec{
			NodeClassRef: &karpv1.NodeClassReference{
				Kind: "VsphereNodeClass",
				Name: class.Name,
			},
			Resources: karpv1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			},
			Requirements: []karpv1.NodeSelectorRequirementWithMinValues{
				{
					Key:      corev1.LabelInstanceTypeStable,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"m-32x", "m-16x", "m-9x", "c-32x", "c-16x", "c-9x", "s-32x", "s-16x", "s-9x", "vsphere-vm.cpu-4.mem-16gb.os-linux"},
				},
			},
		},
	}

	created, err := cp.Create(ctx, nodeClaim)
	require.NoError(t, err)
	require.NotNil(t, created)

	instanceType := created.Labels[corev1.LabelInstanceTypeStable]
	require.NotEmpty(t, instanceType, "expected the instance-type label to be set on the created NodeClaim")

	assert.Equal(t, "c-9x", instanceType)
}
