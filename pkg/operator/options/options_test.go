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

func TestOptions_VMMemoryOverheadPercentDefault(t *testing.T) {
	fs := &coreoptions.FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	opts := &Options{}
	opts.AddFlags(fs)

	require.NoError(t, fs.Parse(nil))
	require.Equal(t, 0.075, opts.VMMemoryOverheadPercent)
}

func TestOptions_VMMemoryOverheadPercentFromEnvAndFlag(t *testing.T) {
	t.Setenv("VM_MEMORY_OVERHEAD_PERCENT", "0.05")

	fs := &coreoptions.FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	opts := &Options{}
	opts.AddFlags(fs)
	require.NoError(t, fs.Parse(nil))
	require.Equal(t, 0.05, opts.VMMemoryOverheadPercent, "env var should set the default")

	fs = &coreoptions.FlagSet{FlagSet: flag.NewFlagSet("test", flag.ContinueOnError)}
	opts = &Options{}
	opts.AddFlags(fs)
	require.NoError(t, fs.Parse([]string{"--vm-memory-overhead-percent=0.1"}))
	require.Equal(t, 0.1, opts.VMMemoryOverheadPercent, "flag should win over env var")
}

func TestOptionsValidation_VMMemoryOverheadPercent(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		wantErr bool
	}{
		{name: "zero is allowed", value: 0},
		{name: "default", value: 0.075},
		{name: "just under one", value: 0.99},
		{name: "negative", value: -0.1, wantErr: true},
		{name: "one", value: 1, wantErr: true},
		{name: "above one", value: 1.5, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&Options{VMMemoryOverheadPercent: tt.value}).validateVMMemoryOverheadPercent()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "vm-memory-overhead-percent")
				return
			}
			require.NoError(t, err)
		})
	}
}
