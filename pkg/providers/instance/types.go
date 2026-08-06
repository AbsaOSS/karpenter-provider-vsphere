package instance

import (
	"context"
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/object"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	config, err := getVMConfig(ctx, vm, []string{"config.uuid"})
	uuid := ""
	if err != nil {
		log.Log.Info(err.Error())
	} else {
		uuid = config.Uuid
	}
	instance := NewInstance(vm, uuid, getImageFromAnnotation(ctx, vm), "", vm.Name(), created, tags)
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
