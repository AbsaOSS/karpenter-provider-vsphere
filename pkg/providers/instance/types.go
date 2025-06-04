package instance

import "time"

type Instance struct {
	LaunchTime time.Time
	ID         string
	State      string
	Image      string
	Name       string
}

func NewInstance(id, image, state, name string, created time.Time) *Instance {
	return &Instance{
		LaunchTime: created,
		State:      state,
		ID:         id,
		Image:      image,
		Name:       name,
	}
}
