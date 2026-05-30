package instance

import (
	"encoding/base64"
	"fmt"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/userdata"
	"github.com/vmware/govmomi/vim25/types"
)

// Config is data used with a VM's guestInfo RPC interface.
type Config []types.BaseOptionValue

const (
	guestInfoIgnitionData     = "guestinfo.ignition.config.data"
	guestInfoIgnitionEncoding = "guestinfo.ignition.config.data.encoding"
	guestInfoUserData         = "guestinfo.userdata"
	guestInfoUserDataEncoding = "guestinfo.userdata.encoding"
	guestInfoMetadata         = "guestinfo.metadata"
	guestInfoMetadataEncoding = "guestinfo.metadata.encoding"
)

func (e *Config) Extract() []types.BaseOptionValue {
	if e == nil {
		return nil
	}
	return *e
}

func CloudConfigHeader(b []byte) []byte {
	return append([]byte("#cloud-config\n"), b...)
}

func (p *DefaultProvider) GetInitData(initData *userdata.InitData, initType *userdata.InitType) ([]types.BaseOptionValue, error) {
	metaDataRaw := []byte(fmt.Sprintf("local-hostname: \"%s\"", initData.NodeName))

	configData := &Config{}
	uData := &userdata.Factory{}
	gen, renderer, err := uData.Build(initType)
	if err != nil {
		return nil, err
	}
	joinData, err := gen.Generate(initData)
	if err != nil {
		panic(err)
	}

	result, err := renderer.Render(
		joinData,
		initData.AdditionalUserData,
	)

	switch initType.Format {
	case v1alpha1.UserDataTypeIgnition:
		configData.SetIgnitionUserData(result)
	case v1alpha1.UserDataTypeCloudConfig:
		configData.SetCloudConfigUserData(result)
	default:
		return nil, fmt.Errorf("unsupported user data format: %s", initType.Format)
	}
	// Set the metadata with the local hostname
	configData.SetMetadata(metaDataRaw)

	return *configData, nil
}

// setUserData sets the user data at the provided key
// as a base64-encoded string.
func (e *Config) setData(userdataKey, encodingKey string, data []byte) {
	*e = append(*e,
		&types.OptionValue{
			Key:   userdataKey,
			Value: e.encode(data),
		},
		&types.OptionValue{
			Key:   encodingKey,
			Value: "base64",
		},
	)
}

func (e *Config) SetMetadata(data []byte) {
	// Set the userdata at the key "guestinfo.userdata.metadata" as a base64-encoded string.
	e.setData(guestInfoMetadata, guestInfoMetadataEncoding, data)
}

// SetIgnitionUserData sets the ignition user data at the key
// "guestinfo.ignition.config.data" as a base64-encoded string.
func (e *Config) SetIgnitionUserData(data []byte) {
	e.setData(guestInfoIgnitionData, guestInfoIgnitionEncoding, data)
}

func (e *Config) SetCloudConfigUserData(data []byte) {
	e.setData(guestInfoUserData, guestInfoUserDataEncoding, data)
}

// encode first attempts to decode the data as many times as necessary
// to ensure it is plain-text before returning the result as a base64
// encoded string.
func (e *Config) encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	for {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			break
		}
		data = decoded
	}
	return base64.StdEncoding.EncodeToString(data)
}
