package main

import (
	"context"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/cloudprovider"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/controllers"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/operator/options"
	"github.com/go-logr/zapr"
	"sigs.k8s.io/karpenter/pkg/cloudprovider/metrics"
	corecontrollers "sigs.k8s.io/karpenter/pkg/controllers"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	coreoperator "sigs.k8s.io/karpenter/pkg/operator"
	"sigs.k8s.io/karpenter/pkg/operator/injection"
	"sigs.k8s.io/karpenter/pkg/operator/logging"
	coreoptions "sigs.k8s.io/karpenter/pkg/operator/options"
)

func main() {
	ctx := injection.WithOptionsOrDie(context.Background(), coreoptions.Injectables...)
	logger := zapr.NewLogger(logging.NewLogger(ctx, "controller"))

	ctx, op := operator.NewOperator(coreoperator.NewOperator())
	logger.V(0).Info("Initial options", "options", options.FromContext(ctx).String())

	vsphereCloudProvider := cloudprovider.New(
		op.InstanceProvider,
		op.GetClient(),
	)

	cloudProvider := metrics.Decorate(vsphereCloudProvider)
	clusterState := state.NewCluster(op.Clock, op.GetClient(), cloudProvider)
	op.WithControllers(ctx, corecontrollers.NewControllers(
		ctx,
		op.Manager,
		op.Clock,
		op.GetClient(),
		op.EventRecorder,
		cloudProvider,
		clusterState,
	)...).
		WithControllers(ctx, controllers.NewControllers(
			ctx,
			op.Manager,
			op.GetClient(),
			op.EventRecorder,
			vsphereCloudProvider,
			op.InstanceProvider,
			op.KubernetesVersionProvider,
			op.InClusterKubernetesInterface,
		)...).
		Start(ctx)
}
