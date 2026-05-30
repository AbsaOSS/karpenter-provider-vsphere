package userdata

type RKE2AirGapGenerator struct{}

func (g *RKE2AirGapGenerator) Generate(input *InitData) (*DistroConfig, error) {
	return getCommon(input, InstallRKE2AirGap)
}
