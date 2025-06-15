package finder

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (t *Provider) getObjectByType(ctx context.Context, tags []*tags.Tag, objT string) (*types.ManagedObjectReference, error) {
	objTagCount := map[types.ManagedObjectReference]int{}
	for _, tag := range tags {
		objs, err := t.TagManager.ListAttachedObjects(ctx, tag.ID)
		if err != nil {
			return nil, err
		}
		for _, obj := range objs {
			if obj.Reference().Type == objT {
				objTagCount[obj.Reference()]++
			}
		}
	}
	for obj, count := range objTagCount {
		if count == len(tags) {
			return &obj, nil
		}
	}
	return nil, fmt.Errorf("no %s objects matching tag count %d", objT, len(tags))
}

func (t *Provider) getObjectByTag(ctx context.Context, taglist map[string]string, typeName string) (object.Reference, error) {
	var vsphereTags []*tags.Tag
	for k, v := range taglist {
		tag, err := t.getTagID(ctx, k, v)
		if err != nil {
			return nil, err
		}
		vsphereTags = append(vsphereTags, tag)
	}
	obj, err := t.getObjectByType(ctx, vsphereTags, typeName)
	if err != nil {
		return nil, err
	}
	return object.NewReference(t.Client, obj.Reference()), nil
}

func (t *Provider) DCByTag(ctx context.Context, tag map[string]string) (*object.Datacenter, error) {
	ref, err := t.getObjectByTag(ctx, tag, "Datacenter")
	return ref.(*object.Datacenter), err
}

func (t *Provider) PoolByTag(ctx context.Context, tag map[string]string) (*object.ResourcePool, error) {
	ref, err := t.getObjectByTag(ctx, tag, "ClusterComputeResource")
	if err != nil {
		return nil, err
	}
	refObj := ref.(*object.ClusterComputeResource)
	return GetRootResourcePool(ctx, refObj)
}

func GetRootResourcePool(ctx context.Context, cluster *object.ClusterComputeResource) (*object.ResourcePool, error) {
	// This returns the root resource pool of the cluster
	return cluster.ResourcePool(ctx)
}

func (t *Provider) NetworkByTag(ctx context.Context, tag map[string]string) (*object.NetworkReference, error) {
	ref, err := t.getObjectByTag(ctx, tag, "Network")
	if err != nil {
		return nil, err
	}
	netObj := ref.(*object.Network)
	var netRef object.NetworkReference = netObj
	return &netRef, nil
}

func (t *Provider) DatastoreByTag(ctx context.Context, tag map[string]string) (*object.Datastore, error) {
	ref, err := t.getObjectByTag(ctx, tag, "Datastore")
	return ref.(*object.Datastore), err
}

func (t *Provider) ImageByTag(ctx context.Context, tag map[string]string) (*object.VirtualMachine, error) {
	ref, err := t.getObjectByTag(ctx, tag, "VirtualMachine")
	if err != nil {
		return nil, err
	}
	vm := ref.(*object.VirtualMachine)
	if !t.isTemplate(ctx, vm) {
		return nil, fmt.Errorf("failed to find VirtualMachine template")
	}
	return vm, nil
}

func (t *Provider) getTagID(ctx context.Context, k, v string) (*tags.Tag, error) {
	return t.TagManager.GetTagForCategory(ctx, v, k)
}

func (t *Provider) TagInstance(ctx context.Context, obj types.ManagedObjectReference, tags map[string]string) error {
	tagIDs, err := t.CreateOrUpdateTags(ctx, tags)
	if err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		err = t.TagManager.AttachTag(ctx, tagID, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Provider) CreateOrUpdateTags(ctx context.Context, instanceTags map[string]string) ([]string, error) {
	tagIDs := make([]string, 0, len(instanceTags))
	for k, v := range instanceTags {
		// Normalize Vsphere tag to fullfil CPI requirements
		if k == "topology.kubernetes.io/zone" {
			k = "k8s-zone"
		}
		category, err := t.CreateOrUpdateCategory(ctx, k)
		if err != nil {
			return nil, err
		}
		tag, err := t.GetOrCreateTag(ctx, v, category)
		if err != nil {
			return nil, fmt.Errorf("failed to create or get tag %s: %w", v, err)
		}
		tagIDs = append(tagIDs, tag)
	}
	return tagIDs, nil
}

func (t *Provider) CreateOrUpdateCategory(ctx context.Context, name string) (string, error) {
	vsphereCategory, err := t.TagManager.GetCategory(ctx, name)
	if err != nil {
		fmt.Println("error getting vsphere category", err)
		vsphereCategory, err := t.TagManager.CreateCategory(ctx, getCategoryObject(name))
		if err != nil {
			return "", fmt.Errorf("failed to create vsphere category %s: %w", name, err)
		}
		return vsphereCategory, nil
	}
	return vsphereCategory.ID, nil
}

func (t *Provider) GetOrCreateTag(ctx context.Context, name, categoryID string) (string, error) {
	tag, err := t.TagManager.GetTagForCategory(ctx, name, categoryID)
	if err != nil {
		fmt.Println("error getting vsphere tag", err)
		id, err := t.TagManager.CreateTag(ctx, &tags.Tag{
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

func (t *Provider) TagsFromVM(ctx context.Context, vm *object.VirtualMachine) (map[string]string, error) {
	tagsAttached, err := t.TagManager.ListAttachedTags(ctx, vm.Reference())
	if err != nil {
		log.FromContext(ctx).Error(err, fmt.Sprintf("failed to list tags for VM %s", vm.Name()))
	}
	return extractTagInfo(ctx, t.TagManager, tagsAttached)

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
