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

type ResPoolSelctorTerm struct {
	// Tags is a map of key/value tags used to select subnets
	// Specifying '*' for a value selects all values for a given tag key.
	// +kubebuilder:validation:XValidation:message="empty tag keys or values aren't supported",rule="self.all(k, k != '' && self[k] != '')"
	// +kubebuilder:validation:MaxProperties:=1
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Name is optional ResourcePoolName
	// +optional
	Name string `json:"name,omitempty"`
}

type DatastoreSelectorTerm struct {
	// Tags is a map of key/value tags used to select subnets
	// Specifying '*' for a value selects all values for a given tag key.
	// +kubebuilder:validation:XValidation:message="empty tag keys or values aren't supported",rule="self.all(k, k != '' && self[k] != '')"
	// +kubebuilder:validation:MaxProperties:=1
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Name is optional DatastoreName
	// +optional
	Name string `json:"name,omitempty"`
}

type NetworkSelectorTerm struct {
	// Tags is a map of key/value tags used to select subnets
	// Specifying '*' for a value selects all values for a given tag key.
	// +kubebuilder:validation:XValidation:message="empty tag keys or values aren't supported",rule="self.all(k, k != '' && self[k] != '')"
	// +kubebuilder:validation:MaxProperties:=1
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Name is optional NetworkName
	// +optional
	Name string `json:"name,omitempty"`
}
type DCSelectorTerm struct {
	// Tags is a map of key/value tags used to select subnets
	// Specifying '*' for a value selects all values for a given tag key.
	// +kubebuilder:validation:XValidation:message="empty tag keys or values aren't supported",rule="self.all(k, k != '' && self[k] != '')"
	// +kubebuilder:validation:MaxProperties:=1
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Name is optional DatacenterName
	// +optional
	Name string `json:"id,omitempty"`
}
type ImageSelectorTerm struct {
	// Tags is a map of key/value tags used to select subnets
	// Specifying '*' for a value selects all values for a given tag key.
	// +kubebuilder:validation:XValidation:message="empty tag keys or values aren't supported",rule="self.all(k, k != '' && self[k] != '')"
	// +kubebuilder:validation:MaxProperties:=1
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
	// Name is optional ImagePattern
	// +optional
	Pattern string `json:"pattern,omitempty"`
}

type VsphereNodeClassSpec struct {
	PoolSelector      ResPoolSelctorTerm    `json:"computeSelector,omitempty"`
	NetworkSelector   NetworkSelectorTerm   `json:"networkSelector,omitempty"`
	DatastoreSelector DatastoreSelectorTerm `json:"datastoreSelector,omitempty"`
	Datacenter        DCSelectorTerm        `json:"dcSelector,omitempty"`
	ImageSelector     ImageSelectorTerm     `json:"imageSelector,omitempty"`
	DiskSize          int64                 `json:"diskSize,omitempty"`
	InstanceTypes     []InstanceType        `json:"instanceTypes,omitempty"`
	UserData          UserData              `json:"userData,omitempty"`
	Tags              map[string]string     `json:"tags,omitempty"`
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
	OS      string `json:"os,omitempty"`
	Arch    string `json:"arch,omitempty"`
	Zone    string `json:"zone,omitempty"`
	Region  string `json:"region,omitempty"`
}

func (nc *VsphereNodeClass) Hash() string {
	return fmt.Sprint(lo.Must(hashstructure.Hash(nc.Spec, hashstructure.FormatV2, &hashstructure.HashOptions{
		SlicesAsSets:    true,
		IgnoreZeroValue: true,
		ZeroNil:         true,
	})))
}
