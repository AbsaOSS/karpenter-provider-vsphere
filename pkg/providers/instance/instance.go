package instance

import (
	"context"
	"fmt"
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	models "github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

type Provider interface {
	Create(context.Context, *v1alpha1.VsphereNodeClass, *karpv1.NodeClaim, []*corecloudprovider.InstanceType) (*Instance, error)
	Get(context.Context, string) (*models.VirtualMachine, error)
	List(context.Context) ([]*models.VirtualMachine, error)
	Delete(context.Context, string) error
	Update(context.Context, string, *models.VirtualMachine) error
}

var _ Provider = (*DefaultProvider)(nil)

type DefaultProvider struct {
	vsphereClient *govmomi.Client
	ClusterName   string
}

func NewDefaultProvider(c *govmomi.Client, clusterName string) *DefaultProvider {
	return &DefaultProvider{
		vsphereClient: c,
		ClusterName:   clusterName,
	}
}

func (p *DefaultProvider) Name() string {
	return "vsphere"
}

func genereateResourcePoolPath(class *v1alpha1.VsphereNodeClass) string {
	return fmt.Sprintf("/%s/host/%s/Resources", class.Spec.DC, class.Spec.ComputeCluster)
}

func (p *DefaultProvider) GenerateVMSpec(ctx context.Context, class *v1alpha1.VsphereNodeClass, name string, instanceType *corecloudprovider.InstanceType) (*types.VirtualMachineCloneSpec, error) {
	locationSpec, err := p.GenerateTarget(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to generate target for VM: %w", err)
	}

	diskAndNet, err := p.GetDeviceSpec(ctx, class, instanceType.Capacity.Storage().Value())
	if err != nil {
		return nil, fmt.Errorf("failed to get device spec: %w", err)
	}
	return &types.VirtualMachineCloneSpec{
		Template: false,
		Location: *locationSpec,
		Config: &types.VirtualMachineConfigSpec{
			Name:         name,
			NumCPUs:      int32(instanceType.Capacity.Cpu().Size()),
			MemoryMB:     instanceType.Capacity.Memory().ToDec().Value(),
			GuestId:      string(types.VirtualMachineGuestOsIdentifierOtherLinux64Guest), // This should be adjusted based on the OS type in the instance type.
			DeviceChange: diskAndNet,
		},
		PowerOn: true,
	}, nil
}

func (p *DefaultProvider) GenerateTarget(ctx context.Context, class *v1alpha1.VsphereNodeClass) (*types.VirtualMachineRelocateSpec, error) {
	f := find.NewFinder(p.vsphereClient.Client, true)
	dc, err := f.Datacenter(ctx, class.Spec.DC)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter: %w", err)
	}
	f.SetDatacenter(dc)
	pool, err := f.ResourcePool(ctx, genereateResourcePoolPath(class))
	if err != nil {
		return nil, fmt.Errorf("failed to find resource pool: %w", err)
	}

	ds, err := f.DatastoreOrDefault(ctx, class.Spec.Datastore)
	if err != nil {
		return nil, fmt.Errorf("failed to find datastore: %w", err)
	}
	poolRef := pool.Reference()
	dsRef := ds.Reference()
	return &types.VirtualMachineRelocateSpec{
		Datastore: &dsRef,
		Pool:      &poolRef,
	}, nil

}
func (p *DefaultProvider) Create(
	ctx context.Context,
	class *v1alpha1.VsphereNodeClass,
	claim *karpv1.NodeClaim,
	instanceTypes []*corecloudprovider.InstanceType) (*Instance, error) {

	VMName := GenerateVMName(p.ClusterName, claim.Name)
	instanceType := instanceTypes[0] // For simplicity, we take the first instance type.
	cloneSpec, err := p.GenerateVMSpec(ctx, class, VMName, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VM spec: %w", err)
	}
	finder := find.NewFinder(p.vsphereClient.Client, true)
	dc, err := finder.Datacenter(ctx, class.Spec.DC)
	if err != nil {
		return nil, fmt.Errorf("failed to find datacenter: %w", err)
	}
	finder.SetDatacenter(dc)
	folder, err := finder.Folder(ctx, fmt.Sprintf("/%s/vm/%s", class.Spec.DC, class.Spec.Path))
	if err != nil {
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}

	vmTemplate, err := finder.VirtualMachine(ctx, class.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM template: %w", err)
	}
	task, err := vmTemplate.Clone(ctx, folder, GenerateVMName(p.ClusterName, claim.Name), *cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to clone VM: %w", err)
	}
	err = task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}
	vm, err := finder.VirtualMachine(ctx, VMName)
	if err != nil {
		return nil, fmt.Errorf("failed to find cloned VM: %w", err)
	}
	creationDate, err := extractCreationDate(ctx, vm)
	if err != nil {
		return nil, fmt.Errorf("failed to extract creation date: %w", err)
	}
	powerState, err := vm.PowerState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get power state: %w", err)
	}
	return NewInstance(vm.UUID(ctx), class.Spec.Image, powerState.Strings(), vm.Name(), *creationDate), err
}

func extractCreationDate(ctx context.Context, vm *object.VirtualMachine) (*time.Time, error) {
	var vmMo models.VirtualMachine
	err := vm.Properties(ctx, vm.Reference(), []string{"config.createDate"}, &vmMo)
	if err != nil {
		return nil, err
	}

	t := vmMo.Config.CreateDate.UTC()
	return &t, nil
}

func GenerateVMName(cluster, claim string) string {
	return fmt.Sprintf("%s-karp-%s", cluster, claim)
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
