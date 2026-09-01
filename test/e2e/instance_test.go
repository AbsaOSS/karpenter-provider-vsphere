package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/test/vcsim"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

const testClusterName = "test-cluster"

// setupInstanceProvider starts an in-process vcsim vCenter (default VPX
// model), marks one of its demo VMs as a template, and wires everything up
// into an instance.DefaultProvider exactly the way the operator does. It
// returns the provider, a NodeClass pointing at that template/pool/
// datastore/network, and a context carrying the options every Create() call
// needs.
func setupInstanceProvider(t *testing.T) (*instance.DefaultProvider, *v1alpha1.VsphereNodeClass, context.Context) {
	t.Helper()
	ctx := context.Background()

	simulator, err := vcsim.New()
	require.NoError(t, err)
	t.Cleanup(simulator.Destroy)

	username := simulator.Username()
	password := simulator.Password()

	sess, err := vsphereclient.NewSession(ctx, simulator.ServerURL().Host, username, password, true)
	require.NoError(t, err)

	findClient := find.NewFinder(sess.Vim, false)
	dc, err := findClient.DefaultDatacenter(ctx)
	require.NoError(t, err)
	findClient.SetDatacenter(dc)

	// Use a standalone-host demo VM as our template. It sits outside the
	// "DC0_C0" resource pool we clone into, which more closely mirrors a
	// real setup where the template lives in its own place.
	templateName := "DC0_H0_VM0"
	templateVM, err := findClient.VirtualMachine(ctx, templateName)
	require.NoError(t, err)
	state, err := templateVM.PowerState(ctx)
	require.NoError(t, err)
	if state == types.VirtualMachinePowerStatePoweredOn {
		task, err := templateVM.PowerOff(ctx)
		require.NoError(t, err)
		require.NoError(t, task.Wait(ctx))
	}
	require.NoError(t, templateVM.MarkAsTemplate(ctx))

	// No sub-folder and ClusterName == testClusterName: newly cloned VMs
	// land directly in the root "/DC0/vm" folder and are named
	// "<testClusterName>-karp-<claimName>", so finder.ListVMs' prefix match
	// picks them up without any extra inventory setup.
	finderProvider := finder.NewDefaultProvider(sess, findClient, dc, "", testClusterName)

	provider := instance.NewDefaultProvider(nil, finderProvider, testClusterName)

	class := &v1alpha1.VsphereNodeClass{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
		Spec: v1alpha1.VsphereNodeClassSpec{
			PoolSelector:      v1alpha1.ResPoolSelctorTerm{Name: "DC0_C0"},
			DatastoreSelector: v1alpha1.DatastoreSelectorTerm{Name: "LocalDS_0"},
			NetworkSelector:   v1alpha1.NetworkSelectorTerm{Name: "VM Network"},
			ImageSelector:     v1alpha1.ImageSelectorTerm{Pattern: templateName},
			UserData:          v1alpha1.UserData{Type: v1alpha1.UserDataTypeCloudConfig},
			Tags: map[string]string{
				"topology.kubernetes.io/zone": "zone-a",
			},
		},
	}

	ctx = options.ToContext(ctx, &options.Options{
		ClusterName:     testClusterName,
		ClusterEndpoint: "https://127.0.0.1:6443",
		JoinToken:       "test-join-token",
		KubeDistro:      string(v1alpha1.RKE2),
		KubeVersion:     "v1.30.0",
	})

	return provider, class, ctx
}

func testInstanceTypes() []*corecloudprovider.InstanceType {
	return []*corecloudprovider.InstanceType{
		{
			Name: "test-type",
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
	}
}

func testNodeClaim(name string) *karpv1.NodeClaim {
	return &karpv1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{karpv1.NodePoolLabelKey: "default"},
		},
	}
}

func TestCreate(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	inst, err := provider.Create(ctx, class, testNodeClaim("claim1"), testInstanceTypes())
	require.NoError(t, err)
	require.NotNil(t, inst)

	assert.Equal(t, fmt.Sprintf("%s-karp-claim1", testClusterName), inst.Name)
	assert.NotEmpty(t, inst.ID, "expected the VM's BIOS UUID to be populated")
	assert.Equal(t, "poweredOn", inst.State)
	assert.Equal(t, "test-type", inst.Type)
	assert.Equal(t, testClusterName, inst.Tags[v1alpha1.ClusterNameTagKey])
	assert.Equal(t, "default", inst.Tags[karpv1.NodePoolLabelKey])
	// class.Spec.Tags should be merged in.
	assert.Equal(t, "zone-a", inst.Tags[corev1.LabelTopologyZone])
}

func TestCreate_UsesFirstInstanceType(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	instanceTypes := append(testInstanceTypes(), &corecloudprovider.InstanceType{
		Name: "second-type",
		Capacity: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	})

	inst, err := provider.Create(ctx, class, testNodeClaim("claim2"), instanceTypes)
	require.NoError(t, err)
	assert.Equal(t, "test-type", inst.Type)
}

// disk.EnableUUID must be true on cloned VMs, otherwise CSI cannot attach
// disks to karpenter-provisioned nodes by UUID (see commit ce8da36).
// vcsim's CloneVMTask does not propagate Config.Flags to the resulting VM
// (only a handful of fields are copied over), so this asserts on the
// generated CloneSpec rather than the round-tripped VM.
func TestCreate_EnablesDiskUUID(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	vmTemplate, err := provider.Finder.ResolveImage(ctx, class.Spec.ImageSelector)
	require.NoError(t, err)

	spec, err := provider.GenerateVMSpec(ctx, class, "test-vm", vmTemplate, testInstanceTypes()[0])
	require.NoError(t, err)

	require.NotNil(t, spec.Config.Flags, "expected the clone spec's flags to be set")
	require.NotNil(t, spec.Config.Flags.DiskUuidEnabled, "expected disk.EnableUUID to be set")
	assert.True(t, *spec.Config.Flags.DiskUuidEnabled, "disk.EnableUUID must be true so CSI can attach disks by UUID")
}

func TestGet(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	created, err := provider.Create(ctx, class, testNodeClaim("claim3"), testInstanceTypes())
	require.NoError(t, err)

	got, err := provider.Get(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)
}

func TestGet_NotFound(t *testing.T) {
	provider, _, ctx := setupInstanceProvider(t)

	_, err := provider.Get(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	created, err := provider.Create(ctx, class, testNodeClaim("claim4"), testInstanceTypes())
	require.NoError(t, err)

	instances, err := provider.List(ctx)
	require.NoError(t, err)

	names := make([]string, 0, len(instances))
	for _, i := range instances {
		names = append(names, i.Name)
	}
	assert.Contains(t, names, created.Name)
}

func TestDelete(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	created, err := provider.Create(ctx, class, testNodeClaim("claim5"), testInstanceTypes())
	require.NoError(t, err)

	require.NoError(t, provider.Delete(ctx, created.ID))

	_, err = provider.Get(ctx, created.ID)
	assert.Error(t, err, "expected the VM to be gone after Delete")

	instances, err := provider.List(ctx)
	require.NoError(t, err)
	for _, i := range instances {
		assert.NotEqual(t, created.Name, i.Name, "deleted VM should not be listed")
	}
}

func TestDelete_NotFound(t *testing.T) {
	provider, _, ctx := setupInstanceProvider(t)

	err := provider.Delete(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestInstanceProvider_ReAuthenticatesAfterSessionInvalidation(t *testing.T) {
	provider, class, ctx := setupInstanceProvider(t)

	created, err := provider.Create(ctx, class, testNodeClaim("claim6"), testInstanceTypes())
	require.NoError(t, err)

	// Simulate vCenter tearing the session down from underneath us; Create,
	// List, Get and Delete must each recover via Session.EnsureValid instead
	// of failing with an authentication error.
	require.NoError(t, session.NewManager(provider.Finder.Session.Vim).Logout(ctx))
	require.NoError(t, provider.Finder.Session.Rest.Logout(ctx))

	_, err = provider.List(ctx)
	require.NoError(t, err, "List should re-authenticate rather than fail")

	_, err = provider.Get(ctx, created.ID)
	require.NoError(t, err, "Get should re-authenticate rather than fail")

	_, err = provider.Create(ctx, class, testNodeClaim("claim7"), testInstanceTypes())
	require.NoError(t, err, "Create should re-authenticate rather than fail")

	require.NoError(t, provider.Delete(ctx, created.ID), "Delete should re-authenticate rather than fail")
}
