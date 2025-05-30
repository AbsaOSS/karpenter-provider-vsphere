package v1alpha1

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
)

var (
	TerminationFinalizer                  = apis.Group + "/termination"
	AnnotationVsphereNodeClassHash        = apis.Group + "/vspherenodeclass-hash"
	AnnotationVsphereNodeClassHashVersion = apis.Group + "/vspherenodeclass-hash-version"
)
