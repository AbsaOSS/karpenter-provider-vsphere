package cloudprovider

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

func (c *CloudProvider) isNodeClassDrifted(ctx context.Context, nodeClaim *karpv1.NodeClaim, _ *karpv1.NodePool, nodeClass *v1alpha1.VsphereNodeClass) (cloudprovider.DriftReason, error) {
	if drifted := c.staticFieldsDrifted(nodeClaim, nodeClass); drifted != "" {
		err := errors.New("current node claim is drifted")
		log.FromContext(ctx).Error(err, "drifted", drifted)
		return drifted, err
	}
	return "", nil
}

func (c *CloudProvider) staticFieldsDrifted(nodeClaim *karpv1.NodeClaim, nodeClass *v1alpha1.VsphereNodeClass) cloudprovider.DriftReason {
	nodeClassHash, foundNodeClassHash := nodeClass.Annotations[v1alpha1.AnnotationVsphereNodeClassHash]
	nodeClassHashVersion, foundNodeClassHashVersion := nodeClass.Annotations[v1alpha1.AnnotationVsphereNodeClassHashVersion]
	nodeClaimHash, foundNodeClaimHash := nodeClaim.Annotations[v1alpha1.AnnotationVsphereNodeClassHash]
	nodeClaimHashVersion, foundNodeClaimHashVersion := nodeClaim.Annotations[v1alpha1.AnnotationVsphereNodeClassHashVersion]

	if !foundNodeClassHash || !foundNodeClaimHash || !foundNodeClassHashVersion || !foundNodeClaimHashVersion {
		return ""
	}
	if nodeClassHashVersion != nodeClaimHashVersion {
		return ""
	}
	return lo.Ternary(nodeClassHash != nodeClaimHash, NodeClassDrift, "")
}
