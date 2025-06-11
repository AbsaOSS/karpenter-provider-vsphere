package instance

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	"github.com/vmware/govmomi/object"
	models "github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

type Provider interface {
	Create(context.Context, *v1alpha1.VsphereNodeClass, *karpv1.NodeClaim, []*corecloudprovider.InstanceType, string) (*Instance, error)
	Get(context.Context, string) (*Instance, error)
	List(context.Context) ([]*Instance, error)
	Delete(context.Context, string) error
}

var _ Provider = (*DefaultProvider)(nil)

type DefaultProvider struct {
	ClusterName string
	VsphereZone string
	kubeClient  kubernetes.Interface
	Finder      *finder.Provider
}

func NewDefaultProvider(kube kubernetes.Interface, finder *finder.Provider, clusterName, zone string) *DefaultProvider {
	return &DefaultProvider{
		ClusterName: clusterName,
		VsphereZone: zone,
		kubeClient:  kube,
		Finder:      finder,
	}
}

func (p *DefaultProvider) Name() string {
	return "vsphere"
}

func (p *DefaultProvider) GenerateVMSpec(ctx context.Context, class *v1alpha1.VsphereNodeClass, name string, instanceType *corecloudprovider.InstanceType) (*types.VirtualMachineCloneSpec, error) {
	locationSpec, err := p.GenerateTarget(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to generate target for VM: %w", err)
	}

	diskAndNet, err := p.GetDeviceSpec(ctx, class, class.Spec.DiskSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get device spec: %w", err)
	}
	initData, err := p.GetInitData(ctx, class, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get init data: %w", err)
	}
	image, err := p.Finder.ResolveImage(ctx, class.Spec.ImageSelector)
	if err != nil {
		return nil, err
	}
	return &types.VirtualMachineCloneSpec{
		Template: false,
		Location: *locationSpec,
		Config: &types.VirtualMachineConfigSpec{
			Name:         name,
			Annotation:   fmt.Sprintf("cloned_from:%s", image.InventoryPath),
			NumCPUs:      int32(instanceType.Capacity.Cpu().Value()),
			MemoryMB:     instanceType.Capacity.Memory().ToDec().Value() * 1024,
			GuestId:      string(types.VirtualMachineGuestOsIdentifierOtherLinux64Guest), // This should be adjusted based on the OS type in the instance type.
			DeviceChange: diskAndNet,
			ExtraConfig:  initData,
		},
		PowerOn: true,
	}, nil
}

func (p *DefaultProvider) GenerateTarget(ctx context.Context, class *v1alpha1.VsphereNodeClass) (*types.VirtualMachineRelocateSpec, error) {
	var relocationSpec types.VirtualMachineRelocateSpec
	dc, err := p.Finder.ResolveDC(ctx, class.Spec.Datacenter)
	if err != nil {
		return nil, err
	}
	pool, err := p.Finder.ResolveResourcePool(ctx, class.Spec.PoolSelector, dc)
	if err != nil {
		return nil, err
	}
	poolRef := pool.Reference()
	relocationSpec.Pool = &poolRef
	datastore, err := p.Finder.ResolveDatastore(ctx, class.Spec.DatastoreSelector)
	dsRef := datastore.Reference()
	if err != nil {
		return nil, err
	}
	relocationSpec.Datastore = &dsRef

	return &relocationSpec, nil
}

func (p *DefaultProvider) Create(
	ctx context.Context,
	class *v1alpha1.VsphereNodeClass,
	claim *karpv1.NodeClaim,
	instanceTypes []*corecloudprovider.InstanceType, poolName string) (*Instance, error) {

	instanceType := instanceTypes[0] // For simplicity, we take the first instance type.
	VMName := GenerateVMName(p.ClusterName, claim.Name)
	instanceTags := map[string]string{
		v1alpha1.ClusterNameTagKey:   p.ClusterName,
		v1alpha1.LabelNodeClass:      class.Name,
		karpv1.NodePoolLabelKey:      poolName,
		corev1.LabelTopologyZone:     p.VsphereZone,
		v1alpha1.LabelInstanceSize:   instanceType.Name,
		v1alpha1.LabelInstanceCPU:    fmt.Sprintf("%d", instanceType.Capacity.Cpu().Value()),
		v1alpha1.LabelInstanceMemory: fmt.Sprintf("%d", utils.GiToMb(instanceType.Capacity.Memory().ToDec().Value())),
	}

	cloneSpec, err := p.GenerateVMSpec(ctx, class, VMName, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VM spec: %w", err)
	}

	vmTemplate, err := p.Finder.ResolveImage(ctx, class.Spec.ImageSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM template: %w", err)
	}
	vmFolder, err := p.Finder.ResolveFolder(ctx)
	if err != nil {
		return nil, err
	}
	task, err := vmTemplate.Clone(ctx, vmFolder, GenerateVMName(p.ClusterName, claim.Name), *cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to clone VM: %w", err)
	}
	vm, err := p.Finder.VMByName(ctx, VMName)
	if err != nil {
		return nil, fmt.Errorf("failed to find cloned VM: %w", err)
	}
	err = p.Finder.TagInstance(ctx, vm.Reference(), instanceTags)
	if err != nil {
		return nil, err
	}

	err = p.Finder.TagInstance(ctx, vm.Reference(), instanceTags)
	if err != nil {
		return nil, err
	}

	creationDate, err := extractCreationDate(ctx, vm)
	if err != nil {
		return nil, fmt.Errorf("failed to extract creation date: %w", err)
	}
	powerState, err := vm.PowerState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get power state: %w", err)
	}
	err = task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}
	return NewInstance(vm, vm.UUID(ctx), vmTemplate.InventoryPath, string(powerState), vm.Name(), *creationDate, instanceTags), err
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

func getImageFromAnnotation(vm *object.VirtualMachine) string {
	var annotation string
	err := vm.Properties(context.Background(), vm.Reference(), []string{"config.config"}, &annotation)
	if err != nil {
		annotation = "image_not_found"
	}
	return strings.TrimPrefix(annotation, "cloned_from:")
}

func (p *DefaultProvider) List(ctx context.Context) ([]*Instance, error) {
	instances := []*Instance{}
	vms, err := p.Finder.ListVMs(ctx)
	//
	if err != nil {
		fmt.Printf("Failed to list VMs: %v\n", err)
	}
	for _, vm := range vms {
		image := getImageFromAnnotation(vm)
		tags, err := p.Finder.TagsFromVM(ctx, vm)
		if err != nil {
			fmt.Printf("Failed to get tags for VM %s: %v\n", vm.Name(), err)
		}
		ps, err := vm.PowerState(ctx)
		if err != nil {
			fmt.Printf("Failed to get power state for VM %s: %v\n", vm.Name(), err)
		}
		creationDate, err := extractCreationDate(ctx, vm)
		if err != nil {
			fmt.Printf("Failed to extract creation date for VM %s: %v\n", vm.Name(), err)
		}
		instances = append(instances, NewInstance(vm, vm.UUID(ctx), image, string(ps), vm.Name(), *creationDate, tags))
	}
	return instances, nil
}

func (p *DefaultProvider) Get(ctx context.Context, vmID string) (*Instance, error) {
	vm, err := p.Finder.GetVMByID(ctx, vmID)
	if err != nil {
		return nil, err
	}
	tags, err := p.Finder.TagsFromVM(ctx, vm)
	if err != nil {
		fmt.Printf("Failed to get tags for VM %s: %v\n", vm.Name(), err)
	}
	instance := NewInstanceFromVM(ctx, vm, time.Now(), tags)
	return instance, nil

}

func (p *DefaultProvider) Delete(ctx context.Context, vmID string) error {
	i, err := p.Get(ctx, vmID)
	if err != nil {
		return err
	}
	vm := i.GetVM()
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return err
	}
	// Wait for the power off task to complete
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	task, err = vm.Destroy(ctx)
	if err != nil {
		return err
	}
	// Wait for the destroy task to complete
	err = task.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
