package controllers

import (
	"context"

	"github.com/awslabs/operatorpkg/controller"
	"github.com/awslabs/operatorpkg/status"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/events"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	nodeclaimgarbagecollection "github.com/absaoss/karpenter-provider-vsphere/pkg/controllers/nodeclaim/garbagecollection"
	nodeclasshash "github.com/absaoss/karpenter-provider-vsphere/pkg/controllers/nodeclass/hash"
	nodeclassstatus "github.com/absaoss/karpenter-provider-vsphere/pkg/controllers/nodeclass/status"
	nodeclasstermination "github.com/absaoss/karpenter-provider-vsphere/pkg/controllers/nodeclass/termination"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/instance"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kubernetesversion"
)

func NewControllers(
	ctx context.Context,
	mgr manager.Manager,
	kubeClient client.Client,
	recorder events.Recorder,
	cloudProvider cloudprovider.CloudProvider,
	instanceProvider instance.Provider,
	kubernetesVersionProvider kubernetesversion.KubernetesVersionProvider,
	inClusterKubernetesInterface kubernetes.Interface,
) []controller.Controller {
	controllers := []controller.Controller{
		nodeclasshash.NewController(kubeClient),
		nodeclassstatus.NewController(kubeClient, kubernetesVersionProvider, inClusterKubernetesInterface),
		nodeclasstermination.NewController(kubeClient, recorder),

		nodeclaimgarbagecollection.NewVirtualMachine(kubeClient, cloudProvider),

		status.NewController[*v1alpha1.VsphereNodeClass](kubeClient, mgr.GetEventRecorderFor("karpenter")),
	}
	return controllers
}
