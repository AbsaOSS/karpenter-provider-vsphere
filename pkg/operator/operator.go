package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	x "net/url"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kubernetesversion"

	"github.com/patrickmn/go-cache"
	"github.com/samber/lo"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
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
}

func NewOperator(ctx context.Context, operator *operator.Operator) (context.Context, *Operator) {
	vsphereClient, err := GetVsphereClient(ctx)
	lo.Must0(err, "creating vsphere client")

	//inClusterConfig := lo.Must(rest.InClusterConfig())
	// for testing purposes load local kubeconfig if available
	inClusterConfig := config.GetConfigOrDie()
	inClusterClient := kubernetes.NewForConfigOrDie(inClusterConfig)

	kubernetesVersionProvider := kubernetesversion.NewKubernetesVersionProvider(
		inClusterClient,
		cache.New(15*time.Minute, 1*time.Minute),
	)
	return ctx, &Operator{
		Operator:                     operator,
		KubernetesVersionProvider:    kubernetesVersionProvider,
		InClusterKubernetesInterface: inClusterClient,
		InstanceProvider:             instance.NewDefaultProvider(vsphereClient, inClusterClient, options.FromContext(ctx).ClusterName, options.FromContext(ctx).VsphereZone),
	}
}

func GetVsphereClient(ctx context.Context) (*instance.VsphereInfo, error) {
	url := &x.URL{
		Scheme: "https",
		Host:   options.FromContext(ctx).VsphereEndpoint,
		Path:   "/sdk",
	}
	soapClient := soap.NewClient(url, options.FromContext(ctx).VsphereInsecure)
	url.User = x.UserPassword(options.FromContext(ctx).VsphereUsername, options.FromContext(ctx).VspherePassword)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create vsphere client")
	}
	vimClient.UserAgent = "karpenter-vsphere"

	c := govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}
	restClient := rest.NewClient(c.Client)
	if err := c.Login(ctx, url.User); err != nil {
		return nil, errors.Wrapf(err, "failed to create client: failed to login")
	}
	if err := restClient.Login(ctx, url.User); err != nil {
		return nil, errors.Wrapf(err, "failed to create client: failed to login to rest client")
	}

	finder := find.NewFinder(c.Client, true)
	dc, err := finder.Datacenter(ctx, options.FromContext(ctx).VsphereDatacenter)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find datacenter %s", options.FromContext(ctx).VsphereDatacenter)
	}
	finder.SetDatacenter(dc)
	poolPath := genereateResourcePoolPath(options.FromContext(ctx).VsphereDatacenter, options.FromContext(ctx).VsphereComputeCluster)
	pool, err := finder.ResourcePool(ctx, poolPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find resource pool %s", poolPath)
	}
	folderPath := fmt.Sprintf("/%s/vm/%s", options.FromContext(ctx).VsphereDatacenter, options.FromContext(ctx).VspherePath)
	path, err := finder.Folder(ctx, folderPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find folder %s", options.FromContext(ctx).VspherePath)
	}
	ds, err := finder.DatastoreOrDefault(ctx, options.FromContext(ctx).VsphereDatastore)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find datastore %s", options.FromContext(ctx).VsphereDatastore)
	}
	return &instance.VsphereInfo{
		PoolRef:      pool.Reference(),
		DatastoreRef: ds.Reference(),
		Datacenter:   dc,
		Folder:       path,
		TagManager:   tags.NewManager(restClient),
		Client:       &c,
		Finder:       finder,
	}, nil
}

func genereateResourcePoolPath(dc, cluster string) string {
	return fmt.Sprintf("/%s/host/%s/Resources", dc, cluster)
}
