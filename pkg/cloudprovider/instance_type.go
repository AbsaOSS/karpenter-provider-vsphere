package cloudprovider

import (
	"context"
	"fmt"
)

type Family string

const (
	FamilyCompute Family = "c"
	FamilyStorage Family = "s"
	FamilyMemory  Family = "m"
)

func (f Family) MemoryToVCPURatio() int {
	switch f {
	case FamilyCompute:
		return 2
	case FamilyStorage:
		return 4
	case FamilyMemory:
		return 8
	default:
		return 0
	}
}

func (f Family) IsValid() bool {
	switch f {
	case FamilyCompute, FamilyStorage, FamilyMemory:
		return true
	default:
		return false
	}
}

// Catalog contains all supported instance profiles.
type InstanceProfile struct {
	Family Family
	Size   int
	VCPU   int
}

func (p InstanceProfile) Name() string {
	return fmt.Sprintf("%s-%dx", p.Family, p.Size)
}

func (p InstanceProfile) Price() float64 {
	return (float64(p.VCPU)*0.25 + float64(p.MemGiB())*0.01)
}

type InstanceProfilesProvider interface {
	List(ctx context.Context) ([]*InstanceProfile, error)
	Get(ctx context.Context, name string) (*InstanceProfile, error)
}

func (p InstanceProfile) MemGiB() int {
	return p.VCPU * p.Family.MemoryToVCPURatio()
}

type InstanceProfilesStaticProvider struct {
}

var InstanceTypeSizes = [...]int{
	1, 2, 4, 9,
	16, 32, 48, 64,
	100, 128, 192, 256,
}

func getProfilesCatalog() map[string]InstanceProfile {
	catalog := make(map[string]InstanceProfile)
	families := []Family{FamilyCompute, FamilyStorage, FamilyMemory}
	for _, family := range families {
		for _, size := range InstanceTypeSizes {
			p := InstanceProfile{
				Family: family,
				Size:   size,
				VCPU:   size,
			}
			catalog[p.Name()] = p
		}
	}
	return catalog
}

var InstanceProfilesCatalog = getProfilesCatalog()

// LookupProfile returns a profile by name.
func (r InstanceProfilesStaticProvider) Get(ctx context.Context, name string) (*InstanceProfile, error) {
	profile, ok := InstanceProfilesCatalog[name]
	if !ok {
		return nil, fmt.Errorf("unknown instance type %q", name)
	}

	p := profile
	return &p, nil
}

func (r InstanceProfilesStaticProvider) List(ctx context.Context) ([]*InstanceProfile, error) {
	profiles := make([]*InstanceProfile, 0, len(InstanceProfilesCatalog))
	for name := range InstanceProfilesCatalog {
		profile := InstanceProfilesCatalog[name]
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}
