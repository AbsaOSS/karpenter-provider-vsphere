package instance

import (
	"time"

	v1alpha1 "github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
)

type Instance struct {
	LaunchTime time.Time
	ID         string
	State      string
	Image      string
	Name       string
	Type       string
	Tags       map[string]string
}

func NewInstance(id, image, state, name string, created time.Time, tags map[string]string) *Instance {
	return &Instance{
		LaunchTime: created,
		State:      state,
		ID:         id,
		Image:      image,
		Name:       name,
		Type:       tags[v1alpha1.LabelInstanceSize],
		Tags:       tags,
	}
}
