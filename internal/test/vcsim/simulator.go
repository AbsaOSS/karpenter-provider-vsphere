// Package vcsim contains tools for running a vCenter simulator.
package vcsim

import (
	"crypto/tls"
	"net/url"

	"github.com/vmware/govmomi/simulator"

	// Register the tagging API endpoints used by the REST client.
	_ "github.com/vmware/govmomi/vapi/simulator"
)

// Simulator binds together a vcsim model and its server.
type Simulator struct {
	model  *simulator.Model
	server *simulator.Server
}

// New creates an in-process VPX vCenter simulator.
func New() (*Simulator, error) {
	model := simulator.VPX()
	if err := model.Create(); err != nil {
		return nil, err
	}
	model.Service.RegisterEndpoints = true
	model.Service.TLS = new(tls.Config)
	server := model.Service.NewServer()
	return &Simulator{model: model, server: server}, nil
}

// Destroy closes the server and removes the simulator model.
func (s *Simulator) Destroy() {
	if s == nil {
		return
	}
	s.server.Close()
	s.model.Remove()
}

// ServerURL returns the simulator server URL.
func (s *Simulator) ServerURL() *url.URL {
	return s.server.URL
}

// Username returns the simulator username.
func (s *Simulator) Username() string {
	return s.server.URL.User.Username()
}

// Password returns the simulator password.
func (s *Simulator) Password() string {
	password, _ := s.server.URL.User.Password()
	return password
}
