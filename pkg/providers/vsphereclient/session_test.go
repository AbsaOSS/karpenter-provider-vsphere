package vsphereclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionURL(t *testing.T) {
	u := sessionURL("vcenter.example.com:443", "user", "p@ssword")

	require.Equal(t, "https", u.Scheme)
	require.Equal(t, "vcenter.example.com:443", u.Host)
	require.Equal(t, "/sdk", u.Path)
	require.Equal(t, "user", u.User.Username())
	password, ok := u.User.Password()
	require.True(t, ok)
	require.Equal(t, "p@ssword", password)
}
