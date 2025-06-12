package finder

import (
	"context"
	"fmt"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
)

func (p *Provider) ResolveResourcePool(ctx context.Context, selector v1alpha1.ResPoolSelctorTerm, dc *object.Datacenter) (*object.ResourcePool, error) {
	if len(selector.Tags) > 0 {
		return p.PoolByTag(ctx, selector.Tags)
	}

	if selector.Name != "" {
		return p.PoolByName(ctx, selector.Name, dc)
	}
	return nil, fmt.Errorf("failed to resolve ResourcePool")
}

func (p *Provider) GetDC(ctx context.Context) (*object.Datacenter, error) {
	return p.FindClient.DatacenterOrDefault(ctx, "*")

}
func (p *Provider) ResolveDC(ctx context.Context, selector v1alpha1.DCSelectorTerm) (*object.Datacenter, error) {
	if len(selector.Tags) > 0 {
		return p.DCByTag(ctx, selector.Tags)
	}

	if selector.Name != "" {
		return p.DCByName(ctx, selector.Name)
	}
	return nil, fmt.Errorf("failed to resolve Datacenter")
}

func (p *Provider) ResolveDatastore(ctx context.Context, selector v1alpha1.DatastoreSelectorTerm) (*object.Datastore, error) {
	if len(selector.Tags) > 0 {
		return p.DatastoreByTag(ctx, selector.Tags)
	}

	if selector.Name != "" {
		return p.DatastoreByName(ctx, selector.Name)
	}
	return nil, fmt.Errorf("failed to resolve Datastore")
}

func (p *Provider) ResolveNetwork(ctx context.Context, selector v1alpha1.NetworkSelectorTerm) (*object.NetworkReference, error) {
	if len(selector.Tags) > 0 {
		return p.NetworkByTag(ctx, selector.Tags)
	}

	if selector.Name != "" {
		return p.NetworkByName(ctx, selector.Name)
	}
	return nil, fmt.Errorf("failed to resolve network")
}

func (p *Provider) ResolveImage(ctx context.Context, selector v1alpha1.ImageSelectorTerm) (*object.VirtualMachine, error) {
	if len(selector.Tags) > 0 {
		return p.ImageByTag(ctx, selector.Tags)
	}
	if selector.Pattern != "" {
		return p.ImageByPattern(ctx, selector.Pattern)
	}
	return nil, fmt.Errorf("failed to resolve image")
}

func (p *Provider) ResolveFolder(ctx context.Context) (*object.Folder, error) {
	return p.GetFolder(ctx, p.Folder)
}

func (p *Provider) isTemplate(ctx context.Context, obj *object.VirtualMachine) bool {
	var vmMo mo.VirtualMachine
	err := obj.Properties(ctx, obj.Reference(), []string{"config.template"}, &vmMo)
	if err != nil {
		return false
	}
	return vmMo.Config.Template
}

func (p *Provider) GetFolder(ctx context.Context, f string) (*object.Folder, error) {
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	p.FindClient.SetDatacenter(dc)
	return p.FindClient.Folder(ctx, f)
}

func (p *Provider) ListVMs(ctx context.Context) ([]*object.VirtualMachine, error) {
	dc, err := p.GetDC(ctx)
	if err != nil {
		return nil, err
	}
	p.FindClient.SetDatacenter(dc)
	folder, err := p.GetFolder(ctx, p.Folder)
	if err != nil {
		return nil, err
	}
	// GetDC and list in given DC or all DCs
	return p.FindClient.VirtualMachineList(ctx, folder.InventoryPath+"/*")

}
