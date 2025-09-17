package options

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	coreoptions "sigs.k8s.io/karpenter/pkg/operator/options"
	"sigs.k8s.io/karpenter/pkg/utils/env"
)

func init() {
	coreoptions.Injectables = append(coreoptions.Injectables, &Options{})
}

type Options struct {
	ClusterName     string
	ClusterEndpoint string
	JoinToken       string
	VsphereEndpoint string
	VsphereUsername string
	VspherePassword string
	VsphereFolder   string
	VsphereInsecure bool
	KubeDistro      string
	KubeVersion     string
}

type optionsKey struct{}

func (o *Options) AddFlags(fs *coreoptions.FlagSet) {
	fs.StringVar(&o.ClusterName, "cluster-name", env.WithDefaultString("CLUSTER_NAME", ""), "[REQUIRED] The name of the kubernetes cluster to use")
	fs.StringVar(&o.ClusterEndpoint, "cluster-endpoint", env.WithDefaultString("CLUSTER_ENDPOINT", ""), "[REQUIRED] Kubernetes API endpoint to use for nodes to join")
	fs.StringVar(&o.JoinToken, "join-token", env.WithDefaultString("JOIN_TOKEN", ""), "[REQUIRED] kubernetes join token")
	fs.StringVar(&o.KubeDistro, "kube-distro", env.WithDefaultString("KUBE_DISTRO", ""), "[REQUIRED] The name of the kubernetes distribution to use")
	fs.StringVar(&o.VsphereEndpoint, "vsphere-endpoint", env.WithDefaultString("GOVC_URL", ""), "[REQUIRED] The vSphere endpoint to use for the vSphere provider")
	fs.StringVar(&o.VsphereUsername, "vsphere-username", env.WithDefaultString("GOVC_USERNAME", ""), "[REQUIRED] The vSphere username to use for the vSphere provider")
	fs.StringVar(&o.VspherePassword, "vsphere-password", env.WithDefaultString("GOVC_PASSWORD", ""), "[REQUIRED] The vSphere password to use for the vSphere provider")
	fs.StringVar(&o.VsphereFolder, "vsphere-path", env.WithDefaultString("VSPHERE_FOLDER", ""), "[REQUIRED] The vSphere path to use for the vSphere provider")
	fs.BoolVar(&o.VsphereInsecure, "vsphere-insecure", env.WithDefaultBool("GOVC_INSECURE", false), "[REQUIRED] The vSphere insecure flag to use for the vSphere provider")
}

func (o *Options) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, o)
}

func ToContext(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

func FromContext(ctx context.Context) *Options {
	retval := ctx.Value(optionsKey{})
	if retval == nil {
		return nil
	}
	return retval.(*Options)
}

func (o *Options) Parse(fs *coreoptions.FlagSet, args ...string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		return fmt.Errorf("parsing flags, %w", err)
	}

	if err := o.Validate(); err != nil {
		return fmt.Errorf("validating options, %w", err)
	}

	return nil
}

func (o *Options) String() string {
	opts, err := json.Marshal(o)
	if err != nil {
		return "couldn't marshal options JSON"
	}

	return string(opts)
}

func (o *Options) Validate() error {
	if o.ClusterEndpoint == "" {
		return fmt.Errorf("--cluster-endpoint is required")
	}
	if o.KubeDistro == "rke2" && o.KubeVersion == "" {
		return errors.New("--kube-distro option requires --kube-version")
	}
	return nil
}
