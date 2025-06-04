package v1alpha1

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
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
	LabelInstanceSize                     = apis.Group + "/instance-size"
	AnnotationVsphereNodeClassHashVersion = apis.Group + "/vspherenodeclass-hash-version"
)
