package finder

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func (p *Provider) PoolByName(ctx context.Context, name string) (*object.ResourcePool, error) {

	poolPath := fmt.Sprintf("host/%s/Resources", name)
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

func (p *Provider) VMByName(ctx context.Context, name string) (*object.VirtualMachine, error) {
	return p.FindClient.VirtualMachine(ctx, name)
}

func (p *Provider) ImageByPattern(ctx context.Context, pattern string) (*object.VirtualMachine, error) {
	vms, err := p.FindClient.VirtualMachineList(ctx, pattern)
	if err != nil {
		return nil, err
	}
	return vms[0], err
}

func (p *Provider) GetVMByID(ctx context.Context, id string) (*object.VirtualMachine, error) {
	ptrBool := false

	vmRef, err := p.IndexClient.FindByUuid(ctx, p.DC, id, true, &ptrBool)
	if err != nil {
		return nil, err
	}
	if vmRef == nil {
		return nil, corecloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("vmRef not found"))
	}
	vm := object.NewVirtualMachine(p.Client, vmRef.Reference())
	return vm, nil
}
