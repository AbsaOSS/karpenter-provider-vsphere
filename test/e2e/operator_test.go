package e2e

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/test/vcsim"
	vsoperator "github.com/absaoss/karpenter-provider-vsphere/pkg/operator"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kwok"
	"github.com/stretchr/testify/require"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	karpoperator "sigs.k8s.io/karpenter/pkg/operator"
)

// Operator test constants
const (
	OperatorClusterName     = "test-cluster"
	OperatorClusterEndpoint = "https://127.0.0.1:6443"
	OperatorJoinToken       = "token"
	OperatorKubeDistro      = "rke2"
	OperatorZone            = "test-zone"
	OperatorRegion          = "test-region"
	OperatorVsphereFolder   = "vm"
	OperatorVsphereDC       = "DC0"
)

func setupContext(
	ctx context.Context,
	server *vcsim.Simulator,
) context.Context {
	return options.ToContext(ctx, &options.Options{
		ClusterName:     OperatorClusterName,
		ClusterEndpoint: OperatorClusterEndpoint,
		JoinToken:       OperatorJoinToken,
		KubeDistro:      OperatorKubeDistro,

		VsphereEndpoint: server.ServerURL().Host,
		VsphereUsername: server.Username(),
		VspherePassword: server.Password(),
		VsphereInsecure: true,

		VsphereFolder: OperatorVsphereFolder,
		VsphereDC:     OperatorVsphereDC,

		Zone:   OperatorZone,
		Region: OperatorRegion,
	})
}

// TestNewOperator_InitializesAllProviders tests the operator initialization using
// vSphere simulator (vcsim). This test ensures that InstanceProfilesProvider is properly
// initialized and not nil. The test would fail with a nil pointer dereference if
// InstanceProfilesProvider initialization were missing, since NewOperator would
// leave the field uninitialized and return a struct with a nil InstanceProfilesProvider.
func TestNewOperator_InitializesAllProviders(t *testing.T) {

	// Create a fake clientset for testing
	fakeClientset := k8sfake.NewSimpleClientset()

	// Pass the fake clientset to the operator
	server := newTestServer(t)

	ctx := setupContext(
		context.Background(),
		server,
	)

	_, op := vsoperator.NewOperatorInt(
		ctx,
		&karpoperator.Operator{},
		fakeClientset,
	)

	require.NotNil(t, op)

	// Embedded Karpenter operator
	require.NotNil(t, op.Operator)

	// Kubernetes-related dependencies
	require.NotNil(t, op.InClusterKubernetesInterface)
	require.NotNil(t, op.KubernetesVersionProvider)

	// vSphere-related dependencies
	require.NotNil(t, op.FinderProvider)
	require.NotNil(t, op.InstanceProvider)

	// Regression check for newly added provider
	require.NotNil(t, op.InstanceProfilesProvider)

	// Ensure the expected implementation was wired in
	_, ok := op.InstanceProfilesProvider.(*kwok.KwokInstanceTypesStaticProvider)
	require.True(t, ok, "expected KwokInstanceTypesStaticProvider")
}
