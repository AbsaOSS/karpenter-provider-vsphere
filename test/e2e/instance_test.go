package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/test/vcsim"
	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// Instance provider test constants
const (
	InstanceTestClusterName  = "test-cluster"
	InstanceTypeName         = "test-type"
	NodeClassName            = "default"
	InstanceZoneValue        = "zone-a"
	InstanceTemplateVMName   = "DC0_H0_VM0"
	InstancePoolSelectorName = "DC0_C0"
	InstanceDatastoreName    = "LocalDS_0"
	InstanceNetworkName      = "VM Network"
	InstanceRegionValue      = "region-a"
	InstanceClusterEndpoint  = "https://127.0.0.1:6443"
	InstanceJoinToken        = "test-join-token"
	InstanceKubeVersion      = "v1.30.0"
)

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
	templateVM, err := findClient.VirtualMachine(ctx, InstanceTemplateVMName)
	require.NoError(t, err)
	state, err := templateVM.PowerState(ctx)
	require.NoError(t, err)
	if state == types.VirtualMachinePowerStatePoweredOn {
		task, err := templateVM.PowerOff(ctx)
		require.NoError(t, err)
		require.NoError(t, task.Wait(ctx))
	}
	require.NoError(t, templateVM.MarkAsTemplate(ctx))

	// No sub-folder and ClusterName == InstanceTestClusterName: newly cloned VMs
	// land directly in the root "/DC0/vm" folder and are named
	// "<InstanceTestClusterName>-karp-<claimName>", so finder.ListVMs' prefix match
	// picks them up without any extra inventory setup.
	finderProvider := finder.NewDefaultProvider(sess, findClient, dc, "", InstanceTestClusterName)

	provider := instance.NewDefaultProvider(nil, finderProvider, InstanceTestClusterName)

	class := &v1alpha1.VsphereNodeClass{
		ObjectMeta: metav1.ObjectMeta{Name: NodeClassName},
		Spec: v1alpha1.VsphereNodeClassSpec{
			PoolSelector:      v1alpha1.ResPoolSelctorTerm{Name: InstancePoolSelectorName},
			DatastoreSelector: v1alpha1.DatastoreSelectorTerm{Name: InstanceDatastoreName},
			NetworkSelector:   v1alpha1.NetworkSelectorTerm{Name: InstanceNetworkName},
			ImageSelector:     v1alpha1.ImageSelectorTerm{Pattern: InstanceTemplateVMName},
			UserData:          v1alpha1.UserData{Type: v1alpha1.UserDataTypeCloudConfig},
			Tags: map[string]string{
				"topology.kubernetes.io/zone": InstanceZoneValue,
			},
		},
	}

	ctx = options.ToContext(ctx, &options.Options{
		ClusterName:     InstanceTestClusterName,
		ClusterEndpoint: InstanceClusterEndpoint,
		JoinToken:       InstanceJoinToken,
		KubeDistro:      string(v1alpha1.RKE2),
		KubeVersion:     InstanceKubeVersion,
		Zone:            InstanceZoneValue,
		Region:          InstanceRegionValue,
	})

	return provider, class, ctx
}

func testInstanceTypes() []*corecloudprovider.InstanceType {
	return []*corecloudprovider.InstanceType{
		testutil.NewInstanceType().
			WithName(InstanceTypeName).
			WithCPU("2").
			WithMemory("512Mi"). // vcsim's default pool caps memory at 961Mi
			WithZone(InstanceZoneValue).
			WithOfferingSchedulingCapacityTypeOnDemand().
			WithPrice(1.0).
			Build(),
	}
}

func testNodeClaim(name string) *karpv1.NodeClaim {
	return &karpv1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{karpv1.NodePoolLabelKey: NodeClassName},
		},
	}
}

func TestCreate(t *testing.T) {
	const expectedState = "poweredOn"
	const claimName = "claim1"
	provider, class, ctx := setupInstanceProvider(t)

	inst, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)
	require.NotNil(t, inst)

	assert.Equal(t, fmt.Sprintf("%s-karp-%s", InstanceTestClusterName, claimName), inst.Name)
	assert.NotEmpty(t, inst.ID, "expected the VM's BIOS UUID to be populated")
	assert.Equal(t, expectedState, inst.State)
	assert.Equal(t, InstanceTypeName, inst.Type)
	assert.Equal(t, InstanceTestClusterName, inst.Tags[v1alpha1.ClusterNameTagKey])
	assert.Equal(t, NodeClassName, inst.Tags[karpv1.NodePoolLabelKey])
	// class.Spec.Tags should be merged in.
	assert.Equal(t, InstanceZoneValue, inst.Tags[corev1.LabelTopologyZone])
	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func TestCreate_UsesFirstInstanceType(t *testing.T) {
	const claimName = "claim2"
	provider, class, ctx := setupInstanceProvider(t)

	instanceTypes := append(testInstanceTypes(), testutil.NewInstanceType().
		WithName("second-type").
		WithCPU("4").
		WithMemory("8Gi").
		Build())

	inst, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), instanceTypes)
	require.NoError(t, err)
	assert.Equal(t, InstanceTypeName, inst.Type)
	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

// disk.EnableUUID must be true on cloned VMs, otherwise CSI cannot attach
// disks to karpenter-provisioned nodes by UUID (see commit ce8da36).
// vcsim's CloneVMTask does not propagate Config.Flags to the resulting VM
// (only a handful of fields are copied over), so this asserts on the
// generated CloneSpec rather than the round-tripped VM.
func TestCreate_EnablesDiskUUID(t *testing.T) {
	const testVMName = "test-vm"
	provider, class, ctx := setupInstanceProvider(t)

	vmTemplate, err := provider.Finder.ResolveImage(ctx, class.Spec.ImageSelector)
	require.NoError(t, err)

	spec, err := provider.GenerateVMSpec(ctx, class, testVMName, vmTemplate, testInstanceTypes()[0])
	require.NoError(t, err)

	require.NotNil(t, spec.Config.Flags, "expected the clone spec's flags to be set")
	require.NotNil(t, spec.Config.Flags.DiskUuidEnabled, "expected disk.EnableUUID to be set")
	assert.True(t, *spec.Config.Flags.DiskUuidEnabled, "disk.EnableUUID must be true so CSI can attach disks by UUID")
}

// RKE2 expects the singular "node-taint" key (see
// https://docs.rke2.io/install/configuration); "node-taints" is silently
// ignored by the agent (see commit b7bfcfd). Unlike the userdata package's
// template-level tests, this asserts on the actual guestinfo.userdata payload
// vcsim recorded on the cloned VM, since ExtraConfig (unlike Flags) is
// propagated by vcsim's CloneVMTask.
func TestCreate_RKE2UserDataUsesSingularNodeTaintKey(t *testing.T) {
	const claimName = "claim-node-taint"
	const extraConfigKey = "config.extraConfig"
	const userdataKey = "guestinfo.userdata"
	const expectedNodeTaintKey = "node-taint:"
	const unexpectedNodeTaintKey = "node-taints:"
	provider, class, ctx := setupInstanceProvider(t)

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)

	vm, err := provider.Finder.GetVMByID(ctx, created.ID)
	require.NoError(t, err)

	var vmMo mo.VirtualMachine
	require.NoError(t, vm.Properties(ctx, vm.Reference(), []string{extraConfigKey}, &vmMo))

	userData := extraConfigValue(vmMo.Config.ExtraConfig, userdataKey)
	require.NotEmpty(t, userData, "expected guestinfo.userdata to be set on the cloned VM")

	decoded, err := base64.StdEncoding.DecodeString(userData)
	require.NoError(t, err)

	assert.Contains(t, string(decoded), expectedNodeTaintKey)
	assert.NotContains(t, string(decoded), unexpectedNodeTaintKey)

	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func extraConfigValue(extraConfig []types.BaseOptionValue, key string) string {
	for _, ov := range extraConfig {
		opt := ov.GetOptionValue()
		if opt.Key == key {
			if s, ok := opt.Value.(string); ok {
				return s
			}
		}
	}
	return ""
}

// commit 69f85d4 added ignition userdata support (butane rendered to
// Ignition JSON); this exercises Create() end-to-end with UserDataType
// Ignition and decodes the real guestinfo.ignition.config.data vcsim
// recorded on the cloned VM.
func TestCreate_IgnitionUserData(t *testing.T) {
	const claimName = "claim-ignition"
	const extraConfigKey = "config.extraConfig"
	const ignitionDataKey = "guestinfo.ignition.config.data"
	provider, class, ctx := setupInstanceProvider(t)
	class.Spec.UserData.Type = v1alpha1.UserDataTypeIgnition

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)

	vm, err := provider.Finder.GetVMByID(ctx, created.ID)
	require.NoError(t, err)

	var vmMo mo.VirtualMachine
	require.NoError(t, vm.Properties(ctx, vm.Reference(), []string{extraConfigKey}, &vmMo))

	ignitionData := extraConfigValue(vmMo.Config.ExtraConfig, ignitionDataKey)
	require.NotEmpty(t, ignitionData, "expected guestinfo.ignition.config.data to be set on the cloned VM")

	decoded, err := base64.StdEncoding.DecodeString(ignitionData)
	require.NoError(t, err)

	var ign map[string]any
	require.NoError(t, json.Unmarshal(decoded, &ign), "expected valid Ignition JSON")

	storage, ok := ign["storage"].(map[string]any)
	require.True(t, ok, "expected an ignition storage section")
	files, ok := storage["files"].([]any)
	require.True(t, ok, "expected ignition storage.files to be a list")
	assert.NotEmpty(t, files, "expected the rendered config.yaml and node-join.sh files to be present")

	systemd, ok := ign["systemd"].(map[string]any)
	require.True(t, ok, "expected an ignition systemd section")
	units, ok := systemd["units"].([]any)
	require.True(t, ok, "expected ignition systemd.units to be a list")
	assert.NotEmpty(t, units, "expected the node-join.service unit to be present")

	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func TestGet(t *testing.T) {
	const claimName = "claim3"
	provider, class, ctx := setupInstanceProvider(t)

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)

	got, err := provider.Get(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, created.ID, got.ID)

	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func TestGet_NotFound(t *testing.T) {
	provider, _, ctx := setupInstanceProvider(t)

	_, err := provider.Get(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	const claimName = "claim4"
	provider, class, ctx := setupInstanceProvider(t)

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)

	instances, err := provider.List(ctx)
	require.NoError(t, err)

	names := make([]string, 0, len(instances))
	for _, i := range instances {
		names = append(names, i.Name)
	}
	assert.Contains(t, names, created.Name)
	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func TestDelete(t *testing.T) {
	const claimName = "claim5"
	provider, class, ctx := setupInstanceProvider(t)

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName), testInstanceTypes())
	require.NoError(t, err)

	require.NoError(t, provider.Delete(ctx, created.ID))

	_, err = provider.Get(ctx, created.ID)
	assert.Error(t, err, "expected the VM to be gone after Delete")

	instances, err := provider.List(ctx)
	require.NoError(t, err)
	for _, i := range instances {
		assert.NotEqual(t, created.Name, i.Name, "deleted VM should not be listed")
	}
	assert.Equal(t, InstanceTypeName, instancetype.Name)
}

func TestDelete_NotFound(t *testing.T) {
	provider, _, ctx := setupInstanceProvider(t)

	err := provider.Delete(ctx, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestInstanceProvider_ReAuthenticatesAfterSessionInvalidation(t *testing.T) {
	const claimName1 = "claim6"
	const claimName2 = "claim7"
	provider, class, ctx := setupInstanceProvider(t)

	created, instancetype, err := provider.Create(ctx, class, testNodeClaim(claimName1), testInstanceTypes())
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

	_, instancetype, err = provider.Create(ctx, class, testNodeClaim(claimName2), testInstanceTypes())
	require.NoError(t, err, "Create should re-authenticate rather than fail")

	require.NoError(t, provider.Delete(ctx, created.ID), "Delete should re-authenticate rather than fail")

	assert.Equal(t, InstanceTypeName, instancetype.Name)
}
