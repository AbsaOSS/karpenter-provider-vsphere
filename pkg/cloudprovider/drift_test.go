package cloudprovider

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNodeClassDrifted(t *testing.T) {
	nodeClass := testutil.NewNodeClass().Build()
	nodeClass.Spec.ImageSelector.Pattern = "v1.35.8"

	nodeClaim := testutil.NewNodeClaim().Build()
	nodeClaim.Annotations = map[string]string{
		v1alpha1.AnnotationVsphereNodeClassHash:        nodeClass.Hash(),
		v1alpha1.AnnotationVsphereNodeClassHashVersion: v1alpha1.VsphereNodeClassHashVersion,
	}

	nodeClass.Spec.ImageSelector.Pattern = "v1.35.10"
	nodeClass.Annotations = map[string]string{
		v1alpha1.AnnotationVsphereNodeClassHash:        nodeClass.Hash(),
		v1alpha1.AnnotationVsphereNodeClassHashVersion: v1alpha1.VsphereNodeClassHashVersion,
	}

	provider := &CloudProvider{}

	reason, err := provider.isNodeClassDrifted(
		context.Background(),
		nodeClaim,
		nil,
		nodeClass,
	)

	require.NoError(t, err)
	assert.Equal(t, NodeClassDrift, reason)
}
