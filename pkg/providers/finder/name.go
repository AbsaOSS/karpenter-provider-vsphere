package finder

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func (p *Provider) PoolByName(ctx context.Context, name string, dc *object.Datacenter) (*object.ResourcePool, error) {
	vfinder := find.NewFinder(p.Client, true)
	vfinder.SetDatacenter(dc)
	poolPath := fmt.Sprintf("/%s/host/%s/Resources", dc.Name(), name)
	pool, err := vfinder.ResourcePool(ctx, poolPath)
	return pool, err
}

func (p *Provider) DatastoreByName(ctx context.Context, name string) (*object.Datastore, error) {
	vfinder := find.NewFinder(p.Client, true)
	datastore, err := vfinder.Datastore(ctx, name)
	return datastore, err
}

func (p *Provider) NetworkByName(ctx context.Context, name string) (*object.NetworkReference, error) {
	vfinder := find.NewFinder(p.Client, true)
	network, err := vfinder.Network(ctx, name)
	return &network, err
}

func (p *Provider) DCByName(ctx context.Context, name string) (*object.Datacenter, error) {
	vfinder := find.NewFinder(p.Client, true)
	dcPath := fmt.Sprintf("/%s", name)
	dc, err := vfinder.Datacenter(ctx, dcPath)
	return dc, err
}

func (p *Provider) VMByName(ctx context.Context, name string) (*object.VirtualMachine, error) {
	vfinder := find.NewFinder(p.Client, true)
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	vfinder.SetDatacenter(dc)
	vm, err := vfinder.VirtualMachine(ctx, name)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (p *Provider) ImageByPattern(ctx context.Context, pattern string) (*object.VirtualMachine, error) {
	vfinder := find.NewFinder(p.Client, true)
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	vfinder.SetDatacenter(dc)

	vms, err := vfinder.VirtualMachineList(ctx, pattern)
	if err != nil {
		return nil, err
	}
	return vms[0], err
}

func (p *Provider) GetVMByID(ctx context.Context, id string) (*object.VirtualMachine, error) {
	ptrBool := false
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	//TODO: move to provider
	searchIndex := object.NewSearchIndex(p.Client)
	vmRef, err := searchIndex.FindByUuid(ctx, dc, id, true, &ptrBool)
	if err != nil {
		return nil, err
	}
	if vmRef == nil {
		return nil, corecloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("vmRef not found"))
	}
	vm := object.NewVirtualMachine(p.Client, vmRef.Reference())
	return vm, nil
}
