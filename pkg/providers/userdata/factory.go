package userdata

import (
	"fmt"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type InitData struct {
	Token              string
	APIEndpoint        string
	KubeVersion        string
	Taints             []corev1.Taint
	NodeName           string
	AdditionalUserData string
}

type InitType struct {
	Format v1alpha1.UserDataType
	Distro v1alpha1.Distro
}

func NewInitData(taints []corev1.Taint, nodeName, endpoint, token, kubeversion string, userdata string) *InitData {
	return &InitData{
		Taints:             taints,
		NodeName:           nodeName,
		APIEndpoint:        endpoint,
		KubeVersion:        kubeversion,
		Token:              token,
		AdditionalUserData: userdata,
	}
}

type Generator interface {
	Generate(input *InitData) (*DistroConfig, error)
}

type Renderer interface {
	Render(data *DistroConfig, additional string) ([]byte, error)
}

type DistroConfig struct {
	Files    []File   `yaml:"write_files"`
	Commands []string `yaml:"runcmd"`
	NodeName string
}
type File struct {
	Path        string `yaml:"path" json:"path"`
	Permissions string `yaml:"permissions" json:"permissions"`
	Owner       string `yaml:"owner" json:"owner"`
	Content     string `yaml:"content" json:"content"`
}

type Factory struct{}

func (f *Factory) Build(initType *InitType) (Generator, Renderer, error) {
	var generator Generator
	var renderer Renderer

	switch initType.Distro {
	case v1alpha1.RKE2:
		generator = &RKE2Generator{}

	case v1alpha1.RKE2AirGapped:
		generator = &RKE2AirGapGenerator{}

	case v1alpha1.KUBEADM:
		generator = &KubeadmGenerator{}

	default:
		return nil, nil, fmt.Errorf("unsupported distro")
	}

	switch initType.Format {
	case v1alpha1.UserDataTypeCloudConfig:
		renderer = &CloudConfigRenderer{}

	case v1alpha1.UserDataTypeIgnition:
		renderer = &ButaneRenderer{}

	default:
		return nil, nil, fmt.Errorf("unsupported format")
	}

	return generator, renderer, nil
}
