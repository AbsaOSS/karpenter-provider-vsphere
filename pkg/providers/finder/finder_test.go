package finder

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestResolveResourcePool_EmptySelector(t *testing.T) {
	provider := &Provider{}

	pool, err := provider.ResolveResourcePool(context.Background(), v1alpha1.ResPoolSelctorTerm{})

	require.Error(t, err)
	require.Nil(t, pool)
}

func TestResolveDatastore_EmptySelector(t *testing.T) {
	provider := &Provider{}

	datastore, err := provider.ResolveDatastore(context.Background(), v1alpha1.DatastoreSelectorTerm{})

	require.Error(t, err)
	require.Nil(t, datastore)
}

func TestResolveNetwork_EmptySelector(t *testing.T) {
	provider := &Provider{}

	network, err := provider.ResolveNetwork(context.Background(), v1alpha1.NetworkSelectorTerm{})

	require.Error(t, err)
	require.Nil(t, network)
}

func TestResolveImage_EmptySelector(t *testing.T) {
	provider := &Provider{}

	image, err := provider.ResolveImage(context.Background(), v1alpha1.ImageSelectorTerm{})

	require.Error(t, err)
	require.Nil(t, image)
}
