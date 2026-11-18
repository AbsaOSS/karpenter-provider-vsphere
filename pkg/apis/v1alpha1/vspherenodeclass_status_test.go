package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetImageTemplateID(t *testing.T) {
	nodeClass := &VsphereNodeClass{}
	nodeClass.Status.ImageTemplateID = "/DC0/vm/template"
	path := nodeClass.GetImageTemplateID()
	require.Equal(t, nodeClass.Status.ImageTemplateID, path)
}
