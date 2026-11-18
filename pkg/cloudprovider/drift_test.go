package cloudprovider

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/testutil"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
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
	reason, err := provider.isNodeClassDrifted(context.Background(), nodeClaim, nil, nodeClass)
	require.NoError(t, err)
	assert.Equal(t, NodeClassDrift, reason)
}

func TestIsNodeClassDriftedImageTemplateID(t *testing.T) {
	for _, test := range []struct {
		name            string
		imageID         string
		imageTemplateID string
		want            cloudprovider.DriftReason
	}{
		{
			name:            "matching image",
			imageID:         "/DC0/vm/ubuntu-v1",
			imageTemplateID: "/DC0/vm/ubuntu-v1",
		},
		{
			name:            "different image",
			imageID:         "/DC0/vm/ubuntu-v1",
			imageTemplateID: "/DC0/vm/ubuntu-v2",
			want:            NodeImageDrift,
		},
		{
			name:            "missing nodeclaim image",
			imageTemplateID: "/DC0/vm/ubuntu-v2",
			want:            NodeImageDrift,
		},
		{
			name:    "unresolved template does not trigger drift",
			imageID: "/DC0/vm/ubuntu-v1",
		},
		{
			name: "both images empty",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			nodeClass := testutil.NewNodeClass().Build()
			nodeClass.Status.ImageTemplateID = test.imageTemplateID
			nodeClaim := testutil.NewNodeClaim().Build()
			nodeClaim.Status.ImageID = test.imageID
			provider := &CloudProvider{}

			// Image drift is independent of spec hashes and their availability.
			for _, withHashes := range []bool{false, true} {
				if withHashes {
					annotations := map[string]string{
						v1alpha1.AnnotationVsphereNodeClassHash:        nodeClass.Hash(),
						v1alpha1.AnnotationVsphereNodeClassHashVersion: v1alpha1.VsphereNodeClassHashVersion,
					}
					nodeClass.Annotations = annotations
					nodeClaim.Annotations = annotations
				} else {
					nodeClass.Annotations = nil
					nodeClaim.Annotations = nil
				}
				reason, err := provider.isNodeClassDrifted(context.Background(), nodeClaim, nil, nodeClass)
				require.NoError(t, err)
				assert.Equal(t, test.want, reason)
			}
		})
	}
}
