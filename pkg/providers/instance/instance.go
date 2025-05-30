package instance

import (
	"context"
	"fmt"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi"
	models "github.com/vmware/govmomi/vim25/mo"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

type Provider interface {
	BeginCreate(context.Context, *v1alpha1.VsphereNodeClass, *karpv1.NodeClaim, []*corecloudprovider.InstanceType) (*VirtualMachinePromise, error)
	Get(context.Context, string) (*models.VirtualMachine, error)
	List(context.Context) ([]*models.VirtualMachine, error)
	Delete(context.Context, string) error
	Update(context.Context, string, *models.VirtualMachine) error
}

var _ Provider = (*DefaultProvider)(nil)

type DefaultProvider struct {
	vsphereClient *govmomi.Client
}

type VirtualMachinePromise struct {
	VM   *models.VirtualMachine
	Wait func() error
}

type createResult struct {
	Task *models.Task
	VM   *models.VirtualMachine
}

func NewDefaultProvider(c *govmomi.Client) *DefaultProvider {
	return &DefaultProvider{
		vsphereClient: c,
	}
}

func (p *DefaultProvider) Name() string {
	return "vsphere"
}

func (p *DefaultProvider) BeginCreate(
	ctx context.Context,
	class *v1alpha1.VsphereNodeClass,
	claim *karpv1.NodeClaim,
	instanceTypes []*corecloudprovider.InstanceType) (*VirtualMachinePromise, error) {
	fmt.Println("Creating VM - not implemented yet")
	// For now, we return a placeholder promise.
	return &VirtualMachinePromise{
		VM:   &models.VirtualMachine{},
		Wait: func() error { return nil },
	}, nil
}

func (p *DefaultProvider) Get(ctx context.Context, vmID string) (*models.VirtualMachine, error) {
	vm := &models.VirtualMachine{}
	return vm, nil
}

func (p *DefaultProvider) List(ctx context.Context) ([]*models.VirtualMachine, error) {
	fmt.Println("Listing all VMs - not implemented yet")
	return nil, nil
}

func (p *DefaultProvider) Delete(ctx context.Context, vmID string) error {
	fmt.Println("Deleting VM - not implemented yet")
	return nil
}

func (p *DefaultProvider) Update(ctx context.Context, vmID string, update *models.VirtualMachine) error {
	fmt.Println("Updating VM - not implemented yet")
	return nil
}
