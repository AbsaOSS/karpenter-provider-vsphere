package cloudprovider

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

var _ cloudprovider.CloudProvider = (*CloudProvider)(nil)

const (
	NodeClassReadinessUnknownReason                              = "NodeClassReadinessUnknown"
	InstanceTypeResolutionFailedReason                           = "InstanceTypeResolutionFailed"
	CreateInstanceFailedReason                                   = "CreateInstanceFailed"
	NodeClassDrift                     cloudprovider.DriftReason = "NodeClassDrift"
)

type CloudProvider struct {
	instanceProvider instance.Provider
	kubeClient       client.Client
}

// Name returns the CloudProvider implementation name.
func (c *CloudProvider) Name() string {
	return "vsphere"
}

func (c *CloudProvider) GetSupportedNodeClasses() []status.Object {
	return []status.Object{&v1alpha1.VsphereNodeClass{}}
}

func New(instanceProvider instance.Provider, kubeClient client.Client) *CloudProvider {
	return &CloudProvider{
		instanceProvider: instanceProvider,
		kubeClient:       kubeClient,
	}
}

func (c *CloudProvider) Delete(ctx context.Context, claim *karpv1.NodeClaim) error {
	if claim.Spec.NodeClassRef == nil || claim.Spec.NodeClassRef.Name == "" {
		return fmt.Errorf("node claim %s/%s does not have a NodeClassRef", claim.Namespace, claim.Name)
	}
	nodeClass := &v1alpha1.VsphereNodeClass{}
	if err := c.kubeClient.Get(context.TODO(), types.NamespacedName{Name: claim.Spec.NodeClassRef.Name}, nodeClass); err != nil {
		return fmt.Errorf("failed to get NodeClass %s for NodeClaim %s/%s: %w", claim.Spec.NodeClassRef.Name, claim.Namespace, claim.Name, err)
	}
	if !nodeClass.DeletionTimestamp.IsZero() {
		return newTerminatingNodeClassError(nodeClass.Name)
	}
	return c.instanceProvider.Delete(context.TODO(), claim.Name)
}

func (c *CloudProvider) Create(ctx context.Context, nodeClaim *karpv1.NodeClaim) (*karpv1.NodeClaim, error) {
	nodeClass, err := c.resolveNodeClassFromNodeClaim(ctx, nodeClaim)
	if err != nil {
		return nil, cloudprovider.NewInsufficientCapacityError(fmt.Errorf("resolving node class, %w", err))
	}
	nodeClassReady := nodeClass.StatusConditions().Get(status.ConditionReady)
	if nodeClassReady.IsFalse() {
		return nil, cloudprovider.NewNodeClassNotReadyError(stderrors.New(nodeClassReady.Message))
	}
	if nodeClassReady.IsUnknown() {
		return nil, cloudprovider.NewCreateError(fmt.Errorf("resolving NodeClass readiness, NodeClass is in Ready=Unknown, %s", nodeClassReady.Message), NodeClassReadinessUnknownReason, "NodeClass is in Ready=Unknown")
	}
	if _, err = nodeClass.GetKubernetesVersion(); err != nil {
		return nil, err
	}

	instanceTypes, err := c.resolveInstanceTypes(nodeClaim, nodeClass)
	if err != nil {
		return nil, cloudprovider.NewCreateError(fmt.Errorf("resolving instance types, %w", err), InstanceTypeResolutionFailedReason, err.Error())
	}
	if len(instanceTypes) == 0 {
		return nil, cloudprovider.NewInsufficientCapacityError(fmt.Errorf("all requested instance types were unavailable during launch"))
	}
	_, err = c.instanceProvider.BeginCreate(ctx, nodeClass, nodeClaim, instanceTypes)
	if err != nil {
		return nil, cloudprovider.NewCreateError(fmt.Errorf("creating instance failed, %w", err), CreateInstanceFailedReason, err.Error())
	}
	return &karpv1.NodeClaim{}, nil
}

func (c *CloudProvider) Get(ctx context.Context, instance string) (*karpv1.NodeClaim, error) {
	return &karpv1.NodeClaim{}, nil
}

func (c *CloudProvider) List(ctx context.Context) ([]*karpv1.NodeClaim, error) {
	return []*karpv1.NodeClaim{}, nil
}

func (c *CloudProvider) RepairPolicies() []cloudprovider.RepairPolicy {
	return []cloudprovider.RepairPolicy{
		// Supported Kubelet Node Conditions
		{
			ConditionType:      corev1.NodeReady,
			ConditionStatus:    corev1.ConditionFalse,
			TolerationDuration: 30 * time.Minute,
		},
	}
}

func (c *CloudProvider) GetInstanceTypes(ctx context.Context, pool *karpv1.NodePool) ([]*cloudprovider.InstanceType, error) {
	//	pool, err := c.resolveNodeClassFromNodePool(ctx, pool)

	return []*cloudprovider.InstanceType{}, nil
}

func (c *CloudProvider) IsDrifted(ctx context.Context, claim *karpv1.NodeClaim) (cloudprovider.DriftReason, error) {
	// For now, we assume no drift detection is implemented.
	return NodeClassDrift, nil
}

func (c *CloudProvider) resolveInstanceTypes(nodeClaim *karpv1.NodeClaim, nodeClass *v1alpha1.VsphereNodeClass) ([]*cloudprovider.InstanceType, error) {
	instanceTypes := []*cloudprovider.InstanceType{}
	for n, t := range nodeClass.Spec.InstanceTypes {
		instanceType := &cloudprovider.InstanceType{
			Name: n,
			Requirements: scheduling.NewRequirements(
				scheduling.NewRequirement(corev1.LabelInstanceTypeStable, corev1.NodeSelectorOpIn, n),
				scheduling.NewRequirement(corev1.LabelArchStable, corev1.NodeSelectorOpIn, t.Arch),
				scheduling.NewRequirement(corev1.LabelOSStable, corev1.NodeSelectorOpIn, t.OS),
			),
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse(t.CPU),
				corev1.ResourceMemory:           resource.MustParse(t.Memory),
				corev1.ResourcePods:             resource.MustParse(t.MaxPods),
				corev1.ResourceEphemeralStorage: resource.MustParse(t.Storage),
			},
			Offerings: []*cloudprovider.Offering{
				{
					Requirements: scheduling.NewRequirements(
						scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn, nodeClass.Spec.ComputeCluster)),
					Price:     float64(0.0),
					Available: true,
				},
			},
		}
		instanceTypes = append(instanceTypes, instanceType)
	}

	reqs := scheduling.NewNodeSelectorRequirementsWithMinValues(nodeClaim.Spec.Requirements...)
	return lo.Filter(instanceTypes, func(i *cloudprovider.InstanceType, _ int) bool {
		return reqs.Compatible(i.Requirements, scheduling.AllowUndefinedWellKnownLabels) == nil &&
			len(i.Offerings.Compatible(reqs).Available()) > 0 &&
			resources.Fits(nodeClaim.Spec.Resources.Requests, i.Allocatable())
	}), nil
}
func (c *CloudProvider) resolveNodeClassFromNodeClaim(ctx context.Context, nodeClaim *karpv1.NodeClaim) (*v1alpha1.VsphereNodeClass, error) {
	nodeClass := &v1alpha1.VsphereNodeClass{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodeClaim.Spec.NodeClassRef.Name}, nodeClass); err != nil {
		return nil, err
	}
	// For the purposes of NodeClass CloudProvider resolution, we treat deleting NodeClasses as NotFound
	if !nodeClass.DeletionTimestamp.IsZero() {
		// For the purposes of NodeClass CloudProvider resolution, we treat deleting NodeClasses as NotFound,
		// but we return a different error message to be clearer to users
		return nil, newTerminatingNodeClassError(nodeClass.Name)
	}
	return nodeClass, nil
}

func (c *CloudProvider) resolveNodeClassFromNodePool(ctx context.Context, nodePool *karpv1.NodePool) (*v1alpha1.VsphereNodeClass, error) {
	nodeClass := &v1alpha1.VsphereNodeClass{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodePool.Spec.Template.Spec.NodeClassRef.Name}, nodeClass); err != nil {
		return nil, err
	}
	// For the purposes of NodeClass CloudProvider resolution, we treat deleting NodeClasses as NotFound
	if !nodeClass.DeletionTimestamp.IsZero() {
		// For the purposes of NodeClass CloudProvider resolution, we treat deleting NodeClasses as NotFound,
		// but we return a different error message to be clearer to users
		return nil, newTerminatingNodeClassError(nodeClass.Name)
	}
	return nodeClass, nil
}

// newTerminatingNodeClassError returns a NotFound error for handling by
func newTerminatingNodeClassError(name string) *errors.StatusError {
	qualifiedResource := schema.GroupResource{Group: apis.Group, Resource: "vspherenodeclasses"}
	err := errors.NewNotFound(qualifiedResource, name)
	err.ErrStatus.Message = fmt.Sprintf("%s %q is terminating, treating as not found", qualifiedResource.String(), name)
	return err
}
