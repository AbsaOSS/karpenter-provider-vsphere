package finder

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func (p *Provider) PoolByName(ctx context.Context, name string, dc *object.Datacenter) (*object.ResourcePool, error) {
	p.FindClient.SetDatacenter(dc)
	poolPath := fmt.Sprintf("/%s/host/%s/Resources", dc.Name(), name)
	pool, err := p.FindClient.ResourcePool(ctx, poolPath)
	return pool, err
}

func (p *Provider) DatastoreByName(ctx context.Context, name string) (*object.Datastore, error) {
	return p.FindClient.Datastore(ctx, name)
}

func (p *Provider) NetworkByName(ctx context.Context, name string) (*object.NetworkReference, error) {
	network, err := p.FindClient.Network(ctx, name)
	return &network, err
}

func (p *Provider) DCByName(ctx context.Context, name string) (*object.Datacenter, error) {
	dcPath := fmt.Sprintf("/%s", name)
	return p.FindClient.Datacenter(ctx, dcPath)
}

func (p *Provider) VMByName(ctx context.Context, name string) (*object.VirtualMachine, error) {
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	p.FindClient.SetDatacenter(dc)
	return p.FindClient.VirtualMachine(ctx, name)
}

func (p *Provider) ImageByPattern(ctx context.Context, pattern string) (*object.VirtualMachine, error) {
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	p.FindClient.SetDatacenter(dc)

	vms, err := p.FindClient.VirtualMachineList(ctx, pattern)
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
