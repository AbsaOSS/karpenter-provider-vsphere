package cloudprovider

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/utils"
	"github.com/awslabs/operatorpkg/status"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
	id, err := utils.ParseInstanceID(claim.Status.ProviderID)
	if err != nil {
		return fmt.Errorf("getting instance ID, %w", err)
	}
	return c.instanceProvider.Delete(context.TODO(), id)
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
	nodePoolName := nodeClaim.Labels[karpv1.NodePoolLabelKey]
	instance, err := c.instanceProvider.Create(ctx, nodeClass, nodeClaim, instanceTypes, nodePoolName)
	if err != nil {
		return nil, cloudprovider.NewCreateError(fmt.Errorf("creating instance failed, %w", err), CreateInstanceFailedReason, err.Error())
	}
	claim := c.instanceToNodeClaim(instance, instanceTypes[0])

	return claim, nil
}

//nolint:gocyclo
func (c *CloudProvider) instanceToNodeClaim(i *instance.Instance, instanceType *cloudprovider.InstanceType) *karpv1.NodeClaim {
	nodeClaim := &karpv1.NodeClaim{}
	labels := map[string]string{}
	annotations := map[string]string{}

	if instanceType != nil {
		labels = utils.GetAllSingleValuedRequirementLabels(instanceType)

		resourceFilter := func(name corev1.ResourceName, value resource.Quantity) bool {
			return !resources.IsZero(value)
		}

		nodeClaim.Status.Capacity = lo.PickBy(instanceType.Capacity, resourceFilter)
		nodeClaim.Status.Allocatable = lo.PickBy(instanceType.Allocatable(), resourceFilter)
	}
	//TODO: Figure out TopologyZone label
	labels[corev1.LabelTopologyZone] = i.Tags[corev1.LabelTopologyZone]

	if v, ok := i.Tags[karpv1.NodePoolLabelKey]; ok {
		labels[karpv1.NodePoolLabelKey] = v
	}

	nodeClaim.Name = GenerateNodeClaimName(i.Name, i.Tags[v1alpha1.ClusterNameTagKey])
	nodeClaim.Labels = labels
	nodeClaim.Annotations = annotations
	nodeClaim.CreationTimestamp = metav1.Time{Time: i.LaunchTime}
	if i.State == "powerOff" {
		nodeClaim.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	}
	nodeClaim.Status.ProviderID = fmt.Sprintf("vsphere://%s", i.ID)
	nodeClaim.Status.ImageID = i.Image

	return nodeClaim
}

func GenerateNodeClaimName(vmName, clusterName string) string {
	return strings.TrimLeft(fmt.Sprintf("%s-karp-", clusterName), vmName)
}

func (c *CloudProvider) Get(ctx context.Context, instance string) (*karpv1.NodeClaim, error) {
	return &karpv1.NodeClaim{}, nil
}

func (c *CloudProvider) List(ctx context.Context) ([]*karpv1.NodeClaim, error) {
	instances, err := c.instanceProvider.List(ctx)
	var nodeClaims []*karpv1.NodeClaim
	if err != nil {
		return nil, fmt.Errorf("listing instances, %w", err)
	}
	for _, instance := range instances {
		instanceType, err := c.resolveInstanceTypeFromInstance(ctx, instance)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to resolve instance type")
			return nil, fmt.Errorf("resolving instance type, %w", err)
		}
		nodeClaim := c.instanceToNodeClaim(instance, instanceType)
		log.FromContext(ctx).Info("converted instance to nodeclaim", "nodeClaimName", nodeClaim.Name)
		nodeClaims = append(nodeClaims, nodeClaim)
	}
	log.FromContext(ctx).Info("listed all nodeclaims", "count", len(nodeClaims))
	return nodeClaims, nil
}

func (c *CloudProvider) resolveInstanceTypeFromInstance(ctx context.Context, instance *instance.Instance) (*cloudprovider.InstanceType, error) {
	nodePool, err := c.resolveNodePoolFromInstance(ctx, instance)
	if err != nil {
		return nil, client.IgnoreNotFound(fmt.Errorf("resolving nodepool, %w", err))
	}
	instanceTypes, err := c.GetInstanceTypes(ctx, nodePool)
	if err != nil {
		return nil, fmt.Errorf("resolving instance types, %w", err)
	}

	instanceType, _ := lo.Find(instanceTypes, func(i *cloudprovider.InstanceType) bool {
		return i.Name == instance.Type
	})

	if instanceType == nil {
		return nil, fmt.Errorf("instance type %s not found in offerings", instance.Type)
	}
	return instanceType, nil
}

func (c *CloudProvider) resolveNodePoolFromInstance(ctx context.Context, instance *instance.Instance) (*karpv1.NodePool, error) {
	nodePoolName := instance.Tags[karpv1.NodePoolLabelKey]
	if nodePoolName == "" {
		return nil, fmt.Errorf("missing nodepool label")
	}

	var nodePool karpv1.NodePool
	if err := c.kubeClient.Get(ctx, client.ObjectKey{Name: nodePoolName}, &nodePool); err != nil {
		return nil, err
	}

	return &nodePool, nil
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
	nodeClass, err := c.resolveNodeClassFromNodePool(ctx, pool)
	if err != nil {
		return nil, fmt.Errorf("resolving node class, %w", err)
	}
	instanceTypes := instanceTypesFromNodeClass(nodeClass)

	return instanceTypes, err
}

func (c *CloudProvider) IsDrifted(ctx context.Context, claim *karpv1.NodeClaim) (cloudprovider.DriftReason, error) {
	nodePoolName, ok := claim.Labels[karpv1.NodePoolLabelKey]
	if !ok {
		return "", nil
	}
	nodePool := &karpv1.NodePool{}
	if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: nodePoolName}, nodePool); err != nil {
		return "", client.IgnoreNotFound(err)
	}
	if nodePool.Spec.Template.Spec.NodeClassRef == nil {
		return "", nil
	}
	nodeClass, err := c.resolveNodeClassFromNodePool(ctx, nodePool)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("NodeClass not found, treating as drifted")
			return NodeClassDrift, nil
		}
		return "", client.IgnoreNotFound(fmt.Errorf("resolving node class, %w", err))
	}
	driftReason, err := c.isNodeClassDrifted(ctx, claim, nodePool, nodeClass)
	if err != nil {
		return "", err
	}
	return driftReason, nil
}

func instanceTypesFromNodeClass(nodeClass *v1alpha1.VsphereNodeClass) []*cloudprovider.InstanceType {
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
			Overhead: &cloudprovider.InstanceTypeOverhead{},
			Offerings: []*cloudprovider.Offering{
				{
					Requirements: scheduling.NewRequirements(
						scheduling.NewRequirement(corev1.LabelTopologyZone, corev1.NodeSelectorOpIn)),
					Price:     float64(0.0),
					Available: true,
				},
			},
		}
		instanceTypes = append(instanceTypes, instanceType)
	}
	return instanceTypes
}
func (c *CloudProvider) resolveInstanceTypes(nodeClaim *karpv1.NodeClaim, nodeClass *v1alpha1.VsphereNodeClass) ([]*cloudprovider.InstanceType, error) {
	instanceTypes := instanceTypesFromNodeClass(nodeClass)
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
