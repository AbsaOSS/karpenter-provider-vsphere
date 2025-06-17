package v1alpha1

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	coreapis "sigs.k8s.io/karpenter/pkg/apis"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

func init() {
	karpv1.RestrictedLabelDomains = karpv1.RestrictedLabelDomains.Insert(RestrictedLabelDomains...)
	karpv1.WellKnownLabels = karpv1.WellKnownLabels.Insert(
		LabelInstanceSize,
	)
}

var (
	RestrictedLabelDomains = []string{
		apis.Group,
	}
	TerminationFinalizer                  = apis.Group + "/termination"
	AnnotationVsphereNodeClassHash        = apis.Group + "/vspherenodeclass-hash"
	LabelNodeClass                        = apis.Group + "/vspherenodeclass"
	LabelInstanceCPU                      = apis.Group + "/instance-cpu"
	LabelInstanceMemory                   = apis.Group + "/instance-memory"
	LabelInstanceSize                     = apis.Group + "/instance-size"
	LabelInstanceType                     = corev1.LabelInstanceTypeStable
	AnnotationVsphereNodeClassHashVersion = apis.Group + "/vspherenodeclass-hash-version"
	NodeClaimTagKey                       = coreapis.Group + "/nodeclaim"
	NodePoolTagKey                        = karpv1.NodePoolLabelKey
	ClusterNameTagKey                     = "karpenter.sh/clustername"
)
