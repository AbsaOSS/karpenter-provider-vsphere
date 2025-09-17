package instance

import (
	"encoding/base64"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/vim25/types"
)

// Config is data used with a VM's guestInfo RPC interface.
type Config []types.BaseOptionValue

type InitData struct {
	Token              string
	APIEndpoint        string
	KubeVersion        string
	Taints             []corev1.Taint
	NodeName           string
	Distro             string
	Format             v1alpha1.UserDataType
	AdditionalUserData string
}

const (
	guestInfoIgnitionData     = "guestinfo.ignition.config.data"
	guestInfoIgnitionEncoding = "guestinfo.ignition.config.data.encoding"
	guestInfoUserData         = "guestinfo.userdata"
	guestInfoUserDataEncoding = "guestinfo.userdata.encoding"
	guestInfoMetadata         = "guestinfo.metadata"
	guestInfoMetadataEncoding = "guestinfo.metadata.encoding"
)

func NewInitData(taints []corev1.Taint, nodeName, endpoint, token, distro, kubeversion string, format v1alpha1.UserDataType, userdata string) *InitData {
	return &InitData{
		Taints:             taints,
		NodeName:           nodeName,
		APIEndpoint:        endpoint,
		KubeVersion:        kubeversion,
		Distro:             distro,
		Token:              token,
		Format:             format,
		AdditionalUserData: userdata,
	}
}

func (e *Config) Extract() []types.BaseOptionValue {
	if e == nil {
		return nil
	}
	return *e
}

func CloudConfigHeader(b []byte) []byte {
	return append([]byte("#cloud-config\n"), b...)
}

func (p *DefaultProvider) GetInitData(initData *InitData) ([]types.BaseOptionValue, error) {
	configData := &Config{}
	// Set the metadata with the local hostname
	metaDataRaw := []byte(fmt.Sprintf("local-hostname: \"%s\"", initData.NodeName))
	selector := fmt.Sprintf("%s:%s", initData.Format, initData.Distro)
	switch t := selector; t {
	// cloud-config userdata
	case "cloud-config:rke2", "cloud-config:rke2airgapped":
		installConfig, err := GetRKE2CloudConfig(initData)
		if err != nil {
			return nil, err
		}
		b, err := yaml.Marshal(installConfig)
		configData.SetCloudConfigUserData(CloudConfigHeader(b))
	//	case "ignition:rke2":
	//		configData.SetIgnitionUserData(buf.Bytes())
	default:
		return nil, fmt.Errorf("unrecognized user-data type: %s", t)
	}
	// Set metadata
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
