package instance

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/userdata"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware/govmomi/find"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	"github.com/vmware/govmomi/object"
	models "github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

const ImageNotFound = "image_not_found"

type Provider interface {
	Create(context.Context, *v1alpha1.VsphereNodeClass, *karpv1.NodeClaim, []*corecloudprovider.InstanceType) (*Instance, error)
	Get(context.Context, string) (*Instance, error)
	List(context.Context) ([]*Instance, error)
	Delete(context.Context, string) error
}

var _ Provider = (*DefaultProvider)(nil)

type DefaultProvider struct {
	ClusterName string
	kubeClient  kubernetes.Interface
	Finder      *finder.Provider
}

func NewDefaultProvider(kube kubernetes.Interface, finder *finder.Provider, clusterName string) *DefaultProvider {
	return &DefaultProvider{
		ClusterName: clusterName,
		kubeClient:  kube,
		Finder:      finder,
	}
}

func (p *DefaultProvider) Name() string {
	return "vsphere"
}

func (p *DefaultProvider) GenerateVMSpec(ctx context.Context, class *v1alpha1.VsphereNodeClass, name string, image *object.VirtualMachine, instanceType *corecloudprovider.InstanceType) (*types.VirtualMachineCloneSpec, error) {
	diskEnableUUID := true
	locationSpec, err := p.GenerateTarget(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to generate target for VM: %w", err)
	}

	diskAndNet, err := p.GetDeviceSpec(ctx, class, class.Spec.DiskSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get device spec: %w", err)
	}

	t := time.Now()
	return &types.VirtualMachineCloneSpec{
		Template: false,
		Location: *locationSpec,
		Config: &types.VirtualMachineConfigSpec{
			Flags: &types.VirtualMachineFlagInfo{
				// disk.EnableUUID = TRUE
				DiskUuidEnabled: &diskEnableUUID,
			},
			Name:         name,
			Annotation:   fmt.Sprintf("cloned_from: %s", image.InventoryPath),
			NumCPUs:      int32(instanceType.Capacity.Cpu().Value()),
			MemoryMB:     utils.InstanceTypeToMegabytes(instanceType.Capacity.Memory()),
			GuestId:      string(types.VirtualMachineGuestOsIdentifierOtherLinux64Guest), // This should be adjusted based on the OS type in the instance type.
			DeviceChange: diskAndNet,
			CreateDate:   &t,
		},
		PowerOn: false,
	}, nil
}

func (p *DefaultProvider) GenerateTarget(ctx context.Context, class *v1alpha1.VsphereNodeClass) (*types.VirtualMachineRelocateSpec, error) {
	var relocationSpec types.VirtualMachineRelocateSpec
	pool, err := p.Finder.ResolveResourcePool(ctx, class.Spec.PoolSelector)
	if err != nil {
		return nil, err
	}
	poolRef := pool.Reference()
	relocationSpec.Pool = &poolRef
	datastore, err := p.Finder.ResolveDatastore(ctx, class.Spec.DatastoreSelector)
	if err != nil {
		return nil, err
	}
	dsRef := datastore.Reference()
	relocationSpec.Datastore = &dsRef

	return &relocationSpec, nil
}

func (p *DefaultProvider) Create(
	ctx context.Context,
	class *v1alpha1.VsphereNodeClass,
	claim *karpv1.NodeClaim,
	instanceTypes []*corecloudprovider.InstanceType) (*Instance, error) {

	if err := p.Finder.Session.EnsureValid(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure vsphere session is valid: %w", err)
	}

	instanceType := instanceTypes[0] // For simplicity, we take the first instance type.
	VMName := GenerateVMName(p.ClusterName, claim.Name)
	instanceTags := map[string]string{
		v1alpha1.ClusterNameTagKey:   p.ClusterName,
		v1alpha1.LabelNodeClass:      class.Name,
		karpv1.NodePoolLabelKey:      claim.Labels[karpv1.NodePoolLabelKey],
		v1alpha1.LabelInstanceSize:   instanceType.Name,
		v1alpha1.LabelInstanceCPU:    fmt.Sprintf("%d", instanceType.Capacity.Cpu().Value()),
		v1alpha1.LabelInstanceMemory: fmt.Sprintf("%d", utils.InstanceTypeToMegabytes(instanceType.Capacity.Memory())),
	}

	maps.Copy(instanceTags, class.Spec.Tags)
	//Default carpenter taint
	taints := []corev1.Taint{
		karpv1.UnregisteredNoExecuteTaint,
	}
	taints = append(taints, claim.Spec.Taints...)
	controllerOpts := options.FromContext(ctx)
	workerInitConfig := userdata.NewInitData(
		taints,
		VMName,
		controllerOpts.ClusterEndpoint,
		controllerOpts.JoinToken,
		controllerOpts.KubeVersion,
		class.Spec.UserData.AdditionalUserdata,
	)
	initType := &userdata.InitType{
		Distro: v1alpha1.Distro(controllerOpts.KubeDistro),
		Format: class.Spec.UserData.Type,
	}

	userData, err := p.GetInitData(workerInitConfig, initType)
	if err != nil {
		return nil, err
	}
	vmTemplate, err := p.Finder.ResolveImage(ctx, class.Spec.ImageSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM template: %w", err)
	}

	cloneSpec, err := p.GenerateVMSpec(ctx, class, VMName, vmTemplate, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VM spec: %w", err)
	}
	// add Init data
	cloneSpec.Config.ExtraConfig = userData
	vmFolder, err := p.Finder.ResolveFolder(ctx)
	if err != nil {
		return nil, err
	}

	vm, err := p.Finder.VMByName(ctx, VMName)
	if err != nil {
		if err.(*find.NotFoundError) == nil {
			return nil, err
		}
	}

	task, err := vmTemplate.Clone(ctx, vmFolder, VMName, *cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to clone VM: %w", err)
	}

	err = task.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}

	vm, err = p.Finder.VMByName(ctx, VMName)
	if err != nil {
		return nil, err
	}

	err = p.Finder.TagInstance(ctx, vm.Reference(), instanceTags)
	if err != nil {
		return nil, err
	}

	creationDate, uuid, err := extractCreationDate(ctx, vm)
	if err != nil {
		return nil, err
	}

	powerOnTask, err := vm.PowerOn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to power on VM: %w", err)
	}
	err = powerOnTask.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("task failed: %w", err)
	}

	powerState, err := vm.PowerState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get power state: %w", err)
	}
	return NewInstance(vm, uuid, vmTemplate.InventoryPath, string(powerState), vm.Name(), *creationDate, instanceTags), err
}

// getVMConfig
func getVMConfig(ctx context.Context, vm *object.VirtualMachine, properties []string) (*types.VirtualMachineConfigInfo, error) {
	vmMo := models.VirtualMachine{
		Config: &types.VirtualMachineConfigInfo{},
	}
	err := vm.Properties(ctx, vm.Reference(), properties, &vmMo)
	if err != nil {
		return nil, err
	}
	return vmMo.Config, nil
}

func extractCreationDate(ctx context.Context, vm *object.VirtualMachine) (*time.Time, string, error) {
	config, err := getVMConfig(ctx, vm, []string{"config.createDate", "config.uuid"})
	if err != nil {
		return nil, "", err
	}

	t := config.CreateDate.UTC()
	return &t, config.Uuid, nil
}

func GenerateVMName(cluster, claim string) string {
	return fmt.Sprintf("%s-karp-%s", cluster, claim)
}

func getImageFromAnnotation(ctx context.Context, vm *object.VirtualMachine) string {
	config, err := getVMConfig(ctx, vm, []string{"config.annotation"})
	if err != nil {
		log.Log.Info(err.Error())
		return ImageNotFound
	}
	return imageFromConfig(config)
}

func imageFromConfig(config *types.VirtualMachineConfigInfo) string {
	if config == nil {
		return ImageNotFound
	}
	image := strings.TrimPrefix(config.Annotation, "cloned_from:")
	return strings.TrimPrefix(image, " ")
}

func belongsToCluster(tags map[string]string, clusterName string) bool {
	return tags[v1alpha1.ClusterNameTagKey] == clusterName
}

func (p *DefaultProvider) List(ctx context.Context) ([]*Instance, error) {
	if err := p.Finder.Session.EnsureValid(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure vsphere session is valid: %w", err)
	}

	instances := []*Instance{}
	vms, err := p.Finder.ListVMs(ctx)
	//
	if err != nil {
		log.FromContext(ctx).Error(err, "")
	}
	if len(vms) < 1 {
		return instances, nil
	}
	for _, vm := range vms {
		ps, err := vm.PowerState(ctx)
		if err != nil {
			log.FromContext(ctx).Error(err, fmt.Sprintf("failed to get power state for VM %s", vm.Name()))
		}
		// skip poweredOff machines
		if ps == "poweredOff" {
			continue
		}
		image := getImageFromAnnotation(ctx, vm)
		tags, err := p.Finder.TagsFromVM(ctx, vm)
		if err != nil {
			log.FromContext(ctx).Error(err, fmt.Sprintf("failed to get tags for VM %s", vm.Name()))
		}

		creationDate, uuid, err := extractCreationDate(ctx, vm)
		if err != nil {
			log.FromContext(ctx).Error(err, fmt.Sprintf("failed to extract creation date for VM %s", vm.Name()))
		}
		// find only VMs belonging to current cluster
		if !belongsToCluster(tags, p.ClusterName) {
			continue
		}
		instances = append(instances, NewInstance(vm, uuid, image, string(ps), vm.Name(), *creationDate, tags))
	}
	return instances, nil
}

func (p *DefaultProvider) Get(ctx context.Context, vmID string) (*Instance, error) {
	if err := p.Finder.Session.EnsureValid(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure vsphere session is valid: %w", err)
	}

	vm, err := p.Finder.GetVMByID(ctx, vmID)
	if err != nil {
		return nil, err
	}
	tags, err := p.Finder.TagsFromVM(ctx, vm)
	if err != nil {
		log.FromContext(ctx).Error(err, fmt.Sprintf("failed to get tags for VM %s", vm.Name()))
	}
	instance := NewInstanceFromVM(ctx, vm, time.Now(), tags)
	return instance, nil

}

func (p *DefaultProvider) Delete(ctx context.Context, vmID string) error {
	if err := p.Finder.Session.EnsureValid(ctx); err != nil {
		return fmt.Errorf("failed to ensure vsphere session is valid: %w", err)
	}

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
