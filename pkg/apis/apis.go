package apis

import (
	_ "embed"

	"github.com/awslabs/operatorpkg/object"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/karpenter/pkg/apis"
)

//go:generate controller-gen crd object:headerFile="../../hack/boilerplate.go.txt" paths="./..." output:crd:artifacts:config=crds
var (
	Group               = "karpenter.vsphere.com"
	VsphereNodeClassCRD []byte
	CRDs                = append(apis.CRDs,
		object.Unmarshal[apiextensionsv1.CustomResourceDefinition](VsphereNodeClassCRD))
)
