package instance

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	models "github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	corecloudprovider "sigs.k8s.io/karpenter/pkg/cloudprovider"
)

type Provider interface {
	Create(context.Context, *v1alpha1.VsphereNodeClass, *karpv1.NodeClaim, []*corecloudprovider.InstanceType, string) (*Instance, error)
	Get(context.Context, string) (*models.VirtualMachine, error)
	List(context.Context) ([]*Instance, error)
	Delete(context.Context, string) error
	Update(context.Context, string, *models.VirtualMachine) error
}

var _ Provider = (*DefaultProvider)(nil)

type VsphereInfo struct {
	PoolRef      types.ManagedObjectReference
	DatastoreRef types.ManagedObjectReference
	Folder       *object.Folder
	Datacenter   *object.Datacenter
	TagManager   *tags.Manager
	Finder       *find.Finder
	Client       *govmomi.Client
}

type DefaultProvider struct {
	VsphereInfo *VsphereInfo
	ClusterName string
	VsphereZone string
	kubeClient  kubernetes.Interface
}

func NewDefaultProvider(v *VsphereInfo, kube kubernetes.Interface, clusterName, zone string) *DefaultProvider {
	return &DefaultProvider{
		VsphereInfo: v,
		ClusterName: clusterName,
		VsphereZone: zone,
		kubeClient:  kube,
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

	diskAndNet, err := p.GetDeviceSpec(ctx, class, instanceType.Capacity.Storage().Value())
	if err != nil {
		return nil, fmt.Errorf("failed to get device spec: %w", err)
	}
	initData, err := p.GetInitData(ctx, class, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get init data: %w", err)
	}
	return &types.VirtualMachineCloneSpec{
		Template: false,
		Location: *locationSpec,
		Config: &types.VirtualMachineConfigSpec{
			Name:         name,
			Annotation:   fmt.Sprintf("cloned_from:%s", class.Spec.Image),
			NumCPUs:      int32(instanceType.Capacity.Cpu().Value()),
			MemoryMB:     instanceType.Capacity.Memory().ToDec().Value() * 1024,
			GuestId:      string(types.VirtualMachineGuestOsIdentifierOtherLinux64Guest), // This should be adjusted based on the OS type in the instance type.
			DeviceChange: diskAndNet,
			ExtraConfig:  initData,
		},
		PowerOn: true,
	}, nil
}

func (p *DefaultProvider) GetInitData(ctx context.Context, class *v1alpha1.VsphereNodeClass, nodeName string) ([]types.BaseOptionValue, error) {
	metaData := &Config{}
	// Set the metadata with the local hostname
	metaDataRaw := []byte(fmt.Sprintf("local-hostname: \"%s\"", nodeName))
	userDataTplBase64 := class.Spec.UserData.TemplateBase64
	decodedUserData, err := base64.StdEncoding.DecodeString(userDataTplBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user data template: %w", err)
	}
	templateValuesSecret, err := p.kubeClient.CoreV1().Secrets(class.Spec.UserData.Values.Namespace).Get(ctx, class.Spec.UserData.Values.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get user data values secret: %w", err)
	}
	values := templateValuesSecret.Data
	strValues := map[string]string{}
	for k, v := range values {
		strValues[k] = string(v)
	}
	tmpl, err := template.New("cloud-data").Parse(string(decodedUserData))
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, strValues)
	if err != nil {
		return nil, fmt.Errorf("failed to execute user data template: %w", err)
	}
	// Ignition userdata
	if class.Spec.UserData.Type == v1alpha1.UserDataTypeCloudInit {
		metaData.SetCloudInitMetadata(buf.Bytes())
	}
	// Cloud-init userdata
	if class.Spec.UserData.Type == v1alpha1.UserDataTypeIgnition {
		metaData.SetIgnitionUserData(buf.Bytes())
	}
	// Set metadata
	metaData.SetMetadata(metaDataRaw)

	return metaData.Extract(), nil

}

type NormalTag struct {
	Key   string
	Value string
}

func (p *DefaultProvider) GenerateTarget(ctx context.Context, class *v1alpha1.VsphereNodeClass) (*types.VirtualMachineRelocateSpec, error) {
	return &types.VirtualMachineRelocateSpec{
		Datastore: &p.VsphereInfo.DatastoreRef,
		Pool:      &p.VsphereInfo.PoolRef,
	}, nil
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
		v1alpha1.LabelInstanceMemory: fmt.Sprintf("%d", GiToMb(instanceType.Capacity.Memory().ToDec().Value())),
	}

	cloneSpec, err := p.GenerateVMSpec(ctx, class, VMName, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VM spec: %w", err)
	}

	vmTemplate, err := p.VsphereInfo.Finder.VirtualMachine(ctx, class.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM template: %w", err)
	}
	task, err := vmTemplate.Clone(ctx, p.VsphereInfo.Folder, GenerateVMName(p.ClusterName, claim.Name), *cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to clone VM: %w", err)
	}
	vm, err := p.VsphereInfo.Finder.VirtualMachine(ctx, VMName)
	if err != nil {
		return nil, fmt.Errorf("failed to find cloned VM: %w", err)
	}

	tagIDs, err := p.CreateOrUpdateTags(ctx, instanceTags)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update tags: %w", err)
	}
	for _, tagID := range tagIDs {
		err = p.VsphereInfo.TagManager.AttachTag(ctx, tagID, vm.Reference())
		if err != nil {
			return nil, fmt.Errorf("failed to attach tag to VM: %w", err)
		}
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
	return NewInstance(vm.UUID(ctx), class.Spec.Image, string(powerState), vm.Name(), *creationDate, instanceTags), err
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

func getImageFromAnnotation(vm *object.VirtualMachine) string {
	var annotation string
	err := vm.Properties(context.Background(), vm.Reference(), []string{"config.config"}, &annotation)
	if err != nil {
		annotation = "image_not_found"
	}
	return strings.TrimPrefix(annotation, "cloned_from:")
}

func extractTagInfo(ctx context.Context, tagManager *tags.Manager, tagIDs []string) (map[string]string, error) {
	tags := make(map[string]string)
	for _, tagID := range tagIDs {
		tag, err := tagManager.GetTag(ctx, tagID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tag %s: %w", tagID, err)
		}
		cat, err := tagManager.GetCategory(ctx, tag.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get category for tag %s: %w", tagID, err)
		}
		if cat.Name == "k8s-zone" {
			// Normalize Vsphere tag to fulfill CPI requirements
			cat.Name = corev1.LabelTopologyZone
		}
		tags[cat.Name] = tag.Name
	}
	return tags, nil

}
func (p *DefaultProvider) List(ctx context.Context) ([]*Instance, error) {
	instances := []*Instance{}
	vms, err := p.VsphereInfo.Finder.VirtualMachineList(ctx, p.VsphereInfo.Folder.InventoryPath+"/*")
	if err != nil {
		fmt.Printf("Failed to list VMs: %v\n", err)
	}
	for _, vm := range vms {
		image := getImageFromAnnotation(vm)
		tagsAttached, err := p.VsphereInfo.TagManager.ListAttachedTags(ctx, vm.Reference())
		if err != nil {
			fmt.Printf("Failed to list tags for VM %s: %v\n", vm.Name(), err)
		}
		tags, err := extractTagInfo(ctx, p.VsphereInfo.TagManager, tagsAttached)
		if err != nil {
			fmt.Printf("Failed to extract tag info for VM %s: %v\n", vm.Name(), err)
		}
		ps, err := vm.PowerState(ctx)
		if err != nil {
			fmt.Printf("Failed to get power state for VM %s: %v\n", vm.Name(), err)
		}
		creationDate, err := extractCreationDate(ctx, vm)
		if err != nil {
			fmt.Printf("Failed to extract creation date for VM %s: %v\n", vm.Name(), err)
		}
		instances = append(instances, NewInstance(vm.UUID(ctx), image, string(ps), vm.Name(), *creationDate, tags))
	}
	return instances, nil
}

func (p *DefaultProvider) Delete(ctx context.Context, vmID string) error {
	ptrBool := false
	vClient := p.VsphereInfo.Client.Client
	searchIndex := object.NewSearchIndex(vClient)
	vmRef, err := searchIndex.FindByUuid(ctx, p.VsphereInfo.Datacenter, vmID, true, &ptrBool)
	if err != nil {
		return err
	}
	if vmRef == nil {
		// VM not found, nothing to delete
		return corecloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("vmRef not found"))
	}
	vm := object.NewVirtualMachine(vClient, vmRef.Reference())
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

func (p *DefaultProvider) Update(ctx context.Context, vmID string, update *models.VirtualMachine) error {
	fmt.Println("Updating VM - not implemented yet")
	return nil
}
