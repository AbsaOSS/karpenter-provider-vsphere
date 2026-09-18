package options

import (
	"testing"

	"flag"

	"github.com/stretchr/testify/require"
	coreoptions "sigs.k8s.io/karpenter/pkg/operator/options"
)

func TestOptions_Parse(t *testing.T) {
	fs := &coreoptions.FlagSet{
		FlagSet: flag.NewFlagSet("test", flag.ContinueOnError),
	}
	opts := &Options{}
	opts.AddFlags(fs)

	err := opts.Parse(
		fs,
		"--cluster-name=test",
		"--vsphere-endpoint=vcenter",
		"--cluster-endpoint=https://example.com",
		"--vsphere-username=username",
		"--vsphere-password=password",
		"--vsphere-dc=DC0",
		"--zone=zone1",
		"--region=region1",
	)

	require.NoError(t, err)

	require.Equal(t, "test", opts.ClusterName)
	require.Equal(t, "vcenter", opts.VsphereEndpoint)
}

func TestOptionsValidation_RequiredArguments(t *testing.T) {
	err := (&Options{}).Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "cluster-endpoint is required")
	require.Contains(t, err.Error(), "vsphere-endpoint is required")
	require.Contains(t, err.Error(), "vsphere-username is required")
	require.Contains(t, err.Error(), "vsphere-password is required")
	require.Contains(t, err.Error(), "vsphere-dc is required")
	require.Contains(t, err.Error(), "cluster-name is required")
	require.Contains(t, err.Error(), "zone is required")
	require.Contains(t, err.Error(), "region is required")
}
