package status

import (
	"context"
	"errors"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	operatorstatus "github.com/awslabs/operatorpkg/status"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/object"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fakeImageProvider struct {
	image    *object.VirtualMachine
	err      error
	selector v1alpha1.ImageSelectorTerm
}

func (p *fakeImageProvider) ResolveImage(_ context.Context, selector v1alpha1.ImageSelectorTerm) (*object.VirtualMachine, error) {
	p.selector = selector
	return p.image, p.err
}

type fakeKubernetesVersionProvider struct{}

func (fakeKubernetesVersionProvider) KubeServerVersion(context.Context) (string, error) {
	return "1.35.0", nil
}

func TestImageReconcile(t *testing.T) {
	for _, selector := range []v1alpha1.ImageSelectorTerm{
		{Pattern: "ubuntu-*"},
		{Tags: map[string]string{"os": "ubuntu"}},
	} {
		nodeClass := &v1alpha1.VsphereNodeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "default", Generation: 1},
			Spec:       v1alpha1.VsphereNodeClassSpec{ImageSelector: selector},
		}
		provider := &fakeImageProvider{image: &object.VirtualMachine{
			Common: object.Common{InventoryPath: "/DC0/vm/templates/ubuntu"},
		}}
		reconciler := NewVsphereImageReconciler(provider)

		_, err := reconciler.Reconcile(context.Background(), nodeClass)
		require.NoError(t, err)
		require.Equal(t, selector, provider.selector)
		require.Equal(t, provider.image.InventoryPath, nodeClass.Status.ImageTemplateID)
		require.True(t, nodeClass.StatusConditions().Get(v1alpha1.ConditionTypeImageReady).IsTrue())

		require.False(t, nodeClass.StatusConditions().Get(operatorstatus.ConditionReady).IsTrue())

		nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeKubernetesVersionReady)
		require.False(t, nodeClass.StatusConditions().Get(operatorstatus.ConditionReady).IsTrue())

		nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeResourcePoolReady)
		nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeDatastoreReady)
		nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeNetworkReady)
		require.True(t, nodeClass.StatusConditions().Get(operatorstatus.ConditionReady).IsTrue())

		provider.image.InventoryPath = "/DC0/vm/templates/ubuntu-new"
		nodeClass.Generation++
		_, err = reconciler.Reconcile(context.Background(), nodeClass)
		require.NoError(t, err)
		require.Equal(t, provider.image.InventoryPath, nodeClass.Status.ImageTemplateID)
		require.Equal(t, nodeClass.Generation, nodeClass.StatusConditions().Get(v1alpha1.ConditionTypeImageReady).ObservedGeneration)
	}
}

func TestImageReconcileFailureAndRecovery(t *testing.T) {
	for _, test := range []struct {
		name  string
		image *object.VirtualMachine
		err   error
	}{
		{name: "lookup failure", err: errors.New("image not found")},
		{name: "nil image"},
		{name: "missing inventory path", image: &object.VirtualMachine{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			nodeClass := &v1alpha1.VsphereNodeClass{}
			nodeClass.Status.ImageTemplateID = "/DC0/vm/old-template"
			nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeImageReady)
			provider := &fakeImageProvider{image: test.image, err: test.err}
			reconciler := NewVsphereImageReconciler(provider)

			_, err := reconciler.Reconcile(context.Background(), nodeClass)
			require.Error(t, err)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
			}
			require.Empty(t, nodeClass.Status.ImageTemplateID)
			condition := nodeClass.StatusConditions().Get(v1alpha1.ConditionTypeImageReady)
			require.True(t, condition.IsFalse())
			require.Equal(t, "ImageResolutionFailed", condition.Reason)
			require.NotEmpty(t, condition.Message)

			provider.err = nil
			provider.image = &object.VirtualMachine{Common: object.Common{InventoryPath: "/DC0/vm/new-template"}}
			_, err = reconciler.Reconcile(context.Background(), nodeClass)
			require.NoError(t, err)
			require.Equal(t, provider.image.InventoryPath, nodeClass.Status.ImageTemplateID)
			require.True(t, nodeClass.StatusConditions().Get(v1alpha1.ConditionTypeImageReady).IsTrue())
		})
	}
}
