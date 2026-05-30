package userdata

import (
	"fmt"
)

type RKE2Generator struct{}

func (r *RKE2Generator) Generate(input *InitData) (*DistroConfig, error) {
	cmd := fmt.Sprintf(installRKE2Cmd, input.KubeVersion)
	return getCommon(input, cmd)
}
