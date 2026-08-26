package e2e

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/test/vcsim"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
)

// setupProvider spins up an in-process vcsim vCenter using the default VPX
// model (one datacenter "DC0", one cluster "DC0_C0" with 3 hosts and a
// resource pool, the "VM Network" port group, the "LocalDS_0" datastore, and
// four demo VMs: DC0_H0_VM{0,1} on the standalone host and DC0_C0_RP0_VM{0,1}
// in the cluster's resource pool) and wires it up to a finder.Provider the
// same way the operator does.
func setupProvider(t *testing.T) (*finder.Provider, context.Context) {
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

	// "DC0_H0" is the standalone-host VM-name prefix in the default vcsim
	// inventory (DC0_H0_VM0 / DC0_H0_VM1). Using it as ClusterName lets
	// ListVMs exercise its prefix-matching logic without any extra
	// inventory setup, and without picking up the DC0_C0_RP0_VM* VMs.
	p := finder.NewDefaultProvider(sess, findClient, dc, "", "DC0_H0")
	return p, ctx
}

func TestResolveResourcePool_ByName(t *testing.T) {
	p, ctx := setupProvider(t)

	pool, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{Name: "DC0_C0"})
	require.NoError(t, err)
	require.NotNil(t, pool)
	assert.Equal(t, "ResourcePool", pool.Reference().Type)
	name, err := pool.ObjectName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Resources", name)

	refObj, err := p.FindClient.ObjectReference(ctx, pool.Reference())
	require.NoError(t, err)
	poolByRef, ok := refObj.(*object.ResourcePool)
	require.True(t, ok)
	assert.Equal(t, "/DC0/host/DC0_C0/Resources", poolByRef.InventoryPath)
}

func TestResolveResourcePool_NotFound(t *testing.T) {
	p, ctx := setupProvider(t)

	_, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{Name: "does-not-exist"})
	assert.Error(t, err)
}

func TestResolveDatastore_ByName(t *testing.T) {
	p, ctx := setupProvider(t)

	ds, err := p.ResolveDatastore(ctx, v1alpha1.DatastoreSelectorTerm{Name: "LocalDS_0"})
	require.NoError(t, err)
	require.NotNil(t, ds)
	assert.Equal(t, "LocalDS_0", ds.Name())
}

func TestResolveNetwork_ByName(t *testing.T) {
	p, ctx := setupProvider(t)

	network, err := p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{Name: "VM Network"})
	require.NoError(t, err)
	require.NotNil(t, network)
}

func TestResolveImage_ByPattern(t *testing.T) {
	p, ctx := setupProvider(t)

	vm, err := p.VMByName(ctx, "DC0_C0_RP0_VM0")
	require.NoError(t, err)
	state, err := vm.PowerState(ctx)
	require.NoError(t, err)
	if state == types.VirtualMachinePowerStatePoweredOn {
		task, err := vm.PowerOff(ctx)
		require.NoError(t, err)
		require.NoError(t, task.Wait(ctx))
	}
	require.NoError(t, vm.MarkAsTemplate(ctx))

	image, err := p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{Pattern: "DC0_C0_RP0_VM0"})
	require.NoError(t, err)
	require.NotNil(t, image)
	name, err := image.ObjectName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "DC0_C0_RP0_VM0", name)
}

func TestResolveFolder(t *testing.T) {
	p, ctx := setupProvider(t)

	folder, err := p.ResolveFolder(ctx)
	require.NoError(t, err)
	require.NotNil(t, folder)
	assert.Equal(t, "/DC0/vm", folder.InventoryPath)
}

func TestListVMs(t *testing.T) {
	p, ctx := setupProvider(t)

	vms, err := p.ListVMs(ctx)
	require.NoError(t, err)

	names := make([]string, 0, len(vms))
	for _, vm := range vms {
		names = append(names, vm.Name())
	}
	// Only the two standalone-host VMs share the "DC0_H0" prefix used as
	// ClusterName; the cluster/resource-pool VMs (DC0_C0_RP0_VM*) must not
	// be picked up.
	assert.ElementsMatch(t, []string{"DC0_H0_VM0", "DC0_H0_VM1"}, names)
}

// --- tag-based resolution ----------------------------------------------

// createAndAttachTag creates a fresh category/tag pair and attaches it to
// ref, mirroring what an operator would do via govc/the vSphere UI before
// pointing a VsphereNodeClass selector at it.
func createAndAttachTag(t *testing.T, ctx context.Context, p *finder.Provider, categoryName, tagName, associableType string, ref types.ManagedObjectReference) {
	t.Helper()

	catID, err := p.Session.Tags.CreateCategory(ctx, &tags.Category{
		Name:            categoryName,
		Cardinality:     "MULTIPLE",
		AssociableTypes: []string{associableType},
	})
	require.NoError(t, err)

	tagID, err := p.Session.Tags.CreateTag(ctx, &tags.Tag{
		Name:       tagName,
		CategoryID: catID,
	})
	require.NoError(t, err)

	require.NoError(t, p.Session.Tags.AttachTag(ctx, tagID, ref))
}

func TestResolveResourcePool_ByTag(t *testing.T) {
	p, ctx := setupProvider(t)

	cluster, err := p.FindClient.ClusterComputeResource(ctx, "DC0_C0")
	require.NoError(t, err)
	createAndAttachTag(t, ctx, p, "pool-category", "pool-tag", "ClusterComputeResource", cluster.Reference())

	pool, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{
		Tags: map[string]string{"pool-category": "pool-tag"},
	})
	require.NoError(t, err)
	require.NotNil(t, pool)
	assert.Equal(t, "ResourcePool", pool.Reference().Type)
	name, err := pool.ObjectName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Resources", name)

	refObj, err := p.FindClient.ObjectReference(ctx, pool.Reference())
	require.NoError(t, err)
	poolByRef, ok := refObj.(*object.ResourcePool)
	require.True(t, ok)
	assert.Equal(t, "/DC0/host/DC0_C0/Resources", poolByRef.InventoryPath)
}

func TestResolveDatastore_ByTag(t *testing.T) {
	p, ctx := setupProvider(t)

	ds, err := p.FindClient.Datastore(ctx, "LocalDS_0")
	require.NoError(t, err)
	createAndAttachTag(t, ctx, p, "ds-category", "ds-tag", "Datastore", ds.Reference())

	resolved, err := p.ResolveDatastore(ctx, v1alpha1.DatastoreSelectorTerm{
		Tags: map[string]string{"ds-category": "ds-tag"},
	})
	require.NoError(t, err)
	require.NotNil(t, resolved)
	name, err := resolved.ObjectName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "LocalDS_0", name)
}

func TestResolveNetwork_ByTag(t *testing.T) {
	p, ctx := setupProvider(t)

	network, err := p.FindClient.Network(ctx, "VM Network")
	require.NoError(t, err)
	createAndAttachTag(t, ctx, p, "net-category", "net-tag", "Network", network.Reference())

	resolved, err := p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{
		Tags: map[string]string{"net-category": "net-tag"},
	})
	require.NoError(t, err)
	require.NotNil(t, resolved)
}

func TestResolveImage_ByTag(t *testing.T) {
	p, ctx := setupProvider(t)

	vm, err := p.VMByName(ctx, "DC0_C0_RP0_VM1")
	require.NoError(t, err)
	state, err := vm.PowerState(ctx)
	require.NoError(t, err)
	if state == types.VirtualMachinePowerStatePoweredOn {
		task, err := vm.PowerOff(ctx)
		require.NoError(t, err)
		require.NoError(t, task.Wait(ctx))
	}
	require.NoError(t, vm.MarkAsTemplate(ctx))
	createAndAttachTag(t, ctx, p, "image-category", "image-tag", "VirtualMachine", vm.Reference())

	image, err := p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{
		Tags: map[string]string{"image-category": "image-tag"},
	})
	require.NoError(t, err)
	require.NotNil(t, image)
	name, err := image.ObjectName(ctx)
	require.NoError(t, err)
	assert.Equal(t, "DC0_C0_RP0_VM1", name)
}

func TestResolveImage_ByTag_RejectsNonTemplateVM(t *testing.T) {
	p, ctx := setupProvider(t)

	// Deliberately do NOT mark this VM as a template.
	vm, err := p.VMByName(ctx, "DC0_C0_RP0_VM1")
	require.NoError(t, err)
	createAndAttachTag(t, ctx, p, "image-category-2", "image-tag-2", "VirtualMachine", vm.Reference())

	_, err = p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{
		Tags: map[string]string{"image-category-2": "image-tag-2"},
	})
	assert.Error(t, err, "expected an error because the tagged VM is not a template")
}

// --- tagging round trip --------------------------------------------------

func TestTagInstanceAndTagsFromVM(t *testing.T) {
	p, ctx := setupProvider(t)

	vm, err := p.VMByName(ctx, "DC0_H0_VM0")
	require.NoError(t, err)

	desired := map[string]string{
		"topology.kubernetes.io/zone": "zone-a",
		"karpenter.sh/clustername":    "test-cluster",
	}
	require.NoError(t, p.TagInstance(ctx, vm.Reference(), desired))

	got, err := p.TagsFromVM(ctx, vm)
	require.NoError(t, err)

	// "topology.kubernetes.io/zone" is normalized to the vSphere-safe
	// category "k8s-zone" on write, and normalized back to
	// corev1.LabelTopologyZone (the same string) on read.
	assert.Equal(t, "zone-a", got[corev1.LabelTopologyZone])
	assert.Equal(t, "test-cluster", got["karpenter.sh/clustername"])
}

func TestCreateOrUpdateTags_Idempotent(t *testing.T) {
	p, ctx := setupProvider(t)

	first, err := p.CreateOrUpdateTags(ctx, map[string]string{"env": "prod"})
	require.NoError(t, err)
	require.Len(t, first, 1)

	// Calling it again for the same category/tag should resolve to the
	// existing tag rather than erroring out or creating a duplicate.
	second, err := p.CreateOrUpdateTags(ctx, map[string]string{"env": "prod"})
	require.NoError(t, err)
	require.Len(t, second, 1)
	assert.Equal(t, first[0], second[0])
}
