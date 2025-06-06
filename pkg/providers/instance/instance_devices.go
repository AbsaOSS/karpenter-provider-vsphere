package instance

import (
	"context"
	"fmt"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const ethCardType = "vmxnet3"

func (d *DefaultProvider) getNetworkSpecs(ctx context.Context, networkName object.NetworkReference, devices object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error) {

	deviceSpecs := []types.BaseVirtualDeviceConfigSpec{}

	// Remove any existing NICs
	for _, dev := range devices.SelectByType((*types.VirtualEthernetCard)(nil)) {
		deviceSpecs = append(deviceSpecs, &types.VirtualDeviceConfigSpec{
			Device:    dev,
			Operation: types.VirtualDeviceConfigSpecOperationRemove,
		})
	}

	// Add new NICs based on the machine config.
	key := int32(-100)

	backing, err := networkName.EthernetCardBackingInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create new ethernet card: %v", err)
	}
	dev, err := object.EthernetCardTypes().CreateEthernetCard(ethCardType, backing)
	if err != nil {
		return nil, fmt.Errorf("unable to create new ethernet card: %v", err)
	}

	// Get the actual NIC object. This is safe to assert without a check
	// because "object.EthernetCardTypes().CreateEthernetCard" returns a
	// "types.BaseVirtualEthernetCard" as a "types.BaseVirtualDevice".
	nic := dev.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()

	// Assign a temporary device key to ensure that a unique one will be
	// generated when the device is created.
	nic.Key = key

	return append(deviceSpecs, &types.VirtualDeviceConfigSpec{
		Device:    dev,
		Operation: types.VirtualDeviceConfigSpecOperationAdd,
	}), err
}

func (p *DefaultProvider) GetDeviceSpec(ctx context.Context, class *v1alpha1.VsphereNodeClass, diskSize int64) ([]types.BaseVirtualDeviceConfigSpec, error) {
	var deviceChange []types.BaseVirtualDeviceConfigSpec
	vmTemplate, err := p.VsphereInfo.Finder.VirtualMachine(ctx, class.Spec.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to find VM template: %w", err)
	}
	network, err := p.VsphereInfo.Finder.Network(ctx, class.Spec.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to find network: %w", err)
	}

	devList, err := vmTemplate.Device(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device list from VM template: %w", err)
	}
	disks := devList.SelectByType((*types.VirtualDisk)(nil))
	if len(disks) == 0 {
		return nil, fmt.Errorf("invalid disk count: %d", len(disks))
	}

	// There is at least one disk
	primaryDisk := disks[0].(*types.VirtualDisk)
	primaryCloneCapacityKB := GiToKb(diskSize)
	primaryDiskConfigSpec, err := getDiskConfigSpec(primaryDisk, primaryCloneCapacityKB)
	if err != nil {
		return nil, fmt.Errorf("error getting disk config spec for primary disk: %v", err)
	}

	netSpec, err := p.getNetworkSpecs(ctx, network, devList)
	if err != nil {
		return nil, fmt.Errorf("failed to get network specs: %w", err)
	}
	return append(deviceChange, append(netSpec, primaryDiskConfigSpec)...), nil

}
func GiToKb(size int64) int64 {
	return size * 1024 * 1024
}
func GiToMb(size int64) int64 {
	return size * 1024
}

func getDiskConfigSpec(disk *types.VirtualDisk, diskCloneCapacityKB int64) (types.BaseVirtualDeviceConfigSpec, error) {
	switch {
	case diskCloneCapacityKB == 0:
		// No disk size specified for the clone. Default to the template disk capacity.
	case diskCloneCapacityKB > 0 && diskCloneCapacityKB >= disk.CapacityInKB:
		disk.CapacityInKB = diskCloneCapacityKB
	case diskCloneCapacityKB > 0 && diskCloneCapacityKB < disk.CapacityInKB:
		return nil, fmt.Errorf(
			"can't resize template disk down, initial capacity is larger: %dKiB > %dKiB",
			disk.CapacityInKB, diskCloneCapacityKB)
	}

	return &types.VirtualDeviceConfigSpec{
		Operation: types.VirtualDeviceConfigSpecOperationEdit,
		Device:    disk,
	}, nil
}
