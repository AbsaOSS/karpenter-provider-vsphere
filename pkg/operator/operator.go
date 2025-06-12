package operator

import (
	"context"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	x "net/url"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kubernetesversion"

	"github.com/patrickmn/go-cache"
	"github.com/samber/lo"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/operator"
)

func init() {
	karpv1.RestrictedLabelDomains = karpv1.RestrictedLabelDomains.Insert(apis.Group)
}

type Operator struct {
	*operator.Operator
	InClusterKubernetesInterface kubernetes.Interface
	KubernetesVersionProvider    kubernetesversion.KubernetesVersionProvider
	InstanceProvider             instance.Provider
	FinderProvider               *finder.Provider
}

func NewOperator(ctx context.Context, operator *operator.Operator) (context.Context, *Operator) {
	vsphereClient, restClient, err := GetVsphereClient(ctx)
	lo.Must0(err, "creating vsphere client")
	tagClient := tags.NewManager(restClient)

	//inClusterConfig := lo.Must(rest.InClusterConfig())
	// for testing purposes load local kubeconfig if available
	inClusterConfig := config.GetConfigOrDie()
	inClusterClient := kubernetes.NewForConfigOrDie(inClusterConfig)

	kubernetesVersionProvider := kubernetesversion.NewKubernetesVersionProvider(
		inClusterClient,
		cache.New(15*time.Minute, 1*time.Minute),
	)

	folder := options.FromContext(ctx).VsphereFolder

	finderProvider := finder.NewDefaultProvider(tagClient, vsphereClient, folder)
	instanceProvider := instance.NewDefaultProvider(
		inClusterClient,
		finderProvider,
		options.FromContext(ctx).ClusterName,
	)
	return ctx, &Operator{
		Operator:                     operator,
		KubernetesVersionProvider:    kubernetesVersionProvider,
		InClusterKubernetesInterface: inClusterClient,
		InstanceProvider:             instanceProvider,
		FinderProvider:               finderProvider,
	}
}

func GetVsphereClient(ctx context.Context) (*vim25.Client, *rest.Client, error) {
	url := &x.URL{
		Scheme: "https",
		Host:   options.FromContext(ctx).VsphereEndpoint,
		Path:   "/sdk",
	}
	soapClient := soap.NewClient(url, options.FromContext(ctx).VsphereInsecure)
	url.User = x.UserPassword(options.FromContext(ctx).VsphereUsername, options.FromContext(ctx).VspherePassword)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create vsphere client")
	}
	vimClient.UserAgent = "karpenter-vsphere"

	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}
	restClient := rest.NewClient(c.Client)
	if err := c.Login(ctx, url.User); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create client: failed to login")
	}
	if err := restClient.Login(ctx, url.User); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create client: failed to login to rest client")
	}

	return c.Client, restClient, nil
}
