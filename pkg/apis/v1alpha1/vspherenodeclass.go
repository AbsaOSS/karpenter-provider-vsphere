package v1alpha1

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionTypeKubernetesVersionReady = "KubernetesVersionReady"
	VsphereNodeClassHashVersion         = "v1"
)

// VsphereNodeClass is the Schema for the VsphereNodeClass API
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=vspherenodeclasses,scope=Cluster,categories=karpenter,shortName={vspherenc,vspherencs}
// +kubebuilder:subresource:status
type VsphereNodeClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VsphereNodeClassSpec   `json:"spec,omitempty"`
	Status            VsphereNodeClassStatus `json:"status,omitempty"`
}

// VsphereNodeClassList contains a list of VsphereNodeClasses
// +kubebuilder:object:root=true
type VsphereNodeClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VsphereNodeClass `json:"items"`
}

type VsphereNodeClassSpec struct {
	Image         string                  `json:"image,omitempty"`
	Network       string                  `json:"network,omitempty"`
	InstanceTypes map[string]InstanceType `json:"instanceTypes,omitempty"`
	UserData      UserData                `json:"userData,omitempty"`
}

type UserDataType string

const (
	UserDataTypeCloudInit UserDataType = "cloud-init"
	UserDataTypeIgnition  UserDataType = "ignition"
)

type UserData struct {
	Type UserDataType `json:"type,omitempty"`
	// +optional
	TemplateBase64 string `json:"templateBase64,omitempty"`
	// +optional
	Values corev1.SecretReference `json:"values,omitempty"`
}

type InstanceType struct {
	CPU     string `json:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty"`
	MaxPods string `json:"maxPods,omitempty"`
	Storage string `json:"storage,omitempty"`
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

func (nc *VsphereNodeClass) Hash() string {
	return fmt.Sprint(lo.Must(hashstructure.Hash(nc.Spec, hashstructure.FormatV2, &hashstructure.HashOptions{
		SlicesAsSets:    true,
		IgnoreZeroValue: true,
		ZeroNil:         true,
	})))
}
