package vsphereclient

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
)

// Session holds long-lived vSphere SOAP and REST clients. vCenter can
// invalidate their session at any time (idle timeout, vpxd restart, HA
// failover, admin-configured absolute session caps, etc), so callers must
// call EnsureValid once at the start of each unit of work (e.g. each
// instance.Provider method) before using Vim/Rest/Tags -- mirroring how
// cluster-api-provider-vsphere's session.GetOrCreate is called once per
// controller Reconcile.
type Session struct {
	mu      sync.Mutex
	session *cache.Session

	Vim  *vim25.Client
	Rest *rest.Client
	Tags *tags.Manager
}

func NewSession(ctx context.Context, endpoint, username, password string, insecure bool) (*Session, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   endpoint,
		Path:   "/sdk",
	}
	u.User = url.UserPassword(username, password)

	cs := &cache.Session{
		URL:         u,
		Insecure:    insecure,
		Passthrough: true, // no file cache: controller runs with readOnlyRootFilesystem
	}

	s := &Session{
		session: cs,
		Vim:     &vim25.Client{},
		Rest:    &rest.Client{},
	}

	if err := cs.Login(ctx, s.Vim, nil); err != nil {
		return nil, fmt.Errorf("failed to create vsphere client: %w", err)
	}
	s.Vim.UserAgent = "karpenter-vsphere"

	if err := cs.Login(ctx, s.Rest, nil); err != nil {
		return nil, fmt.Errorf("failed to create vsphere rest client: %w", err)
	}
	s.Tags = tags.NewManager(s.Rest)

	return s, nil
}

// EnsureValid checks whether the SOAP and REST sessions are still
// authenticated against vCenter, and transparently re-authenticates
func (s *Session) EnsureValid(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	valid, err := vimSessionValid(ctx, s.Vim)
	if err != nil {
		return fmt.Errorf("checking vsphere session: %w", err)
	}
	if !valid {
		if err := s.session.Login(ctx, s.Vim, nil); err != nil {
			return fmt.Errorf("failed to re-authenticate vsphere session: %w", err)
		}
	}

	valid, err = restSessionValid(ctx, s.Rest)
	if err != nil {
		return fmt.Errorf("checking vsphere rest session: %w", err)
	}
	if !valid {
		if err := s.session.Login(ctx, s.Rest, nil); err != nil {
			return fmt.Errorf("failed to re-authenticate vsphere rest session: %w", err)
		}
	}
	return nil
}

func vimSessionValid(ctx context.Context, client *vim25.Client) (bool, error) {
	u, err := session.NewManager(client).UserSession(ctx)
	if err != nil {
		return false, err
	}
	return u != nil, nil
}

func restSessionValid(ctx context.Context, client *rest.Client) (bool, error) {
	u, err := client.Session(ctx)
	if err != nil {
		return false, err
	}
	return u != nil, nil
}
