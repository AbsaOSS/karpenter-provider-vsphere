package instance

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/vapi/tags"
)

func (p *DefaultProvider) CreateOrUpdateTags(ctx context.Context, instanceTags map[string]string) ([]string, error) {
	mgr := p.VsphereInfo.TagManager
	tagIDs := make([]string, 0, len(instanceTags))
	for k, v := range instanceTags {
		// Normalize Vsphere tag to fullfil CPI requirements
		if k == "topology.kubernetes.io/zone" {
			k = "k8s-zone"
		}
		category, err := CreateOrUpdateCategory(ctx, mgr, k)
		if err != nil {
			return nil, err
		}
		tag, err := GetOrCreateTag(ctx, mgr, v, category)
		if err != nil {
			return nil, fmt.Errorf("failed to create or get tag %s: %w", v, err)
		}
		tagIDs = append(tagIDs, tag)
	}
	return tagIDs, nil
}

func CreateOrUpdateCategory(ctx context.Context, mgr *tags.Manager, name string) (string, error) {
	vsphereCategory, err := mgr.GetCategory(ctx, name)
	if err != nil {
		fmt.Println("error getting vsphere category", err)
		vsphereCategory, err := mgr.CreateCategory(ctx, getCategoryObject(name))
		if err != nil {
			return "", fmt.Errorf("failed to create vsphere category %s: %w", name, err)
		}
		return vsphereCategory, nil
	}
	return vsphereCategory.ID, nil
}

func GetOrCreateTag(ctx context.Context, mgr *tags.Manager, name, categoryID string) (string, error) {
	tag, err := mgr.GetTagForCategory(ctx, name, categoryID)
	if err != nil {
		fmt.Println("error getting vsphere tag", err)
		id, err := mgr.CreateTag(ctx, &tags.Tag{
			Description: "karpenter managed tag",
			Name:        name,
			CategoryID:  categoryID,
		})
		if err != nil {
			return "", err
		}
		return id, nil
	}
	return tag.ID, nil
}

func getCategoryObject(name string) *tags.Category {
	return &tags.Category{
		Name:            name,
		Description:     "Karpenter managed category",
		Cardinality:     "MULTIPLE",
		AssociableTypes: []string{"VirtualMachine"},
	}
}
