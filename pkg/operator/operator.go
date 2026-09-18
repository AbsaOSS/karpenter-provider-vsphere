package operator

import (
	"context"
	"time"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	"github.com/vmware/govmomi/find"
	"k8s.io/client-go/kubernetes"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/finder"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kubernetesversion"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kwok"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/vsphereclient"

	"github.com/patrickmn/go-cache"
	"github.com/samber/lo"
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
	InstanceProfilesProvider     kwok.KwokInstanceTypesProvider
}

func NewInClusterClient() kubernetes.Interface {
	inClusterConfig := config.GetConfigOrDie()
	return kubernetes.NewForConfigOrDie(inClusterConfig)
}

func NewOperator(ctx context.Context, operator *operator.Operator) (context.Context, *Operator) {
	return NewOperatorInt(ctx, operator, NewInClusterClient())
}

func NewOperatorInt(ctx context.Context, operator *operator.Operator, inClusterClient kubernetes.Interface) (context.Context, *Operator) {
	optionsFromCtx := options.FromContext(ctx)
	// lo.Must0(optionsFromCtx.Validate(), "validating operator options")
	vsClient, err := vsphereclient.NewSession(
		ctx,
		optionsFromCtx.VsphereEndpoint,
		optionsFromCtx.VsphereUsername,
		optionsFromCtx.VspherePassword,
		optionsFromCtx.VsphereInsecure,
	)
	lo.Must0(err, "creating vsphere session")

	// inClusterConfig := lo.Must(rest.InClusterConfig())
	// for testing purposes load local kubeconfig if available

	kubernetesVersionProvider := kubernetesversion.NewKubernetesVersionProvider(
		inClusterClient,
		cache.New(15*time.Minute, 1*time.Minute),
	)

	folder := options.FromContext(ctx).VsphereFolder
	clusterName := options.FromContext(ctx).ClusterName
	dcName := options.FromContext(ctx).VsphereDC
	findClient := find.NewFinder(vsClient.Vim, true)
	dc, err := findClient.Datacenter(ctx, dcName)
	lo.Must0(err, "finding datacenter")

	finderProvider := finder.NewDefaultProvider(vsClient, findClient, dc, folder, clusterName)
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
		InstanceProfilesProvider:     &kwok.KwokInstanceTypesStaticProvider{},
	}
}
