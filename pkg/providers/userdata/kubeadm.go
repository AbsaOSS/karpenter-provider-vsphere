package userdata

import "fmt"

type KubeadmGenerator struct{}

func (k *KubeadmGenerator) Generate(_ *InitData) (*DistroConfig, error) {
	return &DistroConfig{}, fmt.Errorf("kubeadm not implemented yet")
}
