package e2e

import (
	"context"
	"testing"

	"github.com/absaoss/karpenter-provider-vsphere/internal/test/vcsim"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/govmomi/session"
)

// newTestServer starts an in-process vcsim vCenter (the default VPX model:
// one datacenter, one cluster with 3 hosts, a resource pool, a network, a
// datastore and a handful of demo VMs) with TLS enabled, since NewSession
// always dials "https://". The server and its backing model are torn down
// automatically at the end of the test.
func newTestServer(t *testing.T) *vcsim.Simulator {
	t.Helper()

	simulator, err := vcsim.New()
	require.NoError(t, err)
	t.Cleanup(simulator.Destroy)
	return simulator
}

func TestNewSession(t *testing.T) {
	server := newTestServer(t)
	ctx := context.Background()

	username := server.Username()
	password := server.Password()

	sess, err := vsphereclient.NewSession(ctx, server.ServerURL().Host, username, password, true)
	require.NoError(t, err)
	require.NotNil(t, sess)

	assert.NotNil(t, sess.Vim, "expected an authenticated SOAP/vim25 client")
	assert.NotNil(t, sess.Rest, "expected an authenticated REST client")
	assert.NotNil(t, sess.Tags, "expected a tags manager wired to the REST client")
	assert.Equal(t, "karpenter-vsphere", sess.Vim.UserAgent)

	// Both halves of the session should be immediately usable without any
	// extra login step.
	userSession, err := session.NewManager(sess.Vim).UserSession(ctx)
	require.NoError(t, err)
	require.NotNil(t, userSession, "expected an authenticated SOAP session")
	assert.Equal(t, username, userSession.UserName)

	restSession, err := sess.Rest.Session(ctx)
	require.NoError(t, err)
	require.NotNil(t, restSession, "expected an authenticated REST session")
}

func TestNewSession_ConnectionError(t *testing.T) {
	ctx := context.Background()

	// Nothing listens on this address, so the dial should fail fast and
	// NewSession should surface that as a wrapped error rather than
	// hanging or panicking.
	_, err := vsphereclient.NewSession(ctx, "127.0.0.1:1", "user", "pass", true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create vsphere client")
}

func TestSession_EnsureValid(t *testing.T) {
	server := newTestServer(t)
	ctx := context.Background()

	username := server.Username()
	password := server.Password()

	sess, err := vsphereclient.NewSession(ctx, server.ServerURL().Host, username, password, true)
	require.NoError(t, err)

	t.Run("no-op when the session is already valid", func(t *testing.T) {
		require.NoError(t, sess.EnsureValid(ctx))
	})

	t.Run("re-authenticates after the underlying sessions are invalidated", func(t *testing.T) {
		// Simulate vCenter tearing the session down from underneath us
		// (idle timeout, vpxd restart, HA failover, an admin-forced
		// logout, etc). EnsureValid is what every provider method calls
		// before doing real work, so it must detect this and recover.
		require.NoError(t, session.NewManager(sess.Vim).Logout(ctx))
		require.NoError(t, sess.Rest.Logout(ctx))

		require.NoError(t, sess.EnsureValid(ctx))

		userSession, err := session.NewManager(sess.Vim).UserSession(ctx)
		require.NoError(t, err)
		assert.NotNil(t, userSession, "expected EnsureValid to have re-authenticated the SOAP session")

		restSession, err := sess.Rest.Session(ctx)
		require.NoError(t, err)
		assert.NotNil(t, restSession, "expected EnsureValid to have re-authenticated the REST session")
	})
}
