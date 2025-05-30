// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:defaulter-gen=TypeMeta
// +groupName=karpenter.vsphere.com
package v1alpha1

import (
	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	gv := schema.GroupVersion{Group: apis.Group, Version: "v1alpha1"}
	corev1.AddToGroupVersion(scheme.Scheme, gv)
	scheme.Scheme.AddKnownTypes(gv,
		&VsphereNodeClass{},
		&VsphereNodeClassList{},
	)
}
