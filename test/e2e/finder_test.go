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

// vcsim inventory constants
const (
	FinderDatacenterName       = "DC0"
	FinderClusterName          = "DC0_C0"
	FinderStandaloneHostPrefix = "DC0_H0"
	FinderNetworkName          = "VM Network"
	FinderDatastoreName        = "LocalDS_0"
	FinderDVPGName             = "DC0_DVPG0"
	FinderTemplateVM           = "DC0_H0_VM0"
	FinderRPVM0                = "DC0_C0_RP0_VM0"
	FinderRPVM1                = "DC0_C0_RP0_VM1"
	FinderVMFolderPath         = "/DC0/vm"
	FinderResourcesPath        = "/DC0/host/" + FinderClusterName + "/Resources"
)

// Finder test zone and cluster values
const (
	FinderZoneLabel        = "topology.kubernetes.io/zone"
	FinderZoneValue        = "zone-a"
	FinderClusterNameLabel = "karpenter.sh/clustername"
	FinderClusterValue     = "test-cluster"
)

// setupProvider spins up an in-process vcsim vCenter using the default VPX
// model (one datacenter "DC0", one cluster "DC0_C0" with 3 hosts and a
// resource pool, the "VM Network" port group, the "LocalDS_0" datastore, and
// four demo VMs: DC0_H0_VM{0,1} on the standalone host and DC0_C0_RP0_VM{0,1}
// in the cluster's resource pool) and wires it up to a finder.Provider the
// same way the operator does.
func setupProvider(t *testing.T, cluster string) (*finder.Provider, context.Context) {
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
	dc, err := findClient.Datacenter(ctx, FinderDatacenterName)
	require.NoError(t, err)

	// "DC0_H0" is the standalone-host VM-name prefix in the default vcsim
	// inventory (DC0_H0_VM0 / DC0_H0_VM1). Using it as ClusterName lets
	// ListVMs exercise its prefix-matching logic without any extra
	// inventory setup, and without picking up the DC0_C0_RP0_VM* VMs.
	if cluster == "" {
		cluster = FinderStandaloneHostPrefix
	}
	p := finder.NewDefaultProvider(sess, findClient, dc, "", cluster)
	return p, ctx
}

// withProvider spins up a fresh finder.Provider via setupProvider(t,
// "kube-test") before fn and lets it be torn down (via t.Cleanup) after fn
// returns, so each call gets its own isolated vcsim instance.
func withProvider(t *testing.T, fn func(p *finder.Provider, ctx context.Context)) {
	t.Helper()
	p, ctx := setupProvider(t, "kube-test")
	fn(p, ctx)
}

func TestResolveResourcePool_ByName(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		pool, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{Name: FinderClusterName})
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
		assert.Equal(t, FinderResourcesPath, poolByRef.InventoryPath)
	})
}

func TestResolveResourcePool_NotFound(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		_, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{Name: "does-not-exist"})
		assert.Error(t, err)
	})
}

func TestResolveDatastore_ByName(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		ds, err := p.ResolveDatastore(ctx, v1alpha1.DatastoreSelectorTerm{Name: FinderDatastoreName})
		require.NoError(t, err)
		require.NotNil(t, ds)
		assert.Equal(t, FinderDatastoreName, ds.Name())
	})
}

func TestResolveNetwork_ByName(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		network, err := p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{Name: FinderNetworkName})
		require.NoError(t, err)
		require.NotNil(t, network)
	})
}

func TestResolveImage_ByPattern(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		vm, err := p.VMByName(ctx, FinderRPVM0)
		require.NoError(t, err)
		state, err := vm.PowerState(ctx)
		require.NoError(t, err)
		if state == types.VirtualMachinePowerStatePoweredOn {
			task, err := vm.PowerOff(ctx)
			require.NoError(t, err)
			require.NoError(t, task.Wait(ctx))
		}
		require.NoError(t, vm.MarkAsTemplate(ctx))

		image, err := p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{Pattern: FinderRPVM0})
		require.NoError(t, err)
		require.NotNil(t, image)
		name, err := image.ObjectName(ctx)
		require.NoError(t, err)
		assert.Equal(t, FinderRPVM0, name)
	})
}

func TestResolveFolder(t *testing.T) {
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		folder, err := p.ResolveFolder(ctx)
		require.NoError(t, err)
		require.NotNil(t, folder)
		assert.Equal(t, FinderVMFolderPath, folder.InventoryPath)
	})
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
	const poolCategory = "pool-category"
	const poolTag = "pool-tag"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		cluster, err := p.FindClient.ClusterComputeResource(ctx, FinderClusterName)
		require.NoError(t, err)
		createAndAttachTag(t, ctx, p, poolCategory, poolTag, "ClusterComputeResource", cluster.Reference())

		pool, err := p.ResolveResourcePool(ctx, v1alpha1.ResPoolSelctorTerm{
			Tags: map[string]string{poolCategory: poolTag},
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
		assert.Equal(t, FinderResourcesPath, poolByRef.InventoryPath)
	})
}

func TestResolveDatastore_ByTag(t *testing.T) {
	const dsCategory = "ds-category"
	const dsTag = "ds-tag"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		ds, err := p.FindClient.Datastore(ctx, FinderDatastoreName)
		require.NoError(t, err)
		createAndAttachTag(t, ctx, p, dsCategory, dsTag, "Datastore", ds.Reference())

		resolved, err := p.ResolveDatastore(ctx, v1alpha1.DatastoreSelectorTerm{
			Tags: map[string]string{dsCategory: dsTag},
		})
		require.NoError(t, err)
		require.NotNil(t, resolved)
		name, err := resolved.ObjectName(ctx)
		require.NoError(t, err)
		assert.Equal(t, FinderDatastoreName, name)
	})
}

func TestResolveNetwork_ByTag(t *testing.T) {
	const netCategory = "net-category"
	const netTag = "net-tag"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		network, err := p.FindClient.Network(ctx, FinderNetworkName)
		require.NoError(t, err)
		createAndAttachTag(t, ctx, p, netCategory, netTag, "Network", network.Reference())

		resolved, err := p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{
			Tags: map[string]string{netCategory: netTag},
		})
		require.NoError(t, err)
		require.NotNil(t, resolved)
	})
}

func TestResolveNetwork_ByTag_DistributedVirtualPortgroup(t *testing.T) {
	const dvpgCategory = "dvpg-category"
	const dvpgTag = "dvpg-tag"
	const dvpgKind = "DistributedVirtualPortgroup"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		// The default vcsim VPX model creates one DVS ("DVS0") and one
		// DistributedVirtualPortgroup (FinderDVPGName) per datacenter alongside
		// the plain FinderNetworkName Network; NetworkByTag must match this kind too.
		dvpg, err := p.FindClient.Network(ctx, FinderDVPGName)
		require.NoError(t, err)
		require.Equal(t, dvpgKind, dvpg.Reference().Type)
		createAndAttachTag(t, ctx, p, dvpgCategory, dvpgTag, dvpgKind, dvpg.Reference())

		resolved, err := p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{
			Tags: map[string]string{dvpgCategory: dvpgTag},
		})
		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, dvpgKind, (*resolved).Reference().Type)
	})
}

func TestResolveNetwork_ByTag_AmbiguousAcrossKinds(t *testing.T) {
	const ambiguousNetCategory = "ambiguous-net-category"
	const ambiguousNetTag = "ambiguous-net-tag"
	const ambiguousNetKind = "DistributedVirtualPortgroup"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		network, err := p.FindClient.Network(ctx, FinderNetworkName)
		require.NoError(t, err)
		dvpg, err := p.FindClient.Network(ctx, FinderDVPGName)
		require.NoError(t, err)

		// One tag attached to both a Network and a DistributedVirtualPortgroup:
		// NetworkByTag checks several kinds, so this must be rejected as
		// ambiguous rather than silently picking one.
		catID, err := p.Session.Tags.CreateCategory(ctx, &tags.Category{
			Name:            ambiguousNetCategory,
			Cardinality:     "MULTIPLE",
			AssociableTypes: []string{"Network", ambiguousNetKind},
		})
		require.NoError(t, err)
		tagID, err := p.Session.Tags.CreateTag(ctx, &tags.Tag{Name: ambiguousNetTag, CategoryID: catID})
		require.NoError(t, err)
		require.NoError(t, p.Session.Tags.AttachTag(ctx, tagID, network.Reference()))
		require.NoError(t, p.Session.Tags.AttachTag(ctx, tagID, dvpg.Reference()))

		_, err = p.ResolveNetwork(ctx, v1alpha1.NetworkSelectorTerm{
			Tags: map[string]string{ambiguousNetCategory: ambiguousNetTag},
		})
		assert.Error(t, err, "expected an error because the tag matches both a Network and a DistributedVirtualPortgroup")
	})
}

func TestResolveImage_ByTag(t *testing.T) {
	const imageCategory = "image-category"
	const imageTag = "image-tag"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		vm, err := p.VMByName(ctx, FinderRPVM1)
		require.NoError(t, err)
		state, err := vm.PowerState(ctx)
		require.NoError(t, err)
		if state == types.VirtualMachinePowerStatePoweredOn {
			task, err := vm.PowerOff(ctx)
			require.NoError(t, err)
			require.NoError(t, task.Wait(ctx))
		}
		require.NoError(t, vm.MarkAsTemplate(ctx))
		createAndAttachTag(t, ctx, p, imageCategory, imageTag, "VirtualMachine", vm.Reference())

		image, err := p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{
			Tags: map[string]string{imageCategory: imageTag},
		})
		require.NoError(t, err)
		require.NotNil(t, image)
		name, err := image.ObjectName(ctx)
		require.NoError(t, err)
		assert.Equal(t, FinderRPVM1, name)
	})
}

func TestResolveImage_ByTag_RejectsNonTemplateVM(t *testing.T) {
	const imageCategoryReject = "image-category-2"
	const imageTagReject = "image-tag-2"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		// Deliberately do NOT mark this VM as a template.
		vm, err := p.VMByName(ctx, FinderRPVM1)
		require.NoError(t, err)
		createAndAttachTag(t, ctx, p, imageCategoryReject, imageTagReject, "VirtualMachine", vm.Reference())

		_, err = p.ResolveImage(ctx, v1alpha1.ImageSelectorTerm{
			Tags: map[string]string{imageCategoryReject: imageTagReject},
		})
		assert.Error(t, err, "expected an error because the tagged VM is not a template")
	})
}

// --- tagging round trip --------------------------------------------------

func TestTagInstanceAndTagsFromVM(t *testing.T) {
	const clusterNameLabel = "karpenter.sh/clustername"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		vm, err := p.VMByName(ctx, FinderTemplateVM)
		require.NoError(t, err)

		desired := map[string]string{
			FinderZoneLabel:  FinderZoneValue,
			clusterNameLabel: FinderClusterValue,
		}
		require.NoError(t, p.TagInstance(ctx, vm.Reference(), desired))

		got, err := p.TagsFromVM(ctx, vm)
		require.NoError(t, err)

		// "topology.kubernetes.io/zone" is normalized to the vSphere-safe
		// category "k8s-zone" on write, and normalized back to
		// corev1.LabelTopologyZone (the same string) on read.
		assert.Equal(t, FinderZoneValue, got[corev1.LabelTopologyZone])
		assert.Equal(t, FinderClusterValue, got[clusterNameLabel])
	})
}

func TestCreateOrUpdateTags_Idempotent(t *testing.T) {
	const envKey = "env"
	const envValue = "prod"
	withProvider(t, func(p *finder.Provider, ctx context.Context) {
		first, err := p.CreateOrUpdateTags(ctx, map[string]string{envKey: envValue})
		require.NoError(t, err)
		require.Len(t, first, 1)

		// Calling it again for the same category/tag should resolve to the
		// existing tag rather than erroring out or creating a duplicate.
		second, err := p.CreateOrUpdateTags(ctx, map[string]string{envKey: envValue})
		require.NoError(t, err)
		require.Len(t, second, 1)
		assert.Equal(t, first[0], second[0])
	})
}

func TestListVMs(t *testing.T) {
	p, ctx := setupProvider(t, FinderStandaloneHostPrefix)

	vms, err := p.ListVMs(ctx)
	require.NoError(t, err)

	names := make([]string, 0, len(vms))
	for _, vm := range vms {
		names = append(names, vm.Name())
	}
	// Only the two standalone-host VMs share the FinderStandaloneHostPrefix used as
	// ClusterName; the cluster/resource-pool VMs (DC0_C0_RP0_VM*) must not
	// be picked up.
	assert.ElementsMatch(t, []string{FinderStandaloneHostPrefix + "_VM0", FinderStandaloneHostPrefix + "_VM1"}, names)
}
