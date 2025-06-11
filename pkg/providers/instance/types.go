package instance

import (
	"context"
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/object"
)

type Instance struct {
	LaunchTime time.Time
	ID         string
	State      string
	Image      string
	Name       string
	Type       string
	Tags       map[string]string
	vm         *object.VirtualMachine
}

func NewInstanceFromVM(ctx context.Context, vm *object.VirtualMachine, created time.Time, tags map[string]string) *Instance {
	instance := NewInstance(vm, vm.UUID(ctx), getImageFromAnnotation(vm), "", vm.Name(), created, tags)
	return instance
}
func NewInstance(vm *object.VirtualMachine, id, image, state, name string, created time.Time, tags map[string]string) *Instance {
	return &Instance{
		LaunchTime: created,
		State:      state,
		ID:         id,
		Image:      image,
		Name:       name,
		Type:       tags[v1alpha1.LabelInstanceSize],
		Tags:       tags,
		vm:         vm,
	}
}

func (i *Instance) GetVM() *object.VirtualMachine {
	return i.vm
}
